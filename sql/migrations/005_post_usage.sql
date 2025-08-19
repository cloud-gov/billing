alter table entry
add column amount_microcredits bigint;

update entry
set amount_microcredits = amount * 1e6;

alter table entry
drop column amount;

alter table measurement
add column amount_microcredits bigint;

-- add pricing table
create table price (
	meter text not null,
	natural_id text not null,

	unit_of_measure text,
	microcredits_per_unit bigint,
	valid_during tstzrange,

	constraint fk_resource_kind foreign key (meter, natural_id) references resource_kind(meter, natural_id)
);

-- I did not automatically insert data from these columns into `price` because as of writing, there was no data in those columns.
alter table resource_kind
drop column credits,
drop column amount,
drop column unit_of_measure;

---- create above / drop below ----

update entry
set amount = amount_microcredits / 1e6;

alter table entry
drop column amount_microcredits;

alter table measurement
drop column amount_microcredits;

drop table price;

alter table resource_kind
add column credits integer,
add column amount integer,
add column unit_of_measure text;
