db-up:
	docker compose up --detach --wait

db-down:
	docker compose down

gen: db-up
	go generate ./...

build: gen
	go build .

test: gen
	go test ./...

watchgen:
	find . | grep ".sql" | entr make gen

watch:
	find . | grep ".go" | entr -r go run .

clean: db-down
	go clean
