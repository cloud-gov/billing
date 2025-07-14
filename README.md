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

### Cloud Foundry

As of writing, the application uses the local user's Cloud Foundry session to authenticate. You must sign into Cloud Foundry with `cf login` before starting the application.

### Database

Note that table name `schema_version` is reserved by tern for tracking migration status.

### Environment variables

- Postgres and AWS use their conventional environment variables for configuration. See [Postgres docs](https://www.postgresql.org/docs/current/libpq-envars.html) and [AWS docs](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html).

### River Queue

Billing service uses [River](https://riverqueue.com/docs) for transactional job queuing.

Tips:

- If you have to pass dependencies to a River worker, create a `NewXYZWorker` function that accepts those dependencies and sets them as private package vars. Do not pass dependencies as job args.
  - River job args are serialized to JSON, stored in the database, and deserialized to be run; dependencies like API clients and loggers may not fully serialize their internal state, resulting in nil pointer panics when they are unmarshalled and used.
  - Additionally, dependencies may have sensitive internal information that should not be persisted to the database.

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

### Notes

- Usage data is always persisted to the database, even if partial. The schema is informed by this need. For example, when we take a measurement for a `resource` but do not have a corresponding `resource_kind` in the database, we create an empty `resource_kind` record and will later ask the billing team to fill in the details. Our goal is to never lose usage data.

## References

- [Go.dev: Organizing a Go module](https://go.dev/doc/modules/layout) explains the basic, conventional layout of Go programs.
- [Grafana: How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years) is a source for much of the basic structure of the service.
- [Go Wiki: Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [The Go Blog: Package Names](https://go.dev/blog/package-names)
- [brandur.org: How We Went All In on sqlc/pgx for Postgres + Go](https://brandur.org/sqlc#caveats)
