# Multi-stage build for endless Go application
FROM golang:1.23.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Download source from main branch
RUN git clone https://github.com/abigpotostew/endless.git .

# Download Go dependencies
RUN go mod download
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o endless main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies for SQLite
RUN apk --no-cache add ca-certificates sqlite

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/endless .

# Create data directory for SQLite database
RUN mkdir -p /data && \
    chown -R appuser:appgroup /data /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Set default environment variables
ENV PORT=8080
ENV SQLITE_DB_DIR=/data

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./endless"] 