package dbx

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func UtilUUID(s any) pgtype.UUID {
	switch s := s.(type) {
	case pgtype.UUID:
		return s
	case nil:
		return pgtype.UUID{}
	case string:
		if s == "" {
			return pgtype.UUID{}
		}
	}
	u := pgtype.UUID{}
	if err := u.Scan(s); err != nil {
		panic(fmt.Errorf("failed to convert `%#v` to UUID: %w", s, err))
	}
	return u
}

func UtilTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}
