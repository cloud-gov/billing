build:
	go build .

test:
	go test ./...

watch:
	find . | grep ".go" | entr -r go run .
