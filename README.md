# Billing service

Under development.

## Development

The program has the following structure:

```
sql/      # Source SQL for sqlc to convert into Go.
  migrations/
  queries/
  schema/
internal/ # All non-exported Go code for the service.
  api/    # The HTTP API surface of the application.
  cf/     # Code specifically for interacting with Cloud Foundry.
  db/     # Destination directory for sqlc-generated Go code.
  server/ # Code for managing the web server lifecycle.
  usage/  # Code for reading usage information from Cloud.gov systems.
main.go   # Entrypoint for the server program.
```

## References

- [Go.dev: Organizing a Go module](https://go.dev/doc/modules/layout) explains the basic, conventional layout of Go programs.
- [Grafana: How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years) is a source for much of the basic structure of the service.
