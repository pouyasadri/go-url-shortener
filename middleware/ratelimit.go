package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pouyasadri/go-url-shortener/config"
	"github.com/pouyasadri/go-url-shortener/store"
	"github.com/redis/go-redis/v9"
)

const (
	rateLimitTokensKey = "rate_limit_tokens:"
	rateLimitResetKey  = "rate_limit_reset:"
)

// RateLimitMiddleware enforces per-API-key rate limiting using token bucket algorithm
func RateLimitMiddleware(cfg *config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from context (set by AuthMiddleware)
		apiKeyVal, exists := c.Get("api_key")
		if !exists {
			// No auth, skip rate limiting
			c.Next()
			return
		}

		apiKey := apiKeyVal.(string)
		userIDVal, _ := c.Get("user_id")
		userID := userIDVal.(string)

		ctx := context.Background()
		redisClient := store.GetRedisClient()

		// Check rate limit
		tokensRemaining, resetTime, err := checkRateLimit(ctx, redisClient, apiKey, userID, cfg)
		if err != nil {
			// Log error but don't fail the request
			slog.ErrorContext(ctx, "rate limit check failed",
				slog.String("api_key", apiKey),
				slog.String("error", err.Error()),
			)
			c.Next()
			return
		}

		// Add rate limit info to response headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.RateLimitPerDay))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(tokensRemaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

		// Check if rate limit exceeded
		if tokensRemaining < 0 {
			RespondWithError(c, http.StatusTooManyRequests,
				"rate_limit_exceeded",
				fmt.Sprintf("Rate limit exceeded. Limit: %d requests per day. Reset at: %s",
					cfg.RateLimitPerDay, resetTime.Format(time.RFC3339)))
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit uses a token bucket algorithm stored in Redis
// Returns: (tokens_remaining, reset_time, error)
func checkRateLimit(ctx context.Context, redisClient *redis.Client, apiKey, userID string, cfg *config.SecurityConfig) (int, time.Time, error) {
	tokensKey := rateLimitTokensKey + apiKey
	resetKey := rateLimitResetKey + apiKey

	now := time.Now()

	// Get current tokens and reset time
	tokensStr, err := redisClient.Get(ctx, tokensKey).Result()
	resetStr, resetErr := redisClient.Get(ctx, resetKey).Result()

	var tokens int
	var resetTime time.Time

	// First request or reset expired
	if err == redis.Nil && resetErr == redis.Nil {
		tokens = cfg.RateLimitPerDay
		resetTime = now.Add(cfg.RateLimitWindow)

		// Initialize in Redis
		err := redisClient.Set(ctx, tokensKey, tokens, cfg.RateLimitWindow).Err()
		if err != nil {
			return 0, resetTime, fmt.Errorf("failed to initialize tokens: %w", err)
		}

		err = redisClient.Set(ctx, resetKey, resetTime.Unix(), cfg.RateLimitWindow).Err()
		if err != nil {
			return 0, resetTime, fmt.Errorf("failed to initialize reset time: %w", err)
		}

		tokens--
		return tokens, resetTime, nil
	}

	if err != nil && err != redis.Nil {
		return 0, now, fmt.Errorf("failed to get token count: %w", err)
	}

	if resetErr != nil && resetErr != redis.Nil {
		return 0, now, fmt.Errorf("failed to get reset time: %w", resetErr)
	}

	// Parse reset time
	if resetStr != "" {
		resetUnix, err := strconv.ParseInt(resetStr, 10, 64)
		if err == nil {
			resetTime = time.Unix(resetUnix, 0)
		} else {
			resetTime = now.Add(cfg.RateLimitWindow)
		}
	}

	// Check if window has passed
	if now.After(resetTime) {
		// Reset window
		tokens = cfg.RateLimitPerDay
		resetTime = now.Add(cfg.RateLimitWindow)

		redisClient.Set(ctx, tokensKey, tokens, cfg.RateLimitWindow)
		redisClient.Set(ctx, resetKey, resetTime.Unix(), cfg.RateLimitWindow)

		tokens--
		return tokens, resetTime, nil
	}

	// Normal case: decrement token
	if tokensStr != "" {
		if parsed, err := strconv.Atoi(tokensStr); err == nil {
			tokens = parsed
		}
	}

	tokens--

	// Store updated token count
	err = redisClient.Set(ctx, tokensKey, tokens, cfg.RateLimitWindow).Err()
	if err != nil {
		return tokens, resetTime, fmt.Errorf("failed to update token count: %w", err)
	}

	return tokens, resetTime, nil
}
