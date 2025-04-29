CREATE TABLE customer (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL
);

CREATE TABLE cf_org (
  id GUID PRIMARY KEY,
  name TEXT NOT NULL,
  tier_id TEXT NOT NULL,
  credits_quota BIGINT NOT NULL,
  credits_used BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  CONSTRAINT fk_customer Foreign Key (customer_id) REFERENCES customer(id)
);

CREATE TABLE tier (
  id SERIAL NOT NULL,
  name TEXT NOT NULL,
  tier_credits BIGINT
);
