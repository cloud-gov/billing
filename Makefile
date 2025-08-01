.PHONY: db-up
db-up:
	docker compose up --detach --wait

.PHONY: db-down
db-down:
	docker compose down

.PHONY: gen
gen: db-up
	@# sqlc does not remove generated files on its own when the source file is deleted; remove them manually here. (|| true reports success even when no files were found.)
	@rm internal/db/* || true
	@date
	go generate ./...

.PHONY: build
build: gen
	go build .

.PHONY: test
test: gen
	go test ./...

.PHONY: debug
debug: gen
	@set -a; source docker.env; set +a; dlv debug

.PHONY: watchgen
watchgen:
	@echo "Watching for .sql file changes. Press ctrl+c *twice* to exit, or once to rebuild."
	@while true; do \
		find . -type f -name '*.sql' | entr -d make gen ; \
		sleep 0.5 ; \
	done

# Run entr in a while loop because it exits when files are deleted..
.PHONY: watch
watch:
	@echo "Watching for .go file changes. Press ctrl+c *twice* to exit, or once to rebuild."
	@while true; do \
		set -a; source docker.env; set +a; find . -type f -name '*.go' | entr -d -r go run . ; \
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
	@set -a; source docker.env; PGDATABASE=postgres; set +a; go tool tern migrate --config sql/init/tern.conf --migrations sql/init

.PHONY: db-drop
db-drop:
	@# Drop the database.
	@set -a; source docker.env; PGDATABASE=postgres; set +a; go tool tern migrate --config sql/init/tern.conf --migrations sql/init --destination 0

.PHONY: migrate
migrate: db-init
	@# Migrate to the latest migration.
	@set -a; source docker.env; set +a; go tool tern migrate --config sql/migrations/tern.conf --migrations sql/migrations
	@# Migrate River schema to latest.
	@set -a; source docker.env; set +a; go tool river migrate-up

.PHONY: db-reset
db-reset: db-drop db-init migrate
	@echo "Database reset. Restart app to reconnect."

.PHONY: psql
psql:
	@set -a; source docker.env; set +a; psql
