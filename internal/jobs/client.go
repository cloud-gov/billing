package jobs

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/cloud-gov/billing/internal/usage/reader"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/robfig/cron/v3"
)

func NewClient(conn *pgxpool.Pool, logger *slog.Logger, q dbx.Querier, rdr *reader.Reader) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, NewMeasureUsageWorker(logger, conn, q, rdr))

	schedule, err := cron.ParseStandard("1 * * * *") // Read usage every hour, one minute after the hour.
	if err != nil {
		return nil, fmt.Errorf("parsing cron spec: %w", err)
	}
	riverc, err := river.NewClient(riverpgxv5.New(conn), &river.Config{
		JobTimeout: 10 * time.Minute,
		Logger:     logger,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: runtime.GOMAXPROCS(0)}, // Run as many workers as we have CPU cores available.
		},
		PeriodicJobs: []*river.PeriodicJob{
			river.NewPeriodicJob(
				schedule,
				func() (river.JobArgs, *river.InsertOpts) {
					return MeasureUsageArgs{
						Periodic: true,
					}, nil
				},
				nil,
			),
		},
		Workers: workers,
	})
	return riverc, err
}
