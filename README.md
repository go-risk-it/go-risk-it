# GO Risk-It

A backend application for the Risk-It game.

## Running the application

```bash
make run
```

## Development

### Dev environment setup

Install linters, formatters, and pre-commit hooks with:

```bash
make install
```

### Running the tests

```bash
make test
```

### Code generation

Generate SQLC code for interacting with the database:

```bash
make sqlc
```

Generate mocks for testing:

```bash
make mock
```