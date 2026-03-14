package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/config"
	"github.com/pouyasadri/go-url-shortener/handler"
	"github.com/pouyasadri/go-url-shortener/middleware"
	"github.com/pouyasadri/go-url-shortener/store"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Initialize store (Redis)
	store.InitializeStore()

	// Load security config
	securityCfg := config.LoadSecurityConfig()

	r := gin.Default()

	// Global middleware (applied to all routes, in order)
	// 1. Request ID: adds unique ID for tracing
	r.Use(middleware.RequestIDMiddleware())

	// 2. Structured Logger: logs all requests/responses
	r.Use(middleware.LoggerMiddleware())

	// 3. Recovery: catches panics and returns proper error responses
	r.Use(middleware.RecoveryMiddleware())

	// Health check endpoints (no auth required)
	r.GET("/health", handler.HealthCheck)
	r.GET("/ready", handler.ReadinessCheck)

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the URL Shortener API",
		})
	})

	// Admin group (requires admin API key)
	admin := r.Group("/admin")
	admin.Use(handler.AdminAuthMiddleware())
	{
		admin.POST("/api-keys/generate", handler.GenerateAPIKey)
		admin.GET("/api-keys", handler.ListAPIKeys)
		admin.POST("/api-keys/revoke", handler.RevokeAPIKey)
	}

	// API v1 group (requires authentication)
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())                 // 4. Auth: validates API key
	api.Use(middleware.RateLimitMiddleware(securityCfg)) // 5. Rate Limit: per-key rate limiting
	{
		api.POST("/urls", handler.CreateShortURL)
	}

	// Public redirect endpoint (no auth required)
	r.GET("/:shortUrl", handler.HandleShortURLRedirect)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Starting server", slog.String("port", port))
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Error while running server: %v", err)
	}
}
