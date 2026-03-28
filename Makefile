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
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.11.3
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

# Grafana dashboard generation (requires: brew install go-jsonnet jsonnet-bundler)
dashboards: ## Generate Grafana dashboard JSON from Jsonnet sources
	@echo "Generating dashboards..."
	@for f in grafana/dashboards/*.jsonnet; do \
		out="$${f%.jsonnet}.json"; \
		jsonnet -J grafana/vendor -J grafana/lib "$$f" | python3 -m json.tool > "$$out"; \
		echo "  $$f -> $$out"; \
	done

dashboards-check: ## Verify generated dashboard JSON matches committed files
	@echo "Checking dashboards are up to date..."
	@tmpdir=$$(mktemp -d); \
	for f in grafana/dashboards/*.jsonnet; do \
		out="$${f%.jsonnet}.json"; \
		jsonnet -J grafana/vendor -J grafana/lib "$$f" | python3 -m json.tool > "$$tmpdir/$$(basename $$out)"; \
		if ! diff -q "$$out" "$$tmpdir/$$(basename $$out)" > /dev/null 2>&1; then \
			echo "FAIL: $$out is out of date. Run 'make dashboards' to regenerate."; \
			rm -rf "$$tmpdir"; \
			exit 1; \
		fi; \
	done; \
	rm -rf "$$tmpdir"; \
	echo "OK: all dashboards are up to date."

# Architecture documentation
new-package: ## Scaffold a new package with doc.go (usage: make new-package PKG=internal/logic/game/foo LAYER=Logic)
	@scripts/new-package.sh $(PKG) $(LAYER)

# Architecture diagram generation (requires: d2 CLI — brew install d2)
diagrams: ## Generate architecture diagram from internal packages
	@go run ./cmd/archdiagram/

diagrams-check: ## Verify architecture diagram is up to date
	@echo "Checking architecture diagram is up to date..."
	@tmpdir=$$(mktemp -d); \
	go run ./cmd/archdiagram/ -output "$$tmpdir" 2>/dev/null; \
	failed=0; \
	for f in architecture-diagram.d2 architecture-diagram.svg; do \
		if ! diff -q "docs/$$f" "$$tmpdir/$$f" > /dev/null 2>&1; then \
			echo "FAIL: docs/$$f is out of date. Run 'make diagrams' to regenerate."; \
			failed=1; \
		fi; \
	done; \
	rm -rf "$$tmpdir"; \
	if [ "$$failed" = "1" ]; then exit 1; fi; \
	echo "OK: architecture diagram is up to date."
