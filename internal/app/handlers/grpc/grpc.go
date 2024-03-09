package grpc

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vancho-go/url-shortener/internal/app/base62"
	"github.com/vancho-go/url-shortener/internal/app/handlers/grpc/interceptors"
	"github.com/vancho-go/url-shortener/internal/app/handlers/http/middlewares"
	"github.com/vancho-go/url-shortener/internal/app/models"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"github.com/vancho-go/url-shortener/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"math/rand"
	"strconv"
	"time"
)

// Ping позволяет проверить доступность сервиса.
func (s *URLShortenerServer) Ping(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	db, ok := s.db.(*storage.Database)
	if !ok {
		return nil, status.Error(codes.Internal, "something wrong with storage")
	}

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return nil, status.Error(codes.Internal, "something wrong with storage timeout")
	}

	return nil, nil
}

func (s *URLShortenerServer) AddURL(ctx context.Context, in *proto.AddURLRequest) (*proto.AddURLResponse, error) {
	originalURL := in.OriginalUrl
	userID := ctx.Value(interceptors.UserIDKey).(string)
	if userID == "" {
		return nil, status.Error(codes.Internal, "something wrong")
	}

	shortenURL := base62.Base62Encode(rand.Uint64())
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	for !s.db.IsShortenUnique(ctx, shortenURL) {
		shortenURL = base62.Base62Encode(rand.Uint64())
	}

	ctx, cancel2 := context.WithTimeout(ctx, 1*time.Second)
	defer cancel2()

	err := s.db.AddURL(ctx, originalURL, shortenURL, userID)
	if err != nil {
		if !isUniqueViolationError(err) {
			return nil, status.Error(codes.Internal, "error adding new shorten URL")
		}

		pg, ok := s.db.(*storage.Database)
		if !ok {
			return nil, status.Error(codes.Internal, "internal DB error")
		}
		ctx, cancel3 := context.WithTimeout(ctx, 1*time.Second)
		defer cancel3()
		shortenURL, err = pg.GetShortenURLByOriginal(ctx, string(originalURL))
		if err != nil {
			return nil, status.Error(codes.Internal, "error getting shorten URL")
		}
	}
	var resp proto.AddURLResponse
	resp.Result = s.addr + "/" + shortenURL

	return &resp, nil
}

func (s *URLShortenerServer) AddURLs(ctx context.Context, in *proto.AddURLsRequest) (*proto.AddURLsResponse, error) {
	userID := ctx.Value(interceptors.UserIDKey).(string)
	if userID == "" {
		return nil, status.Error(codes.Internal, "something wrong")
	}

	var response proto.AddURLsResponse
	var batch []models.APIBatchRequest
	const batchSize = 100

	for i, val := range in.IdAndUrl {
		originalURL := val.OriginalUrl
		if originalURL == "" {
			continue
		}
		shortenURL := base62.Base62Encode(rand.Uint64())
		ctxWT, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		for !s.db.IsShortenUnique(ctxWT, shortenURL) {
			shortenURL = base62.Base62Encode(rand.Uint64())
		}

		batch = append(batch, models.APIBatchRequest{
			CorrelationID: val.CorrelationId,
			OriginalURL:   originalURL,
			ShortenURL:    shortenURL,
		})

		if len(batch) == batchSize || i == len(in.IdAndUrl)-1 {
			ctxWT2, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			err := s.db.AddURLs(ctxWT2, userID, batch...)
			if err != nil {
				return nil, status.Error(codes.Internal, "something wrong")
			}
			for _, b := range batch {

				res := proto.AddURLsResponse_Res{
					CorrelationId: b.CorrelationID,
					ShortUrl:      s.addr + "/" + b.ShortenURL,
				}

				response.Result = append(response.Result, &res)
			}
			batch = nil // Сбросить пакет после вставки.
		}
	}
	return &response, nil
}

func (s *URLShortenerServer) GetURL(ctx context.Context, in *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	shortenURL := in.ShortUrl
	ctxWT, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	originalURL, err := s.db.GetURL(ctxWT, shortenURL)
	if err == nil {
		var resp proto.GetURLResponse
		resp.OriginalUrl = originalURL
		return &resp, nil
	}

	if errors.Is(err, storage.ErrDeletedURL) {
		return nil, status.Error(codes.NotFound, "url was deleted")
	}
	return nil, status.Error(codes.NotFound, "url not found")
}

func (s *URLShortenerServer) GetUserURLs(ctx context.Context, in *emptypb.Empty) (*proto.GetUserURLsResponse, error) {
	userID := ctx.Value(interceptors.UserIDKey).(string)
	if userID == "" {
		return nil, status.Error(codes.Internal, "something wrong")
	}

	ctxWT, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	userURLs, err := s.db.GetUserURLs(ctxWT, userID)

	if err != nil {
		middlewares.Log.Error("error getting user urls", zap.Error(err))
		return nil, status.Error(codes.Internal, "error getting user urls")
	}

	if len(userURLs) == 0 {
		return nil, status.Error(codes.NotFound, "not found")
	}

	var resp proto.GetUserURLsResponse

	for _, url := range userURLs {
		res := proto.GetUserURLsResponse_Res{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    s.addr + "/" + url.ShortenURL,
		}

		resp.Result = append(resp.Result, &res)
	}
	return &resp, nil
}

func (s *URLShortenerServer) DeleteURLs(ctx context.Context, in *proto.DeleteURLsRequest) (*emptypb.Empty, error) {
	userID := ctx.Value(interceptors.UserIDKey).(string)
	if userID == "" {
		return nil, status.Error(codes.Internal, "something wrong")
	}

	urlsToDelete := make([]models.DeleteURLRequest, len(in.Urls))

	for pos, url := range in.Urls {
		urlsToDelete[pos].ShortenURL = url
		urlsToDelete[pos].UserID = userID
	}

	ctxWT, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	err := s.db.DeleteUserURLs(ctxWT, urlsToDelete...)
	if err != nil {
		middlewares.Log.Error("error deleting", zap.Error(err))
	}
	return nil, nil
}
func (s *URLShortenerServer) GetStats(ctx context.Context, in *emptypb.Empty) (*proto.GetStatsResponse, error) {
	response, err := s.db.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Error getting stats")
	}

	var resp proto.GetStatsResponse
	resp.Users = strconv.Itoa(response.Users)
	resp.Urls = strconv.Itoa(response.URLs)
	return &resp, nil
}

// isUniqueViolationError проверяет является ли ошибка UniqueViolation.
func isUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}
