# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.23 AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/prod

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.23-alpine

ENV ENVIRONMENT=prod

WORKDIR /src
COPY --from=builder /src/internal/config .
COPY --from=builder /src/internal/data/sqlc/migrations ./migrations
COPY --from=builder /src/map.json .
COPY --from=builder /src/risk-it-server .

RUN printf "%s\n" \
    "DATABASE_HOST=${DATABASE_HOST}" \
    "DATABASE_NAME=${DATABASE_NAME}" \
    "DATABASE_PASSWORD=${DATABASE_PASSWORD}" \
    "DATABASE_PORT=${DATABASE_PORT}" \
    "DATABASE_USER=${DATABASE_USER}" \
    "JWT_SECRET=${JWT_SECRET}" > .env

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
