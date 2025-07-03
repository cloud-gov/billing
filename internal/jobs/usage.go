package jobs

import (
	"context"
	"log/slog"

	"github.com/riverqueue/river"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/usage/reader"
	"github.com/cloud-gov/billing/internal/usage/recorder"
)

var (
	logger  *slog.Logger
	rdr     *reader.Reader
	querier db.Querier
)

type MeasureUsageArgs struct {
}

func (MeasureUsageArgs) Kind() string {
	return "read-record-usage"
}

// MeasureUsageWorker reads and records usage data. Use [NewMeasureUsageWorker] to create an instance for registration.
type MeasureUsageWorker struct {
	river.WorkerDefaults[MeasureUsageArgs]
}

// Work reads usage from all registered meters and persists the reading to the database. It is idempotent if run multiple times within the same hour: For example, at 2:05 and 2:10, but not 2:55 and 1:05. Along with the embedded river.WorkerDefaults, Work fulfills River's Worker interface.
func (u *MeasureUsageWorker) Work(ctx context.Context, job *river.Job[MeasureUsageArgs]) error {
	// If a reading exists from within this hour, the job has already run. Return early so the work function is idempotent.
	exists, err := querier.ReadingExistsInHour(ctx)
	if err != nil || exists {
		return err // If exists && err == nil, nil is returned.
	}

	// Read and record usage
	logger.DebugContext(ctx, "api: reading usage information")
	reading, err := rdr.Read(ctx)
	if err != nil {
		return err
	}

	logger.DebugContext(ctx, "api: recording usage reading")
	err = recorder.RecordReading(ctx, logger, querier, reading)
	return err
}

// NewMeasureUsageWorker stores dependencies required for job execution and returns a new worker.
func NewMeasureUsageWorker(l *slog.Logger, q db.Querier, r *reader.Reader) (*MeasureUsageWorker, error) {
	logger = l
	querier = q
	rdr = r
	return &MeasureUsageWorker{}, nil
}
