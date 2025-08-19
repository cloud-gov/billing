# Generated schema

The database schema is canonically defined by the migrations in `sql/migrations` so that live databases can be migrated to new versions incrementally. `sqlc` can generate models from this without issue. However, is difficult for developers to understand the schema as a whole from the migrations.

For developer convenience, the `db-schema` Make target dumps the schema to this folder, with light formatting applied. Note that because river and tern recommend putting their tables in the same logical database as the application's tables, these appear in the raw `pg_dump` output. The river tables are filtered using `pg_dump`'s `--exclude-table` flag for readability, but some other entities like river's types and functions remain.
