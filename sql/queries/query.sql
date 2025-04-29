-- Start Customer
-- name: GetCustomer :one
SELECT * FROM customer
WHERE id = $1 LIMIT 1;

-- name: ListCustomers :many
SELECT * FROM customer
ORDER BY name;

-- name: CreateCustomer :one
INSERT INTO customer (
  id, name
) VALUES (
  $1, $2
)
RETURNING *;

-- name: UpdateCustomer :exec
UPDATE customer
  set name = $2
WHERE id = $1;

-- name: DeleteCustomer :exec
DELETE FROM customer
WHERE id = $1;

-- START CF_ORG
-- name: GetCF_Org :one
SELECT * FROM cf_org
WHERE id = $1 LIMIT 1;

-- name: ListCF_orgs :many
SELECT * FROM cf_org
ORDER BY name;

-- name: UpdateCF_org :exec
UPDATE cf_org
  set name = $2,
  tier_id = $3,
  credits_quota = $4,
  credits_used = $5
WHERE id = $1;

-- name: DeleteCF_org :exec
DELETE FROM cf_org
WHERE id = $1;

-- name: CreateCF_org :one
INSERT INTO cf_org (
  name, tier_id, credits_quota, credits_used, customer_id
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

