package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/user/ccc/config"
)

// setupTestConfig creates a test configuration.
func setupTestConfig(t *testing.T) *config.Config {
	t.Helper()

	return &config.Config{
		Settings: config.Settings{
			AlwaysThinkingEnabled: true,
			Env: config.Env{
				"API_TIMEOUT":      "30000",
				"DISABLE_TELEMETRY": "1",
			},
		},
		CurrentProvider: "kimi",
		Providers: map[string]config.ProviderConfig{
			"kimi": {
				Env: config.Env{
					"ANTHROPIC_BASE_URL":    "https://api.moonshot.cn/anthropic",
					"ANTHROPIC_AUTH_TOKEN":  "sk-kimi-xxx",
					"ANTHROPIC_MODEL":       "kimi-k2-thinking",
					"ANTHROPIC_SMALL_FAST_MODEL": "kimi-k2-0905-preview",
				},
			},
			"glm": {
				Env: config.Env{
					"ANTHROPIC_BASE_URL":    "https://open.bigmodel.cn/api/anthropic",
					"ANTHROPIC_AUTH_TOKEN":  "sk-glm-xxx",
					"ANTHROPIC_MODEL":       "glm-4.7",
				},
			},
		},
	}
}

func setupTestDir(t *testing.T) func() {
	t.Helper()

	// Save original function
	originalFunc := config.GetDirFunc

	// Set to temp directory
	tmpDir := t.TempDir()
	config.GetDirFunc = func() string {
		return tmpDir
	}

	return func() {
		config.GetDirFunc = originalFunc
	}
}

func TestSwitch(t *testing.T) {
	t.Run("switch to existing provider", func(t *testing.T) {
		cleanup := setupTestDir(t)
		defer cleanup()

		cfg := setupTestConfig(t)

		// Save initial config
		if err := config.Save(cfg); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Switch to glm
		settings, err := Switch(cfg, "glm")
		if err != nil {
			t.Fatalf("Switch() error = %v", err)
		}

		// Verify merged settings
		if settings.Env["ANTHROPIC_BASE_URL"] != "https://open.bigmodel.cn/api/anthropic" {
			t.Errorf("BASE_URL = %s, want glm URL", settings.Env["ANTHROPIC_BASE_URL"])
		}
		// Base env should be preserved
		if settings.Env["API_TIMEOUT"] != "30000" {
			t.Errorf("API_TIMEOUT = %s, want 30000", settings.Env["API_TIMEOUT"])
		}

		// Verify current_provider updated
		if cfg.CurrentProvider != "glm" {
			t.Errorf("CurrentProvider = %s, want glm", cfg.CurrentProvider)
		}

		// Verify settings file was created
		settingsPath := config.GetSettingsPath("glm")
		if _, err := os.Stat(settingsPath); err != nil {
			t.Errorf("Settings file should exist at %s", settingsPath)
		}
	})

	t.Run("switch to non-existing provider", func(t *testing.T) {
		cleanup := setupTestDir(t)
		defer cleanup()

		cfg := setupTestConfig(t)

		_, err := Switch(cfg, "unknown")
		if err == nil {
			t.Fatal("Switch() should error for unknown provider")
		}
		if !strings.Contains(err.Error(), "provider 'unknown' not found") {
			t.Errorf("Error message should mention 'provider 'unknown' not found', got: %v", err)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		_, err := Switch(nil, "kimi")
		if err == nil {
			t.Fatal("Switch() should error for nil config")
		}
	})
}

func TestGetAuthToken(t *testing.T) {
	tests := []struct {
		name     string
		settings *config.Settings
		want     string
	}{
		{
			name: "token exists",
			settings: &config.Settings{
				Env: config.Env{
					"ANTHROPIC_AUTH_TOKEN": "sk-xxx",
				},
			},
			want: "sk-xxx",
		},
		{
			name:     "token does not exist",
			settings: &config.Settings{},
			want:     "PLEASE_SET_ANTHROPIC_AUTH_TOKEN",
		},
		{
			name:     "nil settings",
			settings: nil,
			want:     "PLEASE_SET_ANTHROPIC_AUTH_TOKEN",
		},
		{
			name: "token is empty string",
			settings: &config.Settings{
				Env: config.Env{
					"ANTHROPIC_AUTH_TOKEN": "",
				},
			},
			want: "PLEASE_SET_ANTHROPIC_AUTH_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAuthToken(tt.settings)
			if got != tt.want {
				t.Errorf("GetAuthToken() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestFormatProviderName(t *testing.T) {
	tests := []struct {
		name            string
		providerName    string
		currentProvider string
		want            string
	}{
		{
			name:            "current provider",
			providerName:    "kimi",
			currentProvider: "kimi",
			want:            "kimi (current)",
		},
		{
			name:            "not current provider",
			providerName:    "glm",
			currentProvider: "kimi",
			want:            "glm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatProviderName(tt.providerName, tt.currentProvider)
			if got != tt.want {
				t.Errorf("FormatProviderName() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestListProviders(t *testing.T) {
	t.Run("list providers", func(t *testing.T) {
		cfg := setupTestConfig(t)

		providers := ListProviders(cfg)
		if len(providers) != 2 {
			t.Errorf("ListProviders() returned %d providers, want 2", len(providers))
		}
	})

	t.Run("nil config", func(t *testing.T) {
		providers := ListProviders(nil)
		if len(providers) != 0 {
			t.Errorf("ListProviders(nil) returned %d providers, want 0", len(providers))
		}
	})

	t.Run("empty providers", func(t *testing.T) {
		cfg := &config.Config{
			Providers: map[string]config.ProviderConfig{},
		}
		providers := ListProviders(cfg)
		if len(providers) != 0 {
			t.Errorf("ListProviders() with empty config returned %d providers, want 0", len(providers))
		}
	})
}

func TestValidateProvider(t *testing.T) {
	t.Run("valid provider", func(t *testing.T) {
		cfg := setupTestConfig(t)

		err := ValidateProvider(cfg, "kimi")
		if err != nil {
			t.Errorf("ValidateProvider() error = %v", err)
		}
	})

	t.Run("invalid provider", func(t *testing.T) {
		cfg := setupTestConfig(t)

		err := ValidateProvider(cfg, "unknown")
		if err == nil {
			t.Fatal("ValidateProvider() should error for unknown provider")
		}
		if !strings.Contains(err.Error(), "provider 'unknown' not found") {
			t.Errorf("Error message should mention 'provider 'unknown' not found', got: %v", err)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		err := ValidateProvider(nil, "kimi")
		if err == nil {
			t.Fatal("ValidateProvider() should error for nil config")
		}
	})
}

func TestGetDefaultProvider(t *testing.T) {
	t.Run("returns first provider", func(t *testing.T) {
		cfg := setupTestConfig(t)

		provider := GetDefaultProvider(cfg)
		if provider != "kimi" {
			t.Errorf("GetDefaultProvider() = %s, want kimi", provider)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		provider := GetDefaultProvider(nil)
		if provider != "" {
			t.Errorf("GetDefaultProvider(nil) = %s, want empty string", provider)
		}
	})

	t.Run("empty providers", func(t *testing.T) {
		cfg := &config.Config{
			Providers: map[string]config.ProviderConfig{},
		}
		provider := GetDefaultProvider(cfg)
		if provider != "" {
			t.Errorf("GetDefaultProvider() with empty config = %s, want empty string", provider)
		}
	})
}

func TestGetCurrentProvider(t *testing.T) {
	t.Run("returns current provider", func(t *testing.T) {
		cfg := setupTestConfig(t)

		provider := GetCurrentProvider(cfg)
		if provider != "kimi" {
			t.Errorf("GetCurrentProvider() = %s, want kimi", provider)
		}
	})

	t.Run("current provider not set, returns first", func(t *testing.T) {
		cfg := setupTestConfig(t)
		cfg.CurrentProvider = ""

		provider := GetCurrentProvider(cfg)
		if provider != "kimi" {
			t.Errorf("GetCurrentProvider() with empty current = %s, want kimi", provider)
		}
	})

	t.Run("nil config", func(t *testing.T) {
		provider := GetCurrentProvider(nil)
		if provider != "" {
			t.Errorf("GetCurrentProvider(nil) = %s, want empty string", provider)
		}
	})

	t.Run("current provider invalid, falls back to first", func(t *testing.T) {
		cfg := setupTestConfig(t)
		cfg.CurrentProvider = "invalid"

		provider := GetCurrentProvider(cfg)
		if provider != "kimi" {
			t.Errorf("GetCurrentProvider() with invalid current = %s, want kimi", provider)
		}
	})
}

func TestShortenError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		maxLength  int
		wantContains string
	}{
		{
			name:       "nil error",
			err:        nil,
			maxLength:  40,
			wantContains: "",
		},
		{
			name:       "short error",
			err:        fmt.Errorf("simple error"),
			maxLength:  40,
			wantContains: "simple error",
		},
		{
			name:       "long error with colon",
			err:        fmt.Errorf("failed to read config file: open /path/to/file: no such file or directory"),
			maxLength:  40,
			wantContains: "no such file or directory",
		},
		{
			name:       "error truncated",
			err:        fmt.Errorf("this is a very long error message that should be truncated because it exceeds the maximum length"),
			maxLength:  30,
			wantContains: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShortenError(tt.err, tt.maxLength)
			if tt.wantContains == "" {
				if got != "" {
					t.Errorf("ShortenError() = %s, want empty string", got)
				}
			} else {
				if !strings.Contains(got, tt.wantContains) {
					t.Errorf("ShortenError() = %s, should contain %s", got, tt.wantContains)
				}
			}
		})
	}
}
