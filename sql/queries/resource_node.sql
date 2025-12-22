/*
We can have ambiguous columns because sqlc handles it
noqa: disable=AM04
*/

-- name: GetResourceNode :one
select * from resource_node
where customer_id = $1 and slug = $2;

-- name: ListAncestors :many
select * from resource_node
where path @> subpath(sqlc.arg(path)::ltree, -1);

-- name: ListDescendants :many
select * from resource_node
where subpath(path, -1) <@ sqlc.arg(path)::ltree;

-- name: BulkCreateResourceNodes :exec
-- BulkCreateResourcesNodes creates Resource_Node rows in bulk with the minimum required columns. If a row with the given primary key already exists, that input item is ignored.
insert into resource_node (customer_id, slug, path, resource_natural_id)
select distinct on (rn.customer_id, rn.slug) *
from
  unnest(
    sqlc.arg(customer_id)::uuid[],
    sqlc.arg(slug)::text[],
    sqlc.arg(path)::ltree[],
    sqlc.arg(resource_natural_id)::text[]
  ) as rn (customer_id, slug, path, resource_natural_id)
on conflict (customer_id, slug) do nothing;
