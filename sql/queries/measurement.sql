-- name: CreateMeasurements :copyfrom
INSERT INTO measurement (
	reading_id,
	resource_id,
	value
) VALUES (
	$1, $2, $3
);
