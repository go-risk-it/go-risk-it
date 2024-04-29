# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.22 AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.22-alpine

WORKDIR /src
COPY --from=builder /src/risk-it-server .
COPY --from=builder /src/map.json .
COPY --from=builder /src/internal/config .
COPY --from=builder /src/.env .
COPY --from=builder /src/internal/data/sqlc/migrations ./migrations

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
