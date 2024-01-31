package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/handler"
	"github.com/pouyasadri/go-url-shortener/store"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSONP(200, gin.H{
			"message": "Welcome to the URL Shortener API",
		})
	})

	r.POST("/create-short-url", func(c *gin.Context) {
		handler.CreateShortURL(c)
	})

	r.GET("/:shortUrl", func(c *gin.Context) {
		handler.HandleShortURLRedirect(c)
	})

	store.InitializeStore()

	err := r.Run(":8080")
	if err != nil {
		panic(fmt.Sprintf("Error while running server: %v", err))
	}
}
