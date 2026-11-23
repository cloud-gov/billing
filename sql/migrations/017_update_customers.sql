UPDATE customer
SET name = 'TTS OROS FAC'
WHERE name = 'FAC';

UPDATE customer
SET name = 'DOD SOCOM'
WHERE name = 'SOCOM';

DELETE FROM customer WHERE name IN ('DoD', 'OROS');

---- create above / drop below ----

INSERT INTO customer (id, name) VALUES (gen_random_uuid(), 'OROS');

UPDATE customer
SET name = 'FAC'
WHERE name = 'TTS OROS FAC';

UPDATE customer
SET name = 'SOCOM'
WHERE name = 'DOD SOCOM';
