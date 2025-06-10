-- name: GetResourceKind :one
SELECT * FROM resource_kind
WHERE id = $1 LIMIT 1;

-- name: ListResourceKind :many
SELECT * FROM resource_kind
ORDER BY natural_id;

-- name: UpdateResourceKind :exec
UPDATE resource_kind
  set natural_id = $2,
  credits = $3,
  amount = $4,
  unit_of_measure = $5
  WHERE id = $1;

-- name: DeleteResourceKind :exec
DELETE FROM resource_kind
WHERE id = $1;

-- name: CreateResourceKind :one
INSERT INTO resource_kind (
  natural_id, credits, amount, unit_of_measure
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;
