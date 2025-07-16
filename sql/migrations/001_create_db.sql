CREATE TABLE tier (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  tier_credits BIGINT NOT NULL
);

CREATE TABLE customer (
  id   BIGSERIAL PRIMARY KEY,
  name text      NOT NULL,
  tier_id INT,

  CONSTRAINT fk_tier_id Foreign Key (tier_id) REFERENCES tier(id)
);

CREATE TABLE cf_org (
  id UUID PRIMARY KEY, -- TODO, natural_id instead, since it's defined by CF?
  name TEXT NOT NULL,
  customer_id BIGINT NOT NULL,
  CONSTRAINT fk_customer_id Foreign Key (customer_id) REFERENCES customer(id)
);

CREATE TABLE meter (
  name TEXT NOT NULL CHECK (char_length(trim(name)) > 0) UNIQUE
);
COMMENT ON TABLE meter IS 'A Meter reads usage information from a system in Cloud.gov. It also namespaces natural IDs for resources and resource_kinds; meter + natural_id is a primary key.';

CREATE TABLE resource_kind (
  meter TEXT NOT NULL,
  natural_id TEXT,
  credits INT,
  amount INT,
  unit_of_measure text NOT NULL,

  PRIMARY KEY (meter, natural_id),
  CONSTRAINT fk_meter Foreign Key (meter) REFERENCES meter(name)
);
COMMENT ON TABLE resource_kind IS 'ResourceKind represents a particular kind of billable resource. Note that natural_id can be empty because some meters may only read one kind of resource, and that resource kind may not have a unique identifier in the target system; it is uniquely identified by the meter name only.';

-- Instance of resource
CREATE TABLE resource (
  meter TEXT NOT NULL,
  natural_id TEXT NOT NULL,
  kind_natural_id TEXT,
  cf_org_id UUID NOT NULL,
  PRIMARY KEY (meter, natural_id),
  CONSTRAINT fk_cf_org_id Foreign Key (cf_org_id) REFERENCES cf_org(id),
  CONSTRAINT fk_cf_kind_id Foreign Key (meter, kind_natural_id) REFERENCES resource_kind(meter, natural_id)
);

CREATE TABLE reading (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE measurement (
  reading_id INT NOT NULL,
  meter TEXT NOT NULL,
  resource_natural_id TEXT NOT NULL,
  value INT NOT NULL,
  CONSTRAINT fk_reading_id Foreign Key (reading_id) REFERENCES reading(id),
  CONSTRAINT fk_resource_id Foreign Key (meter, resource_natural_id) REFERENCES resource(meter, natural_id)
);
