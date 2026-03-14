package config

import (
	"os"
	"strconv"
	"time"
)

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	// RequireHTTPS enforces HTTPS for all requests (except localhost)
	RequireHTTPS bool

	// RateLimitPerDay is the number of requests allowed per API key per day
	RateLimitPerDay int

	// RateLimitWindow is the duration for rate limit reset
	RateLimitWindow time.Duration

	// APIKeyPrefix is the prefix for generated keys (sk_live_ or sk_test_)
	APIKeyPrefix string

	// MaxURLLength is the maximum allowed length for a URL
	MaxURLLength int

	// AdminAPIKeyHeader is the header name for admin authentication
	AdminAPIKeyHeader string

	// AdminAPIKey is the actual admin key (if required)
	AdminAPIKey string
}

// LoadSecurityConfig loads configuration from environment variables
func LoadSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RequireHTTPS:      getEnvBool("REQUIRE_HTTPS", false),
		RateLimitPerDay:   getEnvInt("RATE_LIMIT_PER_DAY", 1000),
		RateLimitWindow:   time.Duration(getEnvInt("RATE_LIMIT_WINDOW_HOURS", 24)) * time.Hour,
		APIKeyPrefix:      getEnv("API_KEY_PREFIX", "sk_live_"),
		MaxURLLength:      getEnvInt("MAX_URL_LENGTH", 2048),
		AdminAPIKeyHeader: "X-Admin-Key",
		AdminAPIKey:       getEnv("ADMIN_API_KEY", ""),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}
