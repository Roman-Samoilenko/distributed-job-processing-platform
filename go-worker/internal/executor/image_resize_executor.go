package executor

import (
	"context"
	"fmt"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/config"
	pb "github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/gen"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/jobregistry"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/models"
)

func init() {
	jobregistry.Register(
		models.JobTypeImageResize,
		pb.JobTask_IMAGE_RESIZE,
		func(cfg *config.Config) jobregistry.Executor {
			return NewImageResizeExecutor().Execute
		},
	)
}

type imageResizeExecutor struct{}

func NewImageResizeExecutor() *imageResizeExecutor {
	return &imageResizeExecutor{}
}

func (e *imageResizeExecutor) Execute(ctx context.Context, payload string) (string, error) {
	p, err := models.ParsePayload[models.PayloadImageResize](payload)
	if err != nil {
		return "", fmt.Errorf("failed to parse payload: %w", err)
	}

	// TODO: Реализовать реальное изменение размера изображения

	return fmt.Sprintf("Image resized to %dx%d (simulated)", p.Width, p.Height), nil
}
