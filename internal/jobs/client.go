package jobs

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/cloud-gov/billing/internal/usage/reader"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/robfig/cron/v3"
)

// NewClient creates a new river client with periodic jobs scheduled.
func NewClient(conn *pgxpool.Pool, logger *slog.Logger, q dbx.Querier, rdr *reader.Reader) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, NewMeasureUsageWorker(logger, conn, q, rdr))

	measureUsageSchedule, err := cron.ParseStandard("1 * * * *") // Read usage every hour, one minute after the hour.
	if err != nil {
		return nil, fmt.Errorf("parsing measureUsage cron spec: %w", err)
	}
	postUsageSchedule, err := cron.ParseStandard("1 6 1 * *") // Post usage on the first of every month at 6:01am.
	if err != nil {
		return nil, fmt.Errorf("parsing postUsage cron spec: %w", err)
	}

	return river.NewClient(riverpgxv5.New(conn), &river.Config{
		JobTimeout: 10 * time.Minute,
		Logger:     logger,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: runtime.GOMAXPROCS(0)}, // Run as many workers as we have CPU cores available.
		},
		PeriodicJobs: []*river.PeriodicJob{
			river.NewPeriodicJob(
				measureUsageSchedule,
				func() (river.JobArgs, *river.InsertOpts) {
					return MeasureUsageArgs{
						Periodic: true,
					}, nil
				},
				nil,
			),
			river.NewPeriodicJob(
				postUsageSchedule,
				func() (river.JobArgs, *river.InsertOpts) {
					return PostUsageArgs{
						Periodic: true,
						AsOf:     pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
					}, nil
				},
				nil,
			),
		},
		Workers: workers,
	})
}
