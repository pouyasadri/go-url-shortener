package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RequestEvent represents a single HTTP request captured for analytics
type RequestEvent struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	RequestID       string             `bson:"request_id"`
	Timestamp       time.Time          `bson:"timestamp"`
	Method          string             `bson:"method"`
	Path            string             `bson:"path"`
	StatusCode      int                `bson:"status_code"`
	DurationMs      float64            `bson:"duration_ms"`
	BytesSent       int64              `bson:"bytes_sent"`
	UserID          string             `bson:"user_id"`
	APIKeyID        string             `bson:"api_key_id"`
	ErrorType       *string            `bson:"error_type,omitempty"`
	CreatedShortURL *string            `bson:"created_short_url,omitempty"`
	ClientIP        string             `bson:"client_ip"`
	CreatedAt       time.Time          `bson:"created_at"`
}

// ErrorEvent represents an error that occurred during request processing
type ErrorEvent struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	RequestID   string             `bson:"request_id"`
	Timestamp   time.Time          `bson:"timestamp"`
	ErrorType   string             `bson:"error_type"`
	ErrorDetail string             `bson:"error_detail"`
	StatusCode  int                `bson:"status_code"`
	UserID      string             `bson:"user_id"`
	APIKeyID    string             `bson:"api_key_id"`
	StackTrace  *string            `bson:"stack_trace,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
}

// MetricsHourly represents pre-aggregated metrics for a given hour
type MetricsHourly struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	Hour               time.Time          `bson:"hour"`
	RequestCount       int64              `bson:"request_count"`
	ErrorCount         int64              `bson:"error_count"`
	AvgLatencyMs       float64            `bson:"avg_latency_ms"`
	P50LatencyMs       float64            `bson:"p50_latency_ms"`
	P95LatencyMs       float64            `bson:"p95_latency_ms"`
	P99LatencyMs       float64            `bson:"p99_latency_ms"`
	TotalBytesSent     int64              `bson:"total_bytes_sent"`
	StatusCodeBreak    map[string]int64   `bson:"status_code_break"`
	PathBreakdown      map[string]int64   `bson:"path_breakdown"`
	ErrorTypeBreakdown map[string]int64   `bson:"error_type_breakdown"`
	CreatedAt          time.Time          `bson:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at"`
}

// UserAnalytics represents aggregated stats for a user
type UserAnalytics struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	UserID           string             `bson:"user_id"`
	URLsCreatedTotal int64              `bson:"urls_created_total"`
	APICallsTotal    int64              `bson:"api_calls_total"`
	ErrorCount       int64              `bson:"error_count"`
	LastActive       time.Time          `bson:"last_active"`
	RateLimitHits    int64              `bson:"rate_limit_hits"`
	CreatedAt        time.Time          `bson:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"`
}

// APIKeyAnalytics represents aggregated stats for an API key
type APIKeyAnalytics struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	APIKeyID      string             `bson:"api_key_id"`
	UsageCount    int64              `bson:"usage_count"`
	ErrorCount    int64              `bson:"error_count"`
	ErrorRate     float64            `bson:"error_rate"` // percentage 0-100
	LastUsed      time.Time          `bson:"last_used"`
	RateLimitHits int64              `bson:"rate_limit_hits"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

// URLAnalytics represents aggregated stats for a shortened URL
type URLAnalytics struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	ShortURL      string             `bson:"short_url"`
	LongURL       string             `bson:"long_url"`
	RedirectCount int64              `bson:"redirect_count"`
	CreatedBy     string             `bson:"created_by"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}
