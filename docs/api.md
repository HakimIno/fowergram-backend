# API Documentation

This document describes the API endpoints available in the Fowergram API.

## Base URL

The base URL for all API endpoints is:

```
https://api.fowergram.com
```

For local development:
```
http://localhost:8080
```

## Health Check Endpoints

### Ping Check

Simple health check endpoint used by Kubernetes probes.

```http
GET /ping
```

#### Response

```json
{
    "status": "ok",
    "time": "2024-01-20T15:04:05Z"
}
```

| Status Code | Description |
|------------|-------------|
| 200 | Service is healthy |

### Detailed Health Check

Detailed health check endpoint that shows the status of all services.

```http
GET /health
```

#### Response

```json
{
    "status": "ok",
    "time": "2024-01-20T15:04:05Z",
    "services": {
        "api": "up",
        "db": "up",
        "redis": "up"
    }
}
```

| Status Code | Description |
|------------|-------------|
| 200 | All services are healthy |
| 503 | One or more services are down |

## Authentication

### Login

Authenticate a user and receive an access token.

```http
POST /auth/login
```

#### Request Body

```json
{
    "email": "user@example.com",
    "password": "your-password"
}
```

#### Response

```json
{
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
        "id": "user-id",
        "email": "user@example.com",
        "name": "User Name"
    }
}
```

| Status Code | Description |
|------------|-------------|
| 200 | Successfully authenticated |
| 401 | Invalid credentials |
| 429 | Too many login attempts |

### Register

Register a new user account.

```http
POST /auth/register
```

#### Request Body

```json
{
    "email": "user@example.com",
    "password": "your-password",
    "name": "User Name"
}
```

#### Response

```json
{
    "id": "user-id",
    "email": "user@example.com",
    "name": "User Name"
}
```

| Status Code | Description |
|------------|-------------|
| 201 | Successfully registered |
| 400 | Invalid input |
| 409 | Email already exists |

## Error Responses

All endpoints may return the following error responses:

```json
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable error message"
    }
}
```

### Common Error Codes

| Status Code | Error Code | Description |
|------------|------------|-------------|
| 400 | INVALID_INPUT | The request body is invalid |
| 401 | UNAUTHORIZED | Authentication is required |
| 403 | FORBIDDEN | Not enough permissions |
| 404 | NOT_FOUND | Resource not found |
| 429 | RATE_LIMITED | Too many requests |
| 500 | INTERNAL_ERROR | Internal server error |

## Rate Limiting

The API implements rate limiting to protect against abuse. Rate limits are applied per IP address and/or API token.

| Endpoint | Rate Limit |
|----------|------------|
| /auth/* | 5 requests per minute |
| Other endpoints | 60 requests per minute |

When rate limited, the API will return:
- Status code: 429
- Headers:
  - `X-RateLimit-Limit`: Total requests allowed in the time window
  - `X-RateLimit-Remaining`: Remaining requests in the current window
  - `X-RateLimit-Reset`: Time when the rate limit resets (Unix timestamp)

## Best Practices

1. Always include authentication token in the Authorization header
2. Handle rate limiting by respecting the rate limit headers
3. Implement proper error handling for all possible status codes
4. Use HTTPS in production
5. Keep authentication tokens secure
6. Monitor API health endpoints

## Development Tools

For local development and testing:

1. Use the provided health check endpoints to verify service status
2. Monitor the detailed health check for service dependencies
3. Check rate limiting headers during development
4. Use appropriate environment variables as documented in `environment-variables.md` 