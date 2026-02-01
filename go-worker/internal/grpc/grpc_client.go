package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/config"

	pb "github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/gen"
	"google.golang.org/grpc"
)

type GrpcClient struct {
	Timeout time.Duration
	conn    *grpc.ClientConn
	client  pb.JobStatusServiceClient
}

func NewGrpcClient(cfg config.Config) (*GrpcClient, error) {
	conn, err := grpc.NewClient(cfg.GrpcServerAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &GrpcClient{
		conn:    conn,
		client:  pb.NewJobStatusServiceClient(conn),
		Timeout: cfg.GrpcTimeout,
	}, nil
}

func (gc *GrpcClient) Close() error {
	return gc.conn.Close()
}

func (gc *GrpcClient) SendStatus(ctx context.Context, req *pb.UpdateJobStatusRequest) error {
	ctx, cancel := context.WithTimeout(ctx, gc.Timeout)
	defer cancel()

	resp, err := gc.client.UpdateJobStatus(ctx, req)

	if err != nil {
		return fmt.Errorf("не удалось выполнить вызов gRPC: %w", err)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("gRPC сервер вернул ошибку для %d задачи", req.GetJobId())
	}

	return nil
}
