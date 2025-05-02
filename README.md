# Billing service

Under development.

## Development

The program has the following structure:

```
sql/ # Source SQL for sqlc to convert into Go.
  migrations/
  queries/
  schema/
src/ # All Go code for the service.
  internal/ # Go code that is not publicly exported.
    db/ # Destination directory for sqlc-generated Go code.
```

## References

- [Grafana: How I write HTTP services in Go after 13 years](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years) is a source for much of the basic structure of the service.
