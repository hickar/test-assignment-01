package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/stdlib"
)

type Configuration struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string

	MaxOpenConns        int
	MaxConnLifetime     int
	MaxIdleConns        int
	MaxIdleConnLifetime int
}

func New(ctx context.Context, cfg Configuration) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s dbname=%s sslmode=disable password=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Name,
		cfg.Password,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxConnLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.MaxConnLifetime) * time.Second)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxIdleConnLifetime > 0 {
		db.SetConnMaxIdleTime(time.Duration(cfg.MaxIdleConnLifetime) * time.Second)
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping db: %w", err)
	}

	return db, nil
}
