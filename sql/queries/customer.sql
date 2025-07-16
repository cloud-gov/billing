-- name: GetCustomer :one
SELECT * FROM customer
WHERE id = $1 LIMIT 1;

-- name: ListCustomers :many
SELECT * FROM customer
ORDER BY name;

-- name: UpdateCustomer :exec
UPDATE customer
  set name = $2
WHERE id = $1;

-- name: DeleteCustomer :exec
DELETE FROM customer
WHERE id = $1;

-- name: CreateCustomer :one
-- CreateCustomer adds a customer to the database and creates Accounts for the customer for every AccountType available. Returns the ID of the new Customer.
WITH cust AS (
  INSERT INTO customer (
    name
  )
  VALUES ($1)
  RETURNING id
),
types AS (
  SELECT id
  FROM account_type
),
accts AS (
  INSERT INTO account (customer_id, type)
  SELECT cust.id, types.id
  FROM cust CROSS JOIN types
)
SELECT id
FROM cust;
