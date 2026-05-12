# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.26.2@sha256:b54cbf583d390341599d7bcbc062425c081105cc5ef6d170ced98ef9d047c716 AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/component-test

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.26.2-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1

ENV ENVIRONMENT=component-test

WORKDIR /src
COPY --from=builder /src/component-test/.env .
COPY --from=builder /src/internal/kernel/config .
COPY --from=builder /src/internal/game/internal/data/sqlc/migrations ./migrations/game
COPY --from=builder /src/internal/lobby/internal/data/sqlc/migrations ./migrations/lobby
COPY --from=builder /src/map.json .
COPY --from=builder /src/risk-it-server .

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
