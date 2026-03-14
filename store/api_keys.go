package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pouyasadri/go-url-shortener/models"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis key prefixes for API keys
	apiKeyPrefix         = "api_key:"          // api_key:<hashed_key> -> userID
	apiKeyMetaPrefix     = "api_key_meta:"     // api_key_meta:<key_id> -> JSON
	userAPIKeysPrefix    = "user_api_keys:"    // user_api_keys:<user_id> -> SET of key IDs
	rateLimitPrefix      = "rate_limit:"       // rate_limit:<api_key_id> -> tokens remaining
	rateLimitResetPrefix = "rate_limit_reset:" // rate_limit_reset:<api_key_id> -> reset timestamp
)

// GenerateAPIKey creates a new API key with the given prefix
func GenerateAPIKey(prefix string) (string, error) {
	// Generate 24 random bytes -> 48 hex chars
	randomBytes := make([]byte, 24)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	key := prefix + hex.EncodeToString(randomBytes)
	return key, nil
}

// HashAPIKey creates a SHA256 hash of the API key
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// StoreAPIKey saves a new API key to Redis
func StoreAPIKey(key *models.APIKey) error {
	ctx := context.Background()

	// Hash the key before storing
	hashedKey := HashAPIKey(key.Key)
	key.HashedKey = hashedKey

	// Store the key -> userID mapping (for quick validation)
	// key format: "api_key:<hashed_key>" -> "<user_id>"
	ttl := 365 * 24 * time.Hour // 1 year expiry for safety
	err := storeService.redisClient.Set(ctx,
		fmt.Sprintf("%s%s", apiKeyPrefix, hashedKey),
		key.UserID,
		ttl,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to store API key mapping: %w", err)
	}

	// Store the metadata JSON: "api_key_meta:<id>" -> JSON
	// Format: id|user_id|name|status|created_at|environment
	metadata := fmt.Sprintf("%s|%s|%s|%s|%d|%s",
		key.ID,
		key.UserID,
		key.Name,
		key.Status,
		key.CreatedAt.Unix(),
		key.Environment,
	)

	err = storeService.redisClient.Set(ctx,
		fmt.Sprintf("%s%s", apiKeyMetaPrefix, key.ID),
		metadata,
		ttl,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to store API key metadata: %w", err)
	}

	// Add key ID to user's set of keys
	err = storeService.redisClient.SAdd(ctx,
		fmt.Sprintf("%s%s", userAPIKeysPrefix, key.UserID),
		key.ID,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to add key to user's key set: %w", err)
	}

	slog.InfoContext(ctx, "API key stored",
		slog.String("key_id", key.ID),
		slog.String("user_id", key.UserID),
		slog.String("name", key.Name),
	)

	return nil
}

// ValidateAPIKey checks if a key is valid and returns the user ID
func ValidateAPIKey(key string) (string, error) {
	ctx := context.Background()

	hashedKey := HashAPIKey(key)

	// Retrieve user ID from Redis
	userID, err := storeService.redisClient.Get(ctx,
		fmt.Sprintf("%s%s", apiKeyPrefix, hashedKey),
	).Result()

	if err == redis.Nil {
		return "", fmt.Errorf("invalid API key")
	}
	if err != nil {
		return "", fmt.Errorf("failed to validate API key: %w", err)
	}

	return userID, nil
}

// GetAPIKeyMetadata retrieves metadata for a specific API key ID
func GetAPIKeyMetadata(keyID string) (*models.APIKey, error) {
	ctx := context.Background()

	metadata, err := storeService.redisClient.Get(ctx,
		fmt.Sprintf("%s%s", apiKeyMetaPrefix, keyID),
	).Result()

	if err == redis.Nil {
		return nil, fmt.Errorf("API key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve API key metadata: %w", err)
	}

	// Parse metadata: id|user_id|name|status|created_at|environment
	parts := strings.Split(metadata, "|")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid metadata format")
	}

	createdAtUnix := int64(0)
	if len(parts) > 4 {
		fmt.Sscanf(parts[4], "%d", &createdAtUnix)
	}

	return &models.APIKey{
		ID:          parts[0],
		UserID:      parts[1],
		Name:        parts[2],
		Status:      parts[3],
		CreatedAt:   time.Unix(createdAtUnix, 0),
		Environment: parts[5],
	}, nil
}

// GetUserAPIKeys retrieves all API key IDs for a user
func GetUserAPIKeys(userID string) ([]string, error) {
	ctx := context.Background()

	keyIDs, err := storeService.redisClient.SMembers(ctx,
		fmt.Sprintf("%s%s", userAPIKeysPrefix, userID),
	).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user API keys: %w", err)
	}

	return keyIDs, nil
}

// RevokeAPIKey marks an API key as revoked
func RevokeAPIKey(keyID string) error {
	ctx := context.Background()

	// Get metadata
	key, err := GetAPIKeyMetadata(keyID)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	// Update status to "revoked"
	now := time.Now()
	metadata := fmt.Sprintf("%s|%s|%s|%s|%d|%s",
		key.ID,
		key.UserID,
		key.Name,
		"revoked",
		key.CreatedAt.Unix(),
		key.Environment,
	)

	err = storeService.redisClient.Set(ctx,
		fmt.Sprintf("%s%s", apiKeyMetaPrefix, keyID),
		metadata,
		365*24*time.Hour,
	).Err()

	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	slog.InfoContext(ctx, "API key revoked",
		slog.String("key_id", keyID),
		slog.String("user_id", key.UserID),
		slog.Time("revoked_at", now),
	)

	return nil
}
