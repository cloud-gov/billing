# Billing service

Under development.

## Development

Prerequisite: Docker must be installed and running.

Set up the database:

```sh
make migrate
```

Most development tasks can be achieved using the Make targets defined in [Makefile](./Makefile). See the Makefile for the full list.

```sh
make watchgen # Watch .sql files for changes. On change, regenerate database Go bindings with sqlc. Consider running this in a separate shell at the same time as 'make watch'.
make watch # Watch .go files for changes. On change, recompile and restart the server.
make clean # Shut down the database if it is running and clean binary artifacts.
```

To run the tests:

```sh
make test # Run unit tests
make test-db # Run database tests. See warning below.
```

**Warning**: `make test-db` will run `db-down`, shutting down the postgres container if it is running. This will erase all data in the database.

### Cloud Foundry

As of writing, the application uses the local user's Cloud Foundry session to authenticate. You must sign into Cloud Foundry with `cf login` before starting the application.

### Database

Note that table name `schema_version` is reserved by tern for tracking migration status.

### Dependencies

Why does this project use the packages that it does?

- `jackc/pgx` is used because sqlc supports it, and unlike `database/sql` it supports COPY and the postgres binary protocol, which is faster than the SQL textual protocol.
- `jackc/tern` is used for migrations because we already use another package by `jackc`, `pgx`, and sqlc supports it.
- `coreos/go-oidc` is used for JWT validation over alternatives because it is maintained by a reputable organization (CoreOS) and supports JWKS discovery.

### Environment variables

- Postgres and AWS use their conventional environment variables for configuration. See [Postgres docs](https://www.postgresql.org/docs/current/libpq-envars.html) and [AWS docs](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html).

### River Queue

Billing service uses [River](https://riverqueue.com/docs) for transactional job queuing.

Tips:

- If you have to pass dependencies to a River worker, create a `NewXYZWorker` function that accepts those dependencies and sets them as private package vars. Do not pass dependencies as job args.
  - River job args are serialized to JSON, stored in the database, and deserialized to be run; dependencies like API clients and loggers may not fully serialize their internal state, resulting in nil pointer panics when they are unmarshalled and used.
  - Additionally, dependencies may have sensitive internal information that should not be persisted to the database.

### Testing

Tests follow these naming conventions:

- `TestDB*`: Database tests. Run with `make test-db`, which starts a fresh postgres Docker instance and migrates it to the latest migration.

## Packages

The program has the following structure:

```
sql/          # Source SQL for sqlc to convert into Go.
  init/       # Database creation must be separate from migrations so Tern has something to connect to.
  migrations/ # Schema for billing service tables.
  queries/    # Queries for billing service tables.
internal/     # All non-exported Go code for the service.
  api/        # The HTTP API surface of the application.
  db/         # Destination directory for sqlc-generated Go code.
  jobs/       # River jobs for performing asynchronous work.
  server/     # Code for managing the web server lifecycle.
  usage/      # Code for reading usage information from Cloud.gov systems.
gen.go        # Program-scope go:generate directives.
main.go       # Entrypoint for the server program.
```

## Design Notes

- Usage data is always persisted to the database, even if partial. The schema is informed by this need. For example, when we take a measurement for a `resource` but do not have a corresponding `resource_kind` in the database, we create an empty `resource_kind` record and will later ask the billing team to fill in the details. Our goal is to never lose usage data.

### Time

For business operations like posting usage to customer accounts, use the timezone for `America/New_York`. This aligns with other Cloud.gov business processes; for example, Cloud.gov agreements are considered to execute in Eastern Time.

For all other operations, use UTC. For instance, when a usage reading is taken, the timestamp is captured in UTC.

## Known Limitations

- Currently, credit usage corresponding to each measurement is calculated by multiplying the measurement's value by the applicable price's microcredits_per_unit and dividing by 730, a normalized hours-per-month. This assumes a reading (a collection of measurements) is taken every hour. If a reading fails to be taken due to the application being offline or any other reason, usage is not extrapolated for the gap. For example, suppose readings 1, 2, and 3 were meant to be taken at 1am, 2am, and 3am, covering 3 hours total (midnight to 3am). If reading 2 is skipped, usage will only be calculated for 2 of the three hours, because the current usage posting job does not extrapolate usage for the gap.

## References

- [Go.dev: Organizing a Go module](https://go.dev/doc/modules/layout) explains the basic, conventional layout of Go programs.
- [Grafana: How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years) is a source for much of the basic structure of the service.
- [Go Wiki: Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [The Go Blog: Package Names](https://go.dev/blog/package-names)
- [brandur.org: How We Went All In on sqlc/pgx for Postgres + Go](https://brandur.org/sqlc#caveats)
- [An Elegant DB Schema for Double-Entry Accounting](https://web.archive.org/web/20220901165809/https://www.journalize.io/blog/an-elegant-db-schema-for-double-entry-accounting) is the basis for the double-entry accounting schema. (Our schema has some differences.)
