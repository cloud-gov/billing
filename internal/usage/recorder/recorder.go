package recorder

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/usage/reader"
)

// RecordReading saves a reading to the database.
func RecordReading(ctx context.Context, q db.Querier, r reader.Reading) error {
	dbReading, err := q.CreateReading(ctx, pgxTimestamp(r.Time))
	if err != nil {
		return err
	}

	// todo COPY for resources
	dbMeasurements := make([]db.CreateMeasurementsParams, len(r.Measurements))
	for i, m := range r.Measurements {
		resource, err := q.GetResourceByNaturalID(ctx, db.GetResourceByNaturalIDParams{
			Meter:     m.Meter,
			NaturalID: m.ResourceNaturalID,
		})
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				// Unexpected error.
				// todo, make sure errors.Is works here.
				return err
			}
			// If row is missing, insert it.
			resource, err = q.CreateResource(ctx, db.CreateResourceParams{
				NaturalID:     m.ResourceNaturalID,
				Meter:         m.Meter,
				KindNaturalID: pgxText(m.ResourceKindNaturalID),
				CFOrgID:       pgxUUID(m.OrgID),
			})
			if err != nil {
				return err // TODO, better handling
			}
		}

		dbMeasurements[i] = db.CreateMeasurementsParams{
			ReadingID:  dbReading.ID,
			ResourceID: resource.ID,
			Value:      int32(m.Value),
		}
	}
	_, err = q.CreateMeasurements(ctx, []db.CreateMeasurementsParams{})
	if err != nil {
		return err
	}

	return nil
}

func pgxUUID(s string) pgtype.UUID {
	u := pgtype.UUID{}
	if len(s) > 16 {
		return u
	}
	copy(u.Bytes[:], s)
	return u
}

func pgxText(s *string) pgtype.Text {
	str := pgtype.Text{
		Valid: s != nil,
	}
	if s != nil {
		str.String = *s
	}
	return str
}

func pgxTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}
