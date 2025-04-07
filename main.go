package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/signal"
)

// run sets up dependencies, calls route registration, and starts the server.
// It is separate from main so it can return errors conventionally and main
// can handle them all in one place, and so the io.Writer can be passed as a
// dependency, making it possible to mock and test for outputs.
func run(ctx context.Context, out io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// To use the logger, rename it to "logger" and pass it to your function.
	_ = slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

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
