package grpc

import (
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"github.com/vancho-go/url-shortener/pkg/proto"
)

// URLShortenerServer поддерживает все необходимые методы gRPC сервера.
type URLShortenerServer struct {
	proto.UnimplementedURLShortenerServer
	db   storage.Storager
	addr string
}

// New - конструктор URLShortenerServer.
func New(store storage.Storager, addr string) *URLShortenerServer {
	return &URLShortenerServer{db: store, addr: addr}
}
