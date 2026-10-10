# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.27.2@sha256:5bc7f572bbaa98885a3a1fd9c0aa76b59e3e14e8628bfc316bbfd0c701e4818c AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/component-test

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d

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
