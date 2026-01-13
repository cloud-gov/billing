package api

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/jobs"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/riverqueue/river"
)

// Server contains all business logic for the HTTP API endpoints. It implements [StrictServerInterface] and is input to which handles JSON marshalling writing to the response stream. Use [NewServer] to create a Server.
type Server struct {
	Logger        *slog.Logger
	Querier       db.Querier
	River         *river.Client[pgx.Tx]
	CF            *client.Client
	TokenVerifier *oidc.IDTokenVerifier
}

// Compile-time validation that Server implements [StrictServerInterface].
var _ StrictServerInterface = (*Server)(nil)

// NewServer initializes and returns a [Server].
func NewServer(logger *slog.Logger, q db.Querier, riverc *river.Client[pgx.Tx], cf *client.Client, verifier *oidc.IDTokenVerifier) Server {
	return Server{
		Logger:        logger,
		Querier:       q,
		River:         riverc,
		CF:            cf,
		TokenVerifier: verifier,
	}
}

func (s *Server) CreateTier(ctx context.Context, request CreateTierRequestObject) (CreateTierResponseObject, error) {
	tier, err := s.Querier.CreateTier(ctx, db.CreateTierParams{
		Name:        request.Body.Name,
		TierCredits: int64(request.Body.CreditsPerYear),
	})
	if err != nil {
		return nil, err
	}
	return CreateTier201JSONResponse{
		Id:             string(tier.ID),
		Name:           tier.Name,
		CreditsPerYear: int(tier.TierCredits),
	}, nil
}

func (s *Server) CreateAppUsageJob(ctx context.Context, request CreateAppUsageJobRequestObject) (CreateAppUsageJobResponseObject, error) {
	s.Logger.Debug("api: getting app")
	app, err := s.CF.Applications.Get(ctx, request.Guid)
	if err != nil {
		return nil, fmt.Errorf("getting app: %w", err)
	}
	s.Logger.Debug("api: getting space")
	space, err := s.CF.Spaces.Get(ctx, app.Relationships.Space.Data.GUID)
	if err != nil {
		return nil, fmt.Errorf("getting space: %w", err)
	}

	s.Logger.Debug("api: creating reading")
	reading, err := s.Querier.CreateUniqueReading(ctx, db.CreateUniqueReadingParams{
		CreatedAt: pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
		Periodic:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("creating reading: %w", err)
	}

	s.Logger.Debug("api: upserting resource")
	resource, err := s.Querier.UpsertResource(ctx, db.UpsertResourceParams{
		NaturalID:     app.GUID,
		Meter:         "oneoff",
		KindNaturalID: "",
		CFOrgID:       pgxUUID(space.Relationships.Organization.Data.GUID),
	})
	if err != nil {
		return nil, fmt.Errorf("upserting resource: %w", err)
	}
	s.Logger.Debug("api: creating measurement")
	_, err = s.Querier.CreateMeasurements(ctx, []db.CreateMeasurementsParams{
		{
			ReadingID:         reading.ID,
			Meter:             resource.Meter,
			ResourceNaturalID: resource.NaturalID,
			Value:             1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating measurement: %w", err)
	}
	return CreateAppUsageJob202Response{}, nil
}

func (s *Server) CreateUsageJob(ctx context.Context, request CreateUsageJobRequestObject) (CreateUsageJobResponseObject, error) {

	result, err := s.River.Insert(ctx, jobs.MeasureUsageArgs{}, nil)
	if err != nil {
		return nil, fmt.Errorf("inserting MeasureUsage job to River: %w", err)
	}

	return CreateUsageJob202JSONResponse{
		Id: int(result.Job.ID),
	}, nil
}

func pgxUUID(s string) pgtype.UUID {
	u := pgtype.UUID{}
	u.Scan(s)
	return u
}
