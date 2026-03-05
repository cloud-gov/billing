package dbx

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TimeToTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  t,
		Valid: true,
	}
}

func ToUUID(s any) pgtype.UUID {
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

func ToBlankableUUID(u any) pgtype.UUID {
	uu := ToUUID(u)
	if !uu.Valid {
		if err := uu.Scan("FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF"); err != nil {
			panic(err)
		}
	}
	return uu
}

func UUIDishString(u any) string {
	switch ud := u.(type) {
	case string:
		return ud
	case pgtype.UUID:
		return ud.String()
	}
	return ""
}
