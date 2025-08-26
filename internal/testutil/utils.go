package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func NewPgxUUID() pgtype.UUID {
	u := pgtype.UUID{}
	u.Scan(uuid.NewString())
	return u
}

// NewPgxTimestamptz returns a [pgtype.Timestamptz] based on the provided time.Time, or panics if the argument is not valid.
func NewPgxTimestamptz(at time.Time) pgtype.Timestamptz {
	tz := pgtype.Timestamptz{}
	err := tz.Scan(at)
	if err != nil {
		panic(err)
	}
	return tz
}
