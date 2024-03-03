package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hickar/crtex_test_assignment/account/internal/controllers/kafka"

	"github.com/hickar/crtex_test_assignment/account/internal/config"
	"github.com/hickar/crtex_test_assignment/account/internal/domain"
	"github.com/hickar/crtex_test_assignment/account/internal/repository"
	kconsumer "github.com/hickar/crtex_test_assignment/pkg/kafka/consumer"
	"github.com/hickar/crtex_test_assignment/pkg/postgres"
)

var configPath = flag.String("config", "./config.yaml", "Path to configuration file. Defaults to './config.yaml'")

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %s", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Logger.Level,
	}))

	repo, err := initAccountRepo(ctx, cfg.DB)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize account repository: %s", err))
		os.Exit(1)
	}
	service := domain.NewAccountService(repo)

	kafkaConsumer, err := initKafkaConsumer(cfg.Kafka, service, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize kafka consumer: %s", err))
		os.Exit(1)
	}

	errCh := make(chan error)
	go func() {
		logger.Info("launching kafka consumer")
		if cerr := kafkaConsumer.Run(ctx); cerr != nil {
			errCh <- cerr
		}
	}()

	var stopErr error
	select {
	case <-ctx.Done():
		stopErr = ctx.Err()
	case stopErr = <-errCh:
	}
	if stopErr != nil && !errors.Is(err, context.Canceled) {
		logger.Error(fmt.Sprintf("application stopped with error: %s", stopErr))
		cancel()
		os.Exit(1)
	}

	logger.Info("gracefully shutting down server")
	cancel()
}

func initAccountRepo(ctx context.Context, cfg config.DatabaseConfiguration) (*repository.AccountRepository, error) {
	pgdb, err := postgres.New(ctx, postgres.Configuration{
		Host:                    cfg.Host,
		Port:                    cfg.Port,
		User:                    cfg.User,
		Password:                cfg.Password,
		Name:                    cfg.Name,
		MaxOpenConns:            cfg.MaxConns,
		MaxConnLifetime:         cfg.MaxConnLifetime,
		MaxIdleConnLifetime:     cfg.MaxIdleConnLifetime,
		ConnectionRetries:       cfg.ConnectionRetries,
		ConnectionRetryInterval: cfg.ConnectionRetryInterval,
	})
	if err != nil {
		return nil, err
	}

	return repository.NewAccountRepository(pgdb), nil
}

func initKafkaConsumer(
	cfg config.KafkaConsumerConfiguration,
	service domain.Service,
	logger *slog.Logger,
) (*kconsumer.Consumer, error) {
	handler := kafka.NewAccountHandler(service)
	router := kconsumer.NewTopicRouter()
	router.Handle(
		cfg.Topic,
		handler.NewOrderEvent,
		kconsumer.LoggerMiddleware(logger.With(
			slog.String("module", "kafka_router"),
		)),
	)

	return kconsumer.NewConsumer(
		kconsumer.Configuration{
			BrokerURLs:        cfg.BrokerURLs,
			GroupID:           cfg.GroupID,
			GroupTopics:       cfg.GroupTopics,
			Topic:             cfg.Topic,
			SessionTimeout:    cfg.SessionTimeout,
			HeartbeatInterval: cfg.HeartbeatInterval,
			WorkerCount:       cfg.WorkerCount,
			HandlerTimeout:    cfg.HandlerTimeout,
		},
		router,
	)
}
