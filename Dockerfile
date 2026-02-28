# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /build

# Copy package files
COPY ui/package*.json ./
RUN npm ci

# Copy UI source
COPY ui/ ./

# Build frontend
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.23-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend from previous stage
COPY --from=frontend-builder /build/out ./ui/out

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o mission-control ./cmd/server

# Stage 3: Final minimal image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata sqlite-libs

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /build/mission-control .

# Create data directory for SQLite database
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Set environment variables with sensible defaults
ENV HOST=0.0.0.0 \
    PORT=8080 \
    DATABASE_PATH=/app/data/mission-control.db

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["/app/mission-control"]
