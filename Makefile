help:
	@echo "Commands:"
	@echo "    install: install dependencies and tools for development"
	@echo "    run: spin up postgres DB and server"
	@echo "    pre-commit-check"
	@echo "    test"

install:
	@echo "Installing dependencies and tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install mvdan.cc/gofumpt@latest
	@go install github.com/segmentio/golines@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2
	@echo "Make sure to have pre-commit installed. See https://pre-commit.com/#install"

pre-commit-check:
	pre-commit run --all-files

test:
	@
	go test ./...

sqlc:
	@echo "Building..."
	@docker run --rm -v $(pwd):/src -w /src sqlc/sqlc generate

run:
	@echo "Destroying existing environment..."
	@docker compose --project-name go-risk-it down --remove-orphans
	@echo "Spinning up new environment..."
	@docker compose up --build --detach
