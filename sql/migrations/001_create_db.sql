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
  id UUID PRIMARY KEY, -- TODO, natural_id instead, since it's defined by CF?
  name TEXT NOT NULL,
  tier_id INT NOT NULL,
  credits_quota BIGINT NOT NULL,
  credits_used BIGINT NOT NULL,
  customer_id BIGINT NOT NULL,
  CONSTRAINT fk_customer_id Foreign Key (customer_id) REFERENCES customer(id),
  CONSTRAINT fk_tier_id Foreign Key (tier_id) REFERENCES tier(id)
);

-- A Meter reads usage information from a system in Cloud.gov. It also namespaces natural IDs for resources and resource_kinds; meter + natural_id is a primary key.
CREATE TABLE meter (
  name TEXT NOT NULL CHECK (char_length(trim(name)) > 0) UNIQUE
);

-- Resource type
-- Note that natural_id is nullable because some meters may only read one kind of resource, and that resource may not have a unique identifier in the target system.
CREATE TABLE resource_kind (
  meter TEXT NOT NULL,
  natural_id TEXT,
  credits INT,
  amount INT,
  unit_of_measure text NOT NULL,

  PRIMARY KEY (meter, natural_id),
  CONSTRAINT fk_meter Foreign Key (meter) REFERENCES meter(name)
);

-- Instance of resource
CREATE TABLE resource (
  id SERIAL PRIMARY KEY,
  natural_id TEXT NOT NULL,
  meter TEXT NOT NULL,
  kind_natural_id TEXT,
  cf_org_id UUID NOT NULL,
  CONSTRAINT fk_cf_org_id Foreign Key (cf_org_id) REFERENCES cf_org(id),
  CONSTRAINT fk_cf_kind_id Foreign Key (meter, kind_natural_id) REFERENCES resource_kind(meter, natural_id)
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
