package tasks

import (
	"context"

	"go.uber.org/zap"
)

const (
	TaskBaseLogMonitor = "base-log-monitor"
)

type baseTask struct {
	lg   *zap.Logger
	name string
}

type Task interface {
	Start(ctx context.Context) error
}
