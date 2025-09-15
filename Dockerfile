# Multi-stage build for Burndler
# Stage 1: Frontend Builder
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy package files
COPY frontend/package*.json ./

# Install dependencies (including dev dependencies needed for build)
RUN npm ci

# Copy source code
COPY frontend/ ./

# Build frontend
RUN npm run build

# Stage 2: Backend Builder
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go module files
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download

# Copy backend source
WORKDIR /app
COPY backend/ ./backend/

# Copy frontend build from previous stage
COPY --from=frontend-builder /app/frontend/dist ./backend/internal/static/dist/

# Build the binary
WORKDIR /app/backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /app/burndler \
    cmd/api/main.go

# Stage 3: Production Image
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget && \
    update-ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S burndler && \
    adduser -u 1001 -S burndler -G burndler

# Create directory for the app
WORKDIR /app

# Copy binary from builder stage
COPY --from=backend-builder /app/burndler ./burndler

# Change ownership to non-root user
RUN chown burndler:burndler /app/burndler && \
    chmod +x /app/burndler

# Switch to non-root user
USER burndler

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Set entrypoint
ENTRYPOINT ["./burndler"]