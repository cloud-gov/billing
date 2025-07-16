ALTER TABLE resource
  ADD CONSTRAINT resource_meter_natural_id_uq UNIQUE (meter, natural_id),
  ALTER COLUMN kind_natural_id SET NOT NULL;

ALTER TABLE cf_org
ALTER COLUMN name DROP NOT NULL,
ALTER COLUMN customer_id DROP NOT NULL;

ALTER TABLE resource_kind
ALTER COLUMN unit_of_measure DROP NOT NULL,
ALTER COLUMN natural_id SET NOT NULL; -- Can be empty, but pkeys cannot be null.

CREATE UNIQUE INDEX idx_resource_meter_natural_id ON resource (meter, natural_id);
COMMENT ON INDEX idx_resource_meter_natural_id IS 'Enables efficient deduplicated inserts using BulkCreateResources function.';

CREATE UNIQUE INDEX idx_resource_kind_meter_natural_id ON resource_kind (meter, natural_id);
COMMENT ON INDEX idx_resource_kind_meter_natural_id IS 'Enables efficient deduplicated inserts using BulkCreateResourceKinds function.';

---- create above / drop below ----

ALTER TABLE resource
DROP CONSTRAINT resource_uq,
ALTER COLUMN kind_natural_id DROP NOT NULL;

ALTER TABLE cf_org
ALTER COLUMN name SET NOT NULL,
ALTER COLUMN customer_id SET NOT NULL;

ALTER TABLE resource_kind
ALTER COLUMN unit_of_measure SET NOT NULL,
ALTER COLUMN natural_id DROP NOT NULL;

DROP INDEX idx_resource_meter_natural_id;
DROP INDEX idx_resource_kind_meter_natural_id;
