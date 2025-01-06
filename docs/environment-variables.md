# Environment Variables Documentation

This document describes all environment variables required by the Fowergram API.

## Database Configuration

| Variable | Description | Required | Default | Example |
|----------|-------------|----------|---------|---------|
| DB_HOST | Database host | Yes | - | cloudsql-proxy |
| DB_USER | Database username | Yes | - | postgres |
| DB_PASSWORD | Database password | Yes | - | mysecretpassword |
| DB_NAME | Database name | Yes | - | fowergram |
| DB_PORT | Database port | Yes | 5432 | 5432 |

## Redis Configuration

| Variable | Description | Required | Default | Example |
|----------|-------------|----------|---------|---------|
| REDIS_HOST | Redis host | Yes | - | redis-service |
| REDIS_PORT | Redis port | Yes | 6379 | 6379 |
| REDIS_PASSWORD | Redis password | No | "" | myredispassword |

## Application Configuration

| Variable | Description | Required | Default | Example |
|----------|-------------|----------|---------|---------|
| PORT | HTTP server port | Yes | 8080 | 8080 |
| GIN_MODE | Gin framework mode | Yes | release | release |

## Health Check Endpoints

The application provides two health check endpoints:

1. `/ping`
   - Simple health check
   - Returns 200 OK with timestamp
   - Used by Kubernetes probes

2. `/health`
   - Detailed health check
   - Returns status of all services
   - Used for monitoring

## Example Configuration

```yaml
# Kubernetes ConfigMap example
apiVersion: v1
kind: ConfigMap
metadata:
  name: fowergram-config
data:
  PORT: "8080"
  DB_HOST: "cloudsql-proxy"
  DB_PORT: "5432"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  GIN_MODE: "release"

# Kubernetes Secret example
apiVersion: v1
kind: Secret
metadata:
  name: fowergram-secrets
type: Opaque
data:
  db-user: base64-encoded-value
  db-password: base64-encoded-value
  db-name: base64-encoded-value
```

## Development Setup

For local development, you can use a `.env` file:

```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=fowergram_dev
DB_PORT=5432
REDIS_HOST=localhost
REDIS_PORT=6379
GIN_MODE=debug
PORT=8080
```

## Notes

- All sensitive information should be stored in Kubernetes Secrets
- Use appropriate values for different environments (development, staging, production)
- Keep database and Redis credentials secure
- Monitor health check endpoints for service status 