alter table measurement
add column id uuid primary key default uuid_generate_v7();

---- create above / drop below ----

alter table measurement drop column id;
