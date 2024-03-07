package grpc

import "github.com/vancho-go/url-shortener/pkg/proto"

type URLShortenerServer struct {
	proto.UnimplementedURLShortenerServer
}

func New() *URLShortenerServer {
	return &URLShortenerServer{}
}
