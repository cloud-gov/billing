-- name: CreateReading :one
INSERT INTO reading (
	created_at
) VALUES (
	$1
)
RETURNING *;

-- name: ReadingExistsInHour :one
-- ReadingExistsInHour returns true if at least one Reading was created during the current hour. For instance, if the query is run at 2:55 and a reading was taken at 2:05, the query returns true. If the query is run at 2:55 and a reading was taken at 1:59, the query returns false.
SELECT EXISTS (
    SELECT 1
    FROM   reading
    WHERE  created_at >= date_trunc('hour', now())
           AND created_at <  date_trunc('hour', now()) + INTERVAL '1 hour'
) AS reading_in_current_hour;
