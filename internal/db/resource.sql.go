// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: resource.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const bulkCreateResources = `-- name: BulkCreateResources :exec
INSERT INTO resource (natural_id, meter, kind_natural_id, cf_org_id)
SELECT DISTINCT ON (r.meter, r.natural_id) natural_id, meter, kind_natural_id, cf_org_id
FROM
  UNNEST(
    $1::text[],
    $2::text[],
    $3::text[],
    $4::uuid[]
  ) AS r(natural_id, meter, kind_natural_id, cf_org_id)
ON CONFLICT (meter, natural_id) DO NOTHING
`

type BulkCreateResourcesParams struct {
	NaturalIds     []string
	Meters         []string
	KindNaturalIds []string
	CfOrgIds       []pgtype.UUID
}

// BulkCreateResources creates Resource rows in bulk with the minimum required columns. If a row with the given primary key already exists, that input item is ignored.
// The bulk insert pattern using multiple arrays is sourced from: https://github.com/sqlc-dev/sqlc/issues/218#issuecomment-829263172
func (q *Queries) BulkCreateResources(ctx context.Context, arg BulkCreateResourcesParams) error {
	_, err := q.db.Exec(ctx, bulkCreateResources,
		arg.NaturalIds,
		arg.Meters,
		arg.KindNaturalIds,
		arg.CfOrgIds,
	)
	return err
}

const createResources = `-- name: CreateResources :exec
INSERT INTO resource (
  meter, natural_id, kind_natural_id, cf_org_id
) VALUES (
  $1, $2, $3, $4
) RETURNING meter, natural_id, kind_natural_id, cf_org_id
`

type CreateResourcesParams struct {
	Meter         string
	NaturalID     string
	KindNaturalID string
	CFOrgID       pgtype.UUID
}

func (q *Queries) CreateResources(ctx context.Context, arg CreateResourcesParams) error {
	_, err := q.db.Exec(ctx, createResources,
		arg.Meter,
		arg.NaturalID,
		arg.KindNaturalID,
		arg.CFOrgID,
	)
	return err
}

const deleteResource = `-- name: DeleteResource :exec
DELETE FROM resource
WHERE meter = $1 AND natural_id = $2
`

type DeleteResourceParams struct {
	Meter     string
	NaturalID string
}

func (q *Queries) DeleteResource(ctx context.Context, arg DeleteResourceParams) error {
	_, err := q.db.Exec(ctx, deleteResource, arg.Meter, arg.NaturalID)
	return err
}

const listResources = `-- name: ListResources :many
SELECT meter, natural_id, kind_natural_id, cf_org_id FROM resource
ORDER BY natural_id
`

func (q *Queries) ListResources(ctx context.Context) ([]Resource, error) {
	rows, err := q.db.Query(ctx, listResources)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Resource
	for rows.Next() {
		var i Resource
		if err := rows.Scan(
			&i.Meter,
			&i.NaturalID,
			&i.KindNaturalID,
			&i.CFOrgID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateResource = `-- name: UpdateResource :exec
UPDATE resource
  set kind_natural_id = $3,
    cf_org_id = $4
  WHERE meter = $1 AND natural_id = $2
`

type UpdateResourceParams struct {
	Meter         string
	NaturalID     string
	KindNaturalID string
	CFOrgID       pgtype.UUID
}

func (q *Queries) UpdateResource(ctx context.Context, arg UpdateResourceParams) error {
	_, err := q.db.Exec(ctx, updateResource,
		arg.Meter,
		arg.NaturalID,
		arg.KindNaturalID,
		arg.CFOrgID,
	)
	return err
}

const upsertResource = `-- name: UpsertResource :one
WITH upsert_meter AS (
  INSERT INTO meter(name)
  VALUES ($2)
  ON CONFLICT (name) DO NOTHING
), upsert_org AS (
  INSERT INTO cf_org(id)
  VALUES ($4)
  ON CONFLICT (id) DO NOTHING
), upsert_kind AS (
  INSERT INTO resource_kind(meter, natural_id)
  VALUES ($2, $3)
  ON CONFLICT (meter, natural_id) DO NOTHING
)
INSERT INTO resource (
  natural_id, meter, kind_natural_id, cf_org_id
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (natural_id, meter) DO UPDATE SET
  kind_natural_id = EXCLUDED.kind_natural_id,
  cf_org_id = EXCLUDED.cf_org_id
RETURNING meter, natural_id, kind_natural_id, cf_org_id
`

type UpsertResourceParams struct {
	NaturalID     string
	Meter         string
	KindNaturalID string
	CFOrgID       pgtype.UUID
}

// UpsertResource upserts a Resource and creates minimal rows in foreign tables -- namely meter, cf_org, and resource_kind -- to which Resource has foreign keys. Efficient for single inserts. For bulk inserts, review Bulk* functions.
func (q *Queries) UpsertResource(ctx context.Context, arg UpsertResourceParams) (Resource, error) {
	row := q.db.QueryRow(ctx, upsertResource,
		arg.NaturalID,
		arg.Meter,
		arg.KindNaturalID,
		arg.CFOrgID,
	)
	var i Resource
	err := row.Scan(
		&i.Meter,
		&i.NaturalID,
		&i.KindNaturalID,
		&i.CFOrgID,
	)
	return i, err
}
