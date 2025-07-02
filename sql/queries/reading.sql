-- name: CreateReading :one
INSERT INTO reading (
	created_at
) VALUES (
	$1
)
RETURNING *;
