package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testTime = time.Now()

// TestGetTimeframeHours validates timeframe to hours conversion
func TestGetTimeframeHours(t *testing.T) {
	tests := []struct {
		timeframe string
		expected  int
	}{
		{"24h", 24},
		{"7d", 168},     // 24 * 7
		{"30d", 720},    // 24 * 30
		{"invalid", 24}, // defaults to 24h
	}

	for _, tt := range tests {
		t.Run(tt.timeframe, func(t *testing.T) {
			result := GetTimeframeHours(tt.timeframe)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDashboardResponseStructure validates response structure
func TestDashboardResponseStructure(t *testing.T) {
	resp := DashboardResponse{
		Status:    "ok",
		Timestamp: testTime,
		Timeframe: "24h",
		Data:      map[string]interface{}{},
	}

	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "24h", resp.Timeframe)
	assert.NotNil(t, resp.Data)
}

// TestOverviewDataStructure validates overview data structure
func TestOverviewDataStructure(t *testing.T) {
	overview := OverviewData{
		TotalRequests:   100,
		TotalErrors:     5,
		ErrorRate:       5.0,
		AvgLatencyMs:    25.0,
		ActiveAPIKeys:   10,
		TopUsers:        make([]UserSummary, 0),
		StatusBreakdown: make(map[string]interface{}),
	}

	assert.Equal(t, int64(100), overview.TotalRequests)
	assert.Equal(t, int64(5), overview.TotalErrors)
	assert.Equal(t, 5.0, overview.ErrorRate)
	assert.Equal(t, 0, len(overview.TopUsers))
}

// TestRequestMetricsDataStructure validates request metrics structure
func TestRequestMetricsDataStructure(t *testing.T) {
	metrics := RequestMetricsData{
		TotalRequests:   1000,
		TotalErrors:     50,
		AvgLatencyMs:    50.5,
		P50LatencyMs:    40.0,
		P95LatencyMs:    80.0,
		P99LatencyMs:    100.0,
		StatusCodeBreak: make(map[string]int64),
		TotalBytesSent:  5000,
	}

	assert.Equal(t, int64(1000), metrics.TotalRequests)
	assert.Equal(t, float64(50.5), metrics.AvgLatencyMs)
	assert.Equal(t, float64(100.0), metrics.P99LatencyMs)
}

// TestUserSummaryStructure validates user summary structure
func TestUserSummaryStructure(t *testing.T) {
	user := UserSummary{
		UserID:        "user_123",
		URLsCreated:   10,
		APICallsTotal: 100,
		ErrorCount:    5,
		LastActive:    testTime,
		RateLimitHits: 0,
	}

	assert.Equal(t, "user_123", user.UserID)
	assert.Equal(t, int64(10), user.URLsCreated)
	assert.Equal(t, int64(100), user.APICallsTotal)
}

// TestAPIKeySummaryStructure validates API key summary structure
func TestAPIKeySummaryStructure(t *testing.T) {
	key := APIKeySummary{
		APIKeyID:      "sk_live_abc123",
		UsageCount:    500,
		ErrorCount:    10,
		ErrorRate:     2.0,
		LastUsed:      testTime,
		RateLimitHits: 5,
	}

	assert.Equal(t, "sk_live_abc123", key.APIKeyID)
	assert.Equal(t, int64(500), key.UsageCount)
	assert.Equal(t, 2.0, key.ErrorRate)
}

// TestErrorLogEntryStructure validates error log entry structure
func TestErrorLogEntryStructure(t *testing.T) {
	entry := ErrorLogEntry{
		RequestID:   "req_123",
		Timestamp:   testTime,
		ErrorType:   "ValidationError",
		StatusCode:  400,
		UserID:      "user_123",
		APIKeyID:    "sk_live_abc",
		ErrorDetail: "Invalid URL format",
	}

	assert.Equal(t, "req_123", entry.RequestID)
	assert.Equal(t, "ValidationError", entry.ErrorType)
	assert.Equal(t, 400, entry.StatusCode)
}

// TestURLAnalyticsSummaryStructure validates URL analytics structure
func TestURLAnalyticsSummaryStructure(t *testing.T) {
	url := URLAnalyticsSummary{
		ShortURL:      "abc123",
		LongURL:       "https://example.com/very/long/url",
		RedirectCount: 1000,
		CreatedBy:     "user_123",
		CreatedAt:     testTime,
		LastUsed:      testTime,
	}

	assert.Equal(t, "abc123", url.ShortURL)
	assert.Equal(t, int64(1000), url.RedirectCount)
	assert.Equal(t, "https://example.com/very/long/url", url.LongURL)
}
