db-up:
	docker compose up --detach --wait

db-down:
	docker compose down

gen: db-up
	# sqlc does not remove generated files on its own when the source file is deleted.
	# Do it manually here.
	rm internal/db/*
	go generate ./...

build: gen
	go build .

test: gen
	go test ./...

watchgen:
	find . | grep -E ".sql$$" | entr make gen

watch:
	find . | grep -E ".go$$" | entr -r go run .

clean: db-down
	go clean
