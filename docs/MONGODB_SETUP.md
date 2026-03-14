# MongoDB Setup and Configuration

This document describes the MongoDB setup, indexes, and data retention policies for the URL Shortener analytics system.

## Prerequisites

- MongoDB 4.4 or higher (5.0+ recommended)
- Connection string with appropriate credentials
- Network access to MongoDB instance

## Environment Configuration

Configure MongoDB connection via environment variables:

```bash
# .env
MONGO_URI=mongodb://user:password@localhost:27017/url-shortener-db
MONGO_DB=url-shortener-db
ANALYTICS_ENABLED=true
```

## Connection Details

The application uses MongoDB singleton pattern with connection pooling:

- **Min Pool Size:** 10 connections
- **Max Pool Size:** 100 connections
- **Server Selection Timeout:** 5 seconds
- **Connection Timeout:** 10 seconds

### Single Node Setup

For development/testing:

```bash
docker run --name mongodb -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=admin -e MONGO_INITDB_ROOT_PASSWORD=password mongo:latest
```

Connection string:
```
mongodb://admin:password@localhost:27017/url-shortener-db?authSource=admin
```

### Replica Set Setup

For production (recommended for transactions):

```bash
# Docker Compose example
services:
  mongo:
    image: mongo:5.0
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password
    command: --replSet rs0 --bind_ip_all
    volumes:
      - mongo_data:/data/db

volumes:
  mongo_data:
```

Initialize replica set:
```bash
docker exec mongodb mongosh --eval "rs.initiate({_id: 'rs0', members: [{_id: 0, host: 'localhost:27017'}]})"
```

## Collections and Indexes

The analytics system creates 6 main collections with carefully optimized indexes.

### 1. request_events

Stores individual request event records for 30-day window.

**TTL Index (30 days):**
```javascript
db.request_events.createIndex(
  { created_at: 1 },
  { expireAfterSeconds: 2592000 }  // 30 days in seconds
)
```

**Query Indexes:**
```javascript
// Query by time range and user
db.request_events.createIndex({ created_at: -1, user_id: 1 })

// Query by API key
db.request_events.createIndex({ api_key_id: 1, created_at: -1 })

// Query by path
db.request_events.createIndex({ path: 1, created_at: -1 })

// Query by status code
db.request_events.createIndex({ status_code: 1, created_at: -1 })
```

**Document Structure:**
```json
{
  "_id": ObjectId("..."),
  "request_id": "req_abc123",
  "user_id": "user_123",
  "api_key_id": "sk_live_xyz",
  "method": "GET",
  "path": "/api/v1/urls",
  "status_code": 200,
  "latency_ms": 45,
  "bytes_sent": 1024,
  "created_at": ISODate("2026-03-14T22:30:00Z")
}
```

### 2. error_events

Stores error and failure records with context for 30 days.

**TTL Index (30 days):**
```javascript
db.error_events.createIndex(
  { created_at: 1 },
  { expireAfterSeconds: 2592000 }
)
```

**Query Indexes:**
```javascript
// Query by time and status
db.error_events.createIndex({ created_at: -1, status_code: 1 })

// Query by user
db.error_events.createIndex({ user_id: 1, created_at: -1 })

// Query by error type
db.error_events.createIndex({ error_type: 1, created_at: -1 })
```

**Document Structure:**
```json
{
  "_id": ObjectId("..."),
  "request_id": "req_abc123",
  "user_id": "user_123",
  "api_key_id": "sk_live_xyz",
  "status_code": 400,
  "error_type": "ValidationError",
  "error_detail": "Invalid URL format",
  "path": "/api/v1/urls",
  "method": "POST",
  "created_at": ISODate("2026-03-14T22:30:00Z")
}
```

### 3. metrics_hourly

Pre-computed hourly metrics for fast dashboard queries.

**Query Indexes:**
```javascript
// Query by time range
db.metrics_hourly.createIndex({ hour: -1 })

// Compound queries
db.metrics_hourly.createIndex({ hour: -1, user_id: 1 })
db.metrics_hourly.createIndex({ hour: -1, api_key_id: 1 })
```

**Document Structure:**
```json
{
  "_id": ObjectId("..."),
  "hour": ISODate("2026-03-14T22:00:00Z"),
  "user_id": "user_123",
  "api_key_id": "sk_live_xyz",
  "total_requests": 150,
  "total_errors": 3,
  "total_bytes_sent": 50000,
  "status_code_breakdown": {
    "200": 145,
    "400": 3,
    "500": 2
  },
  "latencies_ms": [12, 15, 18, 25, 30, 35, 40, 45, 50, 55],
  "created_at": ISODate("2026-03-14T22:05:00Z"),
  "updated_at": ISODate("2026-03-14T23:00:00Z")
}
```

### 4. user_analytics

Aggregated user engagement metrics.

**Document Structure:**
```json
{
  "_id": "user_123",
  "urls_created": 50,
  "total_api_calls": 2500,
  "error_count": 15,
  "last_active": ISODate("2026-03-14T22:30:00Z"),
  "rate_limit_hits": 0,
  "first_seen": ISODate("2026-02-01T10:00:00Z"),
  "updated_at": ISODate("2026-03-14T22:30:00Z")
}
```

### 5. api_key_analytics

Aggregated API key usage metrics.

**Document Structure:**
```json
{
  "_id": "sk_live_abc123",
  "user_id": "user_123",
  "usage_count": 5000,
  "error_count": 50,
  "last_used": ISODate("2026-03-14T22:30:00Z"),
  "rate_limit_hits": 10,
  "created_at": ISODate("2026-02-01T10:00:00Z"),
  "updated_at": ISODate("2026-03-14T22:30:00Z")
}
```

### 6. url_analytics

Tracking for shortened URL engagement.

**Document Structure:**
```json
{
  "_id": "abc123",
  "long_url": "https://example.com/very/long/url",
  "redirect_count": 1000,
  "created_by": "user_123",
  "created_at": ISODate("2026-03-10T10:00:00Z"),
  "last_used": ISODate("2026-03-14T22:30:00Z"),
  "updated_at": ISODate("2026-03-14T22:30:00Z")
}
```

## Data Retention Policy

### Request Events & Error Events

- **Retention Period:** 30 days
- **Implementation:** TTL index on `created_at` field
- **Cleanup:** MongoDB automatic index-based deletion
- **Fallback:** Daily manual cleanup job (analytics worker)

Documents automatically expire after 30 days of creation. The TTL index removes documents at the `expireAfterSeconds` time, with potential 1-minute grace period.

### Metrics & Analytics

- **Request Metrics:** Permanent (aggregated form)
- **User Analytics:** Updated daily
- **API Key Analytics:** Updated daily
- **URL Analytics:** Updated on each redirect

These aggregated collections maintain long-term history for trend analysis.

## Index Creation

Indexes are automatically created on startup by the application:

```go
// In db/indexes.go
func InitializeIndexes(ctx context.Context) error {
    // Creates all indexes above
}
```

Manual index creation (if needed):

```bash
# Connect to MongoDB
mongosh mongodb://admin:password@localhost:27017/url-shortener-db?authSource=admin

# Check existing indexes
db.request_events.getIndexes()

# Create index manually
db.request_events.createIndex({ created_at: -1, user_id: 1 })
```

## Query Performance Tuning

### Explain Plans

Verify index usage:

```javascript
// Request events by user in last 24 hours
db.request_events.find({
  user_id: "user_123",
  created_at: { $gte: new Date(Date.now() - 86400000) }
}).explain("executionStats")
```

### Collection Statistics

Monitor collection sizes:

```javascript
db.request_events.stats()
db.metrics_hourly.stats()
```

### Index Statistics

Check index efficiency:

```javascript
db.request_events.aggregate([
  { $indexStats: {} }
])
```

## Monitoring

### Connection Monitoring

```javascript
// View current connections
db.currentOp()

// Kill long-running operation
db.killOp(<opid>)
```

### Storage Usage

```javascript
// Database statistics
db.stats()

// Collection size breakdown
db.getCollectionNames().forEach(name => {
  var stats = db[name].stats();
  print(name + ": " + stats.size + " bytes");
})
```

### Metrics Collection Size

Over time, metrics_hourly collection will grow. With daily aggregation:
- 24 documents per day per user/key combination
- ~1 KB per document
- 1 million documents per year ≈ 1-2 GB (depending on dimensionality)

Implement archival strategy if needed:
```javascript
// Archive metrics older than 1 year
db.metrics_hourly.deleteMany({
  hour: { $lt: new Date(Date.now() - 365*24*60*60*1000) }
})
```

## Backup Strategy

### Point-in-Time Recovery

For production, enable MongoDB backup:

```bash
# Manual backup
mongodump --uri "mongodb://user:pass@localhost:27017/url-shortener-db" --out ./backups

# Restore
mongorestore --uri "mongodb://user:pass@localhost:27017" ./backups
```

### Atlas (Managed MongoDB)

If using MongoDB Atlas:
1. Enable continuous backups (default)
2. Retention: 7-35 days
3. Export to S3 for long-term archival

## Troubleshooting

### High TTL Index Overhead

If TTL index causes performance issues:

1. Check TTL index status:
   ```javascript
   db.request_events.getIndexes()
   ```

2. Rebuild index if needed:
   ```javascript
   db.request_events.dropIndex("created_at_1")
   db.request_events.createIndex(
     { created_at: 1 },
     { expireAfterSeconds: 2592000 }
   )
   ```

### Query Performance Issues

1. Run explain on slow queries:
   ```javascript
   db.request_events.find({...}).explain("executionStats")
   ```

2. Add missing indexes based on query patterns

3. Consider aggregation pipeline optimization

### Replication Lag

For replica sets, monitor lag:

```javascript
rs.printSecondaryReplicationInfo()
```

If lag exceeds acceptable threshold, scale up secondary resources or reduce write rate.

## Production Checklist

- [ ] Authentication enabled (username/password or mTLS)
- [ ] Network isolation (MongoDB not exposed to internet)
- [ ] Backup strategy defined and tested
- [ ] Monitoring and alerting configured
- [ ] TTL index verified on request/error collections
- [ ] Query index coverage confirmed
- [ ] Replica set initialized (if production)
- [ ] Connection pooling configured (10-100)
- [ ] Write concern set to majority for replica sets
- [ ] Read preference set appropriately
