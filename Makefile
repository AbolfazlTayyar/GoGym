BINARY := bin/api
MIGRATIONS_DIR := migrations

.PHONY: build run test test-unit lint migrate-up migrate-down

build:
	go build -o $(BINARY) ./cmd/api

run:
	go run ./cmd/api

# test-unit: short, no containers — safe for tight loops.
test-unit:
	go test -short ./...

# test: full suite, including testcontainers-backed integration tests.
test:
	go test ./...

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$$(go run ./cmd/dsn)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$$(go run ./cmd/dsn)" down 1
