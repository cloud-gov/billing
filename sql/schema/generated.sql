
\restrict jBw11d0AQlC9AAnuPPStB3gB4CvudGHswkl7oHfurFwYYYlC7NH29Ja232oDMmJ

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

CREATE EXTENSION IF NOT EXISTS ltree WITH SCHEMA public;

COMMENT ON EXTENSION ltree IS 'data type for hierarchical tree-like structures';

CREATE TYPE public.river_job_state AS ENUM (
    'available',
    'cancelled',
    'completed',
    'discarded',
    'pending',
    'retryable',
    'running',
    'scheduled'
);

CREATE TYPE public.transaction_type AS ENUM (
    'iaa_pop_start',
    'iaa_pop_end',
    'usage_post'
);

COMMENT ON TYPE public.transaction_type IS 'TransactionType explains why the transaction was made. Each means:
  - iaa_pop_start: The IAA Period of Performance started.
  - iaa_pop_end: The IAA Period of Performance ended.
  - usage_post: Customer usage of was posted, i.e. their account balance was updated to reflect their usage.
';

CREATE FUNCTION public.assert_transaction_balances() RETURNS trigger
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
				MESSAGE = format('ledger error: transaction %s is not balanced; sum(amount_microcredits)=%s', v_tx, v_sum);
	END IF;
	RETURN NULL;
END;
$$;

CREATE FUNCTION public.bounds_month_prev(as_of timestamp with time zone DEFAULT now(), tz text DEFAULT 'America/New_York'::text) RETURNS TABLE(period_start timestamp with time zone, period_end timestamp with time zone)
    LANGUAGE sql IMMUTABLE
    AS $$
	WITH bounds_base AS (
		SELECT date_trunc('month', as_of at time zone tz) AS this_month_local
	)
	SELECT (this_month_local - interval '1 month') AT time zone tz,
	this_month_local AT time zone tz
	FROM bounds_base;
$$;

CREATE FUNCTION public.move_customer(p_customer_id uuid, p_new_parent_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
declare
  v_old_path ltree;
  v_new_parent_path ltree;
  v_new_path ltree;
begin
  select path into v_old_path 
  from customer where id = p_customer_id;

  select path into v_new_parent_path 
  from customer where id = p_new_parent_id;

  if v_new_parent_path <@ v_old_path then
    raise exception 'cannot move customer to its own descendant';
  end if;

  v_new_path := v_new_parent_path || subpath(v_old_path, -1, 1);

  update customer
  set 
    path = v_new_path || subpath(path, nlevel(v_old_path))
  where path <@ v_old_path;

  return true;
end;
$$;

CREATE FUNCTION public.move_resource_node(p_resource_node_id uuid, p_new_parent_id uuid) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
declare
  v_old_path ltree;
  v_new_parent_path ltree;
  v_new_path ltree;
begin
  select path into v_old_path 
  from resource_node where id = p_resource_node_id;

  select path into v_new_parent_path 
  from resource_node where id = p_new_parent_id;

  if v_new_parent_path <@ v_old_path then
    raise exception 'cannot move resource_node to its own descendant';
  end if;

  v_new_path := v_new_parent_path || subpath(v_old_path, -1, 1);

  update resource_node
  set 
    path = v_new_path || subpath(path, nlevel(v_old_path))
  where path <@ v_old_path;

  return true;
end;
$$;

CREATE FUNCTION public.post_usage(as_of timestamp with time zone DEFAULT now()) RETURNS TABLE(transaction_id integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
	ps timestamptz;
	pe timestamptz;
BEGIN
	SELECT period_start, period_end INTO ps, pe FROM bounds_month_prev(as_of);

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
			ON a.type = at.id
			WHERE a.customer_id = it.customer_id
			AND (at.name = 'credit_pool' OR at.name = 'credits_used')
			LIMIT 2
		) AS ac(account_id, normal) ON TRUE
		RETURNING e.transaction_id, e.account_id, e.direction, e.amount_microcredits
	)
	SELECT DISTINCT e.transaction_id
	FROM ins_entries AS e;
END $$;

COMMENT ON FUNCTION public.post_usage(as_of timestamp with time zone) IS 'post_usage returns all entries, plus their associated customer_id. This function must be run in a transaction.';

CREATE FUNCTION public.river_job_state_in_bitmask(bitmask bit, state public.river_job_state) RETURNS boolean
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT CASE state
        WHEN 'available' THEN get_bit(bitmask, 7)
        WHEN 'cancelled' THEN get_bit(bitmask, 6)
        WHEN 'completed' THEN get_bit(bitmask, 5)
        WHEN 'discarded' THEN get_bit(bitmask, 4)
        WHEN 'pending'   THEN get_bit(bitmask, 3)
        WHEN 'retryable' THEN get_bit(bitmask, 2)
        WHEN 'running'   THEN get_bit(bitmask, 1)
        WHEN 'scheduled' THEN get_bit(bitmask, 0)
        ELSE 0
    END = 1;
$$;

CREATE FUNCTION public.update_measurement_microcredits(as_of timestamp with time zone DEFAULT now()) RETURNS bigint
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

CREATE FUNCTION public.uuid_generate_v7() RETURNS uuid
    LANGUAGE plpgsql
    AS $$
begin
  return encode(
    set_bit(
      set_bit(
        overlay(uuid_send(gen_random_uuid())
                placing substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3)
                from 1 for 6
        ),
        52, 1
      ),
      53, 1
    ),
    'hex')::uuid;
end
$$;

SET default_tablespace = '';

SET default_table_access_method = heap;

CREATE TABLE public.account (
    id integer NOT NULL,
    type integer NOT NULL,
    customer_id uuid
);

CREATE SEQUENCE public.account_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.account_id_seq OWNED BY public.account.id;

CREATE TABLE public.account_type (
    id integer NOT NULL,
    name text NOT NULL,
    normal integer
);

CREATE TABLE public.cf_org (
    id uuid NOT NULL,
    name text,
    customer_id uuid
);

CREATE TABLE public.customer (
    old_id bigint NOT NULL,
    name text NOT NULL,
    tier_id integer,
    id uuid DEFAULT public.uuid_generate_v7() NOT NULL,
    path public.ltree,
    slug character varying(50),
    CONSTRAINT valid_path CHECK (((path)::text ~ '^[A-Za-z0-9_]+(\\.[A-Za-z0-9_]+)*$'::text))
);

CREATE SEQUENCE public.customer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.customer_id_seq OWNED BY public.customer.old_id;

CREATE TABLE public.entry (
    transaction_id integer NOT NULL,
    account_id integer NOT NULL,
    direction integer NOT NULL,
    amount_microcredits bigint
);

CREATE TABLE public.measurement (
    reading_id integer NOT NULL,
    meter text NOT NULL,
    resource_natural_id text NOT NULL,
    value integer NOT NULL,
    amount_microcredits bigint,
    transaction_id bigint,
    price_id bigint
);

COMMENT ON COLUMN public.measurement.amount_microcredits IS 'AmountMicrocredits is a denormalized column that is calculated from the Price of the ResourceKind that was applicable when the measurement was taken (based on the time of the Reading). The value is persisted here for simpler rollups and auditing.';

COMMENT ON COLUMN public.measurement.transaction_id IS 'TransactionID is the transaction that accounts for this usage, typically a "post usage" transaction.';

CREATE TABLE public.meter (
    name text NOT NULL,
    CONSTRAINT meter_name_check CHECK ((char_length(TRIM(BOTH FROM name)) > 0))
);

COMMENT ON TABLE public.meter IS 'A Meter reads usage information from a system in Cloud.gov. It also namespaces natural IDs for resources and resource_kinds; meter + natural_id is a primary key.';

CREATE TABLE public.price (
    id integer NOT NULL,
    meter text NOT NULL,
    kind_natural_id text NOT NULL,
    unit_of_measure text NOT NULL,
    microcredits_per_unit bigint NOT NULL,
    unit bigint NOT NULL,
    valid_during tstzrange NOT NULL
);

CREATE SEQUENCE public.price_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.price_id_seq OWNED BY public.price.id;

CREATE TABLE public.reading (
    id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    periodic boolean NOT NULL,
    created_at_utc timestamp with time zone GENERATED ALWAYS AS ((created_at AT TIME ZONE 'UTC'::text)) STORED
);

COMMENT ON COLUMN public.reading.created_at IS 'CreatedAt must be a time in UTC.';

COMMENT ON COLUMN public.reading.periodic IS 'Periodic is true if a reading was taken automatically as part of the periodic usage measurement schedule, or false if it was requested manually.';

COMMENT ON COLUMN public.reading.created_at_utc IS 'CreatedAtUTC supplements CreatedAt, which does not have a timezone. Values must be inserted into CreatedAt in UTC by the client. CreatedAt has a unique index on it to enforce readings being taken at most hourly. Because the index uses functions that are not volatility level IMMUTABLE, it cannot be used on a column with a timezone; hence the supplementary generated column.';

CREATE SEQUENCE public.reading_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.reading_id_seq OWNED BY public.reading.id;

CREATE TABLE public.resource (
    meter text NOT NULL,
    natural_id text NOT NULL,
    kind_natural_id text NOT NULL,
    cf_org_id uuid NOT NULL
);

CREATE TABLE public.resource_kind (
    meter text NOT NULL,
    natural_id text NOT NULL,
    name text
);

COMMENT ON TABLE public.resource_kind IS 'ResourceKind represents a particular kind of billable resource. Note that natural_id can be empty because some meters may only read one kind of resource, and that resource kind may not have a unique identifier in the target system; it is uniquely identified by the meter name only.';

CREATE TABLE public.resource_node (
    path public.ltree,
    slug character varying(50) NOT NULL,
    customer_id uuid NOT NULL,
    CONSTRAINT valid_path CHECK (((path)::text ~ '^[A-Za-z0-9_]+(\\.[A-Za-z0-9_]+)*$'::text))
);

CREATE TABLE public.tier (
    id integer NOT NULL,
    name text NOT NULL,
    tier_credits bigint NOT NULL
);

CREATE SEQUENCE public.tier_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.tier_id_seq OWNED BY public.tier.id;

CREATE TABLE public.transaction (
    id integer NOT NULL,
    occurred_at timestamp with time zone,
    description text,
    type public.transaction_type NOT NULL,
    customer_id uuid
);

COMMENT ON COLUMN public.transaction.customer_id IS 'CustomerID is somewhat redundant because the Entry rows associated with a Transaction are associated with Accounts, which are associated with a Customer. However, we have to create a Transaction before we create an Entry (see post_usage, ins_tx as an example). To join Measurements, Transactions, Entries, and Accounts, Transaction needs a CustomerID.';

CREATE SEQUENCE public.transaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.transaction_id_seq OWNED BY public.transaction.id;

ALTER TABLE ONLY public.account ALTER COLUMN id SET DEFAULT nextval('public.account_id_seq'::regclass);

ALTER TABLE ONLY public.customer ALTER COLUMN old_id SET DEFAULT nextval('public.customer_id_seq'::regclass);

ALTER TABLE ONLY public.price ALTER COLUMN id SET DEFAULT nextval('public.price_id_seq'::regclass);

ALTER TABLE ONLY public.reading ALTER COLUMN id SET DEFAULT nextval('public.reading_id_seq'::regclass);

ALTER TABLE ONLY public.tier ALTER COLUMN id SET DEFAULT nextval('public.tier_id_seq'::regclass);

ALTER TABLE ONLY public.transaction ALTER COLUMN id SET DEFAULT nextval('public.transaction_id_seq'::regclass);

ALTER TABLE ONLY public.account
    ADD CONSTRAINT account_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.account_type
    ADD CONSTRAINT account_type_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.cf_org
    ADD CONSTRAINT cf_org_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_new_id_key UNIQUE (id);

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_old_id_key UNIQUE (old_id);

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_old_id_key1 UNIQUE (old_id);

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.entry
    ADD CONSTRAINT entry_pkey PRIMARY KEY (transaction_id, account_id);

ALTER TABLE ONLY public.meter
    ADD CONSTRAINT meter_name_key UNIQUE (name);

ALTER TABLE ONLY public.price
    ADD CONSTRAINT price_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.reading
    ADD CONSTRAINT reading_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.resource_kind
    ADD CONSTRAINT resource_kind_pkey PRIMARY KEY (meter, natural_id);

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_meter_natural_id_uq UNIQUE (meter, natural_id);

ALTER TABLE ONLY public.resource_node
    ADD CONSTRAINT resource_node_pkey PRIMARY KEY (customer_id, slug);

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_pkey PRIMARY KEY (meter, natural_id);

ALTER TABLE ONLY public.tier
    ADD CONSTRAINT tier_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT transaction_pkey PRIMARY KEY (id);

CREATE INDEX account_customer_type_idx ON public.account USING btree (customer_id, type);

CREATE UNIQUE INDEX account_type_name_uidx ON public.account_type USING btree (name);

CREATE UNIQUE INDEX account_unique ON public.account USING btree (id, type);

CREATE INDEX cf_org_customer_id_idx ON public.cf_org USING btree (customer_id);

CREATE INDEX customer_path_btree_idx ON public.customer USING btree (path);

CREATE INDEX customer_path_gist_idx ON public.customer USING gist (path);

CREATE INDEX customer_path_idx ON public.resource_node USING btree (customer_id, path);

CREATE INDEX idx_reading_created_at ON public.reading USING btree (created_at);

CREATE UNIQUE INDEX idx_resource_kind_meter_natural_id ON public.resource_kind USING btree (meter, natural_id);

COMMENT ON INDEX public.idx_resource_kind_meter_natural_id IS 'Enables efficient deduplicated inserts using BulkCreateResourceKinds function.';

CREATE UNIQUE INDEX idx_resource_meter_natural_id ON public.resource USING btree (meter, natural_id);

COMMENT ON INDEX public.idx_resource_meter_natural_id IS 'Enables efficient deduplicated inserts using BulkCreateResources function.';

CREATE INDEX resource_path_btree_idx ON public.resource_node USING btree (path);

CREATE INDEX resource_path_gist_idx ON public.resource_node USING gist (path);

ALTER TABLE ONLY public.account
    ADD CONSTRAINT fk_customer_id FOREIGN KEY (customer_id) REFERENCES public.customer(id);

ALTER TABLE ONLY public.cf_org
    ADD CONSTRAINT fk_customer_id FOREIGN KEY (customer_id) REFERENCES public.customer(id);

ALTER TABLE ONLY public.resource_node
    ADD CONSTRAINT resource_node_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customer(id);

\unrestrict jBw11d0AQlC9AAnuPPStB3gB4CvudGHswkl7oHfurFwYYYlC7NH29Ja232oDMmJ

