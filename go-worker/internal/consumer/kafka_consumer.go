package consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/config"
	pb "github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/gen"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/jobregistry"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/models"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type kafkaConsumer struct {
	jobChan chan<- models.Job
	reader  *kafka.Reader
}

func NewKafkaConsumer(cfg *config.Config, jobChan chan<- models.Job) Consumer {
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          cfg.KafkaTopic,
		GroupID:        cfg.KafkaGroupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
		StartOffset:    kafka.LastOffset,
	})

	return &kafkaConsumer{
		jobChan: jobChan,
		reader:  reader,
	}
}

func (kc *kafkaConsumer) Start(ctx context.Context) error {
	slog.Info("Kafka consumer started",
		slog.String("topic", kc.reader.Config().Topic),
		slog.String("group", kc.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Kafka consumer stopped by context")
			return ctx.Err()
		default:
			msg, err := kc.reader.FetchMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return fmt.Errorf("kafka fetch error: %w", err)
			}
			if err := kc.processMessage(ctx, msg); err != nil {
				slog.Error("Failed to process message",
					"error", err,
					"offset", msg.Offset,
					"partition", msg.Partition,
				)
				continue
			}
			if err := kc.reader.CommitMessages(ctx, msg); err != nil {
				slog.Error("Failed to commit offset", "error", err)
			}
		}
	}
}

func (kc *kafkaConsumer) Close() error {
	return kc.reader.Close()
}

func (kc *kafkaConsumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var jobTask pb.JobTask
	if err := proto.Unmarshal(msg.Value, &jobTask); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf: %w", err)
	}
	slog.Debug("Received job from Kafka",
		slog.Int64("job_id", jobTask.GetJobId()),
		slog.String("type", jobTask.GetType().String()),
	)

	jobType, err := jobregistry.JobTypeFromProto(jobTask.GetType())
	if err != nil {
		return fmt.Errorf("unknown task type: %w", err)
	}

	job := models.Job{
		ID:        jobTask.GetJobId(),
		Type:      jobType,
		Payload:   jobTask.GetPayload(),
		CreatedAt: jobTask.GetCreatedAt(),
	}

	// Отправка в Worker Pool через канал
	select {
	case kc.jobChan <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
