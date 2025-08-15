package api

import (
	"context"
	"log/slog"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

// Server contains all business logic for the HTTP API endpoints. It implements [StrictServerInterface] and is input to which handles JSON marshalling writing to the response stream. Use [NewServer] to create a Server.
type Server struct {
	Logger  *slog.Logger
	Querier db.Querier
	River   *river.Client[pgx.Tx]
	CF      *client.Client
}

// Compile-time validation that Server implements [StrictServerInterface].
var _ StrictServerInterface = (*Server)(nil)

// NewServer initializes and returns a [Server].
func NewServer(logger *slog.Logger, q db.Querier, riverc *river.Client[pgx.Tx], cf *client.Client) Server {
	return Server{
		Logger:  logger,
		Querier: q,
		River:   riverc,
		CF:      cf,
	}
}

func (s *Server) CreateTier(ctx context.Context, request CreateTierRequestObject) (CreateTierResponseObject, error) {

	return CreateTier201JSONResponse{}, nil
}

func (s *Server) CreateAppUsageJob(ctx context.Context, request CreateAppUsageJobRequestObject) (CreateAppUsageJobResponseObject, error) {
	return CreateAppUsageJob202Response{}, nil
}

func (s *Server) CreateUsageJob(ctx context.Context, request CreateUsageJobRequestObject) (CreateUsageJobResponseObject, error) {
	return CreateUsageJob202Response{}, nil
}
