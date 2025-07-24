package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/riverqueue/river"

	"github.com/cloud-gov/billing/internal/api/middleware"
	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/jobs"
)

// Routes registers all customer-facing HTTP routes for the server.
func Routes(logger *slog.Logger, cf *client.Client, q db.Querier, riverc *river.Client[pgx.Tx], verifier *oidc.IDTokenVerifier) http.Handler {
	mux := chi.NewMux()
	mux.Use(httplog.RequestLogger(logger, &httplog.Options{
		Level: slog.LevelInfo,
	}))

	mux.Mount("/admin", adminMux(logger, cf, q, riverc, verifier))
	return mux
}

// adminMux returns a Handler for admin routes with access restricted to authorized subjects.
func adminMux(logger *slog.Logger, cf *client.Client, q db.Querier, riverc *river.Client[pgx.Tx], verifier *oidc.IDTokenVerifier) http.Handler {
	mux := chi.NewMux()

	hasAdminScope := middleware.NewHasScope(logger, verifier, "usage.admin")
	mux.Use(hasAdminScope)

	mux.Post("/tier", handleCreateTier(q))
	mux.Post("/usage/job", handleCreateUsageJob(riverc))
	mux.Post("/usage/app/{guid}", handleCreateAppUsageJob(logger, cf, q))

	return mux
}

func handleCreateTier(q db.Querier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tier, err := q.CreateTier(r.Context(), db.CreateTierParams{
			Name:        "",
			TierCredits: 0,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, fmt.Sprintf("%v", tier.ID))
	}
}

func handleCreateUsageJob(riverc *river.Client[pgx.Tx]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := riverc.Insert(r.Context(), jobs.MeasureUsageArgs{}, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to insert River job: %v\n", err), http.StatusInternalServerError)
			return
		}
		io.WriteString(w, fmt.Sprintf("Inserted job with ID: %v\nUniqueSkippedAsDuplicate: %v", result.Job.ID, result.UniqueSkippedAsDuplicate))
	}
}

func handleCreateAppUsageJob(logger *slog.Logger, cf *client.Client, q db.Querier) http.HandlerFunc {
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
