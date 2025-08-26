-- name: CreateReading :one
INSERT INTO reading (
	created_at
) VALUES (
	$1
)
RETURNING *;

-- name: CreateUniqueReading :one
-- CreateUniqueReading creates a Reading if one does not exist for the hour specified in created_at. It returns [pgx.ErrNoRows] if a Reading already exists.
INSERT INTO reading (
    created_at, periodic
) VALUES (
    $1, $2
)
ON CONFLICT (date_trunc('hour', created_at))
DO NOTHING
RETURNING *;

-- name: CreateReadingWithID :one
INSERT INTO reading (
	id, created_at, periodic
) VALUES (
	$1, $2, $3
)
RETURNING *;
