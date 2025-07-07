package jobs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/cloud-gov/billing/internal/dbtx"
	"github.com/cloud-gov/billing/internal/usage/reader"
	"github.com/cloud-gov/billing/internal/usage/recorder"
)

var (
	logger  *slog.Logger
	conn    *pgxpool.Pool
	rdr     *reader.Reader
	querier dbtx.Querier
)

type MeasureUsageArgs struct {
}

func (MeasureUsageArgs) Kind() string {
	return "read-record-usage"
}

// MeasureUsageWorker reads and records usage data. Use [NewMeasureUsageWorker] to create an instance for registration with the River client.
type MeasureUsageWorker struct {
	river.WorkerDefaults[MeasureUsageArgs]
}

func (u *MeasureUsageWorker) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		// Unique jobs only exist once for a given set of properties: https://riverqueue.com/docs/unique-jobs
		// TODO: This warrants further testing. I was able to submit this job several times without UniqueSkippedAsDuplicate being returned as true.
		UniqueOpts: river.UniqueOpts{
			ByQueue: true,
		},
	}
}

// Work reads usage from all registered meters and persists the reading to the database. It is idempotent if run multiple times within the same hour: For example, at 2:05 and 2:10, but not 2:55 and 1:05. Along with the embedded river.WorkerDefaults, Work fulfills River's Worker interface.
// Transactional job completion example: https://riverqueue.com/docs/transactional-job-completion
func (u *MeasureUsageWorker) Work(ctx context.Context, job *river.Job[MeasureUsageArgs]) error {
	// If a reading exists from within this hour, the job has already run. Return early so the work function is idempotent.
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txquerier := querier.WithTx(tx)
	exists, err := txquerier.ReadingExistsInHour(ctx)
	if err != nil || exists {
		logger.Debug("MeasureUsageWorker.Work: reading was already recorded for this hour; exiting job")
		return err // If exists && err == nil, nil is returned.
	}

	// Read and record usage
	logger.DebugContext(ctx, "api: reading usage information")
	reading, err := rdr.Read(ctx)
	if err != nil {
		return err
	}

	logger.DebugContext(ctx, "api: recording usage reading")
	err = recorder.RecordReading(ctx, logger, txquerier, reading)
	if err != nil {
		return err
	}
	jobAfter, err := river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("transitioned MeasureUsageWorker job from %q to %q", job.State, jobAfter.State))

	err = tx.Commit(ctx)
	return err
}

// NewMeasureUsageWorker stores dependencies required for job execution and returns a new worker.
func NewMeasureUsageWorker(l *slog.Logger, c *pgxpool.Pool, q dbtx.Querier, r *reader.Reader) (*MeasureUsageWorker, error) {
	logger = l
	conn = c
	querier = q
	rdr = r
	return &MeasureUsageWorker{}, nil
}
