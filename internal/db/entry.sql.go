// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: entry.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const sumEntries = `-- name: SumEntries :many
SELECT
  sum(direction * amount)
FROM
  entry
`

// SumEntries calculates the sum of all entries in the ledger. If the result is not 0, a transaction is imbalanced.
func (q *Queries) SumEntries(ctx context.Context) ([]pgtype.Numeric, error) {
	rows, err := q.db.Query(ctx, sumEntries)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []pgtype.Numeric
	for rows.Next() {
		var sum pgtype.Numeric
		if err := rows.Scan(&sum); err != nil {
			return nil, err
		}
		items = append(items, sum)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
