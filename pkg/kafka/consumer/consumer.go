package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type ConsumerConfiguration struct {
	BrokerURLs        []string
	GroupID           string
	GroupTopics       []string
	Topic             string
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
	WorkerCount       int
	Logger            *slog.Logger
	HandlerTimeout    time.Duration
}

type RouteHandler func(context.Context, *kafka.Message)

type RouteMiddleware func(RouteHandler) RouteHandler

type Consumer struct {
	r      *kafka.Reader
	router MessageRouter
	logger *slog.Logger

	workerCount    int
	handlerTimeout time.Duration
}

type MessageRouter interface {
	Handle(string, RouteHandler, ...RouteMiddleware)
	Route(context.Context, *kafka.Message)
}

func NewConsumer(
	ctx context.Context,
	cfg ConsumerConfiguration,
	router MessageRouter,
) (*Consumer, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           cfg.BrokerURLs,
		GroupID:           cfg.GroupID,
		GroupTopics:       cfg.GroupTopics,
		Topic:             cfg.Topic,
		Dialer:            &kafka.Dialer{},
		MaxBytes:          10e6, // 10 MB
		HeartbeatInterval: cfg.HeartbeatInterval,
		SessionTimeout:    cfg.SessionTimeout,
	})
	err := r.SetOffset(kafka.LastOffset)
	if err != nil {
		return nil, err
	}

	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 8
	}
	if cfg.HandlerTimeout <= 0 {
		cfg.HandlerTimeout = time.Minute
	}

	return &Consumer{
		r:              r,
		router:         router,
		logger:         cfg.Logger,
		workerCount:    cfg.WorkerCount,
		handlerTimeout: cfg.HandlerTimeout,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	messageCh := make(chan *kafka.Message)
	errCh := make(chan error)

	defer close(messageCh)

	go func() {
		for {
			msg, err := c.r.ReadMessage(ctx)
			if err != nil {
				errCh <- fmt.Errorf("error during message reading: %w", err)
				return
			}

			select {
			case messageCh <- &msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	for i := 0; i < c.workerCount; i++ {
		go c.worker(ctx, messageCh)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (c *Consumer) worker(ctx context.Context, messageCh <-chan *kafka.Message) {
	for {
		select {
		case message, ok := <-messageCh:
			if !ok {
				return
			}

			hctx, cancel := context.WithTimeout(ctx, c.handlerTimeout)
			c.router.Route(hctx, message)
			cancel()

		case <-ctx.Done():
			return
		}
	}
}
