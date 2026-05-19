// Package config provides configuration management for ccc.
// Claude settings use dynamic map[string]interface{} to handle arbitrary fields.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetDirFunc is a function that returns the Claude configuration directory.
// This variable allows tests to override the default behavior.
var GetDirFunc = func() string {
	if workDir := os.Getenv("CCC_CONFIG_DIR"); workDir != "" {
		return workDir
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(homeDir, ".claude")
}

// GetDir returns the Claude configuration directory.
func GetDir() string {
	return GetDirFunc()
}

// Config represents the ccc.json configuration structure.
// Settings and Providers use dynamic maps to handle arbitrary Claude settings fields.
type Config struct {
	Settings        map[string]interface{}            `json:"settings"`
	ClaudeArgs      []string                          `json:"claude_args,omitempty"`
	CurrentProvider string                            `json:"current_provider"`
	Providers       map[string]map[string]interface{} `json:"providers"`
}

// GetConfigPath returns the path to ccc.json.
func GetConfigPath() string {
	return filepath.Join(GetDir(), "ccc.json")
}

// GetSettingsPath returns the path to settings.json.
func GetSettingsPath() string {
	return filepath.Join(GetDir(), "settings.json")
}

// Load reads and parses the ccc.json configuration file.
func Load() (*Config, error) {
	configPath := GetConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to ccc.json.
func Save(cfg *Config) error {
	configPath := GetConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveSettings writes the settings to settings.json.
func SaveSettings(settings map[string]interface{}) error {
	settingsPath := GetSettingsPath()

	// Ensure settings directory exists
	settingsDir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// deepCopy creates a deep copy of a map[string]interface{}.
func deepCopy(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return nil
	}

	copied := make(map[string]interface{})
	for k, v := range original {
		switch val := v.(type) {
		case map[string]interface{}:
			copied[k] = deepCopy(val)
		case []interface{}:
			copied[k] = deepCopySlice(val)
		default:
			copied[k] = v
		}
	}
	return copied
}

// deepCopySlice creates a deep copy of a []interface{}.
func deepCopySlice(original []interface{}) []interface{} {
	if original == nil {
		return nil
	}

	copied := make([]interface{}, len(original))
	for i, v := range original {
		switch val := v.(type) {
		case map[string]interface{}:
			copied[i] = deepCopy(val)
		case []interface{}:
			copied[i] = deepCopySlice(val)
		default:
			copied[i] = v
		}
	}
	return copied
}

// DeepMerge recursively merges provider settings into base settings.
// Provider settings override base settings for the same keys.
// This function handles arbitrary Claude settings fields.
func DeepMerge(base, provider map[string]interface{}) map[string]interface{} {
	result := deepCopy(base)
	if result == nil {
		result = make(map[string]interface{})
	}

	for key, value := range provider {
		if existingVal, exists := result[key]; exists {
			// If both are maps, merge them recursively
			if existingMap, ok := existingVal.(map[string]interface{}); ok {
				if newMap, ok := value.(map[string]interface{}); ok {
					result[key] = DeepMerge(existingMap, newMap)
					continue
				}
			}
		}
		// Otherwise, override with provider value
		result[key] = value
	}

	return result
}

// GetEnv extracts the env map from settings.
// Returns nil if env doesn't exist or is not a map.
func GetEnv(settings map[string]interface{}) map[string]interface{} {
	if settings == nil {
		return nil
	}
	if envVal, exists := settings["env"]; exists {
		if envMap, ok := envVal.(map[string]interface{}); ok {
			return envMap
		}
	}
	return nil
}

// GetEnvString extracts a string value from settings.env.
// Returns defaultValue if the key doesn't exist.
func GetEnvString(settings map[string]interface{}, key, defaultValue string) string {
	env := GetEnv(settings)
	if env == nil {
		return defaultValue
	}
	if val, exists := env[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetAuthToken extracts the ANTHROPIC_AUTH_TOKEN from settings.
// Returns a placeholder if the token is not set.
func GetAuthToken(settings map[string]interface{}) string {
	token := GetEnvString(settings, "ANTHROPIC_AUTH_TOKEN", "")
	if token == "" {
		return "PLEASE_SET_ANTHROPIC_AUTH_TOKEN"
	}
	return token
}

// GetBaseURL extracts the ANTHROPIC_BASE_URL from settings.
// Returns empty string if not set.
func GetBaseURL(settings map[string]interface{}) string {
	return GetEnvString(settings, "ANTHROPIC_BASE_URL", "")
}

// GetModel extracts the ANTHROPIC_MODEL from settings.
// Returns empty string if not set.
func GetModel(settings map[string]interface{}) string {
	return GetEnvString(settings, "ANTHROPIC_MODEL", "")
}

// LoadSettings reads the existing settings.json file.
// Returns nil if the file doesn't exist (not an error).
func LoadSettings() (map[string]interface{}, error) {
	settingsPath := GetSettingsPath()
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist is not an error - first run or clean install
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings file: %w", err)
	}

	return settings, nil
}

// FilterUserEnvForSettings filters user-defined env to only keep safe keys.
// It removes keys in managedEnvKeys or with ANTHROPIC_*/CLAUDE_* prefix.
// Returns nil if no keys remain.
func FilterUserEnvForSettings(userEnv map[string]interface{}, managedEnvKeys map[string]bool) map[string]interface{} {
	if userEnv == nil {
		return nil
	}

	filtered := make(map[string]interface{})
	for key, value := range userEnv {
		if managedEnvKeys[key] {
			continue
		}
		if strings.HasPrefix(key, "ANTHROPIC_") || strings.HasPrefix(key, "CLAUDE_") {
			continue
		}
		filtered[key] = value
	}

	if len(filtered) == 0 {
		return nil
	}
	return filtered
}

// MergeEnvMaps merges multiple env maps. Later maps override earlier ones.
// Returns nil if no maps have entries.
func MergeEnvMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		if m == nil {
			continue
		}
		for k, v := range m {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// MergeWithPriority merges multiple settings with priority.
// Priority (highest to lowest):
//  1. userSettings (settings.json - the actual user config)
//  2. providerSettings (provider-specific config)
//  3. baseSettings (ccc.json settings - template)
//
// Returns a new merged map without modifying the inputs.
func MergeWithPriority(baseSettings, providerSettings, userSettings map[string]interface{}) map[string]interface{} {
	// Start with deep copy of base settings
	result := deepCopy(baseSettings)
	if result == nil {
		result = make(map[string]interface{})
	}

	// Merge provider settings into result (provider overrides base)
	result = DeepMerge(result, providerSettings)

	// Merge user settings into result (user overrides all)
	result = DeepMerge(result, userSettings)

	return result
}

// RemoveStopHook removes the Supervisor Stop hook from settings.
// It cleans up supervisor-hook entries in hooks.Stop, and removes
// empty Stop arrays and empty hooks maps.
// User settings like disableAllHooks and allowManagedHooksOnly are left untouched.
// Returns a new map with supervisor hooks removed.
func RemoveStopHook(settings map[string]interface{}) map[string]interface{} {
	// Deep copy to avoid modifying input
	result := deepCopy(settings)
	if result == nil {
		return make(map[string]interface{})
	}

	// Check if hooks exist
	hooksVal, exists := result["hooks"]
	if !exists {
		return result
	}

	hooks, ok := hooksVal.(map[string]interface{})
	if !ok {
		return result
	}

	// Filter Stop hook entries to remove supervisor-hook commands
	stopVal, exists := hooks["Stop"]
	if exists {
		if stopArr, ok := stopVal.([]interface{}); ok {
			filtered := make([]interface{}, 0)
			for _, entry := range stopArr {
				if entryMap, ok := entry.(map[string]interface{}); ok {
					if innerHooks, ok := entryMap["hooks"].([]interface{}); ok {
						// Filter inner hooks to remove supervisor-hook
						innerFiltered := make([]interface{}, 0)
						for _, h := range innerHooks {
							if hMap, ok := h.(map[string]interface{}); ok {
								if cmd, ok := hMap["command"].(string); ok && strings.Contains(cmd, "supervisor-hook") {
									continue
								}
							}
							innerFiltered = append(innerFiltered, h)
						}
						if len(innerFiltered) > 0 {
							entryMap["hooks"] = innerFiltered
							filtered = append(filtered, entryMap)
						}
					} else {
						filtered = append(filtered, entry)
					}
				}
			}
			if len(filtered) == 0 {
				delete(hooks, "Stop")
			} else {
				hooks["Stop"] = filtered
			}
		}
	}

	// Remove hooks field if empty
	if len(hooks) == 0 {
		delete(result, "hooks")
	}

	return result
}
