package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/config"
	grpc2 "github.com/vancho-go/url-shortener/internal/app/handlers/grpc"
	"github.com/vancho-go/url-shortener/internal/app/handlers/grpc/interceptors"
	http2 "github.com/vancho-go/url-shortener/internal/app/handlers/http"
	"github.com/vancho-go/url-shortener/internal/app/handlers/http/middlewares"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"github.com/vancho-go/url-shortener/internal/app/utils"
	"github.com/vancho-go/url-shortener/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
)

// Run запускает приложение.
func Run() error {
	configuration, err := config.ParseServer()
	if err != nil {
		return fmt.Errorf("error parsing server configuration: %v", err)
	}

	err = middlewares.Initialize(configuration.LogLevel)
	if err != nil {
		return errors.New("error initializing logger")
	}

	dbInstance, err := storage.New(*configuration)
	if err != nil {
		return err
	}
	defer dbInstance.Close()

	middlewares.Log.Info("Configuring http compress middleware")
	compressMiddleware := middlewares.GzipMiddleware

	// Канал для передачи потенциальной ошибки от серверов
	errChan := make(chan error, 1)
	// через этот канал сообщим основному потоку, что соединения закрыты
	//idleConnsClosed := make(chan struct{})
	// канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigint := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	r := chi.NewRouter()

	r.Get("/ping", middlewares.RequestLogger(http2.CheckDBConnection(dbInstance)))

	r.Group(func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware)
		r.Get("/{shortenURL}", middlewares.RequestLogger(compressMiddleware(http2.DecodeURL(dbInstance))))
		r.Post("/", middlewares.RequestLogger(compressMiddleware(http2.EncodeURL(dbInstance, configuration.BaseHost))))
	})

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlewares.JWTMiddleware)
			r.Post("/shorten", middlewares.RequestLogger(compressMiddleware(http2.EncodeURLJSON(dbInstance, configuration.BaseHost))))
			r.Post("/shorten/batch", middlewares.RequestLogger(compressMiddleware(http2.EncodeBatch(dbInstance, configuration.BaseHost))))
			r.Get("/user/urls", middlewares.RequestLogger(http2.GetUserURLs(dbInstance, configuration.BaseHost)))
			r.Delete("/user/urls", middlewares.RequestLogger(http2.DeleteURLs(dbInstance)))
		})
		r.Group(func(r chi.Router) {
			r.Use(utils.TrustedSubnetMiddleware(configuration.TrustedSubnet))
			r.Get("/internal/stats", middlewares.RequestLogger(http2.GetStats(dbInstance)))
		})

	})

	r.Mount("/debug", http2.PprofHandler())

	httpSrv := &http.Server{
		Addr:    configuration.ServerHost,
		Handler: r,
	}

	if configuration.EnableHTTPS {
		middlewares.Log.Info("Starting https server", zap.String("address", configuration.ServerHost))
		go func() {
			err = httpSrv.ListenAndServeTLS(path.Join("certs", "cert.pem"), path.Join("certs", "key.pem"))
			if !errors.Is(err, http.ErrServerClosed) {
				errChan <- errors.New("error starting https server")
			}
		}()

	} else {
		middlewares.Log.Info("Starting http server", zap.String("address", configuration.ServerHost))
		go func() {
			err = httpSrv.ListenAndServe()
			if !errors.Is(err, http.ErrServerClosed) {
				errChan <- errors.New("error starting http server")
			}
		}()
	}

	// определяем порт для grpc сервера
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		return fmt.Errorf("error listening grpc port %v", err)
	}

	// создаём gRPC-сервер без зарегистрированной службы
	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.JWTInterceptor),
		grpc.ChainUnaryInterceptor(interceptors.UnaryServerInterceptor),
	)
	// регистрируем сервис
	proto.RegisterURLShortenerServer(grpcSrv, grpc2.New(dbInstance, configuration.BaseHost))

	middlewares.Log.Info("Starting grpc server")
	// получаем запрос gRPC
	go func() {
		if err = grpcSrv.Serve(listen); err != nil {
			errChan <- fmt.Errorf("error starting grpc server %v", err)
		}
	}()

	// ждём завершения процедуры graceful shutdown
	//<-idleConnsClosed

	select {
	case err = <-errChan: // Возврат ошибки от любого сервера, если таковая имеется
		return err
	case <-sigint: // Завершение работы при успешном graceful shutdown
		// получили оповещение о завершении
		// здесь можно освобождать ресурсы перед выходом,
		// например закрыть соединение с базой данных,
		// закрыть открытые файлы
		if err = httpSrv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		fmt.Println("HTTP Server Shutdown gracefully")
		grpcSrv.GracefulStop()
		fmt.Println("gRPC Server Shutdown gracefully")
	}

	return nil
}
