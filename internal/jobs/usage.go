package jobs

import (
	"context"
	"errors"
	"log/slog"

	"github.com/riverqueue/river"
)

var logger *slog.Logger

type MeasureUsageArgs struct {
}

func (MeasureUsageArgs) Kind() string {
	return "read-record-usage"
}

// MeasureUsageWorker reads and records usage data. Use [NewMeasureUsageWorker] to create an instance for registration.
type MeasureUsageWorker struct {
	river.WorkerDefaults[MeasureUsageArgs]
}

func (u *MeasureUsageWorker) Work(ctx context.Context, job *river.Job[MeasureUsageArgs]) error {
	// Read and record usage
	logger.Info("Job ran!")
	return nil
}

func NewMeasureUsageWorker(l *slog.Logger) (*MeasureUsageWorker, error) {
	if l == nil {
		return nil, errors.New("logger cannot be nil!")
	}
	logger = l
	return &MeasureUsageWorker{}, nil
}
