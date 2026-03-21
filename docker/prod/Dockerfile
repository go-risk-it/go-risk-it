# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.23@sha256:60deed95d3888cc5e4d9ff8a10c54e5edc008c6ae3fba6187be6fb592e19e8c0 AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/prod

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.23-alpine@sha256:47d337594bd9e667d35514b241569f95fb6d95727c24b19468813d596d5ae596

ENV ENVIRONMENT=prod

WORKDIR /src
COPY --from=builder /src/internal/config .
COPY --from=builder /src/internal/data/game/sqlc/migrations ./migrations/game
COPY --from=builder /src/internal/data/lobby/sqlc/migrations ./migrations/lobby
COPY --from=builder /src/map.json .
COPY --from=builder /src/risk-it-server .

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
