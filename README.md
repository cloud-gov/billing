# Billing service

The billing service tracks customer usage of Cloud.gov products for billing purposes.

## Development

Prerequisite: Docker must be installed and running.

Copy `docker.env.example` to `docker.env` and fill in any missing values.

Set up the database:

```sh
make migrate
```

Most development tasks can be achieved using the Make targets defined in [Makefile](./Makefile). See the Makefile for the full list.

```sh
make watchgen # Watch .sql files for changes. On change, regenerate database Go bindings with sqlc. Consider running this in a separate shell at the same time as 'make watch'.
make watch # Watch .go files for changes. On change, recompile and restart the server.
make clean # Shut down the database if it is running and clean binary artifacts.
make psql # Connect to the local database.
```

To run the tests:

```sh
make test # Run unit tests.
make test-db # Run database tests.
make psql-testdb # Connect to the test database.
```

Make request to the locally running server:

```sh
make jwt # Get a token from the configured UAA, based on OIDC_ISSUER host. Requires CF_CLIENT_ID and CF_CLIENT_SECRET to be set.
curl -H "Authorization: bearer $(cat jwt.txt)" localhost:8080/some/path # Make a request with the authentication header set.
```

### Cloud Foundry

The application uses service account credentials to authenticate to the Cloud Foundry API (CAPI). Set `CF_CLIENT_ID` and `CF_CLIENT_SECRET` using credentials from CredHub before starting the application with `make watch`.

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

### Collecting data

Usage data is always persisted to the database, even if partial. The schema is informed by this need. For example, when we take a measurement for a `resource` but do not have a corresponding `resource_kind` in the database, we create an empty `resource_kind` record and will later ask the billing team to fill in the details. Our goal is to never lose usage data.

### Time

For business operations like posting usage to customer accounts, use the timezone for `America/New_York`. This aligns with other Cloud.gov business processes; for example, Cloud.gov agreements are considered to execute in Eastern Time.

For all other operations, use UTC. For instance, when a usage reading is taken, the timestamp is captured in UTC.

## Known Limitations

- If we miss a reading, the customer is not charged for that hour. We could fix this by interpolating usage based on measurements taken before and after the gap.
- Prices could be set differently to avoid rounding. Our pricing table prices many resources per month of usage: X credits per 730 hours of use, a normalized month. However, we measure usage hourly. To calculate the credits for a measurement, we solve `measurement.value * price.amount_microcredits / price.unit`. Dividing by unit can force us to round: For instance, `1,000,000 microcredits/month / 730 hours = 1,369.8630136986 microcredits`, which must be rounded. We could avoid this by pricing resources in the same unit we measure: Per hour. Then, we would calculate credits for a measurement with `measurement.value * price.amount_microcredits`, avoiding division.

## References

- [Go.dev: Organizing a Go module](https://go.dev/doc/modules/layout) explains the basic, conventional layout of Go programs.
- [Grafana: How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years) is a source for much of the basic structure of the service.
- [Go Wiki: Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [The Go Blog: Package Names](https://go.dev/blog/package-names)
- [brandur.org: How We Went All In on sqlc/pgx for Postgres + Go](https://brandur.org/sqlc#caveats)
- [An Elegant DB Schema for Double-Entry Accounting](https://web.archive.org/web/20220901165809/https://www.journalize.io/blog/an-elegant-db-schema-for-double-entry-accounting) is the basis for the double-entry accounting schema. (Our schema has some differences.)
