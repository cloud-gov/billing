#!/bin/bash
set -exo pipefail

go mod download

test -z "$(gofmt -l .)"

go vet ./...

go build .

go test -v ./... -skip TestDB
