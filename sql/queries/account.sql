-- name: GetAccountForCustomerAndType :one
SELECT a.*
FROM account a
JOIN customer c on a.customer_id = c.id
WHERE c.name = $1
AND a.type = $2
LIMIT 1;
