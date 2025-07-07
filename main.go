package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/cloud-gov/billing/internal/api"
	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/jobs"
	"github.com/cloud-gov/billing/internal/server"
	"github.com/cloud-gov/billing/internal/usage/meter"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

var (
	ErrCFConfig         = errors.New("parsing Cloud Foundry connection configuration")
	ErrCFClient         = errors.New("creating Cloud Foundry client")
	ErrDBConn           = errors.New("connecting to database")
	ErrRiverClientNew   = errors.New("creating River client")
	ErrRiverClientStart = errors.New("starting River client")
)

// run sets up dependencies, calls route registration, and starts the server.
// It is separate from main so it can return errors conventionally and main
// can handle them all in one place, and so the [io.Writer] can be passed as a
// dependency, making it possible to mock and test for outputs.
func run(ctx context.Context, out io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfconf, err := config.NewFromCFHome()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCFConfig, err)
	}
	cfclient, err := client.New(cfconf)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCFClient, err)
	}

	conn, err := pgxpool.New(ctx, "") // Pass empty connString so PG* environment variables will be used.
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDBConn, err)
	}
	q := db.New(conn)

	logger.Debug("main: initializing meters")
	meters := []reader.Meter{
		meter.NewCFServiceMeter(logger, cfclient.ServiceInstances, cfclient.Spaces),
		meter.NewCFAppMeter(logger, cfclient.Applications, cfclient.Processes),
	}
	rdr := reader.New(meters)

	workers := river.NewWorkers()

	usageWorker, err := jobs.NewMeasureUsageWorker(logger, q, rdr)
	if err != nil {
		return err
	}
	river.AddWorker(workers, usageWorker)

	riverc, err := river.NewClient(riverpgxv5.New(conn), &river.Config{
		JobTimeout: 10 * time.Minute,
		Logger:     logger,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: runtime.GOMAXPROCS(0)}, // Run as many workers as we have CPU cores available.
		},
		Workers: workers,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRiverClientNew, err)
	}

	if err = riverc.Start(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrRiverClientStart, err)
	}

	srv := server.New("", "8080", api.Routes(logger, cfclient, q, riverc), logger)
	srv.ListenAndServe(ctx)
	return nil
}

func main() {
	ctx := context.Background()
	err := run(ctx, os.Stdout)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
