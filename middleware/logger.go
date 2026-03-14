package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware provides structured JSON logging with request/response context
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Extract request ID from context (set by RequestIDMiddleware)
		requestID := ""
		if val, exists := c.Get("request_id"); exists {
			requestID = val.(string)
		}

		// Create request context with structured fields
		logger := slog.Default().With(
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("remote_addr", c.ClientIP()),
		)

		// Log incoming request
		logger.InfoContext(c.Request.Context(), "incoming request")

		// Process request
		c.Next()

		// Calculate response metrics
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		bytesWritten := c.Writer.Size()

		// Log response with metrics
		logAttrs := []slog.Attr{
			slog.Int("status", statusCode),
			slog.Int("bytes", bytesWritten),
			slog.Float64("duration_ms", float64(duration.Milliseconds())),
			slog.String("request_id", requestID),
		}

		// Use different log levels based on status code
		if statusCode >= 500 {
			logger.LogAttrs(c.Request.Context(), slog.LevelError, "response sent", logAttrs...)
		} else if statusCode >= 400 {
			logger.LogAttrs(c.Request.Context(), slog.LevelWarn, "response sent", logAttrs...)
		} else {
			logger.LogAttrs(c.Request.Context(), slog.LevelInfo, "response sent", logAttrs...)
		}
	}
}
