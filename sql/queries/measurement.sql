-- name: CreateMeasurements :copyfrom
INSERT INTO measurement (
	reading_id,
	meter,
	resource_natural_id,
	value
) VALUES (
	$1, $2, $3, $4
);

-- name: BulkCreateMeasurement :exec
INSERT INTO measurement (
	reading_id,
	meter,
	resource_natural_id,
	value
) SELECT * FROM
UNNEST (
	sqlc.arg(reading_id)::int[],
	sqlc.arg(meter)::text[],
	sqlc.arg(resource_natural_id)::text[],
	sqlc.arg(value)::int[]
) AS m(reading_id, meter, resource_natural_id, value);
