package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/pouyasadri/go-url-shortener/store"
	"github.com/redis/go-redis/v9"
)

// RedisCache provides a caching layer using Redis
type RedisCache struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(defaultTTLMinutes int) *RedisCache {
	return &RedisCache{
		client:     store.GetRedisClient(),
		defaultTTL: time.Duration(defaultTTLMinutes) * time.Minute,
	}
}

// Set stores a value in cache with default TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return rc.SetWithTTL(ctx, key, value, rc.defaultTTL)
}

// SetWithTTL stores a value in cache with custom TTL
func (rc *RedisCache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if rc.client == nil {
		return ErrCacheNotAvailable
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	cmd := rc.client.Set(ctx, key, data, ttl)
	if err := cmd.Err(); err != nil {
		slog.Error("Failed to set cache",
			slog.String("key", key),
			slog.String("error", err.Error()))
		return err
	}

	slog.Debug("Cache set",
		slog.String("key", key),
		slog.Duration("ttl", ttl))
	return nil
}

// Get retrieves a value from cache and unmarshals it
func (rc *RedisCache) Get(ctx context.Context, key string, value interface{}) error {
	if rc.client == nil {
		return ErrCacheNotAvailable
	}

	cmd := rc.client.Get(ctx, key)
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		slog.Error("Failed to get cache",
			slog.String("key", key),
			slog.String("error", err.Error()))
		return err
	}

	data, err := cmd.Bytes()
	if err != nil {
		return fmt.Errorf("failed to get cache value: %w", err)
	}

	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	slog.Debug("Cache hit", slog.String("key", key))
	return nil
}

// Delete removes a key from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	if rc.client == nil {
		return ErrCacheNotAvailable
	}

	cmd := rc.client.Del(ctx, key)
	if err := cmd.Err(); err != nil {
		slog.Error("Failed to delete cache",
			slog.String("key", key),
			slog.String("error", err.Error()))
		return err
	}

	slog.Debug("Cache deleted", slog.String("key", key))
	return nil
}

// Exists checks if a key exists in cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if rc.client == nil {
		return false, ErrCacheNotAvailable
	}

	cmd := rc.client.Exists(ctx, key)
	if err := cmd.Err(); err != nil {
		return false, err
	}

	return cmd.Val() > 0, nil
}

// DeletePattern removes all keys matching a pattern
func (rc *RedisCache) DeletePattern(ctx context.Context, pattern string) (int64, error) {
	if rc.client == nil {
		return 0, ErrCacheNotAvailable
	}

	// Use SCAN to find keys matching pattern
	var keys []string
	iter := rc.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) == 0 {
		return 0, nil
	}

	// Delete all found keys
	cmd := rc.client.Del(ctx, keys...)
	if err := cmd.Err(); err != nil {
		slog.Error("Failed to delete cache pattern",
			slog.String("pattern", pattern),
			slog.String("error", err.Error()))
		return 0, err
	}

	count := cmd.Val()
	slog.Debug("Cache pattern deleted",
		slog.String("pattern", pattern),
		slog.Int64("count", count))
	return count, nil
}

// Error definitions
var (
	ErrCacheNotAvailable = fmt.Errorf("cache not available")
	ErrCacheMiss         = fmt.Errorf("cache miss")
)

// Cache key constants
const (
	CachePrefixDashboard      = "dashboard:"
	CacheKeyDashboardOverview = CachePrefixDashboard + "overview"
	CacheKeyDashboardRequests = CachePrefixDashboard + "requests"
	CacheKeyDashboardUsers    = CachePrefixDashboard + "users"
	CacheKeyDashboardAPIKeys  = CachePrefixDashboard + "api-keys"
	CacheKeyDashboardErrors   = CachePrefixDashboard + "errors"
	CacheKeyDashboardURLs     = CachePrefixDashboard + "urls"
)
