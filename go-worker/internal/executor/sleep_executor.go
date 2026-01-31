package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/config"
	pb "github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/gen"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/jobregistry"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/models"
)

func init() {
	jobregistry.Register(
		models.JobTypeSleep,
		pb.JobTask_SLEEP,
		func(cfg *config.Config) jobregistry.Executor {
			exec := NewSleepExecutor()
			return func(ctx context.Context, payload string) (string, error) {
				return exec.Execute(ctx, payload)
			}
		},
	)
}

type sleepExecutor struct {
}

func NewSleepExecutor() TaskExecutor {
	return &sleepExecutor{}
}

func (se *sleepExecutor) Execute(ctx context.Context, payload string) (string, error) {
	p, err := models.ParsePayload[models.PayloadSleep](payload)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга: %w", err)
	}

	duration := time.Duration(p.DurationMs) * time.Millisecond

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-timer.C:
		return fmt.Sprintf("Поспали %d ms", p.DurationMs), nil
	}
}
