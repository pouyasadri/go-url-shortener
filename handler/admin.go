package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pouyasadri/go-url-shortener/middleware"
	"github.com/pouyasadri/go-url-shortener/models"
	"github.com/pouyasadri/go-url-shortener/store"
)

// GenerateAPIKeyRequest is the request body for generating a new API key
type GenerateAPIKeyRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Environment string `json:"environment" binding:"required"`
}

// AdminAuthMiddleware checks for admin API key
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminKey := os.Getenv("ADMIN_API_KEY")
		if adminKey == "" {
			middleware.RespondWithError(c, http.StatusForbidden,
				"forbidden",
				"Admin API key not configured")
			c.Abort()
			return
		}

		providedKey := c.GetHeader("X-Admin-Key")
		if providedKey != adminKey {
			middleware.RespondWithError(c, http.StatusUnauthorized,
				"invalid_admin_key",
				"Invalid or missing admin API key")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateAPIKey creates a new API key for a user
// POST /admin/api-keys/generate
func GenerateAPIKey(c *gin.Context) {
	var req GenerateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest,
			"invalid_request",
			"Invalid request body: "+err.Error())
		return
	}

	// Validate environment
	if req.Environment != "test" && req.Environment != "live" {
		middleware.RespondWithError(c, http.StatusBadRequest,
			"invalid_request",
			"Environment must be 'test' or 'live'")
		return
	}

	// Generate API key
	prefix := "sk_test_"
	if req.Environment == "live" {
		prefix = "sk_live_"
	}

	apiKeyStr, err := store.GenerateAPIKey(prefix)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError,
			"internal_server_error",
			"Failed to generate API key")
		return
	}

	// Create API key object
	apiKey := &models.APIKey{
		ID:          "key_" + uuid.New().String()[:12],
		Key:         apiKeyStr,
		UserID:      req.UserID,
		Name:        req.Name,
		Status:      "active",
		CreatedAt:   time.Now(),
		Environment: req.Environment,
	}

	// Store in Redis
	if err := store.StoreAPIKey(apiKey); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError,
			"internal_server_error",
			"Failed to store API key")
		return
	}

	// Return response (only show the plain key once)
	response := &models.APIKeyResponse{
		ID:          apiKey.ID,
		Key:         apiKey.Key,
		UserID:      apiKey.UserID,
		Name:        apiKey.Name,
		Environment: apiKey.Environment,
		CreatedAt:   apiKey.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// ListAPIKeys returns all API keys for a user
// GET /admin/api-keys?user_id=<user_id>
func ListAPIKeys(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		middleware.RespondWithError(c, http.StatusBadRequest,
			"invalid_request",
			"Missing required query parameter: user_id")
		return
	}

	keyIDs, err := store.GetUserAPIKeys(userID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError,
			"internal_server_error",
			"Failed to retrieve API keys")
		return
	}

	keys := make([]*models.APIKey, 0)
	for _, keyID := range keyIDs {
		key, err := store.GetAPIKeyMetadata(keyID)
		if err != nil {
			continue
		}
		// Don't include the plain key in list responses
		keys = append(keys, key)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"keys":    keys,
	})
}

// RevokeAPIKey revokes an existing API key
// POST /admin/api-keys/revoke
func RevokeAPIKey(c *gin.Context) {
	var req struct {
		KeyID string `json:"key_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest,
			"invalid_request",
			"Invalid request body: "+err.Error())
		return
	}

	if err := store.RevokeAPIKey(req.KeyID); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError,
			"internal_server_error",
			"Failed to revoke API key")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key revoked successfully",
		"key_id":  req.KeyID,
	})
}
