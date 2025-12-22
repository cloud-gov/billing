create extension if not exists ltree;

--
-- LTREE COLUMNS
--
-- Add path & slugs to customer & resource
-- size of slug is restricted to avoid oversize path
--

alter table customer
add path ltree,
add slug varchar(50);

alter table resource
add path        ltree,
add slug        varchar(50),
add customer_id uuid references customer (id);

-- essentially namespace the slugs to customers
create unique index resource_cid_slug_idx on resource (customer_id, slug);

--
-- MOVE FUNCTIONS
--
-- Moves branches while valditing no cirucular refs
--
-- A function per tree; I considered combining with case statements
-- but that seemed overly complicated.
--

create or replace function move_customer(
  p_customer_id   uuid,
  p_new_parent_id uuid
)
returns           boolean as $$
declare
  v_old_path ltree;
  v_new_parent_path ltree;
  v_new_path ltree;
begin
  -- get current path
  select path into v_old_path 
  from customer where id = p_customer_id;

  -- get new parent path
  select path into v_new_parent_path 
  from customer where id = p_new_parent_id;

  -- check for circular reference
  if v_new_parent_path <@ v_old_path then
    raise exception 'cannot move customer to its own descendant';
  end if;

  -- perform the move
  v_new_path := v_new_parent_path || subpath(v_old_path, -1, 1);

  update customer
  set 
    path = v_new_path || subpath(path, nlevel(v_old_path))
  where path <@ v_old_path;

  return true;
end;
$$ language plpgsql;

create or replace function move_resource(
  p_resource_id   uuid,
  p_new_parent_id uuid
)
returns           boolean as $$
declare
  v_old_path ltree;
  v_new_parent_path ltree;
  v_new_path ltree;
begin
  -- get current path
  select path into v_old_path 
  from resource where id = p_resource_id;

  -- get new parent path
  select path into v_new_parent_path 
  from resource where id = p_new_parent_id;

  -- check for circular reference
  if v_new_parent_path <@ v_old_path then
    raise exception 'cannot move resource to its own descendant';
  end if;

  -- perform the move
  v_new_path := v_new_parent_path || subpath(v_old_path, -1, 1);

  update resource
  set 
    path = v_new_path || subpath(path, nlevel(v_old_path))
  where path <@ v_old_path;

  return true;
end;
$$ language plpgsql;


--
-- CHECKS
-- Ensure valid looking paths; e.g., `thing`, or `thing.one`,
-- but not `thing.`, `.thing`, or `thing..one`.
--

alter table customer
add constraint valid_path
check (path::text ~ '^[A-Za-z0-9_]+(\\.[A-Za-z0-9_]+)*$');

alter table resource
add constraint valid_path
check (path::text ~ '^[A-Za-z0-9_]+(\\.[A-Za-z0-9_]+)*$');

--
-- INDEXES
-- https://www.postgresql.org/docs/15/ltree.html#id-1.11.7.32.7
--
-- GiST for tree navigation
-- B-tree for equalities
--

-- customer
create index customer_path_gist_idx on customer using gist (path);
create index customer_path_btree_idx on customer using btree (path);

-- resource
create index resource_path_gist_idx on resource using gist (path);
create index resource_path_btree_idx on resource using btree (path);

-- For customer isolation
create index customer_path_idx on resource (customer_id, path);

---- create above / drop below ----

-- LTREE COLUMNS
-- …and related constraints & indexes
alter table customer drop path, drop slug;
alter table resource drop path, drop slug, drop customer_id;

-- MOVE FUNCTIONS
drop function move_customer;
drop function move_resource;

-- EXTENSIONS
drop extension ltree;
