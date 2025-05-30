CREATE TABLE customer (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL
);

CREATE TABLE tier (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  tier_credits BIGINT
);

CREATE TABLE cf_org (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  tier_id INT NOT NULL,
  credits_quota BIGINT NOT NULL,
  credits_used BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  CONSTRAINT fk_customer_id Foreign Key (customer_id) REFERENCES customer(id),
  CONSTRAINT fk_tier_id Foreign Key (tier_id) REFERENCES tier(id)

);

-- Resource type
CREATE TABLE billable_class (
  id SERIAL PRIMARY KEY,
  native_id TEXT,
  credits INT,
  amount INT, 
  unit_of_measure text NOT NULL
);

-- Instance of resource
CREATE TABLE billable_resource (
  id SERIAL PRIMARY KEY,
  native_id TEXT,
  class_id INT,
  cf_org_id UUID,
  CONSTRAINT fk_cf_org_id Foreign Key (cf_org_id) REFERENCES cf_org(id),
  CONSTRAINT fk_cf_class_id Foreign Key (class_id) REFERENCES billable_class(id)
);




