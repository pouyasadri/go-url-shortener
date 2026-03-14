package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/analytics"
	"github.com/pouyasadri/go-url-shortener/cache"
)

// DashboardResponse is the base response structure for all dashboard endpoints
type DashboardResponse struct {
	Status    string      `json:"status"`
	Timestamp time.Time   `json:"timestamp"`
	Timeframe string      `json:"timeframe"`
	Data      interface{} `json:"data"`
	CachedAt  *time.Time  `json:"cached_at,omitempty"`
}

// OverviewData represents system health snapshot
type OverviewData struct {
	TotalRequests   int64                  `json:"total_requests"`
	TotalErrors     int64                  `json:"total_errors"`
	ErrorRate       float64                `json:"error_rate"`
	AvgLatencyMs    float64                `json:"avg_latency_ms"`
	ActiveAPIKeys   int                    `json:"active_api_keys"`
	TopUsers        []UserSummary          `json:"top_users,omitempty"`
	StatusBreakdown map[string]interface{} `json:"status_breakdown"`
}

// RequestMetricsData represents time-series request data
type RequestMetricsData struct {
	TotalRequests   int64            `json:"total_requests"`
	TotalErrors     int64            `json:"total_errors"`
	AvgLatencyMs    float64          `json:"avg_latency_ms"`
	P50LatencyMs    float64          `json:"p50_latency_ms"`
	P95LatencyMs    float64          `json:"p95_latency_ms"`
	P99LatencyMs    float64          `json:"p99_latency_ms"`
	StatusCodeBreak map[string]int64 `json:"status_code_breakdown"`
	TotalBytesSent  int64            `json:"total_bytes_sent"`
}

// UserSummary represents aggregated user stats
type UserSummary struct {
	UserID        string    `json:"user_id"`
	URLsCreated   int64     `json:"urls_created"`
	APICallsTotal int64     `json:"api_calls_total"`
	ErrorCount    int64     `json:"error_count"`
	LastActive    time.Time `json:"last_active"`
	RateLimitHits int64     `json:"rate_limit_hits"`
}

// APIKeySummary represents aggregated API key stats
type APIKeySummary struct {
	APIKeyID      string    `json:"api_key_id"`
	UsageCount    int64     `json:"usage_count"`
	ErrorCount    int64     `json:"error_count"`
	ErrorRate     float64   `json:"error_rate"`
	LastUsed      time.Time `json:"last_used"`
	RateLimitHits int64     `json:"rate_limit_hits"`
}

// ErrorLogEntry represents a single error event
type ErrorLogEntry struct {
	RequestID   string    `json:"request_id"`
	Timestamp   time.Time `json:"timestamp"`
	ErrorType   string    `json:"error_type"`
	StatusCode  int       `json:"status_code"`
	UserID      string    `json:"user_id"`
	APIKeyID    string    `json:"api_key_id"`
	ErrorDetail string    `json:"error_detail,omitempty"`
}

// URLAnalyticsSummary represents URL engagement data
type URLAnalyticsSummary struct {
	ShortURL      string    `json:"short_url"`
	LongURL       string    `json:"long_url"`
	RedirectCount int64     `json:"redirect_count"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	LastUsed      time.Time `json:"last_used"`
}

// GetTimeframe parses the timeframe query parameter
func GetTimeframe(c *gin.Context) string {
	timeframe := c.DefaultQuery("timeframe", "24h")
	// Validate timeframe
	switch timeframe {
	case "24h", "7d", "30d":
		return timeframe
	default:
		return "24h"
	}
}

// GetTimeframeHours converts timeframe string to hours
func GetTimeframeHours(timeframe string) int {
	switch timeframe {
	case "24h":
		return 24
	case "7d":
		return 24 * 7
	case "30d":
		return 24 * 30
	default:
		return 24
	}
}

// DashboardOverview returns system health snapshot
// GET /admin/dashboard/overview
func DashboardOverview(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	timeframe := GetTimeframe(c)
	cacheKey := cache.CacheKeyDashboardOverview + ":" + timeframe

	// Check cache
	redisCache := cache.NewRedisCache(60)
	var cachedData OverviewData
	var cachedTime *time.Time
	if err := redisCache.Get(ctx, cacheKey, &cachedData); err == nil {
		now := time.Now()
		cachedTime = &now
		c.JSON(http.StatusOK, DashboardResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Timeframe: timeframe,
			Data:      cachedData,
			CachedAt:  cachedTime,
		})
		return
	}

	// Generate data if not cached
	repo, err := analytics.NewRepository()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access analytics",
		})
		return
	}

	// For now, return mock data structure
	// In Day 4, this will query aggregated metrics
	data := OverviewData{
		TotalRequests: 0,
		TotalErrors:   0,
		ErrorRate:     0.0,
		AvgLatencyMs:  0.0,
		ActiveAPIKeys: 0,
		TopUsers:      make([]UserSummary, 0),
		StatusBreakdown: map[string]interface{}{
			"200": 0,
			"400": 0,
			"500": 0,
		},
	}

	// Cache for 1 hour
	_ = redisCache.SetWithTTL(ctx, cacheKey, data, 60*time.Minute)
	_ = repo // Suppress unused warning for now

	c.JSON(http.StatusOK, DashboardResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Timeframe: timeframe,
		Data:      data,
	})
}

// DashboardRequests returns time-series request metrics
// GET /admin/dashboard/requests
func DashboardRequests(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	timeframe := GetTimeframe(c)
	cacheKey := cache.CacheKeyDashboardRequests + ":" + timeframe

	// Check cache
	redisCache := cache.NewRedisCache(60)
	var cachedData RequestMetricsData
	var cachedTime *time.Time
	if err := redisCache.Get(ctx, cacheKey, &cachedData); err == nil {
		now := time.Now()
		cachedTime = &now
		c.JSON(http.StatusOK, DashboardResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Timeframe: timeframe,
			Data:      cachedData,
			CachedAt:  cachedTime,
		})
		return
	}

	// Generate data if not cached
	repo, err := analytics.NewRepository()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access analytics",
		})
		return
	}

	// For now, return mock data structure
	data := RequestMetricsData{
		TotalRequests:   0,
		TotalErrors:     0,
		AvgLatencyMs:    0.0,
		P50LatencyMs:    0.0,
		P95LatencyMs:    0.0,
		P99LatencyMs:    0.0,
		StatusCodeBreak: make(map[string]int64),
		TotalBytesSent:  0,
	}

	// Cache for 1 hour
	_ = redisCache.SetWithTTL(ctx, cacheKey, data, 60*time.Minute)
	_ = repo

	c.JSON(http.StatusOK, DashboardResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Timeframe: timeframe,
		Data:      data,
	})
}

// DashboardUsers returns user engagement statistics
// GET /admin/dashboard/users
func DashboardUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	timeframe := GetTimeframe(c)
	sortBy := c.DefaultQuery("sort", "api_calls") // api_calls, urls_created, last_active
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	cacheKey := cache.CacheKeyDashboardUsers + ":" + timeframe + ":" + sortBy

	// Check cache
	redisCache := cache.NewRedisCache(60)
	var cachedData []UserSummary
	var cachedTime *time.Time
	if err := redisCache.Get(ctx, cacheKey, &cachedData); err == nil {
		now := time.Now()
		cachedTime = &now
		// Respect limit on cached data
		if len(cachedData) > limit {
			cachedData = cachedData[:limit]
		}
		c.JSON(http.StatusOK, DashboardResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Timeframe: timeframe,
			Data:      cachedData,
			CachedAt:  cachedTime,
		})
		return
	}

	// Generate data if not cached
	repo, err := analytics.NewRepository()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access analytics",
		})
		return
	}

	data := make([]UserSummary, 0)

	// Cache for 1 hour
	_ = redisCache.SetWithTTL(ctx, cacheKey, data, 60*time.Minute)
	_ = repo
	_ = sortBy
	_ = limit

	c.JSON(http.StatusOK, DashboardResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Timeframe: timeframe,
		Data:      data,
	})
}

// DashboardAPIKeys returns API key usage analytics
// GET /admin/dashboard/api-keys
func DashboardAPIKeys(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	timeframe := GetTimeframe(c)
	sortBy := c.DefaultQuery("sort", "usage") // usage, error_rate, last_used
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	cacheKey := cache.CacheKeyDashboardAPIKeys + ":" + timeframe + ":" + sortBy

	// Check cache
	redisCache := cache.NewRedisCache(60)
	var cachedData []APIKeySummary
	var cachedTime *time.Time
	if err := redisCache.Get(ctx, cacheKey, &cachedData); err == nil {
		now := time.Now()
		cachedTime = &now
		// Respect limit on cached data
		if len(cachedData) > limit {
			cachedData = cachedData[:limit]
		}
		c.JSON(http.StatusOK, DashboardResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Timeframe: timeframe,
			Data:      cachedData,
			CachedAt:  cachedTime,
		})
		return
	}

	// Generate data if not cached
	repo, err := analytics.NewRepository()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access analytics",
		})
		return
	}

	data := make([]APIKeySummary, 0)

	// Cache for 1 hour
	_ = redisCache.SetWithTTL(ctx, cacheKey, data, 60*time.Minute)
	_ = repo
	_ = sortBy
	_ = limit

	c.JSON(http.StatusOK, DashboardResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Timeframe: timeframe,
		Data:      data,
	})
}

// DashboardErrors returns error log viewer with filtering
// GET /admin/dashboard/errors
func DashboardErrors(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	timeframe := GetTimeframe(c)
	errorType := c.Query("error_type") // filter by error type
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	cacheKey := cache.CacheKeyDashboardErrors + ":" + timeframe
	if errorType != "" {
		cacheKey += ":" + errorType
	}

	// Check cache
	redisCache := cache.NewRedisCache(60)
	var cachedData []ErrorLogEntry
	var cachedTime *time.Time
	if err := redisCache.Get(ctx, cacheKey, &cachedData); err == nil {
		now := time.Now()
		cachedTime = &now
		// Respect limit on cached data
		if len(cachedData) > limit {
			cachedData = cachedData[:limit]
		}
		c.JSON(http.StatusOK, DashboardResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Timeframe: timeframe,
			Data:      cachedData,
			CachedAt:  cachedTime,
		})
		return
	}

	// Generate data if not cached
	repo, err := analytics.NewRepository()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access analytics",
		})
		return
	}

	data := make([]ErrorLogEntry, 0)

	// Cache for 1 hour
	_ = redisCache.SetWithTTL(ctx, cacheKey, data, 60*time.Minute)
	_ = repo
	_ = errorType
	_ = limit

	c.JSON(http.StatusOK, DashboardResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Timeframe: timeframe,
		Data:      data,
	})
}

// DashboardURLs returns URL redirect tracking and engagement
// GET /admin/dashboard/urls
func DashboardURLs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	timeframe := GetTimeframe(c)
	sortBy := c.DefaultQuery("sort", "redirects") // redirects, created, last_used
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	cacheKey := cache.CacheKeyDashboardURLs + ":" + timeframe + ":" + sortBy

	// Check cache
	redisCache := cache.NewRedisCache(60)
	var cachedData []URLAnalyticsSummary
	var cachedTime *time.Time
	if err := redisCache.Get(ctx, cacheKey, &cachedData); err == nil {
		now := time.Now()
		cachedTime = &now
		// Respect limit on cached data
		if len(cachedData) > limit {
			cachedData = cachedData[:limit]
		}
		c.JSON(http.StatusOK, DashboardResponse{
			Status:    "ok",
			Timestamp: time.Now(),
			Timeframe: timeframe,
			Data:      cachedData,
			CachedAt:  cachedTime,
		})
		return
	}

	// Generate data if not cached
	repo, err := analytics.NewRepository()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to access analytics",
		})
		return
	}

	data := make([]URLAnalyticsSummary, 0)

	// Cache for 1 hour
	_ = redisCache.SetWithTTL(ctx, cacheKey, data, 60*time.Minute)
	_ = repo
	_ = sortBy
	_ = limit

	c.JSON(http.StatusOK, DashboardResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Timeframe: timeframe,
		Data:      data,
	})
}
