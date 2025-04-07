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
