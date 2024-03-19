package interceptors

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

// Предполагается, что Log уже инициализирован, как в вашем примере.
var Log, _ = zap.NewDevelopment()

// UnaryServerInterceptor создает gRPC интерсептор для логирования.
func UnaryServerInterceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {
	// Запоминаем время начала обработки запроса.
	startTime := time.Now()

	// Обработка запроса.
	resp, err := handler(ctx, req)

	// Вычисляем продолжительность обработки.
	duration := time.Since(startTime)

	// Логируем детали запроса.
	Log.Info("gRPC request",
		zap.String("method", info.FullMethod),
		zap.String("duration", duration.String()),
		zap.Any("error", err),
	)

	return resp, err
}
