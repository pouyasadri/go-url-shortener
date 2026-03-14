# Analytics API Documentation

The Analytics API provides comprehensive insights into URL shortener usage, performance metrics, and error tracking. All analytics endpoints require admin authentication via the `X-Admin-Key` header.

## Authentication

All analytics dashboard endpoints require admin authentication:

```bash
curl -H "X-Admin-Key: your-admin-key" http://localhost:8080/admin/dashboard/overview
```

The admin key is set via the `ADMIN_API_KEY` environment variable.

## Overview Endpoint

**Endpoint:** `GET /admin/dashboard/overview`

Returns a snapshot of the system's current state with key metrics.

### Query Parameters

- `timeframe` (optional): Time period for metrics. Valid values: `24h`, `7d`, `30d`. Default: `24h`

### Response

```json
{
  "status": "ok",
  "timestamp": "2026-03-14T22:30:00Z",
  "timeframe": "24h",
  "cached_at": "2026-03-14T22:29:00Z",
  "data": {
    "total_requests": 1500,
    "total_errors": 15,
    "error_rate": 1.0,
    "avg_latency_ms": 45.2,
    "active_api_keys": 8,
    "top_users": [
      {
        "user_id": "user_123",
        "urls_created": 50,
        "api_calls_total": 250,
        "error_count": 2,
        "last_active": "2026-03-14T22:25:00Z",
        "rate_limit_hits": 0
      }
    ],
    "status_breakdown": {
      "200": 1485,
      "201": 10,
      "400": 5,
      "401": 0,
      "500": 0
    }
  }
}
```

### Fields

- `total_requests` (int64): Total requests in the timeframe
- `total_errors` (int64): Total error responses (status >= 400)
- `error_rate` (float64): Error rate as percentage
- `avg_latency_ms` (float64): Average response latency in milliseconds
- `active_api_keys` (int64): Number of API keys that made requests
- `top_users` (array): Top 5 users by request volume
- `status_breakdown` (object): HTTP status code distribution

## Requests Endpoint

**Endpoint:** `GET /admin/dashboard/requests`

Returns time-series request metrics with latency percentiles and status breakdown.

### Query Parameters

- `timeframe` (optional): Time period for metrics. Valid values: `24h`, `7d`, `30d`. Default: `24h`
- `sort` (optional): Sort field. Valid values: `requests`, `errors`, `latency`. Default: `requests`

### Response

```json
{
  "status": "ok",
  "timestamp": "2026-03-14T22:30:00Z",
  "timeframe": "24h",
  "cached_at": "2026-03-14T22:29:00Z",
  "data": {
    "total_requests": 1500,
    "total_errors": 15,
    "avg_latency_ms": 45.2,
    "p50_latency_ms": 32.0,
    "p95_latency_ms": 89.5,
    "p99_latency_ms": 145.3,
    "status_code_breakdown": {
      "200": 1485,
      "201": 10,
      "400": 5
    },
    "total_bytes_sent": 5000000
  }
}
```

### Fields

- `total_requests` (int64): Total requests
- `total_errors` (int64): Total error responses
- `avg_latency_ms` (float64): Average latency
- `p50_latency_ms` (float64): 50th percentile latency
- `p95_latency_ms` (float64): 95th percentile latency
- `p99_latency_ms` (float64): 99th percentile latency
- `status_code_breakdown` (object): Status code distribution
- `total_bytes_sent` (int64): Total bytes sent in responses

## Users Endpoint

**Endpoint:** `GET /admin/dashboard/users`

Returns user engagement statistics with sortable and paginated results.

### Query Parameters

- `timeframe` (optional): Time period for metrics. Valid values: `24h`, `7d`, `30d`. Default: `24h`
- `sort` (optional): Sort field. Valid values: `urls_created`, `api_calls`, `errors`. Default: `api_calls`
- `limit` (optional): Number of results (max 1000). Default: `100`
- `offset` (optional): Pagination offset. Default: `0`

### Response

```json
{
  "status": "ok",
  "timestamp": "2026-03-14T22:30:00Z",
  "timeframe": "24h",
  "cached_at": "2026-03-14T22:29:00Z",
  "data": {
    "users": [
      {
        "user_id": "user_123",
        "urls_created": 50,
        "api_calls_total": 250,
        "error_count": 2,
        "last_active": "2026-03-14T22:25:00Z",
        "rate_limit_hits": 0
      }
    ],
    "total_users": 45,
    "limit": 100,
    "offset": 0
  }
}
```

### Fields

- `users` (array): Array of user summaries
  - `user_id` (string): User identifier
  - `urls_created` (int64): Number of URLs created
  - `api_calls_total` (int64): Total API calls made
  - `error_count` (int64): Number of error responses
  - `last_active` (timestamp): Last activity timestamp
  - `rate_limit_hits` (int64): Number of rate limit hits
- `total_users` (int64): Total number of users
- `limit` (int64): Limit used in request
- `offset` (int64): Offset used in request

## API Keys Endpoint

**Endpoint:** `GET /admin/dashboard/api-keys`

Returns API key usage analytics with performance metrics.

### Query Parameters

- `timeframe` (optional): Time period for metrics. Valid values: `24h`, `7d`, `30d`. Default: `24h`
- `sort` (optional): Sort field. Valid values: `usage`, `errors`. Default: `usage`
- `limit` (optional): Number of results (max 1000). Default: `100`

### Response

```json
{
  "status": "ok",
  "timestamp": "2026-03-14T22:30:00Z",
  "timeframe": "24h",
  "cached_at": "2026-03-14T22:29:00Z",
  "data": {
    "api_keys": [
      {
        "api_key_id": "sk_live_abc123",
        "usage_count": 500,
        "error_count": 10,
        "error_rate": 2.0,
        "last_used": "2026-03-14T22:25:00Z",
        "rate_limit_hits": 5
      }
    ],
    "total_keys": 8,
    "limit": 100
  }
}
```

### Fields

- `api_keys` (array): Array of API key summaries
  - `api_key_id` (string): API key identifier
  - `usage_count` (int64): Number of requests
  - `error_count` (int64): Number of errors
  - `error_rate` (float64): Error rate as percentage
  - `last_used` (timestamp): Last usage timestamp
  - `rate_limit_hits` (int64): Number of rate limit hits
- `total_keys` (int64): Total number of API keys
- `limit` (int64): Limit used in request

## Errors Endpoint

**Endpoint:** `GET /admin/dashboard/errors`

Returns structured error log with filtering capabilities.

### Query Parameters

- `timeframe` (optional): Time period for metrics. Valid values: `24h`, `7d`, `30d`. Default: `24h`
- `status_code` (optional): Filter by HTTP status code (e.g., `400`, `500`)
- `user_id` (optional): Filter by user ID
- `limit` (optional): Number of results (max 1000). Default: `100`

### Response

```json
{
  "status": "ok",
  "timestamp": "2026-03-14T22:30:00Z",
  "timeframe": "24h",
  "cached_at": "2026-03-14T22:29:00Z",
  "data": {
    "errors": [
      {
        "request_id": "req_abc123",
        "timestamp": "2026-03-14T22:25:30Z",
        "error_type": "ValidationError",
        "status_code": 400,
        "user_id": "user_123",
        "api_key_id": "sk_live_abc123",
        "error_detail": "Invalid URL format"
      }
    ],
    "total_errors": 15,
    "limit": 100
  }
}
```

### Fields

- `errors` (array): Array of error log entries
  - `request_id` (string): Unique request identifier
  - `timestamp` (timestamp): When error occurred
  - `error_type` (string): Error classification
  - `status_code` (int): HTTP status code
  - `user_id` (string): User who made the request
  - `api_key_id` (string): API key used
  - `error_detail` (string): Error message details
- `total_errors` (int64): Total errors in timeframe
- `limit` (int64): Limit used in request

## URLs Endpoint

**Endpoint:** `GET /admin/dashboard/urls`

Returns redirect tracking and engagement metrics for shortened URLs.

### Query Parameters

- `timeframe` (optional): Time period for metrics. Valid values: `24h`, `7d`, `30d`. Default: `24h`
- `sort` (optional): Sort field. Valid values: `redirects`, `created_at`. Default: `redirects`
- `limit` (optional): Number of results (max 1000). Default: `100`

### Response

```json
{
  "status": "ok",
  "timestamp": "2026-03-14T22:30:00Z",
  "timeframe": "24h",
  "cached_at": "2026-03-14T22:29:00Z",
  "data": {
    "urls": [
      {
        "short_url": "abc123",
        "long_url": "https://example.com/very/long/url",
        "redirect_count": 1000,
        "created_by": "user_123",
        "created_at": "2026-03-10T10:00:00Z",
        "last_used": "2026-03-14T22:25:00Z"
      }
    ],
    "total_urls": 250,
    "limit": 100
  }
}
```

### Fields

- `urls` (array): Array of URL analytics summaries
  - `short_url` (string): Short URL code
  - `long_url` (string): Original long URL
  - `redirect_count` (int64): Number of redirects
  - `created_by` (string): User who created the URL
  - `created_at` (timestamp): Creation timestamp
  - `last_used` (timestamp): Last redirect timestamp
- `total_urls` (int64): Total URLs created
- `limit` (int64): Limit used in request

## Caching

All dashboard endpoints implement Redis-based caching with a default TTL of 60 minutes. The cache includes:

- Endpoint-specific data structures
- Query parameters (timeframe, sort, filter) in cache key
- `cached_at` timestamp in response

If Redis is unavailable, endpoints degrade gracefully and compute metrics on-the-fly without caching.

## Rate Limiting

Admin endpoints are exempt from normal rate limiting but are subject to standard security controls via `X-Admin-Key` header validation.

## Error Responses

All errors follow RFC 7807 Problem Details format:

```json
{
  "type": "about:blank",
  "title": "Internal Server Error",
  "status": 500,
  "detail": "MongoDB connection failed"
}
```

### Common Status Codes

- `200 OK`: Request successful
- `400 Bad Request`: Invalid query parameters
- `401 Unauthorized`: Missing or invalid admin key
- `500 Internal Server Error`: MongoDB or Redis unavailable

## Examples

### Get 7-day overview for top 3 users

```bash
curl -H "X-Admin-Key: sk_admin_xyz" \
  "http://localhost:8080/admin/dashboard/overview?timeframe=7d"
```

### Get error logs filtered by status code 500

```bash
curl -H "X-Admin-Key: sk_admin_xyz" \
  "http://localhost:8080/admin/dashboard/errors?timeframe=24h&status_code=500&limit=50"
```

### Get top 10 API keys by usage (7 days)

```bash
curl -H "X-Admin-Key: sk_admin_xyz" \
  "http://localhost:8080/admin/dashboard/api-keys?timeframe=7d&sort=usage&limit=10"
```

### Get latency percentiles for 30 days

```bash
curl -H "X-Admin-Key: sk_admin_xyz" \
  "http://localhost:8080/admin/dashboard/requests?timeframe=30d"
```
