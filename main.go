package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSONP(200, gin.H{
			"message": "Hey Go URL Shortener",
		})
	})

	err := r.Run(":8080")
	if err != nil {
		panic(fmt.Sprintf("Error while running server: %v", err))
	}
}
