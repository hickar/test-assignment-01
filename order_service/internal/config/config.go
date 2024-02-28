package config

import (
	"log/slog"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Configuration struct {
	GRPCServer GRPCConfiguration     `yaml:"grpc"`
	App        AppConfiguration      `yaml:"app"`
	DB         DatabaseConfiguration `yaml:"db"`
	Logger     LoggerConfiguration   `yaml:"logger"`
}

type GRPCConfiguration struct {
	MaxIdleConnLifetime time.Duration `yaml:"max_idle_connection_lifetime"`
	MaxConnectionAge    time.Duration
	Timeout             time.Duration `yaml:"timeout"`
}

type AppConfiguration struct {
	Name    string `yaml:"name" env:"name"`
	Version string `yaml:"version" env:"version"`
}

type DatabaseConfiguration struct {
	Host                string `yaml:"host" env:"DATABASE_HOST" env-required:"true"`
	Port                int    `yaml:"port" env:"DATABASE_PORT" env-required:"true"`
	User                string `yaml:"user" env:"DATABASE_USER" env-required:"true"`
	Password            string `yaml:"password" env:"DATABASE_PASSWORD" env-required:"true"`
	Name                string `yaml:"name" env:"DATABASE_NAME"`
	MaxOpenConns        int    `yaml:"max_connections"`
	MaxConnLifetime     int    `yaml:"max_connection_lifetime"`
	MaxIdleConns        int    `yaml:"max_idle_connections"`
	MaxIdleConnLifetime int    `yaml:"max_idle_connection_lifetime"`
}

type BrokerConfiguration struct{}

type LoggerConfiguration struct {
	Level slog.Level
}

func LoadConfig(configPath string) (*Configuration, error) {
	var cfg Configuration

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
