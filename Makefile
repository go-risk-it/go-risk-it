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

mock: ## Generate mocks
	@echo "Generating mocks..."
	@rm -rf mocks internal/game/testmocks internal/lobby/testmocks
	@mockery
	@# Relocate game/internal mocks inside the game module (Go internal/ visibility)
	@if [ -d mocks/internal_/game/internal_ ]; then \
		mkdir -p internal/game/testmocks && \
		cp -r mocks/internal_/game/internal_/* internal/game/testmocks/ && \
		rm -rf mocks/internal_/game/internal_; \
	fi
	@# Relocate lobby/internal mocks inside the lobby module (Go internal/ visibility)
	@if [ -d mocks/internal_/lobby/internal_ ]; then \
		mkdir -p internal/lobby/testmocks && \
		cp -r mocks/internal_/lobby/internal_/* internal/lobby/testmocks/ && \
		rm -rf mocks/internal_/lobby/internal_; \
	fi

destroy:
	@echo "Destroying existing environment..."
	@docker compose --project-name go-risk-it down --remove-orphans

run: destroy ## Run the application
	@echo "Spinning up new environment..."
	@docker compose up --build --detach

cp: ## Run component tests
	@echo "Running component tests..."
	@cd component-test; poetry run behave

# Grafana dashboard generation (requires: brew install go-jsonnet)
dashboards: ## Generate Grafana dashboard JSON from Jsonnet sources
	@echo "Generating dashboards..."
	@for f in grafana/dashboards/*.jsonnet; do \
		out="$${f%.jsonnet}.json"; \
		jsonnet -J grafana/lib "$$f" | python3 -m json.tool > "$$out"; \
		echo "  $$f -> $$out"; \
	done

dashboards-check: ## Verify generated dashboard JSON matches committed files
	@echo "Checking dashboards are up to date..."
	@tmpdir=$$(mktemp -d); \
	for f in grafana/dashboards/*.jsonnet; do \
		out="$${f%.jsonnet}.json"; \
		jsonnet -J grafana/lib "$$f" | python3 -m json.tool > "$$tmpdir/$$(basename $$out)"; \
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

# Architecture documentation generation (requires: d2 CLI — brew install d2)
docs: ## Generate architecture docs (D2 diagram + Mermaid + tree + tables)
	@go run ./cmd/archdiagram/

docs-check: ## Verify all generated architecture docs are up to date
	@echo "Checking architecture docs are up to date..."
	@tmpdir=$$(mktemp -d); \
	cp docs/architecture-diagram.d2 "$$tmpdir/architecture-diagram.d2" 2>/dev/null || true; \
	cp docs/architecture.md "$$tmpdir/architecture.md"; \
	cp docs/architecture-components.md "$$tmpdir/architecture-components.md"; \
	cp docs/doc-go-spec.md "$$tmpdir/doc-go-spec.md"; \
	cp README.md "$$tmpdir/README.md"; \
	go run ./cmd/archdiagram/ 2>/dev/null; \
	fail=0; \
	for f in docs/architecture-diagram.d2 docs/architecture.md docs/architecture-components.md docs/doc-go-spec.md README.md; do \
		base=$$(basename "$$f"); \
		if ! diff -q "$$f" "$$tmpdir/$$base" > /dev/null 2>&1; then \
			echo "FAIL: $$f is out of date."; \
			fail=1; \
		fi; \
	done; \
	if [ "$$fail" = "1" ]; then \
		echo "Run 'make docs' to regenerate."; \
		for f in docs/architecture-diagram.d2 docs/architecture.md docs/architecture-components.md docs/doc-go-spec.md README.md; do \
			base=$$(basename "$$f"); \
			cp "$$tmpdir/$$base" "$$f" 2>/dev/null || true; \
		done; \
		rm -rf "$$tmpdir"; \
		exit 1; \
	fi; \
	rm -rf "$$tmpdir"; \
	echo "OK: all architecture docs are up to date."

# Backward compatibility aliases
diagrams: docs ## Alias for 'make docs'

diagrams-check: docs-check ## Alias for 'make docs-check'
