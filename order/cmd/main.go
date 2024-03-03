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

	"github.com/hickar/crtex_test_assignment/order/internal/config"
	grpcHandler "github.com/hickar/crtex_test_assignment/order/internal/controllers/grpc"
	"github.com/hickar/crtex_test_assignment/order/internal/controllers/kafka"
	"github.com/hickar/crtex_test_assignment/order/internal/domain"
	"github.com/hickar/crtex_test_assignment/order/internal/repository"
	"github.com/hickar/crtex_test_assignment/order/proto"
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

	repo, err := initOrderRepo(ctx, cfg.DB)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize order repository: %s", err))
		os.Exit(1)
	}
	service := domain.NewOrderService(repo)

	// Настройка сервера GRPC
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCServer.Port))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to open tcp connection on port %d: %s", cfg.GRPCServer.Port, err))
		os.Exit(1)
	}
	grpcServer := initGRPCServer(cfg.GRPCServer, service, logger)

	// Настройка хэндлеров для сообщений Kafka
	kafkaConsumer, err := initKafkaConsumer(cfg.KafkaConsumer, service, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to initialize kafka consumer: %s", err))
		os.Exit(1)
	}

	errCh := make(chan error)
	go func() {
		logger.Info(fmt.Sprintf("launching server on port %d", cfg.GRPCServer.Port))
		if cerr := grpcServer.Serve(ln); cerr != nil {
			errCh <- cerr
		}
	}()

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
	grpcServer.GracefulStop()
}

func initOrderRepo(ctx context.Context, cfg config.DatabaseConfiguration) (*repository.OrderRepository, error) {
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

	return repository.NewOrderRepository(pgdb), nil
}

func initKafkaConsumer(
	cfg config.KafkaConsumerConfiguration,
	orderService domain.Service,
	logger *slog.Logger,
) (*kconsumer.Consumer, error) {
	kafkaOrderHandler := kafka.NewOrderHandler(orderService)
	kafkaRouter := kconsumer.NewTopicRouter()
	kafkaRouter.Handle(
		cfg.Topic,
		kafkaOrderHandler.NewAccountOrderEvent,
		kconsumer.LoggerMiddleware(logger.With(
			slog.String("module", "kafka_router")),
		),
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
		kafkaRouter,
	)
}

func initGRPCServer(
	cfg config.GRPCConfiguration,
	orderService domain.Service,
	logger *slog.Logger,
) *grpc.Server {
	grpcOrderHandler := grpcHandler.NewOrderHandler(orderService)
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.MaxIdleConnLifetime,
			MaxConnectionAge:  cfg.MaxConnectionAge,
			Timeout:           cfg.Timeout,
		}),
		grpc.UnaryInterceptor(interceptors.LoggerInterceptor(
			logger.With(slog.String("module", "grpc_server")),
		)),
	)
	proto.RegisterOrderServer(grpcServer, grpcOrderHandler)

	return grpcServer
}
