-- name: SumEntries :many
-- SumEntries calculates the sum of all entries in the ledger. If the result is not 0, a transaction is imbalanced.
SELECT
  sum(direction * amount_microcredits / 1e6)
FROM
  entry;

-- name: GetEntry :one
SELECT *
FROM entry
WHERE transaction_id = $1 AND account_id = $2;

-- name: GetEntriesForCustomerAndType :many
SELECT e.*
FROM entry e
JOIN account a ON e.account_id = a.id
JOIN customer c ON a.customer_id = c.id
WHERE c.name = $1
AND a.type = $2;
