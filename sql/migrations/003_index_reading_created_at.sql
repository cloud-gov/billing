CREATE INDEX idx_reading_created_at ON reading (created_at);

CREATE UNIQUE INDEX reading_hourly_uq
    ON reading ( date_trunc('hour', created_at) );
COMMENT ON INDEX reading_hourly_uq IS 'Make readings unique per hour.';

ALTER TABLE reading
ADD COLUMN periodic BOOLEAN NOT NULL;
COMMENT ON COLUMN reading.periodic IS 'Periodic is true if a reading was taken automatically as part of the periodic usage measurement schedule, or false if it was requested manually.';

---- create above / drop below ----

DROP INDEX idx_reading_created_at;
DROP INDEX reading_hourly_uq;

ALTER TABLE reading
DROP COLUMN periodic;
