CREATE INDEX idx_reading_created_at ON reading (created_at);

CREATE UNIQUE INDEX reading_hourly_uq
    ON reading ( date_trunc('hour', created_at) );
COMMENT ON INDEX reading_hourly_uq IS 'Make readings unique per hour.';

---- create above / drop below ----

DROP INDEX idx_reading_created_at;
DROP INDEX reading_hourly_uq;
