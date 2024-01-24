# GO Risk-It

## Development

From the root of the project, run `docker compose up --build` to:

1. Spin up a Postgres DB
2. Run the necessary migrations
3. Generate `sqlc` Go code
4. Build and run the `go-risk-it` application

If you want to just generate the `sqlc` code, run:

```bash
docker run --rm -v $(pwd):/src -w /src sqlc/sqlc generate
```

## Pre-commit
Install:
```bash
go install golang.org/x/tools/cmd/goimports@latest
go install mvdan.cc/gofumpt@latest
Install pre-commit hooks with `pre-commit install`

