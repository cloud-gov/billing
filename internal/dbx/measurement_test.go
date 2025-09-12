package dbx_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/cloud-gov/billing/internal/db"
	"github.com/cloud-gov/billing/internal/dbx"
	. "github.com/cloud-gov/billing/internal/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CFOrg struct {
	// CustomerName is a key we can use to look up the generated ID in testData.CustomerIDs.
	CustomerName string
	db.CFOrg
}

type Transaction struct {
	CustomerName string
	db.Transaction
}

// testData is used to populate the database with rows required to perform a test.
type testData struct {
	Accounts []db.Account
	CFOrgs   []CFOrg
	// CustomerIDs maps from customer names (known in advance) to customer ID (returned by INSERT).
	CustomerIDs   map[string]int64
	Customers     []db.Customer
	Entries       []db.Entry
	Measurements  []db.Measurement
	Meters        []db.Meter
	Prices        []db.Price
	Readings      []db.Reading
	ResourceKinds []db.ResourceKind
	Resources     []db.Resource
	Transactions  []Transaction
}

func TestDBBoundsMonthPrev(t *testing.T) {
	tz, _ := time.LoadLocation("America/New_York")

	testCases := []struct {
		Name                string
		Tz                  *time.Location
		AsOf                pgtype.Timestamptz
		ExpectedPeriodStart pgtype.Timestamptz
		ExpectedPeriodEnd   pgtype.Timestamptz
	}{
		{
			Name:                "AsOf on exclusive upper bound",
			Tz:                  tz,
			AsOf:                PgTimestamptz(time.Date(2025, time.February, 1, 0, 0, 0, 0, tz)),
			ExpectedPeriodStart: PgTimestamptz(time.Date(2025, time.January, 1, 0, 0, 0, 0, tz)),
			ExpectedPeriodEnd:   PgTimestamptz(time.Date(2025, time.February, 1, 0, 0, 0, 0, tz)),
		},
		{
			Name:                "AsOf mid-month",
			Tz:                  tz,
			AsOf:                PgTimestamptz(time.Date(2025, time.February, 15, 0, 0, 0, 0, tz)),
			ExpectedPeriodStart: PgTimestamptz(time.Date(2025, time.January, 1, 0, 0, 0, 0, tz)),
			ExpectedPeriodEnd:   PgTimestamptz(time.Date(2025, time.February, 1, 0, 0, 0, 0, tz)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Arrange
			conn, err := pgxpool.New(t.Context(), "")
			if err != nil {
				t.Fatal("creating database connection failed", err)
			}
			q := newTx(t, conn, false)

			// Act
			result, err := q.BoundsMonthPrev(t.Context(), tc.AsOf)

			// Assert
			if err != nil {
				t.Fatal("error calling the function under test", err)
			}

			if !result.PeriodStart.Time.Equal(tc.ExpectedPeriodStart.Time) {
				t.Fatalf("expected period start %v, got %v", tc.ExpectedPeriodStart.Time, result.PeriodStart.Time)
			}
			if !result.PeriodEnd.Time.Equal(tc.ExpectedPeriodEnd.Time) {
				t.Fatalf("expected period end %v, got %v", tc.ExpectedPeriodEnd.Time, result.PeriodEnd.Time)
			}
		})
	}
}

// newTx creates a new [dbx.Querier] and starts a transaction. Each test should call this function separately so it runs in isolation from the rest. Most callers should `defer rollback()` so database changes are rolled back and tests do not interfere with each other. To commit the results instead -- for example, to debug a failing test -- call `defer commit()`.
//
// Note that isolation is not perfect between tests, even when transactions are rolled back. For example, the sequence used to generate an auto-incrementing ID will not decrement when a transaction is rolled back; the next transaction will start from the incremented sequence. Avoid hardcoding generated IDs for this reason.
func newTx(t *testing.T, conn *pgxpool.Pool, commit bool) dbx.Querier {
	t.Helper()
	// t.Context() can be cancelled before we have a chance to commit, so create a new context instead.
	ctx := context.Background()

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal("begin transaction failed", err)
	}
	// Test acquiring a connection from the pool.
	if err = conn.Ping(ctx); err != nil {
		t.Fatal("database connection ping failed")
	}

	if commit {
		t.Cleanup(func() {
			err = tx.Commit(ctx)
			if err != nil {
				t.Fatal("failed to commit transaction: ", err)
			}
		})
	} else {
		t.Cleanup(func() {
			err = tx.Rollback(ctx)
			// Ignore error if the transaction was already committed
			if err != nil && !errors.Is(pgx.ErrTxClosed, err) {
				t.Fatal("failed to rollback transaction: ", err)
			}
		})
	}

	return dbx.NewQuerier(db.New(conn)).WithTx(tx)
}

func TestDBUpdateMeasurementMicrocredits(t *testing.T) {
	// Arrange
	conn, err := pgxpool.New(t.Context(), "")
	if err != nil {
		t.Fatal("creating database connection failed", err)
	}
	q := newTx(t, conn, false)

	var (
		orgID      = PgUUID()
		meterName  = "meter-1"
		kindID     = "kind-1"
		priceID    = int32(1)
		readingID1 = int32(1)
		readingID2 = int32(2)
		readingID3 = int32(3)
		readingID4 = int32(4)
		readingID5 = int32(5)
		readingID6 = int32(6)
		resourceID = "resource-1"
		tz, _      = time.LoadLocation("America/New_York")
		utc, _     = time.LoadLocation("")
		priceLower = PgTimestamptz(time.Date(2024, time.March, 1, 0, 0, 0, 0, tz))
		priceUpper = PgTimestamptz(time.Date(2026, time.March, 1, 0, 0, 0, 0, tz))
		asOf       = PgTimestamptz(time.Date(2025, time.March, 2, 0, 0, 0, 0, tz))
	)

	td := testData{
		CFOrgs: []CFOrg{
			{
				CFOrg: db.CFOrg{
					ID: orgID,
				},
			},
		},
		Meters: []db.Meter{
			{
				Name: meterName,
			},
		},
		ResourceKinds: []db.ResourceKind{
			{
				Meter:     meterName,
				NaturalID: kindID,
				Name:      PgText(""),
			},
		},
		Prices: []db.Price{
			{
				Meter:               meterName,
				ID:                  priceID,
				KindNaturalID:       kindID,
				MicrocreditsPerUnit: 8,
				UnitOfMeasure:       "hours",
				Unit:                2, // Just for checking the math.
				ValidDuring: pgtype.Range[pgtype.Timestamptz]{
					Lower:     priceLower,
					Upper:     priceUpper,
					LowerType: pgtype.Inclusive,
					UpperType: pgtype.Exclusive,
					Valid:     true,
				},
			},
		},
		Readings: []db.Reading{
			{
				ID:        readingID1,
				CreatedAt: PgTimestamp(time.Date(2025, time.January, 1, 0, 0, 0, 0, utc)),
				// One month before bounds
			},
			{
				ID:        readingID2,
				CreatedAt: PgTimestamp(time.Date(2025, time.February, 1, 0, 0, 0, 0, utc)),
				// Correct first day of bounds, but before start of day ET
			},
			{
				ID:        readingID3,
				CreatedAt: PgTimestamp(time.Date(2025, time.February, 1, 5, 0, 0, 0, utc)),
				// Correct first day of bounds, and at start of day ET, inclusive
			},
			{
				ID:        readingID4,
				CreatedAt: PgTimestamp(time.Date(2025, time.February, 3, 0, 0, 0, 0, utc)),
				// Correct month, mid-month
			},
			{
				ID:        readingID5,
				CreatedAt: PgTimestamp(time.Date(2025, time.March, 1, 0, 0, 0, 0, utc)),
				// Next month in UTC, still previous day in ET
			},
			{
				ID:        readingID6,
				CreatedAt: PgTimestamp(time.Date(2025, time.March, 1, 5, 0, 0, 0, utc)),
				// Next month in UTC and ET
			},
		},
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
				Value:             7,
				ReadingID:         readingID1,
			},
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             7,
				ReadingID:         readingID2,
			},
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             7,
				ReadingID:         readingID3,
			},
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             7,
				ReadingID:         readingID4,
			},
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             7,
				ReadingID:         readingID5,
			},
			{
				Meter:             meterName,
				ResourceNaturalID: resourceID,
				Value:             7,
				ReadingID:         readingID6,
			},
		},
	}

	createTestData(t, q, td)

	// Act
	updated, err := q.UpdateMeasurementMicrocredits(t.Context(), asOf)

	// Assert
	if err != nil {
		t.Fatal("error occured while calling function under test", err)
	}
	if updated.Int64 != 3 {
		t.Fatalf("expected %v rows updated, got %v", 2, updated.Int64)
	}
	ms, err := q.ListMeasurements(t.Context())
	if err != nil {
		t.Fatal("error listing measurements (this is a problem with the test)", err)
	}

	ms1 := ms[measurementFromReadingID(ms, readingID1)]
	if ms1.AmountMicrocredits.Valid {
		t.Logf("expected measurement 1 AmountMicrocredits to be NULL, but was NOT NULL")
		t.Fail()
	}
	ms2 := ms[measurementFromReadingID(ms, readingID2)]
	if ms2.AmountMicrocredits.Valid {
		t.Logf("expected measurement 2 AmountMicrocredits to be NULL, but was NOT NULL")
		t.Fail()
	}
	ms3 := ms[measurementFromReadingID(ms, readingID3)]
	if !ms3.AmountMicrocredits.Valid {
		t.Logf("expected measurement 3 AmountMicrocredits to be NOT NULL, but was NULL")
		t.Fail()
	}
	var expectedAmount int64 = 7 * 8 / 2
	if ms3.AmountMicrocredits.Int64 != expectedAmount {
		t.Logf("expected measurement 3 AmountMicrocredits to be %v, got %v", expectedAmount, ms3.AmountMicrocredits.Int64)
	}
	ms4 := ms[measurementFromReadingID(ms, readingID4)]
	if !ms4.AmountMicrocredits.Valid {
		t.Logf("expected measurement 4 AmountMicrocredits to be NOT NULL, but was NULL")
		t.Fail()
	}
	if ms4.AmountMicrocredits.Int64 != expectedAmount {
		t.Logf("expected measurement 4 AmountMicrocredits to be %v, got %v", expectedAmount, ms4.AmountMicrocredits.Int64)
	}
	ms5 := ms[measurementFromReadingID(ms, readingID5)]
	if !ms5.AmountMicrocredits.Valid {
		t.Logf("expected measurement 5 AmountMicrocredits to be NOT NULL, but was NULL")
		t.Fail()
	}
	if ms5.AmountMicrocredits.Int64 != expectedAmount {
		t.Logf("expected measurement 5 AmountMicrocredits to be %v, got %v", expectedAmount, ms5.AmountMicrocredits.Int64)
	}
	ms6 := ms[measurementFromReadingID(ms, readingID6)]
	if ms6.AmountMicrocredits.Valid {
		t.Logf("expected measurement 6 AmountMicrocredits to be NULL, but was NOT NULL")
		t.Fail()
	}
}

func measurementFromReadingID(m []db.Measurement, id int32) int {
	return slices.IndexFunc(m, func(e db.Measurement) bool {
		return e.ReadingID == id
	})
}
func TestDBPostUsage(t *testing.T) {
	_, _ = time.LoadLocation("America/New_York")

	var (
		customer1Name      = "customer1"
		customer2Name      = "customer2"
		org1ID             = PgUUID()
		org2ID             = PgUUID()
		org3ID             = PgUUID()
		meterName          = "meter-1"
		kindID             = "kind-1"
		readingID1         = int32(1)
		readingID2         = int32(2)
		readingID3         = int32(3)
		readingID4         = int32(4)
		readingID5         = int32(5)
		readingID6         = int32(6)
		resource1ID        = "resource-1"
		resource2ID        = "resource-2"
		amountMicrocredits = PgInt8(56)
		tz, _              = time.LoadLocation("America/New_York")
		utc, _             = time.LoadLocation("")
		asOf               = PgTimestamptz(time.Date(2025, time.March, 1, 0, 0, 0, 0, tz))
		periodEnd          = PgTimestamptz(time.Date(2025, time.March, 1, 0, 0, 0, 0, tz))
	)

	testCases := []struct {
		Name   string
		AsOf   pgtype.Timestamptz
		Before testData
		After  testData
		Return int // Number of records returned.
	}{
		{
			Name: "boundary test: 1 customer, multiple readings over time",
			AsOf: asOf,
			Before: testData{
				CustomerIDs: map[string]int64{},
				Customers: []db.Customer{
					{
						Name: customer1Name,
					},
				},
				CFOrgs: []CFOrg{
					{
						CustomerName: customer1Name,
						CFOrg: db.CFOrg{
							ID: org1ID,
						},
					},
				},
				Meters: []db.Meter{
					{
						Name: meterName,
					},
				},
				ResourceKinds: []db.ResourceKind{
					{
						Meter:     meterName,
						NaturalID: kindID,
						Name:      PgText(""),
					},
				},
				Readings: []db.Reading{
					{
						ID:        readingID1,
						CreatedAt: PgTimestamp(time.Date(2025, time.January, 1, 0, 0, 0, 0, utc)),
						// One month before bounds
					},
					{
						ID:        readingID2,
						CreatedAt: PgTimestamp(time.Date(2025, time.February, 1, 0, 0, 0, 0, utc)),
						// Correct first day of bounds, but before start of day ET
					},
					{
						ID:        readingID3,
						CreatedAt: PgTimestamp(time.Date(2025, time.February, 1, 5, 0, 0, 0, utc)),
						// Correct first day of bounds, and at start of day ET, inclusive
					},
					{
						ID:        readingID4,
						CreatedAt: PgTimestamp(time.Date(2025, time.February, 3, 0, 0, 0, 0, utc)),
						// Correct month, mid-month
					},
					{
						ID:        readingID5,
						CreatedAt: PgTimestamp(time.Date(2025, time.March, 1, 0, 0, 0, 0, utc)),
						// Next month in UTC, still previous day in ET
					},
					{
						ID:        readingID6,
						CreatedAt: PgTimestamp(time.Date(2025, time.March, 1, 5, 0, 0, 0, utc)),
						// Next month in UTC and ET
					},
				},
				Resources: []db.Resource{
					{
						Meter:         meterName,
						NaturalID:     resource1ID,
						KindNaturalID: kindID,
						CFOrgID:       org1ID,
					},
				},
				Measurements: []db.Measurement{
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              7,
						ReadingID:          readingID1,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              7,
						ReadingID:          readingID2,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              7,
						ReadingID:          readingID3,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              7,
						ReadingID:          readingID4,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              7,
						ReadingID:          readingID5,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              7,
						ReadingID:          readingID6,
						AmountMicrocredits: amountMicrocredits,
					},
				},
			},
			After: testData{
				// Accounts are automatically created when the CF Org is created
				Accounts: []db.Account{
					{
						ID:         2,
						CustomerID: 1,
						Type:       201,
					},
					{
						ID:         4,
						CustomerID: 1,
						Type:       401,
					},
				},
				Entries: []db.Entry{
					{
						TransactionID:      1,
						AccountID:          2,
						Direction:          -1,
						AmountMicrocredits: PgInt8(amountMicrocredits.Int64 * 3),
					},
					{
						TransactionID:      1,
						AccountID:          4,
						Direction:          1,
						AmountMicrocredits: PgInt8(amountMicrocredits.Int64 * 3),
					},
				},
				Transactions: []Transaction{
					{
						CustomerName: customer1Name,
						Transaction: db.Transaction{
							OccurredAt:  periodEnd,
							Description: PgText("Monthly usage 2025-02-01--2025-03-01"),
							Type:        db.TransactionTypeUsagePost,
						},
					},
				},
			},
			Return: 1,
		},
		{
			Name: "multiple customers with multiple readings in bounds",
			AsOf: asOf,
			Before: testData{
				CustomerIDs: map[string]int64{},
				Customers: []db.Customer{
					{
						Name: customer1Name,
					},
					{
						Name: customer2Name,
					},
				},
				CFOrgs: []CFOrg{
					{
						CustomerName: customer1Name,
						CFOrg: db.CFOrg{
							ID: org3ID,
						},
					},
					{
						CustomerName: customer2Name,
						CFOrg: db.CFOrg{
							ID: org2ID,
						},
					},
				},
				Meters: []db.Meter{
					{
						Name: meterName,
					},
				},
				ResourceKinds: []db.ResourceKind{
					{
						Meter:     meterName,
						NaturalID: kindID,
						Name:      PgText(""),
					},
				},
				Readings: []db.Reading{
					{
						ID:        readingID1,
						CreatedAt: PgTimestamp(time.Date(2025, time.February, 1, 0, 0, 0, 0, utc)),
						// Correct first day of bounds, but before start of day ET
					},
					{
						ID:        readingID2,
						CreatedAt: PgTimestamp(time.Date(2025, time.February, 1, 5, 0, 0, 0, utc)),
						// Correct first day of bounds, and at start of day ET, inclusive
					},
					{
						ID:        readingID3,
						CreatedAt: PgTimestamp(time.Date(2025, time.March, 1, 0, 0, 0, 0, utc)),
						// Next month in UTC, still previous day in ET
					},
				},
				Resources: []db.Resource{
					{
						Meter:         meterName,
						NaturalID:     resource1ID,
						KindNaturalID: kindID,
						CFOrgID:       org3ID,
					},
					{
						Meter:         meterName,
						NaturalID:     resource2ID,
						KindNaturalID: kindID,
						CFOrgID:       org2ID,
					},
				},
				Measurements: []db.Measurement{
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              2,
						ReadingID:          readingID1,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              2,
						ReadingID:          readingID1,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource1ID,
						Value:              2,
						ReadingID:          readingID2,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource2ID,
						Value:              2,
						ReadingID:          readingID2,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource2ID,
						Value:              2,
						ReadingID:          readingID3,
						AmountMicrocredits: amountMicrocredits,
					},
					{
						Meter:              meterName,
						ResourceNaturalID:  resource2ID,
						Value:              2,
						ReadingID:          readingID3,
						AmountMicrocredits: amountMicrocredits,
					},
				},
			},
			After: testData{
				Accounts: []db.Account{
					{
						ID:         2,
						CustomerID: 1,
						Type:       201,
					},
					{
						ID:         4,
						CustomerID: 1,
						Type:       401,
					},
					{
						ID:         6,
						CustomerID: 2,
						Type:       201,
					},
					{
						ID:         8,
						CustomerID: 2,
						Type:       401,
					},
				},
			},
			Return: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Arrange
			conn, err := pgxpool.New(t.Context(), "")
			if err != nil {
				t.Fatal("creating database connection failed", err)
			}
			q := newTx(t, conn, false)

			createTestData(t, q, tc.Before)
			copyTestDataIDs(&tc.Before, &tc.After)
			// Act
			results, err := q.PostUsage(t.Context(), tc.AsOf)
			if err != nil {
				t.Fatal("error calling function under test: ", err)
			}

			// Assert
			assertDBContains(t, q, tc.After)

			if len(results) != tc.Return {
				t.Logf("expected %v rows returned, got %v", tc.Return, len(results))
				t.Fail()
			}
		})
	}
}

func copyTestDataIDs(before, after *testData) {
	after.CustomerIDs = before.CustomerIDs
}

// createTestData creates a row for each struct in the provided data. It uses the Create methods from q, which may have additional effects.
func createTestData(t *testing.T, q db.Querier, td testData) {
	t.Helper()
	// Order of creation depends on fkey dependencies.
	for _, v := range td.Customers {
		id, err := q.CreateCustomer(t.Context(), v.Name)
		if err != nil {
			t.Fatal("creating customer failed:", err)
		}
		td.CustomerIDs[v.Name] = id
	}
	for _, v := range td.CFOrgs {
		if v.CustomerName != "" {
			customerID, ok := td.CustomerIDs[v.CustomerName]
			if !ok {
				t.Fatal("creating CFOrg: could not look up customer ID by name in CustomerIDs map (did you declare it in testData and include a name?)")
			}
			v.CFOrg.CustomerID = PgInt8(customerID)
		}
		_, err := q.CreateCFOrg(t.Context(), db.CreateCFOrgParams(v.CFOrg))
		if err != nil {
			t.Fatalf("creating CF org failed: %T %v", err, err)
		}
	}
	for _, v := range td.Meters {
		_, err := q.CreateMeter(t.Context(), v.Name)
		if err != nil {
			t.Fatal("creating meter failed:", err)
		}
	}
	for _, v := range td.ResourceKinds {
		_, err := q.CreateResourceKind(t.Context(), db.CreateResourceKindParams{
			Meter:     v.Meter,
			NaturalID: v.NaturalID,
		})
		if err != nil {
			t.Fatal("creating resource kind failed:", err)
		}
	}
	for _, v := range td.Prices {
		_, err := q.CreatePriceWithID(t.Context(), db.CreatePriceWithIDParams(v))
		if err != nil {
			t.Fatal("creating price failed:", err)
		}
	}
	for _, v := range td.Readings {
		_, err := q.CreateReadingWithID(t.Context(), db.CreateReadingWithIDParams{
			ID:        v.ID,
			CreatedAt: v.CreatedAt,
			Periodic:  v.Periodic,
		})
		if err != nil {
			t.Fatal("creating reading failed:", err)
		}
	}
	for _, v := range td.Resources {
		err := q.CreateResources(t.Context(), db.CreateResourcesParams(v))
		if err != nil {
			t.Fatal("creating resource failed:", err)
		}
	}
	for _, v := range td.Measurements {
		q.CreateMeasurement(t.Context(), db.CreateMeasurementParams{
			ReadingID:          v.ReadingID,
			Meter:              v.Meter,
			ResourceNaturalID:  v.ResourceNaturalID,
			Value:              v.Value,
			AmountMicrocredits: v.AmountMicrocredits,
		})
	}
}

// assertDBContains checks if each element in td exists in the database. Missing elements are collected with t.Fail(). The test is only ended with t.Fatal() if a query returns an error.
// TODO: Does not check Meters, Prices, Readings (no Get methods on q yet)
func assertDBContains(t *testing.T, q db.Querier, td testData) {
	t.Helper()
	for _, want := range td.Customers {
		have, err := q.GetCustomer(t.Context(), want.ID)
		if err != nil {
			t.Fatal("getting customer", err)
		}
		if !cmp.Equal(have, want) {
			t.Logf("expected customer %v, got %v", want, have)
			t.Fail()
		}
	}
	for _, want := range td.ResourceKinds {
		have, err := q.GetResourceKind(t.Context(), db.GetResourceKindParams{
			Meter:     want.Meter,
			NaturalID: want.NaturalID,
		})
		if err != nil {
			t.Fatal("getting ResourceKind", err)
		}
		if !cmp.Equal(want, have) {
			t.Logf("expected ResourceKind %v, got %v", want, have)
			t.Fail()
		}
	}
	for _, want := range td.CFOrgs {
		have, err := q.GetCFOrg(t.Context(), want.ID)
		if err != nil {
			t.Fatal("getting CFOrg", err)
		}
		if !cmp.Equal(want, have) {
			t.Logf("expected CFOrg %v, got %v", want, have)
			t.Fail()
		}
	}
	// We don't know which transaction IDs are associated with which customers, so use list instead of getting by ID.
	txns, err := q.ListTransactions(t.Context())
	if err != nil {
		t.Fatal("listing transactions: ", err)
	}
	for _, want := range td.Transactions {
		// Check if a transaction exists that matches every field, besides ID, which we do not know ahead of time.
		have := slices.IndexFunc(txns, func(have db.Transaction) bool {
			// Create a db.Transaction for easy comparison
			custID, ok := td.CustomerIDs[want.CustomerName]
			if !ok {
				t.Fatalf("could not find customer ID for name %v", want.CustomerName)
			}
			want.Transaction.CustomerID = PgInt8(custID)
			return cmp.Equal(want.Transaction, have, cmp.Comparer(func(x, y int32) bool {
				// IDs are the only int32 field we're comparing; always return true
				return true
			}))
		})
		if have == -1 {
			t.Logf("expected Transaction %v, not found amongst: %v", want.Transaction, txns)
			t.Fail()
		}
	}
	for _, want := range td.Entries {
		have, err := q.GetEntry(t.Context(), db.GetEntryParams{
			TransactionID: want.TransactionID,
			AccountID:     want.AccountID,
		})
		if err != nil {
			t.Fatal("getting entry", err)
		}
		if !cmp.Equal(want, have) {
			t.Logf("expected Entry %v, got %v", want, have)
			t.Fail()
		}
	}
}
