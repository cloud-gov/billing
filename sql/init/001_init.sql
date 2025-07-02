---- tern: disable-tx ----
CREATE DATABASE billing;

---- create above / drop below ----

---- tern: disable-tx ----
DROP DATABASE billing WITH (FORCE);
