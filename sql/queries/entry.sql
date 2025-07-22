-- name: SumEntries :many
-- SumEntries calculates the sum of all entries in the ledger. If the result is not 0, a transaction is imbalanced.
SELECT
  sum(direction * amount)
FROM
  entry;
