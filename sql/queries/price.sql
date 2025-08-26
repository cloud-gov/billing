-- name: CreatePriceWithID :one
INSERT INTO price (id, meter, kind_natural_id, unit_of_measure, microcredits_per_unit, valid_during)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
