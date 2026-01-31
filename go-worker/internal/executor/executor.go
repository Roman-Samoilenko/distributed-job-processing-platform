package executor

import (
	"context"
)

type TaskExecutor interface {
	Execute(ctx context.Context, payload string) (string, error)
}
