alter table customer
add column active boolean not null default false;

insert into customer (name)
select 'Pending Pages Customer'
where not exists (select 1 from customer where name = 'Pending Pages Customer');

---- create above / drop below ----

delete from customer where name = 'Pending Pages Customer';

alter table customer drop column active;
