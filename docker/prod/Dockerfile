# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.27@sha256:820153096fda4de6ac1b7194d12e416b7dd8f4e9a33d17e0345b111b945162d8 AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/prod

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.27-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d

ENV ENVIRONMENT=prod

WORKDIR /src
COPY --from=builder /src/internal/kernel/config .
COPY --from=builder /src/internal/game/internal/data/sqlc/migrations ./migrations/game
COPY --from=builder /src/internal/lobby/internal/data/sqlc/migrations ./migrations/lobby
COPY --from=builder /src/map.json .
COPY --from=builder /src/risk-it-server .

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
