package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/models"
	"github.com/pouyasadri/go-url-shortener/store"
)

// AnalyticsEvent represents an event to be published to Redis for async processing
type AnalyticsEvent struct {
	RequestID       string    `json:"request_id"`
	Timestamp       time.Time `json:"timestamp"`
	Method          string    `json:"method"`
	Path            string    `json:"path"`
	StatusCode      int       `json:"status_code"`
	DurationMs      float64   `json:"duration_ms"`
	BytesSent       int64     `json:"bytes_sent"`
	UserID          string    `json:"user_id"`
	APIKeyID        string    `json:"api_key_id"`
	ErrorType       *string   `json:"error_type,omitempty"`
	CreatedShortURL *string   `json:"created_short_url,omitempty"`
	ClientIP        string    `json:"client_ip"`
}

// AnalyticsMiddleware captures request/response metrics and publishes them to Redis
// This middleware should be added early in the middleware chain to capture all requests
func AnalyticsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Extract request ID (set by RequestIDMiddleware)
		requestID := ""
		if val, exists := c.Get("request_id"); exists {
			requestID = val.(string)
		}

		// Extract user ID and API key ID from context (set by AuthMiddleware)
		userID := ""
		if val, exists := c.Get("user_id"); exists {
			userID = val.(string)
		}

		apiKeyID := ""
		if val, exists := c.Get("api_key_id"); exists {
			apiKeyID = val.(string)
		}

		// Process request
		c.Next()

		// Calculate metrics
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		bytesWritten := int64(c.Writer.Size())

		// Extract error type if present
		var errorType *string
		if statusCode >= 400 {
			errValue := c.GetString("error_type")
			if errValue != "" {
				errorType = &errValue
			}
		}

		// Extract created short URL if present (set by CreateShortURL handler)
		var createdURL *string
		if val, exists := c.Get("created_short_url"); exists {
			if url, ok := val.(string); ok {
				createdURL = &url
			}
		}

		// Create analytics event
		event := AnalyticsEvent{
			RequestID:       requestID,
			Timestamp:       startTime,
			Method:          c.Request.Method,
			Path:            c.Request.URL.Path,
			StatusCode:      statusCode,
			DurationMs:      float64(duration.Milliseconds()),
			BytesSent:       bytesWritten,
			UserID:          userID,
			APIKeyID:        apiKeyID,
			ErrorType:       errorType,
			CreatedShortURL: createdURL,
			ClientIP:        c.ClientIP(),
		}

		// Publish event to Redis asynchronously (non-blocking)
		go publishAnalyticsEvent(event)
	}
}

// publishAnalyticsEvent publishes an analytics event to Redis pub/sub
// It's run in a goroutine to avoid blocking request processing
func publishAnalyticsEvent(event AnalyticsEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Serialize event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		slog.Error("Failed to marshal analytics event", slog.String("error", err.Error()))
		return
	}

	// Publish to Redis channel
	client := store.GetRedisClient()
	if client == nil {
		slog.Error("Redis client not available for analytics")
		return
	}

	cmd := client.Publish(ctx, "analytics:events", eventJSON)
	if err := cmd.Err(); err != nil {
		slog.Error("Failed to publish analytics event",
			slog.String("error", err.Error()),
			slog.String("request_id", event.RequestID))
		return
	}

	// Log only in debug mode (verbose)
	// slog.Debug("Analytics event published",
	// 	slog.String("request_id", event.RequestID),
	// 	slog.Int("subscribers", int(cmd.Val())))
}

// SaveAnalyticsEventToDB saves a request event directly to MongoDB (for testing/fallback)
// Used by worker service when processing batched events
func SaveAnalyticsEventToDB(ctx context.Context, event models.RequestEvent) error {
	// This will be implemented in the Day 2 analytics repository package
	// For now, this is a placeholder that the worker service will use
	_ = ctx
	_ = event
	return nil
}
