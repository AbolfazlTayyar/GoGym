BINARY := bin/api

.PHONY: build run test lint migrate-up migrate-down

build:
	go build -o $(BINARY) ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./...

lint:
	golangci-lint run ./...

migrate-up:
	@echo "migrate-up: not implemented yet"

migrate-down:
	@echo "migrate-down: not implemented yet"
