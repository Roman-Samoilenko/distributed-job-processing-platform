package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"go-worker/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 1. Загрузка конфигурации
	cfg, err := config.Load(ctx)
	if err != nil {
		// Fallback logger для критических ошибок
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Настройка slog на основе конфигурации
	var slogHandler slog.Handler
	slogHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     cfg.ParseSlogLevel(),
		AddSource: cfg.DebugMode,
	})
	if cfg.LogFormat == "json" {
		slogHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     cfg.ParseSlogLevel(),
			AddSource: cfg.DebugMode,
		})
	}
	slog.SetDefault(slog.New(slogHandler))

	slog.Info("Starting Job Worker Service",
		slog.String("env", cfg.Environment),
	)

	slog.Info("Configuration loaded", slog.String("config", cfg.String()))

	// 3. Инициализация компонентов
	// TODO: metrics, health

	// TODO: Kafka Consumer, Worker Pool, gRPC Client...

	<-ctx.Done()
	slog.Info("Service shutdown completed gracefully")
}
