package api

import (
	"log/slog"
	"net/http"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httplog/v3"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"

	"github.com/cloud-gov/billing/internal/api/middleware"
	"github.com/cloud-gov/billing/internal/config"
	"github.com/cloud-gov/billing/internal/db"
)

// Routes registers all public HTTP routes for the server.
func Routes(logger *slog.Logger, cf *client.Client, q db.Querier, riverc *river.Client[pgx.Tx], verifier *oidc.IDTokenVerifier, config config.Config) http.Handler {
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

	return mux
}
