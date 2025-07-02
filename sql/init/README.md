# Init Migrations

Create the database required by the billing service here. The `billing` database must already exist for `tern` and `river` to connect to it and run their respective migrations. These separate init migrations achieve that.

Run these migrations with `make db-init`. See Makefile for more.
