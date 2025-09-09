ALTER TABLE transaction
ADD COLUMN customer_id bigint,
ALTER COLUMN occurred_at TYPE timestamptz
USING occurred_at AT TIME ZONE 'UTC';

CREATE OR REPLACE FUNCTION post_usage (
	as_of timestamptz DEFAULT now()
)
RETURNS table(
	customer_id bigint,
	total_amount_microcredits bigint
)
LANGUAGE plpgsql
AS $$
DECLARE
	ps timestamptz;
	pe timestamptz;
BEGIN
	SELECT period_start, period_end INTO ps, pe FROM bounds_month_prev(as_of);

	-- Get total microcredits per customer
	RETURN QUERY
	WITH measurement_totals AS (
		SELECT c.id AS customer_id, sum(m.amount_microcredits) AS total_amount_microcredits
		FROM reading AS rd
		JOIN measurement AS m
		ON rd.id = m.reading_id
		JOIN resource AS r
		ON m.meter = r.meter AND m.resource_natural_id = r.natural_id
		JOIN cf_org AS o
		ON r.cf_org_id = o.id
		JOIN customer AS c
		ON o.customer_id = c.id
		WHERE ps <= rd.created_at_utc
		AND rd.created_at_utc < pe
		AND m.amount_microcredits IS NOT NULL
		GROUP BY c.id
	),
	ins_tx AS (
		INSERT INTO transaction AS txn(customer_id, occurred_at, description, type)
		SELECT
			mt.customer_id,
			pe,
			format('Monthly usage %s-%s', to_char(ps, 'YYYY-MM-DD'), to_char(pe, 'YYYY-MM-DD')),
			'usage_post'
		FROM measurement_totals AS mt
		WHERE mt.total_amount_microcredits <> 0
		RETURNING txn.id, txn.customer_id
	),
	ins_entries AS (
		INSERT INTO entry AS e(transaction_id, account_id, direction, amount_microcredits)
		SELECT
			it.id,
			ac.account_id,
			ac.normal,
			mt.total_amount_microcredits
		FROM ins_tx AS it
		JOIN measurement_totals AS mt
		ON it.customer_id = mt.customer_id
		JOIN LATERAL (
			SELECT a.id, at.normal
			FROM account AS a
			JOIN account_type AS at
			ON a.id = at.id
			WHERE a.customer_id = it.customer_id
			AND (at.name = 'credit_pool' OR at.name = 'credits_used')
			LIMIT 2
		) AS ac(account_id, normal) ON TRUE
		RETURNING e.transaction_id
	)
	SELECT mt.customer_id::bigint, mt.total_amount_microcredits::bigint FROM measurement_totals mt;
END $$;

COMMENT ON FUNCTION post_usage IS 'posts_usage returns ';

---- create above / drop below ----

DROP FUNCTION IF EXISTS post_usage;
