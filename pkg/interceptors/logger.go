package interceptors

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

func LoggerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		logger.Info(
			"request processed",
			slog.String("method", info.FullMethod),
			slog.Int64("response_time_ms", time.Since(start).Milliseconds()),
			slog.Any("error", err),
		)

		return resp, err
	}
}
