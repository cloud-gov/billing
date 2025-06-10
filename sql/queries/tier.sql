-- name: GetTier :one
SELECT * FROM tier
WHERE id = $1 LIMIT 1;

-- name: ListTiers :many
SELECT * FROM tier
ORDER BY name;

-- name: UpdateTier :exec
UPDATE tier
  set name = $2,
  tier_credits = $3
  WHERE id = $1;

-- name: DeleteTier :exec
DELETE FROM tier
WHERE id = $1;

-- name: CreateTier :one
INSERT INTO tier (
  name, tier_credits
) VALUES (
  $1, $2
)
RETURNING *;
