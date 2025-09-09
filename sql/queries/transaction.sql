
-- name: ListTransactions :many
SELECT * FROM transaction
ORDER BY id;

-- name: GetTransaction :one
SELECT * FROM transaction
WHERE id = $1 LIMIT 1;

-- name: CreateTransaction :one
INSERT INTO transaction (
  occurred_at, description, type
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: ListTransactionsWide :many
SELECT *
FROM
  entry
  LEFT JOIN account ON entry.account_id = account.id
  LEFT JOIN account_type ON account.type = account_type.id;
