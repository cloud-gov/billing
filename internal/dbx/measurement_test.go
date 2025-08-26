package dbx_test

import (
	"testing"
	"time"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/dbx"
	"github.com/cloud-gov/billing/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func mkResourceKind(meter, naturalID string) db.ResourceKind {
	return db.ResourceKind{
		Meter:     meter,
		NaturalID: naturalID,
		Name:      pgtype.Text{String: "", Valid: true},
	}
}

func TestDBUpdateMeasurementMicrocredits(t *testing.T) {
	type tdata struct {
		Orgs         []db.CFOrg
		Meters       []db.Meter
		Kinds        []db.ResourceKind
		Prices       []db.Price
		Readings     []db.Reading
		Resources    []db.Resource
		Measurements []db.Measurement
	}

	// Arrange
	conn, err := pgxpool.New(t.Context(), "")
	if err != nil {
		t.Fatal("creating database connection failed", err)
	}
	// Test acquiring a connection from the pool.
	if err = conn.Ping(t.Context()); err != nil {
		t.Fatal("database connection ping failed")
	}
	q := dbx.NewQuerier(db.New(conn))

	var (
		orgID      = testutil.NewPgxUUID()
		meterName  = "meter-1"
		kindID     = "kind-1"
		priceID    = int32(1)
		readingID1 = int32(1)
		readingID2 = int32(2)
		readingID3 = int32(3)
		readingID4 = int32(4)
		resourceID = "resource-1"
		tz, _      = time.LoadLocation("America/New_York")
		priceLower = testutil.NewPgxTimestamptz(time.Date(2024, time.March, 1, 0, 0, 0, 0, tz))
		priceUpper = testutil.NewPgxTimestamptz(time.Date(2026, time.March, 1, 0, 0, 0, 0, tz))
		asOf       = testutil.NewPgxTimestamptz(time.Date(2025, time.March, 1, 0, 0, 0, 0, tz))
	)

	td := tdata{
		Orgs: []db.CFOrg{
			{
				ID: orgID,
			},
		},
		Meters: []db.Meter{
			{
				Name: meterName,
			},
		},

		Kinds: []db.ResourceKind{
			mkResourceKind(meterName, kindID),
		},
		Prices: []db.Price{
			{
				Meter:               meterName,
				ID:                  priceID,
				KindNaturalID:       kindID,
				MicrocreditsPerUnit: pgtype.Int8{Int64: 8, Valid: true},
				ValidDuring: pgtype.Range[pgtype.Timestamptz]{
					Lower: priceLower,
					Upper: priceUpper,
				},
			},
		},
		Readings: []db.Reading{
			{
				ID: readingID1,
				CreatedAt: pgtype.Timestamp{
					Time:  time.Date(2025, time.January, 1, 0, 0, 0, 0, tz),
					Valid: true,
				},
			},
			{
				ID: readingID2,
				CreatedAt: pgtype.Timestamp{
					Time:  time.Date(2025, time.February, 1, 0, 0, 0, 0, tz),
					Valid: true,
				},
			},
			{
				ID: readingID3,
				CreatedAt: pgtype.Timestamp{
					Time:  time.Date(2025, time.February, 3, 0, 0, 0, 0, tz),
					Valid: true,
				},
			},
			{
				ID: readingID4,
				CreatedAt: pgtype.Timestamp{
					Time:  time.Date(2025, time.March, 1, 0, 0, 0, 0, tz),
					Valid: true,
				},
			},
		},
		// stuck. Either insert the things in a specific order or pre-determine the IDs.
		// no problem with pre-determining the IDs, but need to write new sqlc functions
		Resources: []db.Resource{
			{
				Meter:         meterName,
				NaturalID:     resourceID,
				KindNaturalID: kindID,
				CFOrgID:       orgID,
			},
		},
		Measurements: []db.Measurement{
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             8,
				ReadingID:         readingID1,
			},
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             8,
				ReadingID:         readingID2,
			},
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             8,
				ReadingID:         readingID3,
			},
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             8,
				ReadingID:         readingID4,
			},
		},
	}

	for _, i := range td.Orgs {
		_, err := q.CreateCFOrg(t.Context(), i.ID)
		if err != nil {
			t.Fatal("creating CF org failed", err)
		}
	}
	for _, i := range td.Meters {
		_, err := q.CreateMeter(t.Context(), i.Name)
		if err != nil {
			t.Fatal("creating meter failed", err)
		}
	}
	for _, i := range td.Kinds {
		_, err := q.CreateResourceKind(t.Context(), db.CreateResourceKindParams{
			Meter:     i.Meter,
			NaturalID: i.NaturalID,
		})
		if err != nil {
			t.Fatal("creating resource kind failed", err)
		}
	}
	for _, i := range td.Prices {
		_, err := q.CreatePriceWithID(t.Context(), db.CreatePriceWithIDParams(i))
		if err != nil {
			t.Fatal("creating price failed", err)
		}
	}
	for _, i := range td.Readings {
		_, err := q.CreateReadingWithID(t.Context(), db.CreateReadingWithIDParams(i))
		if err != nil {
			t.Fatal("creating reading failed", err)
		}
	}
	for _, i := range td.Resources {
		err = q.CreateResources(t.Context(), db.CreateResourcesParams(i))
		if err != nil {
			t.Fatal("creating resource failed", err)
		}
	}
	for _, i := range td.Measurements {
		q.CreateMeasurement(t.Context(), db.CreateMeasurementParams{
			ReadingID:         i.ReadingID,
			Meter:             i.Meter,
			ResourceNaturalID: i.ResourceNaturalID,
			Value:             i.Value,
		})
	}

	// Act
	updated, err := q.UpdateMeasurementMicrocredits(t.Context(), asOf)

	// Assert
	if err != nil {
		t.Fatal("error occured while calling function under test", err)
	}
	if updated.Int64 != 2 {
		t.Fatalf("expected %v rows updated, got %v", 2, updated.Int64)
	}
}
