alter table resource_node
drop constraint resource_node_slug_uq,
add constraint resource_node_path_uq unique (customer_id, path);

---- create above / drop below ----

alter table resource_node
drop constraint resource_node_path_uq,
add constraint resource_node_slug_uq unique (customer_id, slug);
