package config

import (
	"os"
	"testing"
	"time"
)

// TestLoadSecurityConfigDefaults tests default configuration loading
func TestLoadSecurityConfigDefaults(t *testing.T) {
	// Clear env vars to ensure defaults
	os.Clearenv()

	cfg := LoadSecurityConfig()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"RequireHTTPS", cfg.RequireHTTPS, false},
		{"RateLimitPerDay", cfg.RateLimitPerDay, 1000},
		{"RateLimitWindow", cfg.RateLimitWindow, 24 * time.Hour},
		{"APIKeyPrefix", cfg.APIKeyPrefix, "sk_live_"},
		{"MaxURLLength", cfg.MaxURLLength, 2048},
		{"AdminAPIKeyHeader", cfg.AdminAPIKeyHeader, "X-Admin-Key"},
		{"AdminAPIKey", cfg.AdminAPIKey, ""},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestLoadSecurityConfigCustom tests custom configuration loading
func TestLoadSecurityConfigCustom(t *testing.T) {
	// Set custom env vars
	os.Setenv("REQUIRE_HTTPS", "true")
	os.Setenv("RATE_LIMIT_PER_DAY", "500")
	os.Setenv("RATE_LIMIT_WINDOW_HOURS", "12")
	os.Setenv("API_KEY_PREFIX", "sk_test_")
	os.Setenv("MAX_URL_LENGTH", "4096")
	os.Setenv("ADMIN_API_KEY", "test_admin_key")
	defer os.Clearenv()

	cfg := LoadSecurityConfig()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"RequireHTTPS", cfg.RequireHTTPS, true},
		{"RateLimitPerDay", cfg.RateLimitPerDay, 500},
		{"RateLimitWindow", cfg.RateLimitWindow, 12 * time.Hour},
		{"APIKeyPrefix", cfg.APIKeyPrefix, "sk_test_"},
		{"MaxURLLength", cfg.MaxURLLength, 4096},
		{"AdminAPIKey", cfg.AdminAPIKey, "test_admin_key"},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestLoadSecurityConfigInvalidValues tests handling of invalid environment values
func TestLoadSecurityConfigInvalidValues(t *testing.T) {
	os.Setenv("RATE_LIMIT_PER_DAY", "invalid")
	os.Setenv("RATE_LIMIT_WINDOW_HOURS", "invalid")
	os.Setenv("REQUIRE_HTTPS", "invalid")
	defer os.Clearenv()

	cfg := LoadSecurityConfig()

	// Should use defaults when env values are invalid
	if cfg.RateLimitPerDay != 1000 {
		t.Errorf("RateLimitPerDay: got %d, want 1000 (default)", cfg.RateLimitPerDay)
	}
	if cfg.RateLimitWindow != 24*time.Hour {
		t.Errorf("RateLimitWindow: got %v, want %v (default)", cfg.RateLimitWindow, 24*time.Hour)
	}
	if cfg.RequireHTTPS != false {
		t.Errorf("RequireHTTPS: got %v, want false (default)", cfg.RequireHTTPS)
	}
}
