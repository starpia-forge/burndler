# Backend Configuration

Required environment variables for Burndler backend service.

## Database Configuration

```bash
# PostgreSQL connection
DB_HOST=localhost
DB_PORT=5432
DB_NAME=burndler
DB_USER=burndler
DB_PASSWORD=<secure-password>
DB_SSL_MODE=disable  # Options: disable, require, verify-ca, verify-full

# Connection pool
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_LIFETIME=300s
```

## Storage Configuration

### Storage Mode Selection
```bash
# Storage backend selection
STORAGE_MODE=s3      # Options: s3, local
```

### S3 Storage (Production/Default)
```bash
# S3-compatible storage
S3_ENDPOINT=https://s3.amazonaws.com
S3_REGION=us-east-1
S3_BUCKET=burndler-artifacts
S3_ACCESS_KEY_ID=<access-key>
S3_SECRET_ACCESS_KEY=<secret-key>
S3_USE_SSL=true
S3_PATH_PREFIX=packages/  # Optional prefix for all objects
```

### Local FS Storage (Development/Offline)
```bash
# Local filesystem storage
LOCAL_STORAGE_PATH=/var/lib/burndler/storage
LOCAL_STORAGE_MAX_SIZE=10GB  # Optional size limit
```

## JWT Authentication

```bash
# JWT configuration
JWT_SECRET=<base64-encoded-secret>  # Generate: openssl rand -base64 32
JWT_ISSUER=burndler
JWT_AUDIENCE=burndler-api
JWT_EXPIRATION=24h    # Token lifetime
JWT_REFRESH_EXPIRATION=168h  # Refresh token lifetime

# RBAC roles (hardcoded but documented)
# - Developer: Read/Write access to all operations
# - Engineer: Read-only access, cannot create packages
```

## Server Configuration

```bash
# HTTP server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_MAX_REQUEST_SIZE=100MB

# CORS settings
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.burndler.example
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Content-Type,Authorization
```

## Docker Registry

```bash
# For pulling/resolving images
DOCKER_REGISTRY_URL=https://registry-1.docker.io
DOCKER_REGISTRY_USERNAME=  # Optional, for private registries
DOCKER_REGISTRY_PASSWORD=  # Optional, for private registries
DOCKER_REGISTRY_CACHE_DIR=/tmp/docker-cache
```

## Logging

```bash
# Logging configuration
LOG_LEVEL=info       # Options: debug, info, warn, error
LOG_FORMAT=json      # Options: json, text
LOG_OUTPUT=stdout    # Options: stdout, file
LOG_FILE_PATH=/var/log/burndler/api.log  # If LOG_OUTPUT=file
```

## Build Worker

```bash
# Async build processing
BUILD_WORKER_COUNT=4
BUILD_TIMEOUT=30m
BUILD_TEMP_DIR=/tmp/burndler-builds
BUILD_RETENTION_DAYS=7  # Keep completed builds for N days
```

## Monitoring

```bash
# Metrics and health checks
METRICS_ENABLED=true
METRICS_PORT=9090
HEALTH_CHECK_INTERVAL=30s
```

## Example .env.example

See `.env.example` in the repository root for a complete template.

## Environment-Specific Configurations

### Development
- Use `STORAGE_MODE=local` for offline development
- Set `DB_SSL_MODE=disable` for local PostgreSQL
- Use `LOG_LEVEL=debug` for verbose logging

### Production
- Use `STORAGE_MODE=s3` for scalable artifact storage
- Set `DB_SSL_MODE=verify-full` for secure database connections
- Use `LOG_LEVEL=info` and `LOG_FORMAT=json` for structured logging
- Enable metrics with `METRICS_ENABLED=true`

## Security Notes

1. **Never commit secrets** - Use `.env` files or container secrets
2. **JWT_SECRET must be unique** per environment
3. **Database passwords** should be rotated regularly
4. **S3 credentials** should use IAM roles when possible
5. **Use SSL/TLS** for all production connections