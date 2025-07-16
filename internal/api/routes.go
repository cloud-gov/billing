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
)

// routes registers all routes for the server.
func Routes(logger *slog.Logger, cf *client.Client, q db.Querier, riverc *river.Client[pgx.Tx]) http.Handler {
	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Handle("/usage/job", handleUsageJob(riverc))
	mux.Handle("/usage/app/{guid}", handleUsageApp(logger, cf, q))
	return mux
}

// TODO, how to correctly parameterize the river client?
func handleUsageJob(riverc *river.Client[pgx.Tx]) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result, err := riverc.Insert(r.Context(), jobs.MeasureUsageArgs{}, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert River job: %v\n", err), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, fmt.Sprintf("Inserted job with ID: %v\nUniqueSkippedAsDuplicate: %v", result.Job.ID, result.UniqueSkippedAsDuplicate))
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

func AdminRoutes(logger *slog.Logger, q db.Querier) http.Handler {
	mux := chi.NewMux()
	mux.Mount("/admin", newAdminRouter(logger, q))
	return mux
}

type adminHandler struct {
	Logger  *slog.Logger
	Queries db.Querier
}

func newAdminRouter(logger *slog.Logger, q db.Querier) http.Handler {
	mux := chi.NewMux()
	h := adminHandler{Logger: logger, Queries: q}

	mux.Post("/tier", h.handleCreateTier)

	return mux
}

func (h *adminHandler) handleCreateTier(w http.ResponseWriter, r *http.Request) {
	tier, err := h.Queries.CreateTier(r.Context(), db.CreateTierParams{
		Name:        "",
		TierCredits: 0,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, fmt.Sprintf("%v", tier.ID))
}
