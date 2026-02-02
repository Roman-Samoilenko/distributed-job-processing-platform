package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/config"
	pb "github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/gen"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/jobregistry"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/models"
)

func init() {
	jobregistry.Register(
		models.JobTypeHttpGet,
		pb.JobTask_HTTP_GET,
		func(cfg *config.Config) jobregistry.Executor {
			return NewHttpGetExecutor().Execute
		},
	)
}

type httpGetExecutor struct {
	client *http.Client
}

func NewHttpGetExecutor() *httpGetExecutor {
	return &httpGetExecutor{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *httpGetExecutor) Execute(ctx context.Context, payload string) (string, error) {
	p, err := models.ParsePayload[models.PayloadHttpGet](payload)
	if err != nil {
		return "", fmt.Errorf("failed to parse payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	return fmt.Sprintf("Status: %d, BodyLen: %d", resp.StatusCode, len(body)), nil
}
