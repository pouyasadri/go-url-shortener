package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/shortener"
	"github.com/pouyasadri/go-url-shortener/store"
	"net/http"
)

type URLCreationRequest struct {
	LongURL string `json:"long_url" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
}

func CreateShortURL(c *gin.Context) {
	var createShortURLRequest URLCreationRequest
	if err := c.ShouldBindJSON(&createShortURLRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortUrl := shortener.GenerateShortLink(createShortURLRequest.LongURL, createShortURLRequest.UserID)
	store.SaveUrlMapping(shortUrl, createShortURLRequest.LongURL, createShortURLRequest.UserID)

	host := "http://localhost:8080/"
	c.JSON(200, gin.H{
		"message":   "Short URL created successfully",
		"short_url": host + shortUrl,
	})
}

func HandleShortURLRedirect(c *gin.Context) {
	shortUrl := c.Param("shortUrl")
	longUrl := store.RetrieveLongUrl(shortUrl)
	c.Redirect(http.StatusFound, longUrl)
}
