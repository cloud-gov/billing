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
on conflict (customer_id, resource_natural_id) do update
  set slug = excluded.slug, path = excluded.path;

-- name: GetAppsUsageBySpace :many
select
  subpath(rn.path, 1, -2) as org,
  regexp_replace(
    subpath(rn.path, 1, -1)::text,
    '_?(test|dev|stage|staging|prod)$', -- to aggregate environments
    ''
  ) as space,
  sum(m.amount_microcredits) as total_microcredits,
  sum(m.amount_microcredits) / 1000000 as total_credits
from resource_node as rn
  inner join measurement as m on rn.resource_natural_id = m.resource_natural_id
where
  rn.customer_id = $1
  and rn.path ~ 'apps.usage.cforg%.space%.*{1,}'
group by org, space;
