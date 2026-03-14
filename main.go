package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/handler"
	"github.com/pouyasadri/go-url-shortener/store"
)

func main() {
	store.InitializeStore()

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the URL Shortener API",
		})
	})

	r.POST("/create-short-url", handler.CreateShortURL)
	r.GET("/:shortUrl", handler.HandleShortURLRedirect)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error while running server: %v", err)
	}
}
