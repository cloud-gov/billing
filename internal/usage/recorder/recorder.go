package recorder

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

var (
	ErrReadingExists = errors.New("a reading already exists for the hour of created_at")
)

// RecordReading saves a reading to the database. It returns [ErrReadingExists] if a Reading already exists for the same hour of r.Time.
func RecordReading(ctx context.Context, logger *slog.Logger, q db.Querier, r reader.Reading, periodic bool) error {
	logger.Debug("creating reading in database")

	dbReading, err := q.CreateUniqueReading(ctx, db.CreateUniqueReadingParams{
		CreatedAt: pgxTimestamp(r.Time),
		Periodic:  periodic,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// ErrNoRows is returned if a row already exists for the given hour. The caller may not consider the Reading existing to be an error. Allow them to handle it differently.
			return ErrReadingExists
		}
		return err
	}

	dbMeters := []string{}
	dbCFOrgs := []pgtype.UUID{}
	dbKinds := db.BulkCreateResourceKindsParams{}
	dbResources := db.BulkCreateResourcesParams{}
	dbMeasurements := db.BulkCreateMeasurementParams{}

	discard := 0

	for _, m := range r.Measurements {
		if m.Meter == "" && m.ResourceNaturalID == "" {
			// Empty measurements will make our database code fail. Discard them.
			discard++
			continue
		}

		// We may insert thousands of rows at a time. We only want to insert if a row does not already exist. COPY does not support ON CONFLICT, running an INSERT in a loop is inefficient, and sqlc does not support variable-length INSERTs. As a workaround we write INSERT queries that accept arrays, with one array per column where appropriate.
		dbMeters = append(dbMeters, m.Meter)
		dbCFOrgs = append(dbCFOrgs, pgxUUID(m.OrgID))
		dbKinds.Meters = append(dbKinds.Meters, m.Meter)
		dbKinds.NaturalIds = append(dbKinds.NaturalIds, m.ResourceKindNaturalID)
		dbResources.CfOrgIds = append(dbResources.CfOrgIds, pgxUUID(m.OrgID))
		dbResources.KindNaturalIds = append(dbResources.KindNaturalIds, m.ResourceKindNaturalID)
		dbResources.Meters = append(dbResources.Meters, m.Meter)
		dbResources.NaturalIds = append(dbResources.NaturalIds, m.ResourceNaturalID)
		dbMeasurements.Meter = append(dbMeasurements.Meter, m.Meter)
		dbMeasurements.ReadingID = append(dbMeasurements.ReadingID, dbReading.ID)
		dbMeasurements.ResourceNaturalID = append(dbMeasurements.ResourceNaturalID, m.ResourceNaturalID)
		dbMeasurements.Value = append(dbMeasurements.Value, int32(m.Value))
	}
	if discard > 0 {
		logger.Warn(fmt.Sprintf("discarded %v empty measurements; a meter is returning empty data", discard))
	}

	logger.Debug("creating meters in database")
	err = q.BulkCreateMeters(ctx, dbMeters)
	if err != nil {
		return err
	}
	logger.Debug("creating orgs in database")
	err = q.BulkCreateCFOrgs(ctx, dbCFOrgs)
	if err != nil {
		return err
	}
	logger.Debug("creating resource kinds in database")
	err = q.BulkCreateResourceKinds(ctx, dbKinds)
	if err != nil {
		return err
	}
	logger.Debug("creating resources in database")
	err = q.BulkCreateResources(ctx, dbResources)
	if err != nil {
		return err
	}
	logger.Debug("creating measurements in database")

	// TODO: For some reason, using q.CreateMeasurements, which is implemented with a COPY, does not work here. It works fine for /usage/app/{guid}.
	err = q.BulkCreateMeasurement(ctx, dbMeasurements)
	if err != nil {
		return err
	}
	logger.Debug("created measurements")
	return err
}

func pgxUUID(s string) pgtype.UUID {
	u := pgtype.UUID{}
	u.Scan(s)
	return u
}

func pgxTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}
