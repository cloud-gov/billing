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
	updated bigint;
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
	)
	-- SELECT count(*) INTO updated FROM measurement_totals;
	SELECT mt.customer_id::bigint, mt.total_amount_microcredits::bigint FROM measurement_totals mt;
	-- Select measurements joined to readings, in bounds, join to org, join to customer
	-- Sum up usage by customer for all measurements in bounds
	-- Create a transaction of type usage_post for every customer
	-- [ ] Find the accounts for the customer: 201 credit_pool, 401 credits_used
	-- Create an entry for each account with the credits for that customer
	-- RETURN updated;
END $$;

---- create above / drop below ----

DROP FUNCTION IF EXISTS post_usage;
