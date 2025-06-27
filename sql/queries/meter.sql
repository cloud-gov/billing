-- name: BulkCreateMeters :exec
-- BulkCreateMeters creates Meter rows in bulk with the minimum required columns. If a row with the given primary key already exists, that input item is ignored.
INSERT INTO meter (name)
SELECT DISTINCT name
FROM UNNEST(sqlc.arg(names)::text[]) AS name
ON CONFLICT DO NOTHING;
