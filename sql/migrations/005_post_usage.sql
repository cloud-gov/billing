ALTER TABLE entry
ADD COLUMN amount_microcredits bigint;

UPDATE entry
SET amount_microcredits = amount * 1e6;

ALTER TABLE entry
DROP COLUMN amount;

ALTER TABLE reading
ADD COLUMN created_at_utc timestamptz
GENERATED ALWAYS AS (created_at AT TIME ZONE 'UTC') STORED;
COMMENT ON COLUMN reading.created_at_utc IS 'CreatedAtUTC supplements CreatedAt, which does not have a timezone. Values must be inserted into CreatedAt in UTC by the client. CreatedAt has a unique index on it to enforce readings being taken at most hourly. Because the index uses functions that are not volatility level IMMUTABLE, it cannot be used on a column with a timezone; hence the supplementary generated column.';
COMMENT ON COLUMN reading.created_at IS 'CreatedAt must be a time in UTC.';
CREATE INDEX ON reading (created_at);

ALTER TABLE measurement
ADD COLUMN amount_microcredits bigint,
ADD COLUMN transaction_id bigint,
ADD CONSTRAINT fk_transaction FOREIGN KEY (transaction_id) REFERENCES transaction(id);

COMMENT ON COLUMN measurement.amount_microcredits IS 'AmountMicrocredits is a denormalized column that is calculated from the Price of the ResourceKind that was applicable when the measurement was taken (based on the time of the Reading). The value is persisted here for simpler rollups and auditing.';

COMMENT ON COLUMN measurement.transaction_id IS 'TransactionID is the transaction that accounts for this usage, typically a "post usage" transaction.';

CREATE TABLE price (
	id SERIAL PRIMARY KEY,

	meter text NOT NULL,
	kind_natural_id text NOT NULL,

	unit_of_measure text NOT NULL,
	microcredits_per_unit bigint NOT NULL,
    unit bigint NOT NULL,

	valid_during tstzrange NOT NULL,

	CONSTRAINT fk_resource_kind FOREIGN KEY (meter, kind_natural_id) REFERENCES resource_kind(meter, natural_id)
);

-- I did not automatically insert data from these columns into `price` because as of writing, there was no data in those columns.
ALTER TABLE resource_kind
DROP COLUMN credits,
DROP COLUMN amount,
DROP COLUMN unit_of_measure;

ALTER TABLE resource_kind
ADD COLUMN name text;

ALTER TABLE measurement
ADD COLUMN price_id bigint,
ADD CONSTRAINT fk_price FOREIGN KEY (price_id) REFERENCES price(id);

CREATE OR REPLACE FUNCTION bounds_month_prev(
	as_of timestamptz DEFAULT now(),
	tz text DEFAULT 'America/New_York'
)
RETURNS table (
	period_start timestamptz,
	period_end timestamptz
)
LANGUAGE sql IMMUTABLE
AS $$
	WITH bounds_base AS (
		SELECT date_trunc('month', as_of at time zone tz) AS this_month_local
	)
	SELECT (this_month_local - interval '1 month') AT time zone tz,
	this_month_local AT time zone tz
	FROM bounds_base;
$$;

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

-- TODO create indexes

---- create above / drop below ----
ALTER TABLE reading
DROP COLUMN created_at_utc;

ALTER TABLE entry
ADD COLUMN amount NUMERIC(20,4) NOT NULL;

UPDATE entry
SET amount = amount_microcredits / 1e6;

ALTER TABLE entry
DROP COLUMN IF EXISTS amount_microcredits;

ALTER TABLE measurement
DROP COLUMN IF EXISTS amount_microcredits,
DROP COLUMN IF EXISTS transaction_id,
DROP COLUMN IF EXISTS price_id;

DROP TABLE IF EXISTS price;

ALTER TABLE resource_kind
ADD COLUMN credits integer,
ADD COLUMN amount integer,
ADD COLUMN unit_of_measure text;

ALTER TABLE resource_kind
DROP COLUMN IF EXISTS name;

DROP FUNCTION IF EXISTS bounds_month_prev;
DROP FUNCTION IF EXISTS update_measurement_microcredits;
