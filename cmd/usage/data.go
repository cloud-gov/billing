// Package usage for logging out usage data
package usage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/cloud-gov/billing/internal/config"
	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/cloud-gov/billing/internal/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
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

func GetData() error {
	ctx := context.Background()
	out := os.Stdout

	c, err := config.New()
	if err != nil {
		return fmtErr(ErrBadConfig, err)
	}

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: c.LogLevel,
	}))

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

	return nil
}
