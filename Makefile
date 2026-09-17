BINARY := bin/api
MIGRATIONS_DIR := migrations

.PHONY: build run debug test test-unit lint migrate-up migrate-down

build:
	go build -o $(BINARY) ./cmd/api

run:
	go run ./cmd/api

# debug: runs the api container under headless Delve (port 2345) instead of
# the production build, for the "Debug API (Docker)" launch config in
# .vscode/launch.json — see docker-compose.debug.yml for details.
debug:
	docker compose -f docker-compose.yml -f docker-compose.debug.yml up --build

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
