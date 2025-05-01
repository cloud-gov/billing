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
  CONSTRAINT fk_customer_id Foreign Key (customer_id) REFERENCES customer(id)
);

CREATE TABLE tier (
  id SERIAL PRIMARY KEY ,
  name TEXT NOT NULL,
  tier_credits BIGINT
);

-- Instance of resource
CREATE TABLE billable_resource (
  id SERIAL PRIMARY KEY,
  native_id TEXT,
  class_id TEXT,
  cf_org_id BIGINT,
  CONSTRAINT fk_cf_org_id Foreign Key (cf_org_id) REFERENCES cf_org(id)
);

-- Resource type
CREATE TABLE billable_class (
  id SERIAL PRIMARY KEY,
  native_id TEXT,
  credits INT,
  amount INT, 
  unit_of_measure text NOT NULL
);

