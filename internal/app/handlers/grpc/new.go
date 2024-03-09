package grpc

import (
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"github.com/vancho-go/url-shortener/pkg/proto"
)

type URLShortenerServer struct {
	proto.UnimplementedURLShortenerServer
	db   storage.Storager
	addr string
}

func New(store storage.Storager, addr string) *URLShortenerServer {
	return &URLShortenerServer{db: store, addr: addr}
}
