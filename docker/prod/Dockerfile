# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.27@sha256:3680233e3204827fbdc66088528ae6d4b3d034f51d03a99d454f6de034888244 AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/prod

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.27-alpine@sha256:4cb7ac979db5fcc41cae44b2227ba5ab8a51e8807f40d9ba4dee20a0ad960b5b

ENV ENVIRONMENT=prod

WORKDIR /src
COPY --from=builder /src/internal/kernel/config .
COPY --from=builder /src/internal/game/internal/data/sqlc/migrations ./migrations/game
COPY --from=builder /src/internal/lobby/internal/data/sqlc/migrations ./migrations/lobby
COPY --from=builder /src/map.json .
COPY --from=builder /src/risk-it-server .

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
