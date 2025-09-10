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
