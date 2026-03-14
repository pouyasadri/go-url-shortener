package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/store"
)

// HealthResponse is the response for health check endpoints
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime,omitempty"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}

var startTime = time.Now()

// HealthCheck is a basic liveness probe endpoint
// GET /health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
	})
}

// ReadinessCheck is a readiness probe that checks Redis connectivity
// GET /ready
func ReadinessCheck(c *gin.Context) {
	ctx := c.Request.Context()

	checks := make(map[string]interface{})
	allHealthy := true

	// Check Redis
	if err := store.HealthCheck(ctx); err != nil {
		checks["redis"] = map[string]string{
			"status": "failed",
			"error":  err.Error(),
		}
		allHealthy = false
	} else {
		checks["redis"] = map[string]string{
			"status": "ok",
		}
	}

	status := "ok"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
		Checks:    checks,
	})
}
