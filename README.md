# Billing service

Under development.

## Development

Prerequisite: Docker must be installed and running.

Set up the database:

```sh
make migrate
```

Most development tasks can be achieved using the Make targets defined in [Makefile](./Makefile).

```sh
make watchgen # Watch .sql files for changes. On change, regenerate database Go bindings with sqlc. Consider running this in a separate shell at the same time as 'make watch'.
make watch # Watch .go files for changes. On change, recompile and restart the server.
make clean # Shut down the database if it is running and clean binary artifacts.
```

### Database

Note that table name `schema_version` is reserved by tern for tracking migration status.

### Environment variables

- Postgres and AWS use their conventional environment variables for configuration. See [Postgres docs](https://www.postgresql.org/docs/current/libpq-envars.html) and [AWS docs](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html).

## Packages

The program has the following structure:

```
sql/      # Source SQL for sqlc to convert into Go.
  migrations/
  queries/
internal/ # All non-exported Go code for the service.
  api/    # The HTTP API surface of the application.
  db/     # Destination directory for sqlc-generated Go code.
  server/ # Code for managing the web server lifecycle.
  usage/  # Code for reading usage information from Cloud.gov systems.
gen.go    # Program-scope go:generate directives.
main.go   # Entrypoint for the server program.
```

## References

- [Go.dev: Organizing a Go module](https://go.dev/doc/modules/layout) explains the basic, conventional layout of Go programs.
- [Grafana: How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years) is a source for much of the basic structure of the service.
