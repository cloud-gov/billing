package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/riverqueue/river"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/jobs"
	"github.com/cloud-gov/billing/internal/usage/meter"
	"github.com/cloud-gov/billing/internal/usage/reader"
	"github.com/cloud-gov/billing/internal/usage/recorder"
)

// routes registers all routes for the server.
func Routes(logger *slog.Logger, cf *client.Client, q db.Querier, riverc *river.Client[pgx.Tx]) http.Handler {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Handle("/usage", handleUsage(logger.WithGroup("usage"), cf, q))
	mux.Handle("/usage/job", handleUsageJob(logger, riverc))
	mux.Handle("/usage/app/{guid}", handleUsageApp(logger, cf, q))
	return mux
}

// TODO, how to correctly parameterize the river client?
func handleUsageJob(logger *slog.Logger, riverc *river.Client[pgx.Tx]) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, err := riverc.Insert(r.Context(), jobs.MeasureUsageArgs{}, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert River job: %v\n", err), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, fmt.Sprintf("Inserted job with ID: %v\n", result.Job.ID))
	})
}

// First draft. Later, this will be a scheduled background job.
func handleUsage(logger *slog.Logger, cf *client.Client, q db.Querier) http.HandlerFunc {
	logger.Debug("api: initializing meters")
	meters := []reader.Meter{
		meter.NewCFServiceMeter(logger, cf.ServiceInstances, cf.Spaces),
		meter.NewCFAppMeter(logger, cf.Applications, cf.Processes),
	}
	reader := reader.New(meters)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.DebugContext(ctx, "api: reading usage information")
		reading, err := reader.Read(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.DebugContext(ctx, "api: recording usage reading")
		err = recorder.RecordReading(ctx, logger, q, reading)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.DebugContext(ctx, "api: writing response bytes")
		_, err = fmt.Fprintf(w, "Wrote %v measurements to database.", len(reading.Measurements))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func handleUsageApp(logger *slog.Logger, cf *client.Client, q db.Querier) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger.Debug("api: getting app")
		app, err := cf.Applications.Get(ctx, chi.URLParam(r, "guid"))
		if err != nil {
			http.Error(w, "getting app: "+err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Debug("api: getting space")
		space, err := cf.Spaces.Get(ctx, app.Relationships.Space.Data.GUID)
		if err != nil {
			http.Error(w, "getting space: "+err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Debug("api: creating reading")
		reading, err := q.CreateReading(ctx, pgtype.Timestamp{Time: time.Now().UTC(), Valid: true})
		if err != nil {
			http.Error(w, "creating reading: "+err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Debug("api: upserting resource")
		resource, err := q.UpsertResource(ctx, db.UpsertResourceParams{
			NaturalID:     app.GUID,
			Meter:         "oneoff",
			KindNaturalID: "",
			CFOrgID:       pgxUUID(space.Relationships.Organization.Data.GUID),
		})
		if err != nil {
			http.Error(w, "upserting resource: "+err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Debug("api: creating measurement")
		_, err = q.CreateMeasurements(ctx, []db.CreateMeasurementsParams{
			{
				ReadingID:         reading.ID,
				Meter:             resource.Meter,
				ResourceNaturalID: resource.NaturalID,
				Value:             1,
			},
		})
		if err != nil {
			http.Error(w, "creating measurement: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = io.WriteString(w, "Created measurement.\n")
	})
}

func pgxUUID(s string) pgtype.UUID {
	u := pgtype.UUID{}
	u.Scan(s)
	return u
}
