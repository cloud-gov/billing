// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CFOrg struct {
	ID           uuid.UUID
	Name         string
	TierID       int32
	CreditsQuota int64
	CreditsUsed  int64
	CustomerID   int64
}

type Customer struct {
	ID   int64
	Name string
}

type Measurement struct {
	ReadingID  int32
	ResourceID int32
	Value      int32
}

type Reading struct {
	ID        int32
	CreatedAt time.Time
}

type Resource struct {
	ID        int32
	NaturalID sql.NullString
	KindID    int32
	CFOrgID   uuid.UUID
}

type ResourceKind struct {
	ID            int32
	NaturalID     sql.NullString
	Credits       sql.NullInt32
	Amount        sql.NullInt32
	UnitOfMeasure string
}

type Tier struct {
	ID          int32
	Name        string
	TierCredits int64
}
