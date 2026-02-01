package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/grpc"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/models"
)

type ResultSender struct {
	resultHandler *grpc.ResultHandler
	resultsChan   <-chan models.JobResult
	ctx           context.Context
	wg            sync.WaitGroup
}

func NewResultSender(
	resultHandler *grpc.ResultHandler,
	resultsChan <-chan models.JobResult,
	ctx context.Context,
) *ResultSender {
	return &ResultSender{
		resultHandler: resultHandler,
		resultsChan:   resultsChan,
		ctx:           ctx,
	}
}

// Start запускает обработчик результатов.
func (rs *ResultSender) Start() {
	rs.wg.Add(1)
	go rs.run()
}

// Stop ожидает завершения обработки всех результатов.
func (rs *ResultSender) Stop() {
	rs.wg.Wait()
}

func (rs *ResultSender) run() {
	defer rs.wg.Done()

	for {
		select {
		case result, ok := <-rs.resultsChan:
			if !ok {
				slog.Info("Results channel closed, stopping sender")
				return
			}
			rs.sendResult(result)

		case <-rs.ctx.Done():
			slog.Info("Result sender stopped by context")
			return
		}
	}
}

func (rs *ResultSender) sendResult(result models.JobResult) {
	var err error

	switch result.Status {
	case models.StatusCompleted:
		resultData := map[string]interface{}{
			"result": result.Result,
			"status": string(result.Status),
		}
		err = rs.resultHandler.HandleSuccess(result.JobID, resultData)

	case models.StatusFailed:
		err = rs.resultHandler.HandleFailure(result.JobID, fmt.Errorf("%s", result.Error))

	case models.StatusCreated, models.StatusInProgress:
		slog.Error("Invalid job status in results channel",
			slog.Int64("job_id", result.JobID),
			slog.String("status", string(result.Status)),
		)
		return

	default:
		slog.Error("Unknown job status",
			slog.Int64("job_id", result.JobID),
			slog.String("status", string(result.Status)),
		)
		return
	}

	if err != nil {
		slog.Error("Failed to send result via gRPC",
			slog.Int64("job_id", result.JobID),
			slog.String("error", err.Error()),
		)
		// TODO: отправить в Dead Letter Queue
	} else {
		slog.Debug("Result sent successfully", slog.Int64("job_id", result.JobID))
	}
}
