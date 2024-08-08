# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.22@sha256:2bd56f00ff47baf33e64eae7996b65846c7cb5e0a46e0a882ef179fd89654afa AS builder

WORKDIR /src
COPY go.mod go.sum ./

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

COPY . .

RUN go build -o risk-it-server ./cmd/risk-it-server/component-test

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.22-alpine@sha256:1a478681b671001b7f029f94b5016aed984a23ad99c707f6a0ab6563860ae2f3

WORKDIR /src
COPY --from=builder /src/component-test/.env .
COPY --from=builder /src/internal/config .
COPY --from=builder /src/internal/data/sqlc/migrations ./migrations
COPY --from=builder /src/map.json .
COPY --from=builder /src/risk-it-server .

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
