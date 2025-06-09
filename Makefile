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

watch:
	find . | grep -E "\.go|\.sql" | entr -r "make gen && go run ."

clean: db-down
	go clean
