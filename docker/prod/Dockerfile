# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.27@sha256:ec47ea3725fea0cc88f6de94919aef19f40f12fc1ec0e2d1b3426dbf15ecca04 AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/prod

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.27-alpine@sha256:eb61c3873c97416fff0116553fff65d67a3549f72e11d9758849a689cdd8b523

ENV ENVIRONMENT=prod

WORKDIR /src
COPY --from=builder /src/internal/kernel/config .
COPY --from=builder /src/internal/game/internal/data/sqlc/migrations ./migrations/game
COPY --from=builder /src/internal/lobby/internal/data/sqlc/migrations ./migrations/lobby
COPY --from=builder /src/map.json .
COPY --from=builder /src/risk-it-server .

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
