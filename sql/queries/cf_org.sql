-- name: GetCFOrg :one
SELECT * FROM cf_org
WHERE id = $1 LIMIT 1;

-- name: ListCFOrgs :many
SELECT * FROM cf_org
ORDER BY name;

-- name: UpdateCFOrg :exec
UPDATE cf_org
  set name = $2,
  tier_id = $3,
  credits_quota = $4,
  credits_used = $5
WHERE id = $1;

-- name: DeleteCFOrg :exec
DELETE FROM cf_org
WHERE id = $1;

-- name: CreateCFOrg :one
INSERT INTO cf_org (
  name, tier_id, credits_quota, credits_used, customer_id
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: BulkCreateCFOrgs :exec
-- BulkCreateCFOrgs creates CFOrg rows in bulk with the minimum required columns. If a row with the given primary key already exists, that input item is ignored.
INSERT INTO cf_org (id)
SELECT DISTINCT id
FROM UNNEST(sqlc.arg(ids)::uuid[]) AS id
ON CONFLICT DO NOTHING;
