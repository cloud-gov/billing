// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: cf_org.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createCFOrg = `-- name: CreateCFOrg :one
INSERT INTO cf_org (
  name, tier_id, credits_quota, credits_used, customer_id
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id, name, tier_id, credits_quota, credits_used, customer_id
`

type CreateCFOrgParams struct {
	Name         string
	TierID       int32
	CreditsQuota int64
	CreditsUsed  int64
	CustomerID   int64
}

func (q *Queries) CreateCFOrg(ctx context.Context, arg CreateCFOrgParams) (CFOrg, error) {
	row := q.db.QueryRowContext(ctx, createCFOrg,
		arg.Name,
		arg.TierID,
		arg.CreditsQuota,
		arg.CreditsUsed,
		arg.CustomerID,
	)
	var i CFOrg
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.TierID,
		&i.CreditsQuota,
		&i.CreditsUsed,
		&i.CustomerID,
	)
	return i, err
}

const deleteCFOrg = `-- name: DeleteCFOrg :exec
DELETE FROM cf_org
WHERE id = $1
`

func (q *Queries) DeleteCFOrg(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteCFOrg, id)
	return err
}

const getCFOrg = `-- name: GetCFOrg :one
SELECT id, name, tier_id, credits_quota, credits_used, customer_id FROM cf_org
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetCFOrg(ctx context.Context, id uuid.UUID) (CFOrg, error) {
	row := q.db.QueryRowContext(ctx, getCFOrg, id)
	var i CFOrg
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.TierID,
		&i.CreditsQuota,
		&i.CreditsUsed,
		&i.CustomerID,
	)
	return i, err
}

const listCFOrgs = `-- name: ListCFOrgs :many
SELECT id, name, tier_id, credits_quota, credits_used, customer_id FROM cf_org
ORDER BY name
`

func (q *Queries) ListCFOrgs(ctx context.Context) ([]CFOrg, error) {
	rows, err := q.db.QueryContext(ctx, listCFOrgs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CFOrg
	for rows.Next() {
		var i CFOrg
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.TierID,
			&i.CreditsQuota,
			&i.CreditsUsed,
			&i.CustomerID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCFOrg = `-- name: UpdateCFOrg :exec
UPDATE cf_org
  set name = $2,
  tier_id = $3,
  credits_quota = $4,
  credits_used = $5
WHERE id = $1
`

type UpdateCFOrgParams struct {
	ID           uuid.UUID
	Name         string
	TierID       int32
	CreditsQuota int64
	CreditsUsed  int64
}

func (q *Queries) UpdateCFOrg(ctx context.Context, arg UpdateCFOrgParams) error {
	_, err := q.db.ExecContext(ctx, updateCFOrg,
		arg.ID,
		arg.Name,
		arg.TierID,
		arg.CreditsQuota,
		arg.CreditsUsed,
	)
	return err
}
