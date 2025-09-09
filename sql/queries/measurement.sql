-- name: CreateMeasurement :one
INSERT INTO measurement (
	reading_id,
	meter,
	resource_natural_id,
	value,
	amount_microcredits
) VALUES (
	$1, $2, $3, $4, $5
) RETURNING *;

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

-- name: UpdateMeasurementMicrocredits :one
-- UpdateMeasurementMicrocredits updates the amount of microcredits associated with measurements made in the month preceding as_of based on the prices that were valid for each resource_kind at the time of reading.
SELECT *
FROM update_measurement_microcredits($1);

-- name: BoundsMonthPrev :one
-- BoundsMonthPrev calculates bounds that encapsulate the month previous to the parameter, as_of. The first bound is inclusive and the second is exclusive.
SELECT period_start, period_end
FROM bounds_month_prev($1);

-- name: ListMeasurements :many
SELECT *
FROM measurement;

-- name: PostUsage :many
SELECT customer_id, transaction_id, account_id, direction, total_amount_microcredits
FROM post_usage($1);
