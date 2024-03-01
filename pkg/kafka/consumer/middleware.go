package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

func LoggerMiddleware(logger *slog.Logger) RouteMiddleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx context.Context, message *kafka.Message) {
			start := time.Now()
			next(ctx, message)

			logger.Info(
				"kafka message processed",
				slog.String("topic", message.Topic),
				slog.String("key", string(message.Key)),
				slog.Int("response_time_ms", int(time.Since(start).Milliseconds())),
			)
		}
	}
}
