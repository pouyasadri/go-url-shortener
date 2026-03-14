package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/middleware"
	"github.com/pouyasadri/go-url-shortener/shortener"
	"github.com/pouyasadri/go-url-shortener/store"
)

// URLCreationRequest is the expected JSON body for POST /create-short-url or POST /api/v1/urls
type URLCreationRequest struct {
	LongURL string `json:"long_url" binding:"required"`
	Alias   string `json:"alias,omitempty"` // Optional custom alias
}

func CreateShortURL(c *gin.Context) {
	var req URLCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest,
			"invalid_request",
			"Invalid request body: "+err.Error())
		return
	}

	// Get user ID from auth middleware (via API key)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, http.StatusUnauthorized,
			"missing_credentials",
			"User ID not found in context")
		return
	}
	userID := userIDVal.(string)

	// Validate that long_url is a well-formed absolute URL
	if _, err := url.ParseRequestURI(req.LongURL); err != nil {
		middleware.RespondWithError(c, http.StatusBadRequest,
			"invalid_url",
			"Invalid URL format")
		return
	}

	shortUrl, err := shortener.GenerateShortLink(req.LongURL, userID)
	if err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError,
			"internal_server_error",
			"Failed to generate short URL")
		return
	}

	if err := store.SaveUrlMapping(shortUrl, req.LongURL, userID); err != nil {
		middleware.RespondWithError(c, http.StatusInternalServerError,
			"internal_server_error",
			"Failed to save URL mapping")
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	shortURLFull := fmt.Sprintf("%s://%s/%s", scheme, host, shortUrl)

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Short URL created successfully",
		"short_url": shortURLFull,
		"user_id":   userID,
	})
}

func HandleShortURLRedirect(c *gin.Context) {
	shortUrl := c.Param("shortUrl")
	longUrl, err := store.RetrieveLongUrl(shortUrl)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "short URL not found or expired"})
		return
	}
	c.Redirect(http.StatusFound, longUrl)
}
