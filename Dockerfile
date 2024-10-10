# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.22@sha256:b274ff14d8eb9309b61b1a45333bf0559a554ebcf6732fa2012dbed9b01ea56f AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/component-test

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.22-alpine@sha256:f56a8a4a1aea41bc4694728b69c219af1523aea15690cbbed82dc9bac81e6603

WORKDIR /src
COPY --from=builder /src/component-test/.env .
COPY --from=builder /src/internal/config .
COPY --from=builder /src/internal/data/sqlc/migrations ./migrations
COPY --from=builder /src/map.json .
COPY --from=builder /src/risk-it-server .

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
