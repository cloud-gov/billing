
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

SET default_tablespace = '';

SET default_table_access_method = heap;

CREATE TABLE public.account (
    id integer NOT NULL,
    customer_id bigint,
    type integer
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
    customer_id bigint
);

CREATE TABLE public.customer (
    id bigint NOT NULL,
    name text NOT NULL,
    tier_id integer
);

CREATE SEQUENCE public.customer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.customer_id_seq OWNED BY public.customer.id;

CREATE TABLE public.entry (
    transaction_id integer NOT NULL,
    account_id integer NOT NULL,
    amount numeric(20,4) NOT NULL,
    direction integer
);

CREATE TABLE public.measurement (
    reading_id integer NOT NULL,
    meter text NOT NULL,
    resource_natural_id text NOT NULL,
    value integer NOT NULL
);

CREATE TABLE public.meter (
    name text NOT NULL,
    CONSTRAINT meter_name_check CHECK ((char_length(TRIM(BOTH FROM name)) > 0))
);

COMMENT ON TABLE public.meter IS 'A Meter reads usage information from a system in Cloud.gov. It also namespaces natural IDs for resources and resource_kinds; meter + natural_id is a primary key.';

CREATE TABLE public.reading (
    id integer NOT NULL,
    created_at timestamp without time zone NOT NULL,
    periodic boolean NOT NULL
);

COMMENT ON COLUMN public.reading.periodic IS 'Periodic is true if a reading was taken automatically as part of the periodic usage measurement schedule, or false if it was requested manually.';

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
    credits integer,
    amount integer,
    unit_of_measure text
);

COMMENT ON TABLE public.resource_kind IS 'ResourceKind represents a particular kind of billable resource. Note that natural_id can be empty because some meters may only read one kind of resource, and that resource kind may not have a unique identifier in the target system; it is uniquely identified by the meter name only.';

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
    occurred_at timestamp without time zone,
    description text,
    type public.transaction_type NOT NULL
);

CREATE SEQUENCE public.transaction_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.transaction_id_seq OWNED BY public.transaction.id;

ALTER TABLE ONLY public.account ALTER COLUMN id SET DEFAULT nextval('public.account_id_seq'::regclass);

ALTER TABLE ONLY public.customer ALTER COLUMN id SET DEFAULT nextval('public.customer_id_seq'::regclass);

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
    ADD CONSTRAINT customer_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.entry
    ADD CONSTRAINT entry_pkey PRIMARY KEY (transaction_id, account_id);

ALTER TABLE ONLY public.meter
    ADD CONSTRAINT meter_name_key UNIQUE (name);

ALTER TABLE ONLY public.reading
    ADD CONSTRAINT reading_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.resource_kind
    ADD CONSTRAINT resource_kind_pkey PRIMARY KEY (meter, natural_id);

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_meter_natural_id_uq UNIQUE (meter, natural_id);

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT resource_pkey PRIMARY KEY (meter, natural_id);

ALTER TABLE ONLY public.tier
    ADD CONSTRAINT tier_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.transaction
    ADD CONSTRAINT transaction_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX account_unique ON public.account USING btree (id, type);

CREATE INDEX idx_reading_created_at ON public.reading USING btree (created_at);

CREATE UNIQUE INDEX idx_resource_kind_meter_natural_id ON public.resource_kind USING btree (meter, natural_id);

COMMENT ON INDEX public.idx_resource_kind_meter_natural_id IS 'Enables efficient deduplicated inserts using BulkCreateResourceKinds function.';

CREATE UNIQUE INDEX idx_resource_meter_natural_id ON public.resource USING btree (meter, natural_id);

COMMENT ON INDEX public.idx_resource_meter_natural_id IS 'Enables efficient deduplicated inserts using BulkCreateResources function.';

CREATE UNIQUE INDEX reading_hourly_uq ON public.reading USING btree (date_trunc('hour'::text, created_at));

COMMENT ON INDEX public.reading_hourly_uq IS 'Make readings unique per hour.';

CREATE CONSTRAINT TRIGGER transaction_balances_chk AFTER INSERT OR DELETE OR UPDATE ON public.entry DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION public.assert_transaction_balances();

ALTER TABLE ONLY public.entry
    ADD CONSTRAINT entry_transaction_id_fkey FOREIGN KEY (transaction_id) REFERENCES public.transaction(id);

ALTER TABLE ONLY public.entry
    ADD CONSTRAINT fk_account_id FOREIGN KEY (account_id) REFERENCES public.account(id);

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT fk_cf_kind_id FOREIGN KEY (meter, kind_natural_id) REFERENCES public.resource_kind(meter, natural_id);

ALTER TABLE ONLY public.resource
    ADD CONSTRAINT fk_cf_org_id FOREIGN KEY (cf_org_id) REFERENCES public.cf_org(id);

ALTER TABLE ONLY public.cf_org
    ADD CONSTRAINT fk_customer_id FOREIGN KEY (customer_id) REFERENCES public.customer(id);

ALTER TABLE ONLY public.account
    ADD CONSTRAINT fk_customer_id FOREIGN KEY (customer_id) REFERENCES public.customer(id);

ALTER TABLE ONLY public.resource_kind
    ADD CONSTRAINT fk_meter FOREIGN KEY (meter) REFERENCES public.meter(name);

ALTER TABLE ONLY public.measurement
    ADD CONSTRAINT fk_reading_id FOREIGN KEY (reading_id) REFERENCES public.reading(id);

ALTER TABLE ONLY public.measurement
    ADD CONSTRAINT fk_resource_id FOREIGN KEY (meter, resource_natural_id) REFERENCES public.resource(meter, natural_id);

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT fk_tier_id FOREIGN KEY (tier_id) REFERENCES public.tier(id);

ALTER TABLE ONLY public.account
    ADD CONSTRAINT fk_type_id FOREIGN KEY (type) REFERENCES public.account_type(id);

