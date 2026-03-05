-- name: CreateCFOrg :one
INSERT INTO cf_org (id, name, customer_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCFOrg :one
SELECT * FROM cf_org
WHERE id = $1;

-- name: ListCFOrgs :many
SELECT * FROM cf_org
ORDER BY name;

-- name: UpdateCFOrg :exec
UPDATE cf_org
SET name = $2
WHERE id = $1;

-- name: DeleteCFOrg :exec
DELETE FROM cf_org
WHERE id = $1;

-- name: BulkCreateCFOrgs :exec
-- BulkCreateCFOrgs creates CFOrg rows in bulk with the minimum required columns. If a row with the given primary key already exists, that input item is ignored.
INSERT INTO cf_org (id, name)
SELECT
  id,
  name
FROM
  UNNEST(
    sqlc.arg(ids)::uuid[],
    sqlc.arg(names)::text[]
  ) AS o (id, name)
ON CONFLICT (id) DO UPDATE
  SET name = excluded.name;
