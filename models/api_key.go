package models

import "time"

// APIKey represents an API key in the system
type APIKey struct {
	// ID is the unique identifier for the API key (internal use)
	ID string `json:"id"`

	// Key is the actual API key that users provide in Authorization header
	// Format: sk_live_<random> or sk_test_<random>
	Key string `json:"key"`

	// HashedKey is the bcrypt hash of the API key (stored in Redis)
	HashedKey string `json:"-"`

	// UserID is the owner of this API key
	UserID string `json:"user_id"`

	// Name is a human-readable name for the API key (e.g., "Production API Key")
	Name string `json:"name"`

	// Status is either "active" or "revoked"
	Status string `json:"status"`

	// CreatedAt is when the API key was created
	CreatedAt time.Time `json:"created_at"`

	// LastUsedAt is when the API key was last used (optional)
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`

	// RevokedAt is when the API key was revoked (optional)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Environment indicates if this is for "test" or "live"
	Environment string `json:"environment"`
}

// APIKeyResponse is the response sent after generating a new API key
// (includes the plain key that's only shown once)
type APIKeyResponse struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"created_at"`
}

// RateLimitInfo holds rate limit state for an API key
type RateLimitInfo struct {
	APIKeyID        string    `json:"api_key_id"`
	UserID          string    `json:"user_id"`
	TokensRemaining int       `json:"tokens_remaining"`
	ResetAt         time.Time `json:"reset_at"`
	LastRequestAt   time.Time `json:"last_request_at"`
}
