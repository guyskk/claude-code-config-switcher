package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/guyskk/ccc/internal/config"
)

// writeSettingsJSON writes a settings.json into the test config dir.
func writeSettingsJSON(t *testing.T, content string) {
	t.Helper()
	path := filepath.Join(config.GetDir(), "settings.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}
}

func TestBuildProviderSettingsJSON(t *testing.T) {
	tests := []struct {
		name     string
		envMap   map[string]interface{}
		wantNil  bool
		wantKeys []string
	}{
		{
			name:    "nil map returns empty string",
			envMap:  nil,
			wantNil: true,
		},
		{
			name:    "empty map returns empty string",
			envMap:  map[string]interface{}{},
			wantNil: true,
		},
		{
			name:     "single key",
			envMap:   map[string]interface{}{"ANTHROPIC_BASE_URL": "https://example.com"},
			wantKeys: []string{"ANTHROPIC_BASE_URL"},
		},
		{
			name: "multiple keys",
			envMap: map[string]interface{}{
				"ANTHROPIC_BASE_URL":   "https://example.com",
				"ANTHROPIC_AUTH_TOKEN": "sk-test",
			},
			wantKeys: []string{"ANTHROPIC_BASE_URL", "ANTHROPIC_AUTH_TOKEN"},
		},
		{
			name:     "special characters are escaped",
			envMap:   map[string]interface{}{"KEY": `value with "quotes" and \backslash`},
			wantKeys: []string{"KEY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildProviderSettingsJSON(tt.envMap)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if result != "" {
					t.Errorf("expected empty string, got: %s", result)
				}
				return
			}

			if result == "" {
				t.Fatal("expected non-empty JSON, got empty string")
			}

			// Verify it's valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(result), &parsed); err != nil {
				t.Fatalf("result is not valid JSON: %v\nresult: %s", err, result)
			}

			// Verify it has "env" key
			envVal, ok := parsed["env"]
			if !ok {
				t.Fatal("JSON should have 'env' key")
			}

			envMap, ok := envVal.(map[string]interface{})
			if !ok {
				t.Fatal("'env' should be a map")
			}

			// Verify expected keys are present
			for _, key := range tt.wantKeys {
				if _, exists := envMap[key]; !exists {
					t.Errorf("expected key %q in env map, got keys: %v", key, envMap)
				}
			}
		})
	}
}

func TestBuildProviderSettingsJSON_ExpandsEnvVars(t *testing.T) {
	// Set a test env var
	t.Setenv("CCC_TEST_BASE_URL", "https://expanded.example.com")

	envMap := map[string]interface{}{
		"ANTHROPIC_BASE_URL": "${CCC_TEST_BASE_URL}",
	}

	result, err := buildProviderSettingsJSON(envMap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	envResult := parsed["env"].(map[string]interface{})
	got := envResult["ANTHROPIC_BASE_URL"].(string)
	if got != "https://expanded.example.com" {
		t.Errorf("expected env var to be expanded, got: %s", got)
	}
}
