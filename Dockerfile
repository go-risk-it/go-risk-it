# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.21 AS builder

WORKDIR /src
COPY ./ .

# Ensure binary is statically compiled
ENV CGO_ENABLED=0

RUN go mod download

RUN go build -o risk-it-server ./cmd/risk-it-server

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.21-alpine

WORKDIR /src
COPY --from=builder /src/risk-it-server .

# Command to run the executable
ENTRYPOINT ["./risk-it-server"]
