-- name: GetResource :one
SELECT * FROM resource
WHERE id = $1 LIMIT 1;

-- name: ListResources :many
SELECT * FROM resource
ORDER BY natural_id;

-- name: UpdateResource :exec
UPDATE resource
  set natural_id = $2,
  kind_id = $3,
  cf_org_id = $4
  WHERE id = $1;

-- name: DeleteResource :exec
DELETE FROM resource
WHERE id = $1;

-- name: CreateResource :one
INSERT INTO resource (
  natural_id, kind_id, cf_org_id
) VALUES (
  $1, $2, $3
)
RETURNING *;
