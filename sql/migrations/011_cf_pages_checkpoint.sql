create table pages_checkpoint (
  id int primary key default 1,
  last_ingested_charge_period_end_utc timestamptz not null,
  constraint pages_checkpoint_singleton check (id = 1)
);

---- create above / drop below ----

drop table pages_checkpoint;
