DATABASE_URL ?= postgres://lending:lending@localhost:5432/lending?sslmode=disable
HTTP_ADDR    ?= :8080

.PHONY: help run build test test-integration vet tidy db-up db-down migrate-up migrate-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

run: ## Run the server (auto-applies migrations on boot)
	DATABASE_URL="$(DATABASE_URL)" HTTP_ADDR="$(HTTP_ADDR)" go run ./cmd/server

build: ## Build the server binary into ./bin
	go build -o bin/server ./cmd/server

test: ## Run the fast suite (unit + httptest)
	go test ./...

test-integration: ## Run the tagged pgx integration tests (needs Docker)
	go test -tags=integration ./...

vet: ## Static analysis
	go vet ./...
	go vet -tags=integration ./...

tidy: ## Tidy go.mod / go.sum
	go mod tidy

db-up: ## Start the local Postgres via docker compose
	docker compose up -d postgres

db-down: ## Stop and remove the local Postgres (and its volume)
	docker compose down -v

migrate-up: ## Apply migrations with the golang-migrate CLI
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Roll back the last migration with the golang-migrate CLI
	migrate -path migrations -database "$(DATABASE_URL)" down 1
