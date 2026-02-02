package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/config"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/consumer"
	_ "github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/executor"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/grpc"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/jobregistry"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/models"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/worker"
	"github.com/joho/godotenv"
)

func main() {

	// 1. Загрузка конфигурации ДО создания контекста
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("No .env file found or error loading it", "error", err)
	}
	cfg, err := config.Load(context.Background())
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// 2. Настройка логирования
	setupLogging(&cfg)

	// 3. Создаем контекст для graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("Starting Job Worker Service", slog.String("env", cfg.Environment))

	// 4. Валидация реестра
	if err := jobregistry.ValidateAllTypesRegistered(); err != nil {
		slog.Error("Registry validation failed", "error", err)
		panic(err.Error())
	}

	// 5. Инициализация компонентов
	components, err := initializeComponents(ctx, &cfg)
	if err != nil {
		slog.Error("Failed to initialize components", "error", err)
		panic(err.Error())
	}

	// 6. Запуск компонентов
	startComponents(ctx, components)

	// 7. Ожидание сигнала завершения
	<-ctx.Done()
	slog.Info("Shutdown signal received")

	// 8. Graceful shutdown с таймаутом
	shutdownComponents(components)
}

func setupLogging(cfg *config.Config) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level:     cfg.ParseSlogLevel(),
		AddSource: cfg.DebugMode,
	}

	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	slog.SetDefault(slog.New(handler))
}

// components содержит все компоненты системы.
type components struct {
	grpcClient   *grpc.GrpcClient
	workerPool   *worker.WorkerPool
	consumer     consumer.Consumer
	resultSender *worker.ResultSender
	jobsChan     chan models.Job
	resultsChan  chan models.JobResult
}

func initializeComponents(ctx context.Context, cfg *config.Config) (*components, error) {
	c := &components{
		jobsChan:    make(chan models.Job, cfg.JobsChannelBuffer),
		resultsChan: make(chan models.JobResult, cfg.ResultsChannelBuffer),
	}

	// gRPC клиент
	grpcClient, err := grpc.NewGrpcClient(cfg)
	if err != nil {
		return nil, err
	}
	c.grpcClient = grpcClient

	// Executor'ы
	executors := jobregistry.CreateExecutors(cfg)

	// Worker Pool
	c.workerPool = worker.NewWorkerPool(cfg, c.jobsChan, c.resultsChan, executors, ctx)

	// Result Sender
	resultHandler := grpc.NewResultHandler(c.grpcClient)
	c.resultSender = worker.NewResultSender(resultHandler, c.resultsChan, ctx)

	// Kafka Consumer
	c.consumer = consumer.NewKafkaConsumer(cfg, c.jobsChan)

	return c, nil
}

func startComponents(ctx context.Context, c *components) {
	// Запуск Worker Pool
	slog.Info("Starting worker pool")
	c.workerPool.Start()

	// Запуск Result Sender
	slog.Info("Starting result sender")
	c.resultSender.Start()

	// Запуск Kafka Consumer в отдельной горутине
	slog.Info("Starting Kafka consumer")
	go func() {
		if err := c.consumer.Start(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("Kafka consumer error", "error", err)
		}
	}()

	slog.Info("All components started successfully")
}

func shutdownComponents(c *components) {
	const shutdownTimeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	slog.Info("Starting graceful shutdown", slog.Duration("timeout", shutdownTimeout))

	// Останавливаем Kafka consumer (перестаем принимать новые задачи)
	slog.Info("Closing Kafka consumer")
	if err := c.consumer.Close(); err != nil {
		slog.Error("Error closing consumer", "error", err)
	}

	// Закрываем канал задач (воркеры завершат обработку текущих)
	slog.Info("Closing jobs channel")
	close(c.jobsChan)

	// Ждем завершения всех воркеров
	slog.Info("Waiting for workers to finish")
	c.workerPool.Stop()

	// Закрываем канал результатов
	slog.Info("Closing results channel")
	close(c.resultsChan)

	// Ждем отправки всех результатов
	slog.Info("Waiting for result sender to finish")
	c.resultSender.Stop()

	// Закрываем gRPC соединение
	slog.Info("Closing gRPC client")
	if err := c.grpcClient.Close(); err != nil {
		slog.Error("Error closing gRPC client", "error", err)
	}

	select {
	case <-ctx.Done():
		slog.Warn("Shutdown timeout exceeded")
	default:
		slog.Info("Graceful shutdown completed successfully")
	}
}
