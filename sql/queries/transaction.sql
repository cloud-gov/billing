
-- name: ListTransactions :many
SELECT * FROM transactions
ORDER BY transaction_id;

-- name: GetTransactions :one
SELECT * FROM transactions
WHERE WHERE id = $1 LIMIT 1;

-- name: CreateTransaction :one
INSERT INTO transactions (
  transaction_date, resource_id, cf_ord_id, description, direction, amount, transaction_type_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetTransactionType :one
SELECT * FROM transaction_type
WHERE id = $1 LIMIT 1;

-- name: ListTransactionType :many
SELECT * FROM transaction_type
ORDER BY name;

-- name: UpdateTransactionType :exec
UPDATE transaction_type
  set name = $2,
  WHERE id = $1;

-- name: DeleteTransactionType :exec
DELETE FROM transaction_type
WHERE id = $1;

-- name: CreateTransactionType :one
INSERT INTO transaction_type (
  name
) VALUES (
  $1
)
RETURNING *;