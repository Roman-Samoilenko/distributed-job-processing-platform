package grpc

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/gen"
)

type ResultHandler struct {
	grpcClient *GrpcClient
}

func NewResultHandler(grpcClient *GrpcClient) *ResultHandler {
	return &ResultHandler{grpcClient: grpcClient}
}

// Обработка успешного выполнения задачи
func (h *ResultHandler) HandleSuccess(jobID int64, result interface{}) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}

	req := &pb.UpdateJobStatusRequest{
		JobId:  jobID,
		Status: pb.UpdateJobStatusRequest_COMPLETED,
		Result: string(resultJSON),
	}

	return h.grpcClient.SendStatus(context.Background(), req)
}

// Обработка ошибки выполнения задачи
func (h *ResultHandler) HandleFailure(jobID int64, err error) error {
	req := &pb.UpdateJobStatusRequest{
		JobId:        jobID,
		Status:       pb.UpdateJobStatusRequest_FAILED,
		Result:       "",
		ErrorMessage: err.Error(),
	}

	return h.grpcClient.SendStatus(context.Background(), req)
}

// Специфичная логика для разных типов задач
func (h *ResultHandler) HandleHTTPResult(jobID int64, statusCode int, body string) error {
	result := map[string]interface{}{
		"status_code":  statusCode,
		"body_length":  len(body),
		"completed_at": time.Now().Unix(),
	}

	return h.HandleSuccess(jobID, result)
}

func (h *ResultHandler) HandleImageResizeResult(jobID int64, newSize string) error {
	result := map[string]interface{}{
		"new_size":     newSize,
		"completed_at": time.Now().Unix(),
	}

	return h.HandleSuccess(jobID, result)
}
