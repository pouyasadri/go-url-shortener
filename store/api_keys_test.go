package store

import (
	"strings"
	"testing"
)

// TestGenerateAPIKey tests API key generation
func TestGenerateAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		prefix    string
		wantError bool
	}{
		{
			name:      "generate_live_key",
			prefix:    "sk_live_",
			wantError: false,
		},
		{
			name:      "generate_test_key",
			prefix:    "sk_test_",
			wantError: false,
		},
		{
			name:      "empty_prefix",
			prefix:    "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateAPIKey(tt.prefix)
			if (err != nil) != tt.wantError {
				t.Errorf("GenerateAPIKey() error = %v, wantErr %v", err, tt.wantError)
				return
			}
			if !tt.wantError && !strings.HasPrefix(key, tt.prefix) {
				t.Errorf("GenerateAPIKey() returned key with wrong prefix. got %s, want prefix %s", key, tt.prefix)
			}
			// Check minimum length (prefix + 48 hex chars)
			expectedMinLength := len(tt.prefix) + 48
			if len(key) < expectedMinLength {
				t.Errorf("GenerateAPIKey() key too short. got len=%d, want min=%d", len(key), expectedMinLength)
			}
		})
	}
}

// TestHashAPIKey tests API key hashing
func TestHashAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wantLen   int
		stability bool // same key should produce same hash
	}{
		{
			name:      "hash_valid_key",
			key:       "sk_live_abc123xyz",
			wantLen:   64, // SHA256 hex = 64 chars
			stability: true,
		},
		{
			name:      "hash_empty_key",
			key:       "",
			wantLen:   64,
			stability: true,
		},
		{
			name:      "hash_long_key",
			key:       "sk_live_" + strings.Repeat("x", 1000),
			wantLen:   64,
			stability: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashAPIKey(tt.key)

			// Check length
			if len(hash) != tt.wantLen {
				t.Errorf("HashAPIKey() returned wrong length. got %d, want %d", len(hash), tt.wantLen)
			}

			// Check stability (same input = same output)
			if tt.stability {
				hash2 := HashAPIKey(tt.key)
				if hash != hash2 {
					t.Errorf("HashAPIKey() not stable. hash1=%s, hash2=%s", hash, hash2)
				}
			}

			// Check it's hex encoded
			for _, c := range hash {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("HashAPIKey() returned non-hex character: %c", c)
				}
			}
		})
	}
}

// TestHashAPIKeyUniqueness tests that different keys produce different hashes
func TestHashAPIKeyUniqueness(t *testing.T) {
	key1, _ := GenerateAPIKey("sk_live_")
	key2, _ := GenerateAPIKey("sk_live_")

	hash1 := HashAPIKey(key1)
	hash2 := HashAPIKey(key2)

	if hash1 == hash2 {
		t.Errorf("Different keys produced same hash. key1=%s, key2=%s, hash=%s", key1, key2, hash1)
	}
}
