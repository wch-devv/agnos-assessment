# ==========================================
# Stage 1: Build Go binary
# ==========================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates for downloading dependencies
RUN apk add --no-cache git ca-certificates

# Copy go.mod and go.sum first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/server ./cmd/server

# ==========================================
# Stage 2: Minimal Runtime Image
# ==========================================
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/server /app/server

# Set ownership
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["/app/server"]
