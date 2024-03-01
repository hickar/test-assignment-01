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

	"github.com/hickar/crtex_test_assignment/account_service/internal/config"
	"github.com/hickar/crtex_test_assignment/account_service/internal/controllers"
	"github.com/hickar/crtex_test_assignment/account_service/internal/domain"
	"github.com/hickar/crtex_test_assignment/account_service/internal/repository"
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
	accountRepo, err := initAccountRepo(ctx, cfg.DB)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize account repository: %s", err))
		os.Exit(1)
	}
	accountService := domain.NewAccountService(accountRepo)
	accountKafkaController := controllers.NewAccountKafkaHandler(accountService)

	kafkaRouter := kconsumer.NewConsumerTopicRouter()
	kafkaRouter.Handle(
		cfg.Kafka.Topic,
		accountKafkaController.NewOrderEvent,
		kconsumer.LoggerMiddleware(logger),
	)

	kafkaConsumer, err := kconsumer.NewConsumer(
		ctx,
		kconsumer.ConsumerConfiguration{
			BrokerURLs:        cfg.Kafka.BrokerURLs,
			GroupID:           cfg.Kafka.GroupID,
			GroupTopics:       cfg.Kafka.GroupTopics,
			Topic:             cfg.Kafka.Topic,
			SessionTimeout:    cfg.Kafka.SessionTimeout,
			HeartbeatInterval: cfg.Kafka.HeartbeatInterval,
			WorkerCount:       cfg.Kafka.WorkerCount,
			Logger:            logger,
			HandlerTimeout:    cfg.Kafka.HandlerTimeout,
		},
		kafkaRouter,
	)
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
	if stopErr != nil && errors.Is(err, context.Canceled) {
		logger.Error(fmt.Sprintf("application stopped with error: %s", err))
		cancel()
		os.Exit(1)
	}

	logger.Info("gracefully shutting down server")

	cancel()
}

func initAccountRepo(ctx context.Context, cfg config.DatabaseConfiguration) (*repository.AccountRepository, error) {
	pgdb, err := postgres.New(ctx, postgres.Configuration{
		Host:                cfg.Host,
		Port:                cfg.Port,
		User:                cfg.User,
		Password:            cfg.Password,
		Name:                cfg.Name,
		MaxOpenConns:        cfg.MaxOpenConns,
		MaxConnLifetime:     cfg.MaxConnLifetime,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnLifetime: cfg.MaxIdleConnLifetime,
	})
	if err != nil {
		return nil, err
	}

	return repository.NewAccountRepository(pgdb), nil
}
