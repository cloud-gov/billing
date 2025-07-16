-- Accounting contains convenience queries that may not have practical use in the application.

-- name: AccountingEquation :many
-- Output the two sides of the standard accounting equation as two rows, for all defined top-level account types. For instance, if only 'liabilities' and 'expenses' are defined, the output is
--   liabilities
--   expenses
SELECT string_agg(name, ' + ') AS expression
FROM account_type
WHERE id % 100 = 0
GROUP BY id, normal;
