ALTER TABLE transaction
ADD COLUMN customer_id bigint,
ALTER COLUMN occurred_at TYPE timestamptz
USING occurred_at AT TIME ZONE 'UTC';

COMMENT ON COLUMN transaction.customer_id IS 'CustomerID is somewhat redundant because the Entry rows associated with a Transaction are associated with Accounts, which are associated with a Customer. However, we have to create a Transaction before we create an Entry (see post_usage, ins_tx as an example). To join Measurements, Transactions, Entries, and Accounts, Transaction needs a CustomerID.';

CREATE OR REPLACE FUNCTION post_usage (
	as_of timestamptz DEFAULT now()
)
RETURNS table(
	transaction_id integer
)
LANGUAGE plpgsql
AS $$
DECLARE
	ps timestamptz;
	pe timestamptz;
BEGIN
	SELECT period_start, period_end INTO ps, pe FROM bounds_month_prev(as_of);

	RETURN QUERY
	-- Step 1: Calculate total credits for measurements in period
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
	-- Step 2: Create a transaction row for each customer with nonzero usage
	ins_tx AS (
		INSERT INTO transaction AS txn(customer_id, occurred_at, description, type)
		SELECT
			mt.customer_id,
			pe,
			format('Monthly usage %s--%s', to_char(ps, 'YYYY-MM-DD'), to_char(pe, 'YYYY-MM-DD')),
			'usage_post'
		FROM measurement_totals AS mt
		WHERE mt.total_amount_microcredits <> 0
		RETURNING txn.id, txn.customer_id
	),
	-- Step 3: Insert two entries for each transaction with credits calculated earlier
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
			ON a.type = at.id
			WHERE a.customer_id = it.customer_id
			AND (at.name = 'credit_pool' OR at.name = 'credits_used')
			LIMIT 2
		) AS ac(account_id, normal) ON TRUE
		RETURNING e.transaction_id, e.account_id, e.direction, e.amount_microcredits
	)
	-- Step 4: Return transaction IDs created by the function
	SELECT DISTINCT e.transaction_id
	FROM ins_entries AS e;
END $$;

COMMENT ON FUNCTION post_usage IS 'post_usage returns all entries, plus their associated customer_id. This function must be run in a transaction.';

CREATE OR REPLACE FUNCTION assert_transaction_balances()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
	v_tx bigint := COALESCE(NEW.transaction_id, OLD.transaction_id);
	v_sum bigint;
BEGIN
	SELECT (COALESCE(SUM(e.amount_microcredits * e.direction), 0))
	INTO v_sum
	FROM entry e
	WHERE transaction_id = v_tx;

	IF v_sum <> 0 THEN
		RAISE EXCEPTION
			USING
				ERRCODE = '23514', -- check violation
				MESSAGE = format('ledger error: transaction %s is not balanced; sum(amount_microcredits)=%s', v_tx, v_sum);
	END IF;
	RETURN NULL;
END;
$$;

ALTER TABLE entry
ALTER COLUMN direction SET NOT NULL;

ALTER TABLE account
ALTER COLUMN customer_id SET NOT NULL,
ALTER COLUMN type SET NOT NULL;

-- Indexes so post_usage is fast.
-- Filter by month and join to measurements
CREATE INDEX IF NOT EXISTS reading_created_at_utc_idx
	ON reading (created_at_utc);

-- Join m -> rd
CREATE INDEX IF NOT EXISTS measurement_reading_id_idx
	ON measurement (reading_id);

-- Join m -> r
CREATE INDEX IF NOT EXISTS measurement_meter_resource_natural_idx
	ON measurement (meter, resource_natural_id);

-- Join r -> o
CREATE INDEX IF NOT EXISTS resource_cf_org_id_idx
	ON resource (cf_org_id);

-- Join o -> c
CREATE INDEX IF NOT EXISTS cf_org_customer_id_idx
	ON cf_org (customer_id);

-- LATERAL lookup of the two accounts for a customer
CREATE INDEX IF NOT EXISTS account_customer_type_idx
	ON account (customer_id, type);

-- Resolve "credit_pool"/"credits_used" quickly
CREATE UNIQUE INDEX IF NOT EXISTS account_type_name_uidx
	ON account_type (name);

---- create above / drop below ----

DROP FUNCTION IF EXISTS post_usage;

-- Restore previous version of assert_transaction_balances from migration 004
CREATE OR REPLACE FUNCTION assert_transaction_balances()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
	PERFORM 1
	FROM entry e
	GROUP BY e.transaction_id
	HAVING SUM(e.amount) <> 0;

	IF FOUND THEN
		RAISE EXCEPTION
		  'ledger error: at least one transaction is not balanced (sum(amount) <> 0)';
	END IF;

	RETURN NULL;
END;
$$;

ALTER TABLE entry
ALTER COLUMN direction DROP NOT NULL;

ALTER TABLE account
ALTER COLUMN customer_id DROP NOT NULL,
ALTER COLUMN type DROP NOT NULL;

DROP INDEX IF EXISTS reading_created_at_utc_idx;
DROP INDEX IF EXISTS measurement_reading_id_idx;
DROP INDEX IF EXISTS measurement_meter_resource_natural_idx;
DROP INDEX IF EXISTS resource_cf_org_id_idx;
DROP INDEX IF EXISTS cf_org_customer_id_idx;
DROP INDEX IF EXISTS account_customer_type_idx;
DROP INDEX IF EXISTS account_type_name_uidx;
