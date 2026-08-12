.PHONY: run build test lint proto docker-up docker-down migrate

BINARY   := finance-api
CMD_PATH := ./cmd/api

## run: start the API server locally (requires Postgres + Redis running)
run:
	go run $(CMD_PATH)

## build: compile the binary
build:
	go build -o $(BINARY) $(CMD_PATH)

## test: run all tests
test:
	go test ./... -v -race

## lint: run static analysis
lint:
	golangci-lint run ./...

## proto: generate Go code from proto definitions (requires protoc + plugins)
proto:
	protoc \
		--go_out=./proto \
		--go_opt=paths=source_relative \
		--go-grpc_out=./proto \
		--go-grpc_opt=paths=source_relative \
		proto/finance.proto

## docker-up: start all infrastructure services
docker-up:
	docker compose up -d --build

## docker-down: stop all services and remove volumes
docker-down:
	docker compose down -v

## migrate: apply SQL migrations manually (requires psql)
migrate:
	psql "$$POSTGRES_DSN" -f migrations/001_initial.sql

## tidy: tidy Go modules
tidy:
	go mod tidy
