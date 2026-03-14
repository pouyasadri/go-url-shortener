package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// StorageService is a wrapper around the raw Redis client.
type StorageService struct {
	redisClient *redis.Client
}

var (
	storeService = &StorageService{}
	ctx          = context.Background()
)

// Note that in a real world usage, the cache duration shouldn't have
// an expiration time; an LRU policy config should be set where the
// values that are retrieved less often are purged automatically from
// the cache and stored back in RDBMS whenever the cache is full.
const CacheDuration = 6 * time.Hour

// InitializeStore connects to Redis using environment variables
// (REDIS_ADDR, REDIS_PASSWORD, REDIS_DB) with sensible defaults,
// and returns the initialized store pointer.
func InitializeStore() *StorageService {
	addr := getEnv("REDIS_ADDR", "localhost:6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if parsed, err := strconv.Atoi(dbStr); err == nil {
			db = parsed
		}
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis at %s: %v", addr, err)
	}
	log.Printf("Redis connected successfully: pong = %s", pong)

	storeService.redisClient = redisClient
	return storeService
}

// SaveUrlMapping persists the short URL → original URL mapping in Redis.
func SaveUrlMapping(shortUrl string, originalUrl string, userId string) error {
	// Store the original URL as the value; tag with userId for traceability.
	value := fmt.Sprintf("%s|%s", originalUrl, userId)
	err := storeService.redisClient.Set(ctx, shortUrl, value, CacheDuration).Err()
	if err != nil {
		return fmt.Errorf("failed to save URL mapping (shortUrl=%s): %w", shortUrl, err)
	}
	return nil
}

// RetrieveLongUrl returns the original URL for the given short URL.
// Returns an error if the key does not exist or Redis fails.
func RetrieveLongUrl(shortUrl string) (string, error) {
	value, err := storeService.redisClient.Get(ctx, shortUrl).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("short URL not found: %s", shortUrl)
	}
	if err != nil {
		return "", fmt.Errorf("failed to retrieve URL for shortUrl=%s: %w", shortUrl, err)
	}
	// Value is stored as "originalUrl|userId"; extract the original URL.
	for i := len(value) - 1; i >= 0; i-- {
		if value[i] == '|' {
			return value[:i], nil
		}
	}
	// Fallback: value stored without userId (backwards compatibility).
	return value, nil
}

// GetRedisClient returns the underlying Redis client for direct access
func GetRedisClient() *redis.Client {
	return storeService.redisClient
}

// HealthCheck verifies that Redis is still connected
func HealthCheck(ctx context.Context) error {
	if storeService.redisClient == nil {
		return fmt.Errorf("store not initialized")
	}
	_, err := storeService.redisClient.Ping(ctx).Result()
	return err
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
