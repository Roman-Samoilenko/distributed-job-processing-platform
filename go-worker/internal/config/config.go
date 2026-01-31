package config

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	// Kafka
	KafkaBrokers     string   `env:"KAFKA_BROKERS,required"`
	KafkaBrokersList []string `env:"-"` // Заполняется в main после парсинга
	KafkaTopic       string   `env:"KAFKA_TOPIC,default=job_requests"`
	KafkaGroupID     string   `env:"KAFKA_GROUP_ID,required"`
	KafkaClientID    string   `env:"KAFKA_CLIENT_ID,default=go-worker"`

	// gRPC
	GrpcServerAddress string        `env:"GRPC_SERVER_ADDRESS,required"`
	GrpcTimeout       time.Duration `env:"GRPC_TIMEOUT,default=5s"`

	// Worker Pool
	WorkerPoolSize    int           `env:"WORKER_POOL_SIZE,default=10"`
	MaxJobTimeout     time.Duration `env:"MAX_JOB_TIMEOUT,default=30s"`
	JobsChannelBuffer int           `env:"JOBS_CHANNEL_BUFFER,default=100"`

	// Logging
	LogLevel  string `env:"LOG_LEVEL,default=info"`
	LogFormat string `env:"LOG_FORMAT,default=json"`

	// Health
	HealthPort int `env:"HEALTH_PORT,default=8765"`

	// Other
	Environment string `env:"ENVIRONMENT,default=production"`
	DebugMode   bool   `env:"DEBUG_MODE,default=false"`
}

func Load(ctx context.Context) (Config, error) {
	var cfg Config

	if err := envconfig.Process(ctx, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to process environment: %w", err)
	}

	return cfg, nil
}

func (c Config) ParseSlogLevel() slog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (c Config) String() string {
	return fmt.Sprintf(
		"kafka=%s, topic=%s, group=%s, grpc=%s, workers=%d",
		c.KafkaBrokers, c.KafkaTopic, c.KafkaGroupID,
		c.GrpcServerAddress, c.WorkerPoolSize,
	)
}
