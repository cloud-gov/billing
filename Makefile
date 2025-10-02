# Use bash so we can use `source`.
SHELL := /bin/bash

.PHONY: db-up
db-up:
	docker compose up --detach --wait db

.PHONY: db-down
db-down:
	docker compose down db

.PHONY: gen
gen: db-up
	@# sqlc does not remove generated files on its own when the source file is deleted; remove them manually here. (|| true reports success even when no files were found.)
	@rm internal/db/* || true
	@date
	go generate ./...
	@echo "Done."

.PHONY: build
build: gen
	go build .

.PHONY: test
test: gen
	go test ./... -skip TestDB

.PHONY: debug
debug: gen
	@set -a; source docker.env; set +a; dlv debug

.PHONY: debug-test
debug-test-db:
	@# Equivalent to `make db-up` and `make db-down`, but for the ephemeral database
	@docker compose down --volumes test-db-ephemeral
	@docker compose up --detach --wait test-db-ephemeral

	@# Equivalent to `make db-init`, but for the ephemeral database
	@set -a; source docker.env; PGDATABASE=postgres PGPORT=5433; set +a; \
	go tool tern migrate --config sql/init/tern.conf --migrations sql/init

	@# Equivalent to `make db-migrate`, but for the ephemeral database.
	@# Migrate to the latest migration.
	@set -a; source docker.env; PGPORT=5433; set +a; \
	go tool tern migrate --config sql/migrations/tern.conf --migrations sql/migrations
	@# Migrate River schema to latest.
	@set -a; source docker.env; PGPORT=5433; set +a; \
	go tool river migrate-up

	@echo "Running database tests (TestDB*)..."
	@# Edit and use like: dlv test ./internal/dbx -- -test.run TestDBPostUsage
	@set -a; source docker.env; PGPORT=5433; set +a; \
	dlv test ./internal/dbx -- -test.run TestDBPostUsage

.PHONY: watchgen
watchgen:
	@echo "Watching for .sql file changes. Press ctrl+c *twice* to exit, or once to rebuild."
	@while true; do \
		find . -type f -name '*.sql' | entr -d make gen ; \
		sleep 0.5 ; \
	done

# Run entr in a while loop because it exits when files are deleted.
# Run targets db-up and db-init before running.
.PHONY: watch
watch:
	@echo "Watching for .go file changes. Press ctrl+c *twice* to exit, or once to rebuild."
	@while true; do \
		set -a; source docker.env; set +a; \
		find . -type f -name '*.go' \
		| entr -d -r go run . ; \
		sleep 0.5 ; \
	done

.PHONY: watchtest
watchtest:
	@echo "Running unit tests every time .go files change. Press ctrl+c *twice* to exit, or once to rebuild."
	@while true; do \
		sleep 0.5 ; \
		set -a; source docker.env; set +a; find . -type f -name '*.go' | entr -d go test ./... ; \
	done

.PHONY: clean
clean: db-down
	go clean

.PHONY: db-init
db-init:
	@# Initialize the database.
	@set -a; source docker.env; PGDATABASE=postgres; set +a; \
	go tool tern migrate --config sql/init/tern.conf --migrations sql/init

.PHONY: db-drop
db-drop:
	@# Drop the database.
	@set -a; source docker.env; PGDATABASE=postgres; set +a; \
	go tool tern migrate --config sql/init/tern.conf --migrations sql/init --destination 0

.PHONY: db-migrate
db-migrate: db-init
	@# Migrate to the latest migration.
	@set -a; source docker.env; set +a; \
	go tool tern migrate --config sql/migrations/tern.conf --migrations sql/migrations
	@# Migrate River schema to latest.
	@set -a; source docker.env; set +a; \
	go tool river migrate-up

.PHONY: db-remigrate
db-remigrate: db-init
	@# Redo the latest migration in Tern.
	@set -a; source docker.env; set +a; \
	go tool tern migrate -d -+1 --config sql/migrations/tern.conf --migrations sql/migrations

.PHONY: db-reset
db-reset: db-drop db-init db-migrate
	@echo "Database reset. Restart app to reconnect."

.PHONY: psql
psql:
	@set -a; source docker.env; set +a; psql

.PHONY: psql-testdb
psql-testdb:
	@set -a; source docker.env; PGPORT=5433; set +a; psql

.PHONY: db-schema
db-schema:
	@set -a; source docker.env; set +a; \
	pg_dump --schema-only --exclude-table='river*' --exclude-table="schema_version" --no-owner \
	| grep --invert-match "\-\-" \
	| cat -s \
	> sql/schema/generated.sql

.PHONY: test-db
test-db:
	@# Equivalent to `make db-up` and `make db-down`, but for the ephemeral database
	@docker compose down --volumes test-db-ephemeral
	@docker compose up --detach --wait test-db-ephemeral

	@# Equivalent to `make db-init`, but for the ephemeral database
	@set -a; source docker.env; PGDATABASE=postgres PGPORT=5433; set +a; \
	go tool tern migrate --config sql/init/tern.conf --migrations sql/init

	@# Equivalent to `make db-migrate`, but for the ephemeral database.
	@# Migrate to the latest migration.
	@set -a; source docker.env; PGPORT=5433; set +a; \
	go tool tern migrate --config sql/migrations/tern.conf --migrations sql/migrations
	@# Migrate River schema to latest.
	@set -a; source docker.env; PGPORT=5433; set +a; \
	go tool river migrate-up

	@echo "Running database tests (TestDB*)..."
	@# Disable caching with -count=1, since go does not cache bust when .sql files change
	@set -a; source docker.env; PGPORT=5433; set +a; \
	go test ./... -run TestDB -count=1

# Run from inside a container.
# Ephemeral database must already be up via docker compose.
.PHONY: test-db-ci
test-db-ci:
	@# Equivalent to `make db-init`, but for the ephemeral database
	@# Intended to be run inside a container.
	PGDATABASE=postgres \
	go tool tern migrate --config sql/init/tern.conf --migrations sql/init

	@# Equivalent to `make db-migrate`, but for the ephemeral database.
	@# Migrate to the latest migration.
	go tool tern migrate --config sql/migrations/tern.conf --migrations sql/migrations
	@# Migrate River schema to latest.
	go tool river migrate-up

	@echo "Running database tests (TestDB*)..."
	@# Disable caching with -count=1, since go does not cache bust when .sql files change
	go test ./... -run TestDB -count=1

.PHONY: jwt
jwt:
	@set -a; source docker.env; set +a; \
	uaac target $${OIDC_ISSUER%/oauth/token}; \
	uaac token client get $$CF_CLIENT_ID -s $$CF_CLIENT_SECRET --scope "usage.admin"; \
	uaac context billing | grep access_token | awk '{print $$2}' > jwt.txt

	@# JWTs are encoded with base64url, which `base64 -d` cannot directly parse.
	@# Steps:
	@# - Split on '.' and get the middle (claims) section of the token
	@# - Replace URL-safe chars with base64 equivalents
	@# - Add padding as needed
	@# - Decode base64url to unicode
	@# - Pretty print with fromjson
	@echo "Token saved to jwt.txt. Expires at:"
	@cat jwt.txt | jq -Rr 'split(".")[1] | gsub("-"; "+") | gsub("_"; "/") | . + (["","===","==","="][length % 4]) | @base64d | fromjson | .exp | todate'
