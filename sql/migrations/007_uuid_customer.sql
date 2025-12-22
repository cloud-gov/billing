-- can't get uuidv7 as extension or built-in easily in cloud.gov brokered RDS;
-- - it's available by default in pgsql v18, but we're stuck on v15 for now
-- - pg_uuidv7 clang extension is not among the trusted extensions, see ^1
--   - uuid-ossp is, but it doesn't support uuidv7 either.
-- - so probably the thing to do is implement in db itself
--   - what probably makes the most sense and won't be a hassle for everyone
--   - is to just implement uuidv7 in pl/pgsql, as in this gist:
--   - https://gist.github.com/kjmph/5bd772b2c2df145aa645b837da7eca74
-- - an alternative aws discusses is to...
--   - enable pg_tle and pl/rust
--   - install their example plrust uuid7 implementation
--   - they walk through that in ^2
--   - and discuss pg_tle in more depth in ^3
-- [1] https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/PostgreSQL.Concepts.General.FeatureSupport.Extensions.html#PostgreSQL.Concepts.General.Extensions.Trusted
-- [2] https://aws.amazon.com/blogs/database/implement-uuidv7-in-amazon-rds-for-postgresql-using-trusted-language-extensions/
-- [3] https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/PostgreSQL_trusted_language_extension.html

-- from https://gist.github.com/kjmph/5bd772b2c2df145aa645b837da7eca74
create or replace function uuid_generate_v7()
returns uuid
as $$
begin
  -- use random v4 uuid as starting point (which has the same variant we need)
  -- then overlay timestamp
  -- then set version 7 by flipping the 2 and 1 bit in the version 4 string
  return encode(
    set_bit(
      set_bit(
        overlay(uuid_send(gen_random_uuid())
                placing substring(int8send(floor(extract(epoch from clock_timestamp()) * 1000)::bigint) from 3)
                from 1 for 6
        ),
        52, 1
      ),
      53, 1
    ),
    'hex')::uuid;
end
$$
language plpgsql
volatile;

alter table customer add new_id uuid default uuid_generate_v7() unique;

alter table cf_org add new_id uuid constraint cfk references customer (new_id);
alter table account add new_id uuid constraint cfk references customer (new_id);

update cf_org set new_id = c.new_id from customer as c
where customer_id = c.id;

update account set new_id = c.new_id from customer as c
where customer_id = c.id;

alter table cf_org drop constraint fk_customer_id;
alter table account drop constraint fk_customer_id;

alter table customer drop constraint customer_pkey;
alter table customer rename id to old_id; -- keep old_id in case of rollback
alter table customer rename new_id to id;
alter table customer add primary key (id), add unique (old_id);

alter table cf_org drop customer_id;
alter table account drop customer_id;

alter table cf_org rename new_id to customer_id;
alter table account rename new_id to customer_id;

alter table cf_org rename constraint cfk to fk_customer_id;
alter table account rename constraint cfk to fk_customer_id;

-- recreate indexes from 006
create index if not exists cf_org_customer_id_idx
on cf_org (customer_id);
create index if not exists account_customer_type_idx
on account (customer_id, type);

---- create above / drop below ----

alter table cf_org
add old_id bigint constraint cfk references customer (old_id);
alter table account
add old_id bigint constraint cfk references customer (old_id);

update cf_org set old_id = c.old_id from customer as c
where customer_id = c.id;

update account set old_id = c.old_id from customer as c
where customer_id = c.id;

alter table cf_org drop constraint fk_customer_id;
alter table account drop constraint fk_customer_id;

alter table customer
drop constraint customer_pkey,
drop column id;

alter table customer rename old_id to id;
alter table customer add primary key (id);

alter table cf_org drop customer_id;
alter table account drop customer_id;

alter table cf_org rename old_id to customer_id;
alter table account rename old_id to customer_id;

alter table cf_org rename constraint cfk to fk_customer_id;
alter table account rename constraint cfk to fk_customer_id;

-- recreate indexes from 006
create index if not exists cf_org_customer_id_idx
on cf_org (customer_id);
create index if not exists account_customer_type_idx
on account (customer_id, type);

drop function uuid_generate_v7;
