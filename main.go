package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/jackc/pgx/v5"

	"github.com/cloud-gov/billing/internal/api"
	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/server"
)

var (
	ErrCFConfig = errors.New("parsing Cloud Foundry connection configuration")
	ErrCFClient = errors.New("creating Cloud Foundry client")
	ErrDBConn   = errors.New("connecting to database")
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
	conn, err := pgx.Connect(ctx, "") // Pass empty connString so PG* environment variables will be used.
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDBConn, err)
	}
	q := db.New(conn)

	srv := server.New("", "8080", api.Routes(logger, cfclient, q), logger)
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
