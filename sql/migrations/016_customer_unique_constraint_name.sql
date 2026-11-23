alter table customer
add constraint customer_name_unique unique (name);

---- create above / drop below ----

alter table customer
drop constraint customer_name_unique;
