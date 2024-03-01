package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/hickar/crtex_test_assignment/order_service/internal/config"
	grpcHandler "github.com/hickar/crtex_test_assignment/order_service/internal/controllers/grpc"
	"github.com/hickar/crtex_test_assignment/order_service/internal/controllers/kafka"
	"github.com/hickar/crtex_test_assignment/order_service/internal/domain"
	"github.com/hickar/crtex_test_assignment/order_service/internal/repository"
	"github.com/hickar/crtex_test_assignment/order_service/proto"
	"github.com/hickar/crtex_test_assignment/pkg/interceptors"
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

	orderRepo, err := initOrderRepo(ctx, cfg.DB)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize order repository: %s", err))
		os.Exit(1)
	}
	orderService := domain.NewOrderService(orderRepo)

	// Настройка сервера GRPC
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCServer.Port))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to open tcp connection on port %d: %s", cfg.GRPCServer.Port, err))
		os.Exit(1)
	}

	grpcOrderHandler := grpcHandler.NewOrderHandler(orderService)
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.GRPCServer.MaxIdleConnLifetime,
			MaxConnectionAge:  cfg.GRPCServer.MaxConnectionAge,
			Timeout:           cfg.GRPCServer.Timeout,
		}),
		grpc.UnaryInterceptor(interceptors.LoggerInterceptor(logger.With(slog.String("module", "grpc_server")))),
	)
	proto.RegisterOrderServer(grpcServer, grpcOrderHandler)

	errCh := make(chan error)
	go func() {
		logger.Info(fmt.Sprintf("launching server on port %d", cfg.GRPCServer.Port))
		if cerr := grpcServer.Serve(ln); cerr != nil {
			errCh <- cerr
		}
	}()

	// Настройка хэндлеров для сообщений Kafka
	kafkaOrderHandler := kafka.NewKafkaOrderHandler(orderService)
	kafkaRouter := kconsumer.NewConsumerTopicRouter()
	kafkaRouter.Handle(
		cfg.KafkaConsumer.Topic,
		kafkaOrderHandler.NewAccountOrderEvent,
		kconsumer.LoggerMiddleware(logger),
	)

	kafkaConsumer, err := kconsumer.NewConsumer(
		ctx,
		kconsumer.ConsumerConfiguration{
			BrokerURLs:        cfg.KafkaConsumer.BrokerURLs,
			GroupID:           cfg.KafkaConsumer.GroupID,
			GroupTopics:       cfg.KafkaConsumer.GroupTopics,
			Topic:             cfg.KafkaConsumer.Topic,
			SessionTimeout:    cfg.KafkaConsumer.SessionTimeout,
			HeartbeatInterval: cfg.KafkaConsumer.HeartbeatInterval,
			WorkerCount:       cfg.KafkaConsumer.WorkerCount,
			Logger:            logger,
			HandlerTimeout:    cfg.KafkaConsumer.HandlerTimeout,
		},
		kafkaRouter,
	)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize kafka consumer: %s", err))
		os.Exit(1)
	}

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
	grpcServer.GracefulStop()
}

func initOrderRepo(ctx context.Context, cfg config.DatabaseConfiguration) (*repository.OrderRepository, error) {
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

	return repository.NewOrderRepository(pgdb), nil
}
