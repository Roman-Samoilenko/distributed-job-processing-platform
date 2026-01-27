//go:build !test

package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"log/slog"

	env "github.com/caarlos0/env/v6"
)

// Config содержит всю конфигурацию сервиса
type Config struct {
	// Kafka Configuration
	KafkaBrokers  []string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	KafkaTopic    string   `env:"KAFKA_TOPIC" envDefault:"job_requests"`
	KafkaGroupID  string   `env:"KAFKA_GROUP_ID" envDefault:"job-workers-group-v1"`
	KafkaClientID string   `env:"KAFKA_CLIENT_ID" envDefault:"go-worker"`

	// gRPC
	GrpcServerAddress string        `env:"GRPC_SERVER_ADDRESS" envDefault:"localhost:9090"`
	GrpcTimeout       time.Duration `env:"GRPC_TIMEOUT" envDefault:"5s"`

	// Worker Pool
	WorkerPoolSize    int           `env:"WORKER_POOL_SIZE" envDefault:"10"`
	MaxJobTimeout     time.Duration `env:"MAX_JOB_TIMEOUT" envDefault:"30s"`
	JobsChannelBuffer int           `env:"JOBS_CHANNEL_BUFFER" envDefault:"100"`

	// Logging
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"json"`

	// Health Check
	HealthPort int `env:"HEALTH_PORT" envDefault:"8080"`

	// Metrics
	MetricsPort int    `env:"METRICS_PORT" envDefault:"9090"`
	MetricsPath string `env:"METRICS_PATH" envDefault:"/metrics"`

	// Development
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	DebugMode   bool   `env:"DEBUG_MODE" envDefault:"false"`
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	// Kafka
	if len(c.KafkaBrokers) == 0 {
		return fmt.Errorf("KAFKA_BROKERS must contain at least one broker")
	}
	for _, broker := range c.KafkaBrokers {
		if broker == "" {
			return fmt.Errorf("KAFKA_BROKERS contains empty broker")
		}
	}
	if c.KafkaTopic == "" {
		return fmt.Errorf("KAFKA_TOPIC cannot be empty")
	}
	if c.KafkaGroupID == "" {
		return fmt.Errorf("KAFKA_GROUP_ID cannot be empty")
	}

	// gRPC
	if c.GrpcServerAddress == "" {
		return fmt.Errorf("GRPC_SERVER_ADDRESS cannot be empty")
	}

	// Worker Pool
	if c.WorkerPoolSize <= 0 {
		return fmt.Errorf("WORKER_POOL_SIZE must be positive")
	}
	if c.MaxJobTimeout <= 0 {
		return fmt.Errorf("MAX_JOB_TIMEOUT must be positive")
	}
	if c.JobsChannelBuffer < 0 {
		return fmt.Errorf("JOBS_CHANNEL_BUFFER cannot be negative")
	}

	// Ports
	if c.HealthPort <= 0 || c.HealthPort > 65535 {
		return fmt.Errorf("HEALTH_PORT must be valid port (1-65535)")
	}
	if c.MetricsPort <= 0 || c.MetricsPort > 65535 {
		return fmt.Errorf("METRICS_PORT must be valid port (1-65535)")
	}

	// Logging
	switch strings.ToLower(c.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error")
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return fmt.Errorf("LOG_FORMAT must be json or text")
	}

	return nil
}

// Load загружает конфигурацию из переменных окружения
func Load(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// ParseSlogLevel конвертирует строковый уровень в slog.Level
func (c *Config) ParseSlogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// String возвращает безопасное строковое представление
func (c *Config) String() string {
	return fmt.Sprintf(`
=== Job Worker Configuration ===
Kafka:
  Brokers: [%s]
  Topic: %s
  Group ID: %s
  Client ID: %s

gRPC:
  Server: %s
  Timeout: %v

Worker Pool:
  Size: %d
  Job Timeout: %v
  Jobs Buffer: %d

Logging:
  Level: %s
  Format: %s

Ports:
  Health: %d
  Metrics: %d (%s)

Environment: %s (Debug: %t)`,
		strings.Join(c.KafkaBrokers, ", "),
		c.KafkaTopic,
		c.KafkaGroupID,
		c.KafkaClientID,
		c.GrpcServerAddress,
		c.GrpcTimeout,
		c.WorkerPoolSize,
		c.MaxJobTimeout,
		c.JobsChannelBuffer,
		c.LogLevel,
		c.LogFormat,
		c.HealthPort,
		c.MetricsPort,
		c.MetricsPath,
		c.Environment,
		c.DebugMode,
	)
}

// Example возвращает пример .env файла
func (c *Config) Example() string {
	return `# Kafka Configuration
KAFKA_BROKERS=kafka:9092,localhost:9092
KAFKA_TOPIC=job_requests
KAFKA_GROUP_ID=job-workers-group-v1
KAFKA_CLIENT_ID=go-worker-1

# gRPC
GRPC_SERVER_ADDRESS=localhost:9090
GRPC_TIMEOUT=5s

# Worker Pool
WORKER_POOL_SIZE=10
MAX_JOB_TIMEOUT=30s
JOBS_CHANNEL_BUFFER=100

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# Health Check
HEALTH_PORT=8080

# Metrics
METRICS_PORT=9090
METRICS_PATH=/metrics

# Development
ENVIRONMENT=development
DEBUG_MODE=true
`
}
