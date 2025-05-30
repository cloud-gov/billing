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

-- START Tier
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

-- START billable resource
-- name: GetBillableResource :one
SELECT * FROM billable_resource
WHERE id = $1 LIMIT 1;

-- name: ListBillableResources :many
SELECT * FROM billable_resource
ORDER BY native_id;

-- name: UpdateBillableResource :exec
UPDATE billable_resource
  set native_id = $2,
  class_id = $3,
  cf_org_id = $4
  WHERE id = $1;

-- name: DeleteBillableResouce :exec
DELETE FROM billable_resource
WHERE id = $1;

-- name: CreateBillableResource :one
INSERT INTO billable_resource (
  native_id, class_id, cf_org_id
) VALUES (
  $1, $2, $3 
)
RETURNING *;

-- START Billable Class
-- name: GetBillableClass :one
SELECT * FROM billable_class
WHERE id = $1 LIMIT 1;

-- name: ListBillableClass :many
SELECT * FROM billable_class
ORDER BY native_id;

-- name: UpdateBillableClass :exec
UPDATE billable_class
  set native_id = $2,
  credits = $3,
  amount = $4,
  unit_of_measure = $5
  WHERE id = $1;

-- name: DeleteBillableClass :exec
DELETE FROM billable_class
WHERE id = $1;

-- name: CreateBillableClass :one
INSERT INTO billable_class (
  native_id, credits, amount, unit_of_measure
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

