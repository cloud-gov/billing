package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/cloud-gov/billing/internal/usage/reader"
	"github.com/cloud-gov/billing/internal/usage/recorder"
)

var (
	logger  *slog.Logger
	conn    *pgxpool.Pool
	rdr     *reader.Reader
	querier dbx.Querier
)

type MeasureUsageArgs struct {
	// Periodic is true if a reading was taken automatically as part of the periodic usage measurement schedule, or false if it was requested manually.
	Periodic bool
}

func (MeasureUsageArgs) Kind() string {
	return "measure-usage"
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

// Work reads usage from all registered meters and persists the reading to the database if no reading exists for the current hour. It is idempotent if run multiple times within the same hour: For example, at 2:05 and 2:10, but not 2:55 and 1:05. Along with the embedded river.WorkerDefaults, Work fulfills River's Worker interface.
// Transactional job completion example: https://riverqueue.com/docs/transactional-job-completion
func (u *MeasureUsageWorker) Work(ctx context.Context, job *river.Job[MeasureUsageArgs]) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txquerier := querier.WithTx(tx)

	logger.DebugContext(ctx, "measure-usage job: reading usage information")
	// TODO: This is an expensive operation. It can be avoided if we try inserting a Reading into the database with q.CreateUniqueReading before calling Read(), but as written, the Reading is only returned when all its meters have read usage. If we upsert the Reading earlier, we can mark the job Complete early.
	reading, err := rdr.Read(ctx)
	if err != nil {
		return err
	}

	logger.DebugContext(ctx, "measure-usage job: recording usage reading")
	err = recorder.RecordReading(ctx, logger, txquerier, reading, job.Args.Periodic)
	if err != nil && !errors.Is(err, recorder.ErrReadingExists) {
		// If err is ErrReadingExists, a Reading was already recorded for this hour. We can continue completing the job. Other errors are unexpected and are returned.
		return err
	}
	jobAfter, err := river.JobCompleteTx[*riverpgxv5.Driver](ctx, tx, job)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("measure-usage job: transitioned job from %q to %q", job.State, jobAfter.State))

	err = tx.Commit(ctx)
	return err
}

// NewMeasureUsageWorker stores dependencies required for job execution and returns a new worker.
func NewMeasureUsageWorker(l *slog.Logger, c *pgxpool.Pool, q dbx.Querier, r *reader.Reader) *MeasureUsageWorker {
	logger = l
	conn = c
	querier = q
	rdr = r
	return &MeasureUsageWorker{}
}
