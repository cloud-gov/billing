// testutil provides functions that are reused in multiple tests. You can import it with a period (. "github.com/cloud-gov/billing/internal/testutil") to shorten calls.
package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func PgUUID() pgtype.UUID {
	u := pgtype.UUID{}
	u.Scan(uuid.NewString())
	return u
}

// PgTimestamptz returns a [pgtype.Timestamptz] based on the provided time.Time, or panics if the argument is not valid.
func PgTimestamptz(at time.Time) pgtype.Timestamptz {
	tz := pgtype.Timestamptz{}
	err := tz.Scan(at)
	if err != nil {
		panic(err)
	}
	return tz
}

// PgTimestamp returns a [pgtype.Timestamp] based on the provided time.Time, or panics if the argument is not valid.
func PgTimestamp(at time.Time) pgtype.Timestamp {
	ts := pgtype.Timestamp{}
	err := ts.Scan(at)
	if err != nil {
		panic(err)
	}
	return ts
}

func PgInt8(v int64) pgtype.Int8 {
	return pgtype.Int8{
		Int64: v,
		Valid: true,
	}
}

func PgInt4(v int32) pgtype.Int4 {
	return pgtype.Int4{
		Int32: v,
		Valid: true,
	}
}
