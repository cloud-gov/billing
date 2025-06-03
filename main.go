package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"

	"github.com/cloud-gov/billing/internal/api"
	"github.com/cloud-gov/billing/internal/server"
)

// run sets up dependencies, calls route registration, and starts the server.
// It is separate from main so it can return errors conventionally and main
// can handle them all in one place, and so the [io.Writer] can be passed as a
// dependency, making it possible to mock and test for outputs.
func run(ctx context.Context, out io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfconf, err := config.NewFromCFHome()
	if err != nil {
		return err
	}
	cfclient, err := client.New(cfconf)
	if err != nil {
		return err
	}

	// db := db.NewMockDB()
	// logger.Info("running with in-memory mock database")

	srv := server.New("", "8080", api.Routes(logger, cfclient), logger)
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
