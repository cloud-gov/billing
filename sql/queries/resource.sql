  -- name: GetResource :one
SELECT * FROM resource
WHERE meter = $1 AND natural_id = $2 LIMIT 1;

-- name: ListResources :many
SELECT * FROM resource
ORDER BY natural_id;

-- name: UpdateResource :exec
UPDATE resource
  set kind_natural_id = $3,
    cf_org_id = $4
  WHERE meter = $1 AND natural_id = $2;

-- name: DeleteResource :exec
DELETE FROM resource
WHERE meter = $1 AND natural_id = $2;

-- name: CreateResources :exec
INSERT INTO resource (
  meter, natural_id, kind_natural_id, cf_org_id
) VALUES (
  $1, $2, $3, $4
) RETURNING *;

-- name: UpsertResource :one
-- UpsertResource upserts a Resource and creates minimal rows in foreign tables -- namely meter, cf_org, and resource_kind -- to which Resource has foreign keys. Efficient for single inserts. For bulk inserts, review Bulk* functions.
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
RETURNING *;

-- name: BulkCreateResources :exec
-- BulkCreateResources creates Resource rows in bulk with the minimum required columns. If a row with the given primary key already exists, that input item is ignored.
-- The bulk insert pattern using multiple arrays is sourced from: https://github.com/sqlc-dev/sqlc/issues/218#issuecomment-829263172
INSERT INTO resource (natural_id, meter, kind_natural_id, cf_org_id)
SELECT DISTINCT ON (r.meter, r.natural_id) *
FROM
  UNNEST(
    sqlc.arg(natural_ids)::text[],
    sqlc.arg(meters)::text[],
    sqlc.arg(kind_natural_ids)::text[],
    sqlc.arg(cf_org_ids)::uuid[]
  ) AS r(natural_id, meter, kind_natural_id, cf_org_id)
ON CONFLICT (meter, natural_id) DO NOTHING;
