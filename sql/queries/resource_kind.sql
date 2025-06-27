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

-- name: BulkCreateResourceKinds :exec
-- BulkCreateResourceKinds creates ResourceKind rows in bulk with the minimum required columns. If a row with the given primary key already exists, that input item is ignored.
-- The bulk insert pattern using multiple arrays is sourced from: https://github.com/sqlc-dev/sqlc/issues/218#issuecomment-829263172
INSERT INTO resource_kind (meter, natural_id)
SELECT DISTINCT ON (r.meter, r.natural_id) *
FROM
  UNNEST(
    sqlc.arg(meters)::text[],
    sqlc.arg(natural_ids)::text[]
  ) AS r(meter, natural_id)
ON CONFLICT (meter, natural_id) DO NOTHING;
