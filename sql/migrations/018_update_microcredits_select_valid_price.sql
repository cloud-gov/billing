CREATE OR REPLACE FUNCTION update_measurement_microcredits(
	as_of timestamptz DEFAULT now()
)
RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
	ps timestamptz;
	pe timestamptz;
	updated bigint;
BEGIN
	SELECT period_start, period_end INTO ps, pe FROM bounds_month_prev(as_of);

	WITH measurement_amounts AS (
		SELECT
			r.meter AS meter,
			r.natural_id AS resource_natural_id,
			rd.id AS reading_id,
			sum(p.microcredits_per_unit * m.value / p.unit) AS amount_microcredits,
			p.id AS price_id
		FROM reading rd
		JOIN measurement AS m
		ON rd.id = m.reading_id
		JOIN resource AS r
		ON m.meter = r.meter AND m.resource_natural_id = r.natural_id
		JOIN price AS p
		ON r.meter = p.meter AND r.kind_natural_id = p.kind_natural_id AND p.valid_during @> rd.created_at_utc
		WHERE ps <= rd.created_at_utc
		AND rd.created_at_utc < pe
		AND m.amount_microcredits IS NULL
		GROUP BY
			r.meter,
			r.natural_id,
			rd.id,
			p.id
	),
	update_measurements AS (
		UPDATE measurement AS m
		SET
			amount_microcredits = ma.amount_microcredits,
			price_id = ma.price_id
		FROM measurement_amounts AS ma
		WHERE
			m.meter = ma.meter AND
			m.resource_natural_id = ma.resource_natural_id AND
			m.reading_id = ma.reading_id
		RETURNING 1
	)
	SELECT count(*) INTO updated FROM update_measurements;
	RETURN updated;
END $$;

---- create above / drop below ----

CREATE OR REPLACE FUNCTION update_measurement_microcredits(
	as_of timestamptz DEFAULT now()
)
RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
	ps timestamptz;
	pe timestamptz;
	updated bigint;
BEGIN
	SELECT period_start, period_end INTO ps, pe FROM bounds_month_prev(as_of);

	WITH measurement_amounts AS (
		SELECT
			r.meter AS meter,
			r.natural_id AS resource_natural_id,
			rd.id AS reading_id,
			sum(p.microcredits_per_unit * m.value / p.unit) AS amount_microcredits,
			p.id AS price_id
		FROM reading rd
		JOIN measurement AS m
		ON rd.id = m.reading_id
		JOIN resource AS r
		ON m.meter = r.meter AND m.resource_natural_id = r.natural_id
		JOIN price AS p
		ON r.meter = p.meter AND r.kind_natural_id = p.kind_natural_id
		WHERE ps <= rd.created_at_utc
		AND rd.created_at_utc < pe
		AND m.amount_microcredits IS NULL
		GROUP BY
			r.meter,
			r.natural_id,
			rd.id,
			p.id
	),
	update_measurements AS (
		UPDATE measurement AS m
		SET
			amount_microcredits = ma.amount_microcredits,
			price_id = ma.price_id
		FROM measurement_amounts AS ma
		WHERE
			m.meter = ma.meter AND
			m.resource_natural_id = ma.resource_natural_id AND
			m.reading_id = ma.reading_id
		RETURNING 1
	)
	SELECT count(*) INTO updated FROM update_measurements;
	RETURN updated;
END $$;
