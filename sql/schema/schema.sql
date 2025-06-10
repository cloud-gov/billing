CREATE TABLE customer (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL
);

CREATE TABLE tier (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  tier_credits BIGINT NOT NULL
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
CREATE TABLE resource_kind (
  id SERIAL PRIMARY KEY,
  natural_id TEXT,
  credits INT,
  amount INT, 
  unit_of_measure text NOT NULL
);

-- Instance of resource
CREATE TABLE resource (
  id SERIAL PRIMARY KEY,
  natural_id TEXT,
  kind_id INT NOT NULL,
  cf_org_id UUID NOT NULL,
  CONSTRAINT fk_cf_org_id Foreign Key (cf_org_id) REFERENCES cf_org(id),
  CONSTRAINT fk_cf_kind_id Foreign Key (kind_id) REFERENCES resource_kind(id)
);

CREATE TABLE reading (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE measurement (
  reading_id INT NOT NULL,
  resource_id INT NOT NULL,
  value INT NOT NULL,
  CONSTRAINT fk_reading_id Foreign Key (reading_id) REFERENCES reading(id),
  CONSTRAINT fk_resource_id Foreign Key (resource_id) REFERENCES resource(id)
);
