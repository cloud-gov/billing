# Billing and Usage CLI

## Building the CLI

```shell
go build
```

## Using the CLI

To see the CLI options:

```shell
./usage -h
```

To produce a JSON report of usage:

1. Create an SSH tunnel to the billing database
1. Create a `.env` with the database connection information and other necessary variables

    ```env
    # fill with your tunnel's connection parameters
    export PGHOST=localhost
    export PGPORT=
    export PGDATABASE=
    export PGUSER=
    export PGPASSWORD=
    
    # blank values are fine
    export CF_API_URL=
    export CF_CLIENT_ID=
    export CF_CLIENT_SECRET=
    export OIDC_ISSUER=
    ```

1. Set the environment variables in your shell:

    ```shell
    source .env
    ```

1. Run the usage CLI and stream output to a JSON file:

    ```shell
    ./usage -cname 'some customer' | jq | tee report.json
    ```
