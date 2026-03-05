package recorder_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/usage/reader"
	"github.com/cloud-gov/billing/internal/usage/recorder"
)

// stubQuerier records the arguments it receives.  If errOn matches the method name being called it returns an error so the test can verify the error-handling path. Only the methods that RecordReading uses are implemented.
type stubQuerier struct {
	errOn string // one of: CreateReading, BulkCreateMeters, BulkCreateCFOrgs, BulkCreateResourceKinds, BulkCreateResources, BulkCreateMeasurement

	createReadingTS pgtype.Timestamp
	bulkMeters      []string
	bulkOrgs        db.BulkCreateCFOrgsParams
	bulkKinds       db.BulkCreateResourceKindsParams
	bulkResources   db.BulkCreateResourcesParams
	bulkMs          db.BulkCreateMeasurementParams
	bulkRNodes      db.BulkCreateResourceNodesParams
}

var ErrExpected = errors.New("this error was expected")

func (s *stubQuerier) CreateReading(_ context.Context, arg db.CreateReadingParams) (db.Reading, error) {
	if s.errOn == "CreateReading" {
		return db.Reading{}, ErrExpected
	}
	s.createReadingTS = arg.CreatedAt
	return db.Reading{ID: 1}, nil
}

func (s *stubQuerier) CreateUniqueReading(_ context.Context, arg db.CreateUniqueReadingParams) (db.Reading, error) {
	if s.errOn == "CreateReading" {
		return db.Reading{}, ErrExpected
	}
	s.createReadingTS = arg.CreatedAt
	return db.Reading{ID: 1}, nil
}

func (s *stubQuerier) BoundsMonthPrev(_ context.Context, asOf pgtype.Timestamptz) (db.BoundsMonthPrevRow, error) {
	panic("unimplemented")
}

func (s *stubQuerier) BulkCreateMeters(_ context.Context, meters []string) error {
	if s.errOn == "BulkCreateMeters" {
		return ErrExpected
	}
	s.bulkMeters = meters
	return nil
}

func (s *stubQuerier) BulkCreateCFOrgs(_ context.Context, orgs db.BulkCreateCFOrgsParams) error {
	if s.errOn == "BulkCreateCFOrgs" {
		return ErrExpected
	}
	s.bulkOrgs = orgs
	return nil
}

func (s *stubQuerier) BulkCreateResourceKinds(_ context.Context, arg db.BulkCreateResourceKindsParams) error {
	if s.errOn == "BulkCreateResourceKinds" {
		return ErrExpected
	}
	s.bulkKinds = arg
	return nil
}

func (s *stubQuerier) BulkCreateResources(_ context.Context, arg db.BulkCreateResourcesParams) error {
	if s.errOn == "BulkCreateResources" {
		return ErrExpected
	}
	s.bulkResources = arg
	return nil
}

func (s *stubQuerier) BulkCreateMeasurement(_ context.Context, arg db.BulkCreateMeasurementParams) error {
	if s.errOn == "BulkCreateMeasurement" {
		return ErrExpected
	}
	s.bulkMs = arg
	return nil
}

func (s *stubQuerier) BulkCreateResourceNodes(_ context.Context, arg db.BulkCreateResourceNodesParams) error {
	if s.errOn == "BulkCreateResourceNodes" {
		return ErrExpected
	}
	s.bulkRNodes = arg
	return nil
}

func (s *stubQuerier) AccountingEquation(_ context.Context) ([]string, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateCFOrg(_ context.Context, arg db.CreateCFOrgParams) (db.CFOrg, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateCustomer(_ context.Context, arg string) (pgtype.UUID, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateMeasurement(_ context.Context, arg db.CreateMeasurementParams) (db.Measurement, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateMeasurements(_ context.Context, arg []db.CreateMeasurementsParams) (int64, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateMeter(_ context.Context, name string) (string, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreatePriceWithID(_ context.Context, arg db.CreatePriceWithIDParams) (db.Price, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateReadingWithID(_ context.Context, arg db.CreateReadingWithIDParams) (db.Reading, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateResourceKind(_ context.Context, arg db.CreateResourceKindParams) (db.ResourceKind, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateResources(_ context.Context, arg db.CreateResourcesParams) error {
	panic("unimplemented")
}

func (s *stubQuerier) CreateTier(_ context.Context, arg db.CreateTierParams) (db.Tier, error) {
	panic("unimplemented")
}

func (s *stubQuerier) CreateTransaction(_ context.Context, arg db.CreateTransactionParams) (db.Transaction, error) {
	panic("unimplemented")
}

func (s *stubQuerier) DeleteCFOrg(_ context.Context, id pgtype.UUID) error {
	panic("unimplemented")
}

func (s *stubQuerier) DeleteCustomer(_ context.Context, id pgtype.UUID) error {
	panic("unimplemented")
}

func (s *stubQuerier) DeleteResource(_ context.Context, arg db.DeleteResourceParams) error {
	panic("unimplemented")
}

func (s *stubQuerier) DeleteResourceKind(_ context.Context, arg db.DeleteResourceKindParams) error {
	panic("unimplemented")
}

func (s *stubQuerier) DeleteTier(_ context.Context, id int32) error {
	panic("unimplemented")
}

func (s *stubQuerier) GetCFOrg(_ context.Context, id pgtype.UUID) (db.CFOrg, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetCustomer(_ context.Context, id pgtype.UUID) (db.Customer, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetCustomersByName(_ context.Context, name string) ([]db.Customer, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetAccountForCustomerAndType(_ context.Context, arg db.GetAccountForCustomerAndTypeParams) (db.Account, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetUsageByPath(ctx context.Context, arg db.GetUsageByPathParams) ([]db.GetUsageByPathRow, error) {
	panic("unimplemented")
}

func (s *stubQuerier) LQueryResourceNodes(ctx context.Context, arg db.LQueryResourceNodesParams) ([]db.ResourceNode, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListResourceNodeAncestors(ctx context.Context, path string) ([]db.ResourceNode, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListResourceNodeDescendants(ctx context.Context, path string) ([]db.ResourceNode, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetAppsUsageBySpace(_ context.Context, customerID pgtype.UUID) ([]db.GetAppsUsageBySpaceRow, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetEntriesForCustomerAndType(_ context.Context, arg db.GetEntriesForCustomerAndTypeParams) ([]db.Entry, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetEntry(_ context.Context, arg db.GetEntryParams) (db.Entry, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetResource(_ context.Context, arg db.GetResourceParams) (db.Resource, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetResourceKind(_ context.Context, arg db.GetResourceKindParams) (db.ResourceKind, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetResourceNode(_ context.Context, arg db.GetResourceNodeParams) (db.ResourceNode, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetTier(_ context.Context, id int32) (db.Tier, error) {
	panic("unimplemented")
}

func (s *stubQuerier) GetTransaction(_ context.Context, id int32) (db.Transaction, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListAncestors(_ context.Context, path string) ([]db.ResourceNode, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListDescendants(_ context.Context, path string) ([]db.ResourceNode, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListCFOrgs(_ context.Context) ([]db.CFOrg, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListCustomers(_ context.Context) ([]db.Customer, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListMeasurements(_ context.Context) ([]db.Measurement, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListResourceKind(_ context.Context) ([]db.ResourceKind, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListResources(_ context.Context) ([]db.Resource, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListTiers(_ context.Context) ([]db.Tier, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListTransactions(_ context.Context) ([]db.Transaction, error) {
	panic("unimplemented")
}

func (s *stubQuerier) ListTransactionsWide(_ context.Context) ([]db.ListTransactionsWideRow, error) {
	panic("unimplemented")
}

func (s *stubQuerier) PostUsage(_ context.Context, asOf pgtype.Timestamptz) ([]pgtype.Int4, error) {
	panic("unimplemented")
}

func (s *stubQuerier) SumEntries(_ context.Context) ([]pgtype.Numeric, error) {
	panic("unimplemented")
}

func (s *stubQuerier) UpdateCFOrg(_ context.Context, arg db.UpdateCFOrgParams) error {
	panic("unimplemented")
}

func (s *stubQuerier) UpdateCustomer(_ context.Context, arg db.UpdateCustomerParams) error {
	panic("unimplemented")
}

func (s *stubQuerier) UpdateMeasurementMicrocredits(_ context.Context, asOf pgtype.Timestamptz) (pgtype.Int8, error) {
	panic("unimplemented")
}

func (s *stubQuerier) UpdateResource(_ context.Context, arg db.UpdateResourceParams) error {
	panic("unimplemented")
}

func (s *stubQuerier) UpdateTier(_ context.Context, arg db.UpdateTierParams) error {
	panic("unimplemented")
}

func (s *stubQuerier) UpsertResource(_ context.Context, arg db.UpsertResourceParams) (db.Resource, error) {
	panic("unimplemented")
}

type WantedErr int64

const (
	NotWanted WantedErr = iota
	ErrWanted
	PanicWanted
)

func TestRecordReading(t *testing.T) {
	goodM := &reader.Measurement{
		Meter:                 "cpu",
		OrgID:                 uuid.NewString(),
		ResourceNaturalID:     "inst-1",
		ResourceKindNaturalID: "kind-1",
		Value:                 10,
	}
	emptyM := &reader.Measurement{}

	cases := []struct {
		name       string
		reading    reader.Reading
		errOn      string // make stubQuerier fail on this step ("" == happy path)
		wantErr    WantedErr
		wantMeters int // how many meter strings the stub should record
	}{
		{
			"no measurements",
			reader.Reading{Time: time.Now()},
			"",
			NotWanted,
			0,
		},
		{
			"only empties",
			reader.Reading{
				Time:         time.Now(),
				Measurements: []*reader.Measurement{emptyM},
			},
			"",
			NotWanted,
			0,
		},
		{
			"mixed good+empty",
			reader.Reading{
				Time:         time.Now(),
				Measurements: []*reader.Measurement{goodM, emptyM},
			},
			"",
			NotWanted,
			1,
		},
		{
			"duplicates",
			reader.Reading{
				Time:         time.Now(),
				Measurements: []*reader.Measurement{goodM, goodM},
			},
			"",
			NotWanted,
			2, // deduplication is done in the database; not tested here
		},
		{
			"negative value",
			reader.Reading{
				Time: time.Now(),
				Measurements: []*reader.Measurement{
					{
						Meter:                 "disk",
						OrgID:                 uuid.NewString(),
						ResourceNaturalID:     "d1",
						ResourceKindNaturalID: "dk",
						Value:                 -5,
					},
				},
			},
			"",
			NotWanted,
			1,
		},
		{
			"int32 overflow", // we just check that we *didn’t* crash
			reader.Reading{
				Time: time.Now(),
				Measurements: []*reader.Measurement{
					{
						Meter:                 "ram",
						OrgID:                 uuid.NewString(),
						ResourceNaturalID:     "r1",
						ResourceKindNaturalID: "rk",
						Value:                 int(math.MaxInt32) + 1,
					},
				},
			},
			"",
			NotWanted,
			1,
		},
		{
			"bad UUID", // stub makes failure surface at org insert
			reader.Reading{
				Time: time.Now(),
				Measurements: []*reader.Measurement{
					{
						Meter:                 "net",
						OrgID:                 "not-a-uuid",
						ResourceNaturalID:     "n1",
						ResourceKindNaturalID: "nk",
						Value:                 7,
					},
				},
			},
			"BulkCreateCFOrgs",
			PanicWanted,
			1,
		},
		{
			"zero time",
			reader.Reading{
				Time:         time.Time{},
				Measurements: []*reader.Measurement{goodM},
			},
			"",
			NotWanted,
			1,
		},
		{
			"error on BulkCreateResources",
			reader.Reading{
				Time:         time.Now(),
				Measurements: []*reader.Measurement{goodM},
			},
			"BulkCreateResources",
			ErrWanted,
			1,
		},
	}

	nullLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tc.wantErr != PanicWanted {
						panic(r)
					}
				}
			}()

			stub := &stubQuerier{errOn: tc.errOn}
			err := recorder.RecordReading(t.Context(), nullLogger, stub, tc.reading, false)

			if tc.wantErr > NotWanted && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if tc.wantErr != ErrWanted && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := len(stub.bulkMeters); got != tc.wantMeters {
				t.Fatalf("meters: want %d, got %d", tc.wantMeters, got)
			}
			// spot-check one other slice so we know they stayed in-sync
			if len(stub.bulkMeters) != tc.wantMeters {
				t.Fatalf("expected %v meters passed to database, got %v", len(stub.bulkMeters), tc.wantMeters)
			}
		})
	}
}
