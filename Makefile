DATABASE_URL ?= postgres://lending:lending@localhost:5432/lending?sslmode=disable
HTTP_ADDR    ?= :8080

# Pinned linter versions — everyone gets identical results and the bar cannot
# drift. Bump these deliberately.
GOLANGCI_LINT_VERSION ?= v2.13.0
GO_ARCH_LINT_VERSION  ?= v1.17.0

# Where `go install` drops tool binaries: it honors $GOBIN when set, and falls
# back to $GOPATH/bin otherwise.
GOBIN_DIR := $(shell go env GOBIN)
ifeq ($(GOBIN_DIR),)
GOBIN_DIR := $(shell go env GOPATH)/bin
endif

.PHONY: help run build test test-integration vet fmt golangci arch-lint lint tidy tidy-check hooks-install pre-commit db-up db-down migrate-up migrate-down

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

fmt: ## Auto-format Go code with gofmt (rewrites files in place)
	gofmt -w .

golangci: ## Run golangci-lint only (installs the pinned version)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	$(GOBIN_DIR)/golangci-lint run ./...

arch-lint: ## Run go-arch-lint only (installs the pinned version)
	go install github.com/fe3dback/go-arch-lint@$(GO_ARCH_LINT_VERSION)
	$(GOBIN_DIR)/go-arch-lint check

lint: golangci arch-lint ## Run all linters (golangci-lint + go-arch-lint)

tidy-check: ## Fail if go.mod / go.sum are not tidy
	go mod tidy
	git diff --exit-code -- go.mod go.sum

hooks-install: ## Install the pre-commit git hooks
	pre-commit install

pre-commit: ## Run the whole pre-commit bar across all files
	pre-commit run --all-files

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
