package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to every request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request already has ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate: req_<12-char-uuid>
			requestID = "req_" + uuid.New().String()[:12]
		}

		// Store in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
