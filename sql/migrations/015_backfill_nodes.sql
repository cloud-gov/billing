with
  customer_orgs as (
    select
      o.id as org_id,
      o.name as org_name,
      n.slug as org_slug,
      n.path as org_path,
      c.id as customer_id,
      c.name as customer_name,
      c.path as customer_path
    from cf_org as o
      inner join customer as c on o.customer_id = c.id
      inner join resource_node as n on o.id = n.resource_natural_id::uuid
  ),

  missing_resources as (
    select distinct on (r.natural_id)
      r.natural_id,
      r.cf_org_id,
      co.org_path,
      co.customer_id,
      co.customer_name,
      coalesce(
        nullif(rk.name, ''),
        nullif(rk.natural_id, ''),
        'nil'
      ) as kind_name,

      -- would require CF API lookup on resource_kind,
      -- but these are deleted resources so it may be lost
      'rsrc_bfDelRsrc_' || r.natural_id as slug,

      -- could potentially find org name through CF API
      lower(regexp_replace(
        coalesce(co.org_name, 'bfDelRsrc_nil'), '\W+', '_'
      )) as org_name
    from resource as r
      inner join customer_orgs as co on r.cf_org_id = co.org_id
      inner join resource_kind as rk on r.kind_natural_id = rk.natural_id
    where r.natural_id not in
      (select rn.resource_natural_id from resource_node as rn)
  )

insert into resource_node (resource_natural_id, customer_id, slug, path)
select
  r.natural_id,
  r.customer_id,
  r.slug,
  regexp_replace(
  concat_ws(
    '.',
    coalesce(
      nullif(r.org_path::text, ''),
      ('apps.usage.cforg_' || r.org_name)
    ),
    concat(
      'kind_bfDelRsrc_',
      regexp_replace(r.kind_name, '\W+', '_', 'g')
    ),
    r.slug
  ), '[^a-zA-Z0-9_.]', '', 'g')::ltree as path
from missing_resources as r;

---- create above / drop below ----

delete from resource_node
where path ~ 'bfDelRsrc%';
