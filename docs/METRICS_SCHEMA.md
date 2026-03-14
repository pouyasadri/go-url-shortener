# Metrics Schema Documentation

Complete reference for all analytics data models and their relationships in the URL Shortener analytics system.

## Data Models

### AnalyticsEvent

The core event structure captured by the analytics middleware for every request.

```go
type AnalyticsEvent struct {
    RequestID    string        // Unique request identifier
    UserID       string        // User ID from Bearer token
    APIKeyID     string        // API key used for request
    Method       string        // HTTP method (GET, POST, etc.)
    Path         string        // Request path
    StatusCode   int           // HTTP response status code
    LatencyMs    int64         // Request duration in milliseconds
    BytesSent    int64         // Response body size in bytes
    ErrorMessage *string       // Error message if request failed (nil for success)
    Timestamp    time.Time     // When event occurred
}
```

**Publishing:** Events are published to Redis channel `analytics:events` asynchronously (< 1ms overhead).

**Processing:** Event processor batches events (1000 items or 5 minutes) before writing to MongoDB.

### RequestEvent

MongoDB document representing a single HTTP request.

**Collection:** `request_events`

**TTL:** 30 days (automatic deletion)

```go
type RequestEvent struct {
    ID         primitive.ObjectID `bson:"_id,omitempty"`
    RequestID  string             `bson:"request_id"`
    UserID     string             `bson:"user_id"`
    APIKeyID   string             `bson:"api_key_id"`
    Method     string             `bson:"method"`
    Path       string             `bson:"path"`
    StatusCode int                `bson:"status_code"`
    LatencyMs  int64              `bson:"latency_ms"`
    BytesSent  int64              `bson:"bytes_sent"`
    CreatedAt  time.Time          `bson:"created_at"`
}
```

**Indexes:**
- `created_at` (TTL, 30 days)
- `created_at DESC, user_id ASC`
- `api_key_id ASC, created_at DESC`
- `path ASC, created_at DESC`
- `status_code ASC, created_at DESC`

**Query Use Cases:**
- Find all requests for a user in last 24h
- Find errors (status >= 400) for analysis
- Aggregate latencies for percentile calculation
- Group by API key for usage metrics

### ErrorEvent

MongoDB document representing an error or failed request.

**Collection:** `error_events`

**TTL:** 30 days (automatic deletion)

```go
type ErrorEvent struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    RequestID   string             `bson:"request_id"`
    UserID      string             `bson:"user_id"`
    APIKeyID    string             `bson:"api_key_id"`
    StatusCode  int                `bson:"status_code"`
    ErrorType   string             `bson:"error_type"` // e.g., ValidationError, RateLimitError
    ErrorDetail string             `bson:"error_detail"`
    Path        string             `bson:"path"`
    Method      string             `bson:"method"`
    CreatedAt   time.Time          `bson:"created_at"`
}
```

**Indexes:**
- `created_at` (TTL, 30 days)
- `created_at DESC, status_code ASC`
- `user_id ASC, created_at DESC`
- `error_type ASC, created_at DESC`

**Query Use Cases:**
- Find all 5xx errors in last hour
- Error trends by type
- User error history
- Error dashboard filtering

### MetricsHourly

Pre-computed hourly aggregations for dashboard performance.

**Collection:** `metrics_hourly`

**Retention:** Permanent (can be archived if needed)

**Computation:** Cron job runs at top of each hour, aggregates past hour's events

```go
type MetricsHourly struct {
    ID                  primitive.ObjectID     `bson:"_id,omitempty"`
    Hour                time.Time              `bson:"hour"` // Hour truncated (e.g., 22:00:00)
    UserID              string                 `bson:"user_id"`
    APIKeyID            string                 `bson:"api_key_id"`
    TotalRequests       int64                  `bson:"total_requests"`
    TotalErrors         int64                  `bson:"total_errors"`
    TotalBytesSent      int64                  `bson:"total_bytes_sent"`
    StatusCodeBreakdown map[string]int64       `bson:"status_code_breakdown"` // e.g., {"200": 145, "400": 3}
    LatenciesMs         []int64                `bson:"latencies_ms"` // Sorted latency values for percentile calc
    CreatedAt           time.Time              `bson:"created_at"`
    UpdatedAt           time.Time              `bson:"updated_at"`
}
```

**Indexes:**
- `hour DESC`
- `hour DESC, user_id ASC`
- `hour DESC, api_key_id ASC`

**Computation Algorithm:**

1. Run at beginning of hour (e.g., 23:00 for 22:00 hour)
2. Query `request_events` for previous hour
3. Group by user_id, api_key_id
4. For each group:
   - Count total requests and errors
   - Sum bytes sent
   - Build status code breakdown
   - Collect all latency values
5. Sort latencies and calculate percentiles (p50, p95, p99)
6. Upsert into `metrics_hourly` (idempotent)

**Percentile Calculation:**

Using linear interpolation on sorted latency array:
```
p95_idx = 0.95 * (len(latencies) - 1)
lower = latencies[floor(p95_idx)]
upper = latencies[ceil(p95_idx)]
p95 = lower + (upper - lower) * (p95_idx % 1)
```

### UserAnalytics

Aggregated user engagement metrics (updated from request events).

**Collection:** `user_analytics`

**Document ID:** User ID (e.g., `user_123`)

**Retention:** Permanent

```go
type UserAnalytics struct {
    ID             string    `bson:"_id"`
    URLsCreated    int64     `bson:"urls_created"`
    TotalAPICalls  int64     `bson:"total_api_calls"`
    ErrorCount     int64     `bson:"error_count"`
    LastActive     time.Time `bson:"last_active"`
    RateLimitHits  int64     `bson:"rate_limit_hits"`
    FirstSeen      time.Time `bson:"first_seen"`
    UpdatedAt      time.Time `bson:"updated_at"`
}
```

**Update Strategy:**
- Updated by metrics aggregator job
- Queries all request_events for user
- Counts URLs by filtering `method == "POST" && path == "/api/v1/urls"`
- Counts errors by `status_code >= 400`
- Tracks rate limit errors (429 status)

### APIKeyAnalytics

Aggregated API key usage metrics.

**Collection:** `api_key_analytics`

**Document ID:** API Key ID (e.g., `sk_live_abc123`)

**Retention:** Permanent

```go
type APIKeyAnalytics struct {
    ID            string    `bson:"_id"`
    UserID        string    `bson:"user_id"`
    UsageCount    int64     `bson:"usage_count"`
    ErrorCount    int64     `bson:"error_count"`
    LastUsed      time.Time `bson:"last_used"`
    RateLimitHits int64     `bson:"rate_limit_hits"`
    CreatedAt     time.Time `bson:"created_at"`
    UpdatedAt     time.Time `bson:"updated_at"`
}
```

**Update Strategy:**
- Updated by metrics aggregator job
- Queries all request_events for api_key_id
- Counts errors and rate limit hits

### URLAnalytics

Tracking for shortened URL engagement and reach.

**Collection:** `url_analytics`

**Document ID:** Short URL code (e.g., `abc123`)

**Retention:** Permanent (mirrors url_mappings in Redis)

```go
type URLAnalytics struct {
    ID            string    `bson:"_id"`
    LongURL       string    `bson:"long_url"`
    RedirectCount int64     `bson:"redirect_count"`
    CreatedBy     string    `bson:"created_by"`
    CreatedAt     time.Time `bson:"created_at"`
    LastUsed      time.Time `bson:"last_used"`
    UpdatedAt     time.Time `bson:"updated_at"`
}
```

**Update Strategy:**
- Incremented on every redirect (HandleShortURLRedirect)
- Updated with LastUsed timestamp
- RedirectCount tracked for engagement metrics

## Dashboard Response Models

### DashboardResponse

Generic wrapper for all dashboard responses.

```go
type DashboardResponse struct {
    Status    string                 `json:"status"` // "ok" or "error"
    Timestamp time.Time              `json:"timestamp"`
    Timeframe string                 `json:"timeframe"` // "24h", "7d", "30d"
    CachedAt  *time.Time             `json:"cached_at,omitempty"`
    Data      interface{}            `json:"data"`
}
```

**CachedAt:** Set if response was retrieved from Redis cache. Indicates when cache was generated.

### OverviewData

System snapshot with key metrics.

```go
type OverviewData struct {
    TotalRequests   int64                  `json:"total_requests"`
    TotalErrors     int64                  `json:"total_errors"`
    ErrorRate       float64                `json:"error_rate"` // Percentage
    AvgLatencyMs    float64                `json:"avg_latency_ms"`
    ActiveAPIKeys   int64                  `json:"active_api_keys"`
    TopUsers        []UserSummary          `json:"top_users"` // Top 5
    StatusBreakdown map[string]interface{} `json:"status_breakdown"`
}
```

### RequestMetricsData

Time-series request metrics with percentiles.

```go
type RequestMetricsData struct {
    TotalRequests    int64              `json:"total_requests"`
    TotalErrors      int64              `json:"total_errors"`
    AvgLatencyMs     float64            `json:"avg_latency_ms"`
    P50LatencyMs     float64            `json:"p50_latency_ms"`
    P95LatencyMs     float64            `json:"p95_latency_ms"`
    P99LatencyMs     float64            `json:"p99_latency_ms"`
    StatusCodeBreak  map[string]int64   `json:"status_code_breakdown"`
    TotalBytesSent   int64              `json:"total_bytes_sent"`
}
```

**Latency Calculation:**
- Aggregates latencies from metrics_hourly
- Sorts and calculates percentiles
- Uses linear interpolation for precision

### UserSummary

User engagement metrics for dashboard.

```go
type UserSummary struct {
    UserID        string    `json:"user_id"`
    URLsCreated   int64     `json:"urls_created"`
    APICallsTotal int64     `json:"api_calls_total"`
    ErrorCount    int64     `json:"error_count"`
    LastActive    time.Time `json:"last_active"`
    RateLimitHits int64     `json:"rate_limit_hits"`
}
```

### APIKeySummary

API key usage metrics for dashboard.

```go
type APIKeySummary struct {
    APIKeyID      string    `json:"api_key_id"`
    UsageCount    int64     `json:"usage_count"`
    ErrorCount    int64     `json:"error_count"`
    ErrorRate     float64   `json:"error_rate"` // Percentage
    LastUsed      time.Time `json:"last_used"`
    RateLimitHits int64     `json:"rate_limit_hits"`
}
```

### ErrorLogEntry

Structured error log for dashboard.

```go
type ErrorLogEntry struct {
    RequestID   string    `json:"request_id"`
    Timestamp   time.Time `json:"timestamp"`
    ErrorType   string    `json:"error_type"`
    StatusCode  int       `json:"status_code"`
    UserID      string    `json:"user_id"`
    APIKeyID    string    `json:"api_key_id"`
    ErrorDetail string    `json:"error_detail"`
}
```

### URLAnalyticsSummary

URL engagement metrics for dashboard.

```go
type URLAnalyticsSummary struct {
    ShortURL      string    `json:"short_url"`
    LongURL       string    `json:"long_url"`
    RedirectCount int64     `json:"redirect_count"`
    CreatedBy     string    `json:"created_by"`
    CreatedAt     time.Time `json:"created_at"`
    LastUsed      time.Time `json:"last_used"`
}
```

## Data Flow

```
Client Request
       ↓
[Analytics Middleware]
    - Creates AnalyticsEvent
    - Publishes to Redis (analytics:events channel)
    - Returns immediately (< 1ms overhead)
       ↓
[Redis Pub/Sub]
   ↓→→→→→→→ Event Processor Goroutine
           - Subscribes to analytics:events
           - Buffers events (max 1000 or 5 min timeout)
           - Batch inserts into request_events
           - Publishes request/error event counts
           - Handles MongoDB connection errors gracefully
       ↓
[MongoDB request_events & error_events]
   ↓→→→→→→→ Metrics Aggregator (Cron: hourly at :00)
           - Queries events for last hour
           - Calculates percentiles
           - Groups by user/key
           - Upserts into metrics_hourly
       ↓
[MongoDB metrics_hourly, user_analytics, api_key_analytics]
   ↓→→→→→→→ Dashboard Handler
           - Checks Redis cache
           - If miss: Queries MongoDB aggregations
           - Merges with query parameters
           - Caches response (60 min TTL)
           - Returns JSON response
       ↓
[Admin Dashboard API]
   Response with CachedAt timestamp
```

## Query Patterns

### Get Hourly Metrics (24 hours)

```javascript
db.metrics_hourly.find({
  hour: {
    $gte: new Date(Date.now() - 24*60*60*1000),
    $lt: new Date()
  }
})
.sort({ hour: -1 })
.limit(24)
```

### Get User Errors (24 hours)

```javascript
db.error_events.find({
  user_id: "user_123",
  created_at: {
    $gte: new Date(Date.now() - 24*60*60*1000)
  }
})
.sort({ created_at: -1 })
.limit(100)
```

### Calculate Error Rate

```javascript
db.request_events.aggregate([
  {
    $match: {
      created_at: {
        $gte: new Date(Date.now() - 24*60*60*1000)
      }
    }
  },
  {
    $facet: {
      total: [{ $count: "count" }],
      errors: [
        { $match: { status_code: { $gte: 400 } } },
        { $count: "count" }
      ]
    }
  }
])
```

### Get Top URLs by Redirects

```javascript
db.url_analytics.find({})
.sort({ redirect_count: -1 })
.limit(10)
```

## Storage Estimates

**Per Day (assuming 100k requests):**
- request_events: ~100 MB
- error_events: ~5 MB (5% error rate)
- metrics_hourly: ~500 KB
- **Total per day:** ~105 MB

**30-Day Retention:**
- Raw events: ~3.15 GB
- Metrics: ~15 MB
- **Total hot storage:** ~3.2 GB

**Growth over 1 year:**
- Compressed: ~500 GB (depends on compression ratio)
- With archival strategy: ~50 GB hot + archived cold storage

**Index Overhead:**
- ~30% of data size (rough estimate)
- Mainly from TTL index on timestamp
- Sortable indexes on frequently queried fields

## Cache Strategy

**Cache Keys:**
```
analytics:dashboard:overview:<timeframe>
analytics:dashboard:requests:<timeframe>:<sort>
analytics:dashboard:users:<timeframe>:<sort>:<limit>
analytics:dashboard:api-keys:<timeframe>:<sort>:<limit>
analytics:dashboard:errors:<timeframe>:<filter>
analytics:dashboard:urls:<timeframe>:<sort>:<limit>
```

**TTL:** 60 minutes (configurable)

**Invalidation:** 
- Manual: When metrics aggregator completes (future enhancement)
- Automatic: Expires after TTL

**Fallback:** If Redis unavailable, compute on-the-fly (graceful degradation)
