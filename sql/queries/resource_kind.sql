-- name: GetResourceKind :one
SELECT * FROM resource_kind
WHERE meter = $1 AND natural_id = $2 LIMIT 1;

-- name: ListResourceKind :many
SELECT * FROM resource_kind
ORDER BY natural_id;

-- name: UpdateResourceKind :exec
UPDATE resource_kind
  set credits = $3,
  amount = $4,
  unit_of_measure = $5
  WHERE meter = $1 AND natural_id = $2;

-- name: DeleteResourceKind :exec
DELETE FROM resource_kind
WHERE meter = $1 AND natural_id = $2;

-- name: CreateResourceKind :one
INSERT INTO resource_kind (
  meter, natural_id, credits, amount, unit_of_measure
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;
