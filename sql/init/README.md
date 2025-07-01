# Init Migrations

Create databases required by the billing service here. The `billing` and `river` databases must already exist for `tern` and `river` to connect to them and run their respective migrations. These separate init migrations achieve that.

Run these migrations with `make`. See Makefile for targets.
