# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go mod files first for caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Add ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /build/server .

EXPOSE 8080

CMD ["./server"]
