package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/store"
)

// AuthMiddleware validates API key from Authorization header
// Expected format: Authorization: Bearer sk_live_xxx
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			RespondWithError(c, http.StatusUnauthorized,
				"missing_credentials",
				"Missing Authorization header. Use: Authorization: Bearer <api_key>")
			c.Abort()
			return
		}

		// Parse "Bearer <key>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			RespondWithError(c, http.StatusUnauthorized,
				"invalid_auth_format",
				"Invalid Authorization format. Use: Authorization: Bearer <api_key>")
			c.Abort()
			return
		}

		apiKey := parts[1]
		if apiKey == "" {
			RespondWithError(c, http.StatusUnauthorized,
				"missing_credentials",
				"API key is empty")
			c.Abort()
			return
		}

		// Validate API key
		userID, err := store.ValidateAPIKey(apiKey)
		if err != nil {
			RespondWithError(c, http.StatusUnauthorized,
				"invalid_api_key",
				"The provided API key is invalid or revoked")
			c.Abort()
			return
		}

		// Store user ID in context for later use
		c.Set("user_id", userID)
		c.Set("api_key", apiKey)

		c.Next()
	}
}

// OptionalAuthMiddleware is like AuthMiddleware but doesn't abort if auth is missing
// This is useful for endpoints that work both with and without authentication
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth provided, but that's OK
			c.Next()
			return
		}

		// Parse "Bearer <key>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, ignore and continue
			c.Next()
			return
		}

		apiKey := parts[1]
		if apiKey == "" {
			// Empty key, ignore and continue
			c.Next()
			return
		}

		// Try to validate API key
		userID, err := store.ValidateAPIKey(apiKey)
		if err != nil {
			// Invalid key, ignore and continue
			c.Next()
			return
		}

		// Store user ID in context if validation succeeded
		c.Set("user_id", userID)
		c.Set("api_key", apiKey)

		c.Next()
	}
}
