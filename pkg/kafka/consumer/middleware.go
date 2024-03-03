package consumer

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

func LoggerMiddleware(logger *slog.Logger) RouteMiddleware {
	return func(next RouteHandler) RouteHandler {
		return func(ctx context.Context, message *kafka.Message) error {
			start := time.Now()
			err := next(ctx, message)
			respTime := time.Since(start).Milliseconds()

			if err == nil {
				logger.Info(
					"kafka message successfully processed",
					slog.String("topic", message.Topic),
					slog.String("key", string(message.Key)),
					slog.Int64("response_time_ms", respTime),
				)
			}
			if err != nil {
				logger.Error(
					"kafka message processing failed",
					slog.String("topic", message.Topic),
					slog.String("key", string(message.Key)),
					slog.Int64("response_time_ms", respTime),
					slog.Any("error", err),
				)
			}

			return err
		}
	}
}
