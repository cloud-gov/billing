package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/cloud-gov/billing/internal/usage/meter"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

// routes registers all routes for the server.
func Routes(logger *slog.Logger, cf *client.Client) http.Handler {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Handle("/meter", handleMeter(logger.WithGroup("meter"), cf))
	return mux
}

// First draft. Later, this will be a scheduled background job.
func handleMeter(logger *slog.Logger, cf *client.Client) http.HandlerFunc {
	logger.Debug("meter: initializing meters")
	meters := []reader.Meter{
		meter.NewCFServiceMeter(logger, cf.ServiceInstances, cf.Spaces),
		meter.NewCFAppMeter(logger, cf.Applications, cf.Processes),
	}
	reader := reader.New(meters)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.DebugContext(ctx, "meter: reading usage information")
		readings, err := reader.Read(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(ctx, "meter: marshalling JSON")
		b, err := json.Marshal(readings)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.DebugContext(ctx, "meter: writing response bytes")
		_, err = w.Write(b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
