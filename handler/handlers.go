package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/shortener"
	"github.com/pouyasadri/go-url-shortener/store"
)

// URLCreationRequest is the expected JSON body for POST /create-short-url.
type URLCreationRequest struct {
	LongURL string `json:"long_url" binding:"required"`
	UserID  string `json:"user_id"  binding:"required"`
}

func CreateShortURL(c *gin.Context) {
	var req URLCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that long_url is a well-formed absolute URL.
	if _, err := url.ParseRequestURI(req.LongURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid long_url: must be a valid URL"})
		return
	}

	shortUrl, err := shortener.GenerateShortLink(req.LongURL, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate short URL"})
		return
	}

	if err := store.SaveUrlMapping(shortUrl, req.LongURL, req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save URL mapping"})
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
