create table pages_org (
  id uuid primary key default uuid_generate_v7(),
  natural_id text not null unique,
  customer_id uuid not null,
  constraint fk_customer_id foreign key (customer_id) references customer (id)
);

create index if not exists pages_org_customer_id_idx
on pages_org (customer_id);

create table pages_node (
  natural_id text primary key,
  name       text not null
);

---- create above / drop below ----

drop table pages_node;
drop table pages_org;
