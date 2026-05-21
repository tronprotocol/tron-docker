# Build stage
FROM golang:1.25-alpine@sha256:8d22e29d960bc50cd025d93d5b7c7d220b1ee9aa7a239b3c8f55a57e987e8d45 AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o slack_sr_monitor main.go

# Final stage
FROM alpine:3.20@sha256:d9e853e87e55526f6b2917df91a2115c36dd7c696a35be12163d44e6e2a4b6bc

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S monitor && adduser -S -G monitor monitor

WORKDIR /home/monitor

# Create logs directory owned by the non-root user
RUN mkdir -p logs && chown -R monitor:monitor /home/monitor

# Copy the binary from builder
COPY --from=builder --chown=monitor:monitor /app/slack_sr_monitor .

USER monitor

# Command to run
CMD ["./slack_sr_monitor"]
