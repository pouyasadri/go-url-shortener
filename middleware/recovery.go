package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware recovers from panics and returns proper error responses
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID for correlation
				requestID := ""
				if val, exists := c.Get("request_id"); exists {
					requestID = val.(string)
				}

				// Log the panic with stack trace
				logger := slog.Default().With(
					slog.String("request_id", requestID),
					slog.String("panic", fmt.Sprintf("%v", err)),
				)

				logger.ErrorContext(c.Request.Context(), "panic recovered",
					slog.String("stack_trace", string(debug.Stack())),
				)

				// Return RFC 7807 error response
				if !c.Writer.Written() {
					RespondWithError(c, http.StatusInternalServerError,
						"internal_server_error",
						"An unexpected error occurred. Please try again later.")
				}
			}
		}()
		c.Next()
	}
}
