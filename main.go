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
	cfconfig "github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/robfig/cron/v3"

	"github.com/cloud-gov/billing/internal/api"
	"github.com/cloud-gov/billing/internal/config"
	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/cloud-gov/billing/internal/jobs"
	"github.com/cloud-gov/billing/internal/migrate"
	"github.com/cloud-gov/billing/internal/server"
	"github.com/cloud-gov/billing/internal/usage/meter"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

var (
	ErrBadConfig        = errors.New("reading config from environment")
	ErrCFClient         = errors.New("creating Cloud Foundry client")
	ErrCFConfig         = errors.New("parsing Cloud Foundry connection configuration")
	ErrCrontab          = errors.New("parsing crontab for periodic job execution")
	ErrDBConn           = errors.New("connecting to database")
	ErrDBMigration      = errors.New("migrating the database")
	ErrOIDCProvider     = errors.New("discovering OIDC provider")
	ErrRiverClientNew   = errors.New("creating River client")
	ErrRiverClientStart = errors.New("starting River client")
)

func fmtErr(outer, inner error) error {
	return fmt.Errorf("%w: %w", outer, inner)
}

// run sets up dependencies, migrates the database to the latest
// migration, calls route registration, and starts the server. It is separate
// from main so it can return errors conventionally and main can handle them
// all in one place, and so the [io.Writer] can be passed as a dependency,
// making it possible to mock and test for outputs.
func run(ctx context.Context, out io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	c, err := config.New()
	if err != nil {
		return fmtErr(ErrBadConfig, err)
	}

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: c.LogLevel,
	}))
	logger.Debug("run: initializing CF client")
	cfconf, err := cfconfig.New(c.CFApiUrl,

		cfconfig.ClientCredentials(c.CFClientId, c.CFClientSecret))
	if err != nil {
		return fmtErr(ErrCFConfig, err)
	}
	cfclient, err := client.New(cfconf)
	if err != nil {
		return fmtErr(ErrCFClient, err)
	}

	logger.Debug("run: initializing database")
	conn, err := pgxpool.New(ctx, "") // Pass empty connString so PG* environment variables will be used.
	if err != nil {
		return fmtErr(ErrDBConn, err)
	}

	logger.Debug("run: migrating the database")
	err = migrate.Migrate(ctx, conn)
	if err != nil {
		return fmtErr(ErrDBMigration, err)
	}

	q := dbx.NewQuerier(db.New(conn))

	logger.Debug("run: initializing meters")
	meters := []reader.Meter{
		meter.NewCFServiceMeter(logger, cfclient.ServiceInstances, cfclient.Spaces),
		meter.NewCFAppMeter(logger, cfclient.Applications, cfclient.Processes),
	}
	rdr := reader.New(meters)

	logger.Debug("run: initializing OIDC provider for JWT verification")
	oidcProvider, err := oidc.NewProvider(ctx, c.Issuer)
	if err != nil {
		return fmtErr(ErrOIDCProvider, err)
	}
	verifier := oidcProvider.Verifier(&oidc.Config{ClientID: c.CFClientId}) // todo check alg

	logger.Debug("run: initializing River workers and client")
	workers := river.NewWorkers()
	river.AddWorker(workers, jobs.NewMeasureUsageWorker(logger, conn, q, rdr))

	schedule, err := cron.ParseStandard("1 * * * *") // Read usage every hour, one minute after the hour.
	if err != nil {
		return ErrCrontab
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
					return jobs.MeasureUsageArgs{
						Periodic: true,
					}, nil
				},
				nil,
			),
		},
		Workers: workers,
	})
	if err != nil {
		return fmtErr(ErrRiverClientNew, err)
	}

	logger.Debug("run: starting River server")
	if err = riverc.Start(ctx); err != nil {
		return fmtErr(ErrRiverClientStart, err)
	}

	logger.Debug("run: starting web server")
	srv := server.New(c.Host, c.Port, api.Routes(logger, cfclient, q, riverc, verifier, c), logger)
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
