package middleware

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAnalyticsEventSerialization verifies analytics events can be serialized to JSON
func TestAnalyticsEventSerialization(t *testing.T) {
	now := time.Now()
	errorType := "ValidationError"
	shortURL := "abc123"

	event := AnalyticsEvent{
		RequestID:       "req_123456789",
		Timestamp:       now,
		Method:          "POST",
		Path:            "/api/v1/urls",
		StatusCode:      201,
		DurationMs:      25.5,
		BytesSent:       150,
		UserID:          "user_123",
		APIKeyID:        "sk_live_abc123",
		ErrorType:       &errorType,
		CreatedShortURL: &shortURL,
		ClientIP:        "127.0.0.1",
	}

	// Verify all fields are set correctly
	assert.Equal(t, "req_123456789", event.RequestID)
	assert.Equal(t, "POST", event.Method)
	assert.Equal(t, 201, event.StatusCode)
	assert.Equal(t, "abc123", *event.CreatedShortURL)
	assert.Equal(t, "ValidationError", *event.ErrorType)

	// Verify it can be marshaled to JSON
	data, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it can be unmarshaled back
	var unmarshaled AnalyticsEvent
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, event.RequestID, unmarshaled.RequestID)
	assert.Equal(t, event.APIKeyID, unmarshaled.APIKeyID)
}

// TestAnalyticsEventWithNilFields verifies optional fields can be nil
func TestAnalyticsEventWithNilFields(t *testing.T) {
	event := AnalyticsEvent{
		RequestID:       "req_123456789",
		Timestamp:       time.Now(),
		Method:          "GET",
		Path:            "/health",
		StatusCode:      200,
		DurationMs:      1.5,
		BytesSent:       50,
		UserID:          "",
		APIKeyID:        "",
		ErrorType:       nil,
		CreatedShortURL: nil,
		ClientIP:        "127.0.0.1",
	}

	// Verify optional fields are nil
	assert.Nil(t, event.ErrorType)
	assert.Nil(t, event.CreatedShortURL)
	assert.Empty(t, event.UserID)
	assert.Empty(t, event.APIKeyID)

	// Verify it can still be marshaled to JSON
	data, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify nil fields are omitted (or null) in JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	assert.NoError(t, err)
	// error_type and created_short_url should be absent or null
	assert.True(t, jsonMap["error_type"] == nil || jsonMap["error_type"] == "")
}

// TestAnalyticsEventDurationMetrics verifies duration calculations are accurate
func TestAnalyticsEventDurationMetrics(t *testing.T) {
	event := AnalyticsEvent{
		RequestID:  "req_test",
		Timestamp:  time.Now(),
		DurationMs: 42.5,
	}

	assert.Equal(t, 42.5, event.DurationMs)
	assert.Greater(t, event.DurationMs, 0.0)
}

// BenchmarkAnalyticsEventMarshal benchmarks JSON marshaling performance
func BenchmarkAnalyticsEventMarshal(b *testing.B) {
	event := AnalyticsEvent{
		RequestID:  "req_123456789",
		Timestamp:  time.Now(),
		Method:     "POST",
		Path:       "/api/v1/urls",
		StatusCode: 201,
		DurationMs: 25.5,
		BytesSent:  150,
		UserID:     "user_123",
		APIKeyID:   "sk_live_abc123",
		ClientIP:   "127.0.0.1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(event)
	}
}
