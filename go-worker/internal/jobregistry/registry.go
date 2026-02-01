package jobregistry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/config"
	pb "github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/gen"
	"github.com/Roman-Samoilenko/distributed-job-processing-platform/go-worker/internal/models"
)

type Executor func(ctx context.Context, payload string) (string, error)
type ExecutorFactory func(cfg *config.Config) Executor

var (
	jobTypeToProto     = make(map[models.JobType]pb.JobTask_TaskType)
	protoTypeToJobType = make(map[pb.JobTask_TaskType]models.JobType)
	executorFactories  = make(map[models.JobType]ExecutorFactory)

	mu sync.RWMutex
)

func Register(jobType models.JobType, protoType pb.JobTask_TaskType, factory ExecutorFactory) {
	mu.Lock()
	defer mu.Unlock()

	if !jobType.IsValid() {
		panic(fmt.Sprintf("Задача типа %v не поддерживается", jobType))
	}
	if _, exists := jobTypeToProto[jobType]; exists {
		panic(fmt.Sprintf("%v уже зарегистрирован", jobType))
	}
	if _, exists := protoTypeToJobType[protoType]; exists {
		panic(fmt.Sprintf("%v уже зарегистрирован", protoType))
	}

	jobTypeToProto[jobType] = protoType
	protoTypeToJobType[protoType] = jobType

	executorFactories[jobType] = factory
}

func JobTypeFromProto(protoType pb.JobTask_TaskType) (models.JobType, error) {
	mu.RLock()
	defer mu.RUnlock()

	jobType, exists := protoTypeToJobType[protoType]
	if !exists {
		return "", fmt.Errorf("protoType %v не зарегистрирован", protoType)
	}
	return jobType, nil
}

func CreateExecutors(cfg *config.Config) map[models.JobType]Executor {
	mu.RLock()
	defer mu.RUnlock()

	res := make(map[models.JobType]Executor, len(models.AllJobTypes))

	for _, jt := range models.AllJobTypes {
		factory, exists := executorFactories[jt]
		if !exists {
			slog.Warn("Фабрика для этого типа не найдена", "фабрика", jt)
			continue
		}
		res[jt] = factory(cfg)
	}
	slog.Info(fmt.Sprintf("Создано %d executor'ов", len(res)))

	return res
}

func ValidateAllTypesRegistered() error {
	mu.RLock()
	defer mu.RUnlock()

	for _, jt := range models.AllJobTypes {
		_, exists := executorFactories[jt]
		if !exists {
			return fmt.Errorf("%s не зарегестрирован", jt)
		}
	}
	return nil
}
