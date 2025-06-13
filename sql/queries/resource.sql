-- name: GetResource :one
SELECT * FROM resource
WHERE id = $1 LIMIT 1;

-- name: GetResourceByNaturalID :one
SELECT * FROM resource
WHERE meter = $1 AND natural_id = $2 LIMIT 1;

-- name: ListResources :many
SELECT * FROM resource
ORDER BY natural_id;

-- name: UpdateResource :exec
UPDATE resource
  set meter = $2,
  natural_id = $3,
  kind_natural_id = $4,
  cf_org_id = $5
  WHERE id = $1;

-- name: DeleteResource :exec
DELETE FROM resource
WHERE id = $1;

-- name: CreateResource :one
INSERT INTO resource (
  natural_id, meter, kind_natural_id, cf_org_id
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;
