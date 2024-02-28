package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hickar/crtex_test_assignment/order_service/internal/config"
	grpcHandler "github.com/hickar/crtex_test_assignment/order_service/internal/controllers/grpc"
	"github.com/hickar/crtex_test_assignment/order_service/internal/domain"
	"github.com/hickar/crtex_test_assignment/order_service/internal/repository"
	"github.com/hickar/crtex_test_assignment/shared/postgres"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var configPath = flag.String("config", "./config.yaml", "Path to configuration file. Defaults to './config.yaml'")

func main() {
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)

	pgdb, err := postgres.New(ctx, postgres.Configuration{
		Host:                cfg.DB.Host,
		Port:                cfg.DB.Port,
		User:                cfg.DB.User,
		Password:            cfg.DB.Password,
		MaxOpenConns:        cfg.DB.MaxOpenConns,
		MaxConnLifetime:     cfg.DB.MaxConnLifetime,
		MaxIdleConns:        cfg.DB.MaxIdleConns,
		MaxIdleConnLifetime: cfg.DB.MaxIdleConnLifetime,
	})
	if err != nil {
		log.Fatalf("failed to initialize database connection: %w", err)
	}

	orderRepo := repository.NewOrderRepository(pgdb)
	orderService := domain.NewOrderService(orderRepo)

	grpcOrderHandler := grpcHandler.NewOrderHandler(orderService)
	grpcServer := grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: cfg.GRPCServer.MaxIdleConnLifetime,
		MaxConnectionAge:  cfg.GRPCServer.MaxConnectionAge,
		Timeout:           cfg.GRPCServer.Timeout,
	}))

	fmt.Println("Hello world")
	time.Sleep(time.Hour)
}
