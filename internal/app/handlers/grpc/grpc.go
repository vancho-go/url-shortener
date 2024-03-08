package grpc

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *URLShortenerServer) Ping(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}
