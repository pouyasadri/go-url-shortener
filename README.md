# Go URL Shortener

A production-ready URL shortener service built with **Go 1.25**, **Gin**, and **Redis**. Features clean error handling, environment-based configuration, Docker support with multi-stage builds, and comprehensive testing.

## Overview

The service provides a RESTful API to shorten long URLs and redirect users from short URLs to their original destinations. It uses Redis as the primary storage backend for mapping short URLs to original URLs with automatic TTL expiration (6 hours by default).

## Features

- ✅ Fast URL generation using SHA-256 hashing and Base58 encoding
- ✅ Deterministic short codes (same input always produces same output)
- ✅ 6-hour TTL on all stored mappings
- ✅ Environment-based configuration
- ✅ Comprehensive error handling (no panics in production)
- ✅ Input validation (URL format checking)
- ✅ Docker & docker-compose support with multi-stage builds
- ✅ Health checks included
- ✅ Makefile with 20+ targets for easy development
- ✅ Unit tests with 100% pass rate
- ✅ Multi-platform builds (amd64, arm64)

## Project Structure

```
go-url-shortener/
├── main.go                          # Application entry point
├── handler/
│   └── handlers.go                  # HTTP request handlers
├── shortener/
│   ├── shorturl_generator.go        # Short URL generation logic
│   └── shorturl_generator_test.go   # Unit tests
├── store/
│   ├── store_service.go             # Redis storage layer
│   └── store_service_test.go        # Integration tests
├── Dockerfile                       # Multi-stage Docker build
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

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | HTTP server port |
| `REDIS_ADDR` | `localhost:6379` | Redis server address (use `redis:6379` in Docker) |
| `REDIS_PASSWORD` | `` (empty) | Redis authentication password |
| `REDIS_DB` | `0` | Redis database number (0-15) |

Example `.env`:
```env
SERVER_PORT=8080
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

For Docker environments, `.env` variables are automatically passed to the containers via `docker-compose.yml`.

## API Usage

### Create a Short URL

**Request:**
```bash
curl -X POST http://localhost:8080/create-short-url \
  -H "Content-Type: application/json" \
  -d '{
    "long_url": "https://www.example.com/very/long/path?param=value",
    "user_id": "user-123"
  }'
```

**Response (201 Created):**
```json
{
  "message": "Short URL created successfully",
  "short_url": "http://localhost:8080/jTa4L57P"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid JSON, missing fields, or malformed URL
- `500 Internal Server Error`: Redis connection failure

### Redirect to Original URL

**Request:**
```bash
curl -L http://localhost:8080/jTa4L57P
```

**Response:**
- `302 Found`: Redirects to the original long URL
- `404 Not Found`: Short URL not found or expired (TTL expired)

### Health Check

**Request:**
```bash
curl http://localhost:8080/
```

**Response (200 OK):**
```json
{
  "message": "Welcome to the URL Shortener API"
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

This project has been thoroughly refactored and modernized:

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
