.PHONY: db-up
db-up:
	docker compose up --detach --wait

.PHONY: db-down
db-down:
	docker compose down

.PHONY: gen
gen: db-up
	# sqlc does not remove generated files on its own when the source file is deleted.
	# Do it manually here.
	rm internal/db/*
	go generate ./...

.PHONY: build
build: gen
	go build .

.PHONY: test
test: gen
	go test ./...

.PHONY: watchgen
watchgen:
	@echo "Watching for .sql file changes. Press ctrl+c *twice* to exit."
	@while true; do \
		find . -type f -name '*.sql' | entr -d make gen ; \
		sleep 0.5 ; \
	done


	find . | grep -E ".sql$$" | entr make gen

# Run entr in a while loop because it exits when files are deleted..
.PHONY: watch
watch:
	@echo "Watching for .go file changes. Press ctrl+c *twice* to exit."
	@while true; do \
		find . -type f -name '*.go' | entr -d -r go run . ; \
		sleep 0.5 ; \
	done

.PHONY: clean
clean: db-down
	go clean
