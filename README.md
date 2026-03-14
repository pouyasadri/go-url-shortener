# Go URL Shortener

A production-ready URL shortener service built with **Go 1.25**, **Gin**, and **Redis**. Features clean error handling, environment-based configuration, Docker support with multi-stage builds, and comprehensive testing.

## Overview

The service provides a RESTful API to shorten long URLs and redirect users from short URLs to their original destinations. It uses Redis as the primary storage backend for mapping short URLs to original URLs with automatic TTL expiration (6 hours by default).

## Features

- ✅ **Authentication**: API key-based authentication (Bearer tokens)
- ✅ **Rate Limiting**: Per-API-key rate limiting (1000 req/day default)
- ✅ **Admin API**: Generate, list, and revoke API keys
- ✅ **Request Tracing**: Unique X-Request-ID for each request
- ✅ **Structured Logging**: JSON-formatted logs with request context
- ✅ **Panic Recovery**: Automatic recovery from panics with proper error responses
- ✅ **Health Checks**: Liveness (/health) and readiness (/ready) probes
- ✅ **RFC 7807 Errors**: Standard problem detail error responses
- ✅ Fast URL generation using SHA-256 hashing and Base58 encoding
- ✅ Deterministic short codes (same input always produces same output)
- ✅ 6-hour TTL on all stored mappings
- ✅ Environment-based configuration
- ✅ Comprehensive error handling (no panics in production)
- ✅ Input validation (URL format checking)
- ✅ Docker & docker-compose support with multi-stage builds
- ✅ Makefile with 20+ targets for easy development
- ✅ Unit tests with 100% pass rate
- ✅ Multi-platform builds (amd64, arm64)

## Architecture

### Phase 2: Analytics Pipeline

```
Client Request
       ↓
[Analytics Middleware] ← Captures request metrics (< 1ms overhead)
       ↓
[Redis Pub/Sub] → event published to analytics:events channel
       ↓
[Event Processor Worker] ← Subscribes to events
    - Buffers events (1000 items or 5 min timeout)
    - Batch writes to MongoDB
       ↓
[MongoDB Collections]
    - request_events (30-day TTL)
    - error_events (30-day TTL)
    - metrics_hourly (permanent)
    - user_analytics (permanent)
    - api_key_analytics (permanent)
    - url_analytics (permanent)
       ↓
[Metrics Aggregator] ← Hourly cron job
    - Calculates latency percentiles (p50, p95, p99)
    - Groups metrics by user/API key
    - Upserts into metrics_hourly
       ↓
[Dashboard Endpoints] ← Admin API
    - Checks Redis cache (60 min TTL)
    - Falls back to MongoDB if cache miss
    - Returns JSON with CachedAt timestamp
```

### Services

- **Main App**: HTTP API server on port 8080
- **Analytics Worker**: Background service for event processing and aggregation
- **Redis**: Event streaming and dashboard response caching
- **MongoDB**: Persistent analytics storage with 30-day raw event retention

All services are orchestrated via `docker-compose.yml` for easy deployment.

## Project Structure

```
go-url-shortener/
├── main.go                          # Application entry point
├── handler/
│   ├── handlers.go                  # HTTP request handlers
│   ├── admin.go                     # Admin API key management endpoints
│   ├── admin_dashboard.go           # Admin dashboard endpoints (6 endpoints)
│   ├── admin_dashboard_test.go      # Dashboard endpoint tests
│   └── health.go                    # Health check endpoints
├── middleware/
│   ├── request_id.go                # Request ID generation for tracing
│   ├── logger.go                    # Structured JSON logging
│   ├── recovery.go                  # Panic recovery middleware
│   ├── auth.go                      # API key authentication
│   ├── ratelimit.go                 # Token bucket rate limiting
│   ├── error_handler.go             # RFC 7807 error responses
│   ├── analytics.go                 # Analytics event capture & publishing
│   └── analytics_test.go            # Analytics middleware tests
├── shortener/
│   ├── shorturl_generator.go        # Short URL generation logic
│   └── shorturl_generator_test.go   # Unit tests
├── store/
│   ├── store_service.go             # Redis storage layer
│   ├── store_service_test.go        # Integration tests
│   ├── api_keys.go                  # API key management in Redis
│   └── api_keys_test.go             # API key tests
├── config/
│   ├── security.go                  # Security configuration
│   ├── security_test.go             # Configuration tests
│   └── analytics.go                 # Analytics configuration
├── models/
│   ├── api_key.go                   # API key data models
│   └── analytics.go                 # Analytics data models
├── db/
│   ├── mongo.go                     # MongoDB client (singleton)
│   └── indexes.go                   # MongoDB index definitions
├── analytics/
│   └── repository.go                # MongoDB data access layer
├── cache/
│   └── redis_cache.go               # Redis caching abstraction
├── cmd/analytics-worker/
│   ├── main.go                      # Worker service entry point
│   ├── event_processor.go           # Event buffering & batch writes
│   ├── metrics_aggregator.go        # Hourly metrics calculation
│   └── cleanup_job.go               # Data retention cleanup
├── docs/
│   ├── ANALYTICS_API.md             # Analytics API documentation
│   ├── MONGODB_SETUP.md             # MongoDB setup and configuration
│   └── METRICS_SCHEMA.md            # Data model and schema reference
├── Dockerfile                       # Multi-stage Docker build (main app)
├── Dockerfile.worker                # Multi-stage Docker build (worker)
├── docker-compose.yml               # Docker Compose orchestration
├── Makefile                         # Build automation
├── .env.example                     # Environment variables template
├── .env                             # Local environment (git ignored)
└── README.md                        # This file
```


## Prerequisites

### For Local Development
- **Go** 1.25 or higher
- **Redis** 7.0 or higher
- **Make** (optional, for using Makefile targets)

### For Docker
- **Docker** 20.10+ with buildx support
- **Docker Compose** 2.0+

## Installation

### Option 1: Local Development

1. Clone the repository:
```bash
git clone https://github.com/pouyasadri/go-url-shortener.git
cd go-url-shortener
```

2. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env if needed (defaults point to localhost:6379)
```

3. Download dependencies:
```bash
go mod download
```

4. Start Redis (if not already running):
```bash
# Using Docker
docker run -d -p 6379:6379 redis:7-alpine

# Or if Redis is installed locally
redis-server
```

5. Run the application:
```bash
go run main.go
# Or use the Makefile
make run
```

### Option 2: Docker & Docker Compose

1. Clone the repository:
```bash
git clone https://github.com/pouyasadri/go-url-shortener.git
cd go-url-shortener
```

2. Start the application with docker-compose:
```bash
docker-compose up -d
# Or use the Makefile
make docker-up
```

The application will be available at `http://localhost:8080`.

## Configuration

All configuration is managed through environment variables in the `.env` file.

### Core Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `REDIS_ADDR` | `localhost:6379` | Redis server address (use `redis:6379` in Docker) |
| `REDIS_PASSWORD` | `` (empty) | Redis authentication password |
| `REDIS_DB` | `0` | Redis database number (0-15) |

### Security Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `REQUIRE_HTTPS` | `false` | Enforce HTTPS for all requests |
| `RATE_LIMIT_PER_DAY` | `1000` | API requests allowed per key per day |
| `RATE_LIMIT_WINDOW_HOURS` | `24` | Rate limit reset window in hours |
| `API_KEY_PREFIX` | `sk_live_` | Prefix for generated API keys |
| `MAX_URL_LENGTH` | `2048` | Maximum allowed URL length |
| `ADMIN_API_KEY` | `` (empty) | Admin key for API key management endpoints |

Example `.env`:
```env
PORT=8080
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

REQUIRE_HTTPS=false
RATE_LIMIT_PER_DAY=1000
RATE_LIMIT_WINDOW_HOURS=24
API_KEY_PREFIX=sk_live_
MAX_URL_LENGTH=2048
ADMIN_API_KEY=dev_admin_key_12345
```

For Docker environments, `.env` variables are automatically passed to the containers via `docker-compose.yml`.

## API Usage

### Health Checks

#### Liveness Probe
**Request:**
```bash
curl http://localhost:8080/health
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "timestamp": "2026-03-14T13:20:00Z"
}
```

#### Readiness Probe
**Request:**
```bash
curl http://localhost:8080/ready
```

**Response (200 OK):**
```json
{
  "status": "ok",
  "timestamp": "2026-03-14T13:20:00Z",
  "uptime": "1h30m45s",
  "checks": {
    "redis": {
      "status": "ok"
    }
  }
}
```

### Admin API: Generate API Key

Generate a new API key for a user (requires `ADMIN_API_KEY` header).

**Request:**
```bash
curl -X POST http://localhost:8080/admin/api-keys/generate \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: dev_admin_key_12345" \
  -d '{
    "user_id": "user-123",
    "name": "Production API Key",
    "environment": "live"
  }'
```

**Response (201 Created):**
```json
{
  "id": "key_a1b2c3d4e5f6",
  "key": "sk_live_abc123xyz...",
  "user_id": "user-123",
  "name": "Production API Key",
  "environment": "live",
  "created_at": "2026-03-14T13:20:00Z"
}
```

**Note**: The `key` is only shown once. Store it securely.

### Admin API: List API Keys

**Request:**
```bash
curl -X GET "http://localhost:8080/admin/api-keys?user_id=user-123" \
  -H "X-Admin-Key: dev_admin_key_12345"
```

**Response (200 OK):**
```json
{
  "user_id": "user-123",
  "keys": [
    {
      "id": "key_a1b2c3d4e5f6",
      "user_id": "user-123",
      "name": "Production API Key",
      "status": "active",
      "environment": "live",
      "created_at": "2026-03-14T13:20:00Z"
    }
  ]
}
```

### Admin API: Revoke API Key

**Request:**
```bash
curl -X POST http://localhost:8080/admin/api-keys/revoke \
  -H "Content-Type: application/json" \
  -H "X-Admin-Key: dev_admin_key_12345" \
  -d '{
    "key_id": "key_a1b2c3d4e5f6"
  }'
```

**Response (200 OK):**
```json
{
  "message": "API key revoked successfully",
  "key_id": "key_a1b2c3d4e5f6"
}
```

### Create a Short URL

Requires authentication with a valid API key.

**Request:**
```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk_live_abc123xyz..." \
  -d '{
    "long_url": "https://www.example.com/very/long/path?param=value"
  }'
```

**Response (201 Created):**
```json
{
  "message": "Short URL created successfully",
  "short_url": "http://localhost:8080/jTa4L57P",
  "user_id": "user-123"
}
```

**Response Headers:**
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1710426000
X-Request-ID: req_a1b2c3d4e5f6
```

**Error Responses:**
- `400 Bad Request`: Invalid JSON or malformed URL
- `401 Unauthorized`: Missing or invalid API key
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Redirect to Original URL

No authentication required.

**Request:**
```bash
curl -L http://localhost:8080/jTa4L57P
```

**Response:**
- `302 Found`: Redirects to the original long URL
- `404 Not Found`: Short URL not found or expired (TTL expired)

### Error Responses

All errors follow the **RFC 7807 Problem Details** format:

**Example 401 Unauthorized:**
```json
{
  "type": "https://api.url-shortener.dev/errors/invalid_api_key",
  "title": "Invalid API Key",
  "status": 401,
  "detail": "The provided API key is invalid or revoked",
  "instance": "/api/v1/urls",
  "trace_id": "req_a1b2c3d4e5f6"
}
```

**Example 429 Too Many Requests:**
```json
{
  "type": "https://api.url-shortener.dev/errors/rate_limit_exceeded",
  "title": "Rate Limit Exceeded",
  "status": 429,
  "detail": "Rate limit exceeded. Limit: 1000 requests per day. Reset at: 2026-03-15T13:20:00Z",
  "instance": "/api/v1/urls",
  "trace_id": "req_a1b2c3d4e5f6"
}
```

## Makefile Commands

The project includes a comprehensive Makefile with 20+ targets. Use `make help` to see all available commands.

### Build Commands
```bash
make build              # Build the Go application locally
make clean              # Remove build artifacts
make all                # Clean, lint, test, and build
```

### Test Commands
```bash
make test              # Run all unit tests with coverage
make test-short        # Run tests in short mode
```

### Code Quality
```bash
make fmt               # Format code with gofmt
make vet               # Run go vet analysis
make lint              # Run all linters (fmt + vet)
```

### Docker Commands
```bash
make docker-build      # Build Docker image with buildx (supports amd64, arm64)
make docker-build-push # Build and push to Docker registry
make docker-up         # Start docker-compose stack
make docker-down       # Stop and remove containers
make docker-logs       # View logs from all services
make docker-logs-app   # View application logs only
make docker-logs-redis # View Redis logs only
make docker-restart    # Restart services
make docker-clean      # Remove containers, volumes, and images
```

### Development Commands
```bash
make dev-up           # Start development environment
make dev-down         # Stop development environment
make dev-logs         # View development logs
make run              # Build and run locally
make .env             # Create .env from .env.example
```

### Utility Commands
```bash
make info             # Display project information
make version          # Show Go version
make deps             # Download and tidy dependencies
```

## Docker Support

### Building the Docker Image

The project uses **docker buildx** for multi-platform builds supporting both `linux/amd64` and `linux/arm64`.

```bash
# Build and load image locally
make docker-build

# Build and push to Docker registry
make docker-build-push

# Or manually with docker buildx
docker buildx build --platform linux/amd64,linux/arm64 -t url-shortener:latest .
```

### Running with Docker Compose

The `docker-compose.yml` orchestrates both the application and Redis services:

```bash
# Start the stack
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the stack
docker-compose down

# Remove all data (volumes)
docker-compose down -v
```

**Services:**
- **app**: Go URL Shortener application on port 8080
- **redis**: Redis 7 Alpine on port 6379 with persistent storage

### Dockerfile Details

Uses a multi-stage build to minimize image size:
1. **Stage 1 (builder)**: Go 1.25 Alpine - compiles the application
2. **Stage 2 (runtime)**: Alpine Linux - runs the compiled binary

Benefits:
- Small final image size (~15-20 MB)
- No Go toolchain in production image
- Security: minimal dependencies
- Health checks included with wget

## Testing

Run the test suite:

```bash
# All tests with coverage
make test

# Short mode (skips integration tests)
make test-short
```

### Test Coverage

- **`shortener/shorturl_generator_test.go`**: Unit tests for SHA-256 hash generation and Base58 encoding
- **`store/store_service_test.go`**: Integration tests for Redis operations (requires Redis running)

All tests pass with the current codebase.

## Recent Updates

### Phase 2: Analytics & Observability (Mar 2026)
- ✅ **MongoDB Integration**: Production-ready MongoDB connection with pooling (10-100 connections)
- ✅ **Analytics Middleware**: Captures all request metrics with Redis pub/sub (< 1ms overhead)
- ✅ **Event Streaming**: Asynchronous event publishing to Redis channels
- ✅ **Background Worker**: Separate service for event processing and metrics aggregation
- ✅ **Event Processor**: Buffers events (1000 items or 5 min timeout) and batch writes to MongoDB
- ✅ **Metrics Aggregator**: Hourly cron job for percentile calculation and metric aggregation
- ✅ **Data Retention**: 30-day TTL on raw events with automatic cleanup
- ✅ **Indexed Storage**: Optimized MongoDB indexes for common query patterns
- ✅ **6 Admin Dashboard Endpoints**:
  - `GET /admin/dashboard/overview` - System snapshot with key metrics
  - `GET /admin/dashboard/requests` - Time-series metrics with latency percentiles
  - `GET /admin/dashboard/users` - User engagement statistics (sortable, paginated)
  - `GET /admin/dashboard/api-keys` - API key usage analytics
  - `GET /admin/dashboard/errors` - Error log viewer with filtering
  - `GET /admin/dashboard/urls` - URL redirect tracking and engagement
- ✅ **Redis Caching**: Dashboard responses cached for 60 minutes
- ✅ **Graceful Degradation**: Analytics optional; app works without MongoDB/Redis
- ✅ **Comprehensive Documentation**: Analytics API, MongoDB setup, metrics schema guides
- ✅ **Docker Multi-Service**: Analytics worker runs as separate container
- ✅ **All Tests Passing**: 20+ tests covering analytics, middleware, config, and core logic

### Phase 1: Security & Authentication Foundation (Mar 2026)
- ✅ **API Key Authentication**: Bearer token-based auth system with Redis storage
- ✅ **Rate Limiting**: Token bucket algorithm per API key (1000 req/day default)
- ✅ **Admin API**: Generate, list, and revoke API keys
- ✅ **Request Tracing**: Unique X-Request-ID for every request
- ✅ **Structured Logging**: JSON-formatted logs with slog and context propagation
- ✅ **Panic Recovery**: Automatic recovery middleware with proper error responses
- ✅ **Health Checks**: Liveness (/health) and readiness (/ready) probes with Redis check
- ✅ **RFC 7807 Errors**: Standard problem detail error response format
- ✅ **Middleware Stack**: Request ID → Logger → Recovery → Auth → RateLimit → Handler
- ✅ **Configuration**: Security settings loaded from environment variables
- ✅ **Unit Tests**: Config and API key generation tests (no Redis dependency)

### Code Quality Improvements
- ✅ Go 1.21.6 → **1.25.0**
- ✅ All dependencies updated to latest stable versions
- ✅ Removed all `panic()` statements — proper error handling throughout
- ✅ Removed `os.Exit()` — replaced with error returns
- ✅ Fixed broken test file compilation errors
- ✅ Added input validation for URLs
- ✅ Dynamic base URL construction (works behind reverse proxies)
- ✅ User ID now properly stored with URL mapping

### New Infrastructure
- ✅ Production-ready Dockerfile with multi-stage builds
- ✅ Docker Compose for local development
- ✅ Comprehensive Makefile with 20+ targets
- ✅ Environment-based configuration system
- ✅ .gitignore for secrets and build artifacts
- ✅ Health checks in Docker

### Dependency Updates
| Package | Before | After |
|---------|--------|-------|
| gin-gonic/gin | v1.9.1 | v1.12.0 |
| redis/go-redis | v9.4.0 | v9.18.0 |
| testify | v1.8.4 | v1.11.1 |
| All others | — | Updated |

## Performance Notes

- **Short code generation**: O(1) — hash and encode
- **URL lookup**: O(1) — direct Redis GET
- **Storage**: 6-hour TTL on all mappings (configurable)
- **Concurrency**: Full Gin support for concurrent requests
- **Database**: Redis stores ~100 bytes per mapping (~1KB at typical density)

## Contributing

Contributions are welcome! Please ensure:
1. All tests pass: `make test`
2. Code is formatted: `make fmt`
3. No linting issues: `make vet`
4. New features include tests

## License

This project is provided as-is for educational and commercial use.

## Author

Created by [pouyasadri](https://github.com/pouyasadri)

## Support

For issues or questions, please open an issue on the [GitHub repository](https://github.com/pouyasadri/go-url-shortener).
