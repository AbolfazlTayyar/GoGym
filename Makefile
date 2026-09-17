BINARY := bin/api
MIGRATIONS_DIR := migrations
SWAG_VERSION := v1.16.6

.PHONY: build run debug test test-unit lint migrate-up migrate-down swagger

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

# swagger: (re)generates docs/swagger from swag annotations on handlers
# (cmd/api/main.go for general API info, individual handlers for endpoints).
# Rerun this after adding or changing annotated handlers — the generated
# files are committed, so a stale run means stale docs shipped in the UI.
swagger:
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/api/main.go -o docs/swagger --parseDependency
