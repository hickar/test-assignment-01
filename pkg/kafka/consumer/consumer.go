package consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Configuration struct {
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

type RouteHandler func(context.Context, *kafka.Message) error

type RouteMiddleware func(RouteHandler) RouteHandler

type Consumer struct {
	r      *kafka.Reader
	router MessageRouter

	workerCount    int
	handlerTimeout time.Duration
}

type MessageRouter interface {
	Handle(string, RouteHandler, ...RouteMiddleware)
	Route(context.Context, *kafka.Message)
}

func NewConsumer(
	cfg Configuration,
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
			case <-ctx.Done():
				return
			case messageCh <- &msg:
			}
		}
	}()

	for i := 0; i < c.workerCount; i++ {
		go c.worker(ctx, messageCh)
	}

	var closeErr error
	select {
	case <-ctx.Done():
		closeErr = ctx.Err()
	case closeErr = <-errCh:
	}

	return errors.Join(closeErr, c.r.Close())
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
