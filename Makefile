MAKEFILE := $(lastword $(MAKEFILE_LIST))

# Default target
.DEFAULT_GOAL := help

help: ## Print this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


install: ## Install dependencies and tools
	@echo "Installing dependencies and tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/segmentio/golines@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.62.2
	@pre-commit install

pre-commit-check: ## Run pre-commit checks
	pre-commit run --all-files

test: ## Run tests
	go test ./...

sqlc: ## Generate SQLC code to interact with the database
	@echo "Building..."
	@docker compose run --rm sqlc

mock: destroy ## Generate mocks
	@echo "Building..."
	@rm -rf mocks
	@docker compose run --rm mockery

destroy:
	@echo "Destroying existing environment..."
	@docker compose --project-name go-risk-it down --remove-orphans

run: destroy ## Run the application
	@echo "Spinning up new environment..."
	@docker compose up --build --detach

cp: ## Run component tests
	@echo "Running component tests..."
	@cd component-test; poetry run behave
