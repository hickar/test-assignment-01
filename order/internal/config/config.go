package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Configuration struct {
	GRPCServer    GRPCConfiguration          `yaml:"grpc"`
	DB            DatabaseConfiguration      `yaml:"db"`
	Logger        LoggerConfiguration        `yaml:"logger"`
	KafkaConsumer KafkaConsumerConfiguration `yaml:"kafka_consumer"`
}

type GRPCConfiguration struct {
	Port                int           `yaml:"port"`
	MaxIdleConnLifetime time.Duration `yaml:"max_idle_connection_lifetime"`
	MaxConnectionAge    time.Duration `yaml:"max_connection_age"`
	Timeout             time.Duration `yaml:"timeout"`
}

type DatabaseConfiguration struct {
	Host                    string        `yaml:"host" env:"DATABASE_HOST"`
	Port                    int           `yaml:"port" env:"DATABASE_PORT"`
	User                    string        `yaml:"user" env:"DATABASE_USER"`
	Password                string        `yaml:"password" env:"DATABASE_PASSWORD"`
	Name                    string        `yaml:"name" env:"DATABASE_NAME"`
	MaxConns                int           `yaml:"max_connections"`
	MaxConnLifetime         time.Duration `yaml:"max_connection_lifetime"`
	MaxIdleConnLifetime     time.Duration `yaml:"max_idle_connection_lifetime"`
	ConnectionRetries       int           `yaml:"connection_retries"`
	ConnectionRetryInterval time.Duration `yaml:"connection_retry_interval"`
}

type KafkaConsumerConfiguration struct {
	BrokerURLs        []string      `yaml:"broker_urls"`
	GroupID           string        `yaml:"group_id"`
	GroupTopics       []string      `yaml:"group_topics"`
	Topic             string        `yaml:"topic"`
	SessionTimeout    time.Duration `yaml:"session_timeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	HandlerTimeout    time.Duration `yaml:"handler_timeout"`
	WorkerCount       int           `yaml:"worker_count"`
}

type LoggerConfiguration struct {
	Level slog.Level
}

func LoadConfig(configPath string) (*Configuration, error) {
	var cfg Configuration

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("failed to read environment variables into config: %w", err)
	}

	return &cfg, nil
}
