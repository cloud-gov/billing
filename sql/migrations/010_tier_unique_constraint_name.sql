alter table tier
add constraint name_unique unique (name);

---- create above / drop below ----

alter table tier
drop constraint name_unique;
