# === Stage 1: Build Stage ===
# Use a Golang image as the base image for the build stage
FROM golang:1.21 AS builder

WORKDIR /src
COPY ./src .
RUN go mod download go-risk-it

# Ensure binary is statically compiled
ENV CGO_ENABLED=0
RUN go build -o main .

# === Stage 2: Runtime Stage ===
# Use a lightweight Golang image as the base image for the runtime stage
FROM golang:1.21-alpine

WORKDIR /src
COPY --from=builder /src/main .

# Command to run the srclication
ENTRYPOINT ["./main"]
