alter table resource_node
drop constraint resource_node_slug_uq;

---- create above / drop below ----

alter table resource_node
add constraint resource_node_slug_uq unique (customer_id, slug);
