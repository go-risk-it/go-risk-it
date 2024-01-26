install:
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/segmentio/golines@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2
	@echo "Make sure to have pre-commit installed. See https://pre-commit.com/#install"

pre-commit-check:
	pre-commit run --all-files

test:
	go test ./...

run:
	docker compose up --build --detach
