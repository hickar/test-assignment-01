package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Configuration struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string

	MaxOpenConns        int
	MaxConnLifetime     time.Duration
	MaxIdleConnLifetime time.Duration

	ConnectionRetries       int
	ConnectionRetryInterval time.Duration
}

func New(ctx context.Context, cfg Configuration) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)
	connCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		connCfg.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.MaxConnLifetime > 0 {
		connCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxIdleConnLifetime > 0 {
		connCfg.MaxConnIdleTime = cfg.MaxIdleConnLifetime
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, connCfg)
	if err != nil {
		return nil, err
	}

	if cfg.ConnectionRetries <= 0 {
		cfg.ConnectionRetries = 3
	}
	if cfg.ConnectionRetryInterval <= 0 {
		cfg.ConnectionRetryInterval = time.Second * 3
	}

	for r := 0; r < cfg.ConnectionRetries; r++ {
		if err = dbpool.Ping(ctx); err == nil {
			return dbpool, nil
		}

		time.Sleep(cfg.ConnectionRetryInterval)
	}

	return dbpool, fmt.Errorf("failed to ping database after %d retries: %w", cfg.ConnectionRetries, err)
}
