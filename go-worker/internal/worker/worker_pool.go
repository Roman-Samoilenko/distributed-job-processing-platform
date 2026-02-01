package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/config"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/jobregistry"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/models"
)

type WorkerPool struct {
	numWorkers  int
	jobTimeout  time.Duration
	jobsChan    <-chan models.Job
	resultsChan chan<- models.JobResult
	executors   map[models.JobType]jobregistry.Executor

	wg  sync.WaitGroup
	ctx context.Context
}

func NewWorkerPool(
	cfg *config.Config,
	jobsChan <-chan models.Job,
	resultsChan chan<- models.JobResult,
	executors map[models.JobType]jobregistry.Executor,
	ctx context.Context,
) *WorkerPool {
	return &WorkerPool{
		numWorkers:  cfg.WorkerPoolSize,
		jobTimeout:  cfg.MaxJobTimeout,
		jobsChan:    jobsChan,
		resultsChan: resultsChan,
		executors:   executors,
		wg:          sync.WaitGroup{},
		ctx:         ctx,
	}
}

func (wp *WorkerPool) Start() {
	for i := range wp.numWorkers {
		wp.wg.Add(1)
		go wp.runWorker(i)
	}
}

func (wp *WorkerPool) Stop() {
	wp.wg.Wait()
}

func (wp *WorkerPool) runWorker(id int) {
	defer wp.wg.Done()
	slog.Debug("Worker started", slog.Int("worker_id", id))

	for job := range wp.jobsChan {
		select {
		case <-wp.ctx.Done():
			return
		default:
		}

		slog.Debug("Processing job",
			slog.Int("worker_id", id),
			slog.Int64("job_id", job.ID),
			slog.String("type", string(job.Type)),
		)

		result := wp.process(job)

		select {
		case wp.resultsChan <- result:
		case <-wp.ctx.Done():
			return
		}
	}

	slog.Debug("Worker stopped", slog.Int("worker_id", id))
}

func (wp *WorkerPool) process(job models.Job) models.JobResult {
	ctx, cancel := context.WithTimeout(wp.ctx, wp.jobTimeout)
	defer cancel()

	exec, exists := wp.executors[job.Type]
	if !exists {
		return models.JobResult{
			JobID:  job.ID,
			Status: models.StatusFailed,
			Error:  fmt.Sprintf("unknown job type: %s", job.Type),
		}
	}

	output, err := exec(ctx, job.Payload)

	if err != nil {
		slog.Error("Job failed",
			slog.Int64("job_id", job.ID),
			slog.String("error", err.Error()),
		)
		return models.JobResult{
			JobID:  job.ID,
			Status: models.StatusFailed,
			Error:  err.Error(),
		}
	}

	return models.JobResult{
		JobID:  job.ID,
		Status: models.StatusCompleted,
		Result: output,
	}
}
