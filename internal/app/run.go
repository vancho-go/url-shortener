package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/auth"
	"github.com/vancho-go/url-shortener/internal/app/compress"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"github.com/vancho-go/url-shortener/internal/app/utils"
	"github.com/vancho-go/url-shortener/pkg/logger"
	"go.uber.org/zap"
	"log"
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

	err = logger.Initialize(configuration.LogLevel)
	if err != nil {
		return errors.New("error initializing logger")
	}

	dbInstance, err := storage.New(*configuration)
	if err != nil {
		return err
	}
	defer dbInstance.Close()

	logger.Log.Info("Configuring http compress middleware")
	compressMiddleware := compress.GzipMiddleware

	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})
	// канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigint := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	r := chi.NewRouter()

	r.Get("/ping", logger.RequestLogger(handlers.CheckDBConnection(dbInstance)))

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)
		r.Get("/{shortenURL}", logger.RequestLogger(compressMiddleware(handlers.DecodeURL(dbInstance))))
		r.Post("/", logger.RequestLogger(compressMiddleware(handlers.EncodeURL(dbInstance, configuration.BaseHost))))
	})

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.JWTMiddleware)
			r.Post("/shorten", logger.RequestLogger(compressMiddleware(handlers.EncodeURLJSON(dbInstance, configuration.BaseHost))))
			r.Post("/shorten/batch", logger.RequestLogger(compressMiddleware(handlers.EncodeBatch(dbInstance, configuration.BaseHost))))
			r.Get("/user/urls", logger.RequestLogger(handlers.GetUserURLs(dbInstance, configuration.BaseHost)))
			r.Delete("/user/urls", logger.RequestLogger(handlers.DeleteURLs(dbInstance)))
		})
		r.Group(func(r chi.Router) {
			r.Use(utils.TrustedSubnetMiddleware(configuration.TrustedSubnet))
			r.Get("/internal/stats", logger.RequestLogger(handlers.GetStats(dbInstance)))
		})

	})

	r.Mount("/debug", handlers.PprofHandler())

	srv := &http.Server{
		Addr:    configuration.ServerHost,
		Handler: r,
	}

	// запускаем горутину обработки пойманных прерываний
	go func() {
		<-sigint
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()

	if configuration.EnableHTTPS {
		logger.Log.Info("Starting https server", zap.String("address", configuration.ServerHost))
		err = srv.ListenAndServeTLS(path.Join("certs", "cert.pem"), path.Join("certs", "key.pem"))
		if !errors.Is(err, http.ErrServerClosed) {
			return errors.New("error starting https server")
		}

	} else {
		logger.Log.Info("Starting http server", zap.String("address", configuration.ServerHost))
		err = srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			return errors.New("error starting http server")
		}
	}

	// ждём завершения процедуры graceful shutdown
	<-idleConnsClosed
	// получили оповещение о завершении
	// здесь можно освобождать ресурсы перед выходом,
	// например закрыть соединение с базой данных,
	// закрыть открытые файлы
	fmt.Println("Server Shutdown gracefully")

	return nil
}
