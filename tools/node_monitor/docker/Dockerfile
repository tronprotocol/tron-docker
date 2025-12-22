# Multi-stage build for tron node monitor

# Builder stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Enable Go modules and download dependencies first (better layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o node_monitor .

# Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/node_monitor /app/node_monitor

# Default config path (can be overridden by volume)
COPY node_monitor.yml /app/node_monitor.yml

EXPOSE 9098

ENTRYPOINT ["/app/node_monitor"]
CMD ["-config", "/app/node_monitor.yml"]
