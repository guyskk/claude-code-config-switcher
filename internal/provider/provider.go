// Package provider handles provider switching and configuration merging.
package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/guyskk/ccc/internal/config"
)

// EnvPair represents a single environment variable key-value pair.
type EnvPair struct {
	Key   string
	Value string
}

// SwitchResult contains the result of switching providers.
// It includes the merged env variables that should be passed to the child process.
type SwitchResult struct {
	// Settings is the merged settings (without env) that was saved to settings.json
	Settings map[string]interface{}
	// EnvVars contains the merged environment variables (base.env + provider.env)
	// that should be passed to the claude subprocess
	EnvVars []EnvPair
	// ProviderEnv contains the merged base + provider env map,
	// used to build the --settings JSON for claude CLI.
	ProviderEnv map[string]interface{}
}

// SwitchWithHook switches to the specified provider and cleans up supervisor hooks.
// It generates settings.json with merged configuration from:
//  1. Existing settings.json (user config - highest priority)
//  2. ccc.json settings (base template)
//  3. Provider settings (provider-specific)
//
// It also removes any leftover supervisor artifacts (slash commands, state, logs).
// Returns the merged env that should be passed to the claude subprocess.
func SwitchWithHook(cfg *config.Config, providerName string) (*SwitchResult, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Check if provider exists
	providerSettings, exists := cfg.Providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found in configuration", providerName)
	}

	// Load existing settings.json (user's actual configuration)
	userSettings, err := config.LoadSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}

	// Extract env from each source before merging (to distinguish user env from ccc env)
	userEnvMap := config.GetEnv(userSettings)
	baseEnvMap := config.GetEnv(cfg.Settings)
	providerEnvMap := config.GetEnv(providerSettings)

	// Merge settings with priority: user > provider > base
	mergedSettings := config.MergeWithPriority(cfg.Settings, providerSettings, userSettings)

	// Remove Supervisor Stop hook and related fields
	cleanedSettings := config.RemoveStopHook(mergedSettings)

	// Remove merged env from settings, replace with user's original env.
	// Conflicting keys are no longer filtered here -- they are overridden
	// by the --settings CLI parameter when launching claude (higher priority).
	delete(cleanedSettings, "env")
	if userEnvMap != nil && len(userEnvMap) > 0 {
		cleanedSettings["env"] = userEnvMap
	}

	// Save merged settings to settings.json
	settingsPath := config.GetSettingsPath()
	settingsData, err := json.MarshalIndent(cleanedSettings, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings: %w", err)
	}
	if err := os.WriteFile(settingsPath, settingsData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write settings: %w", err)
	}

	// Clean up any leftover supervisor artifacts
	cleanupSupervisorArtifacts()

	// Update current_provider in ccc.json
	cfg.CurrentProvider = providerName
	if err := config.Save(cfg); err != nil {
		return nil, fmt.Errorf("failed to update current provider: %w", err)
	}

	// Extract env map for subprocess: only base + provider env (not user env)
	subprocessEnvMap := config.MergeEnvMaps(baseEnvMap, providerEnvMap)

	// Convert env map to EnvPair slice
	envVars := envMapToPairs(subprocessEnvMap)

	return &SwitchResult{
		Settings:    cleanedSettings,
		EnvVars:     envVars,
		ProviderEnv: subprocessEnvMap,
	}, nil
}

// cleanupSupervisorArtifacts removes leftover supervisor files:
//   - slash command files (supervisor.md, supervisoroff.md)
//   - state files (supervisor-*.json) and log files (supervisor-*.log)
func cleanupSupervisorArtifacts() {
	commandsDir := config.GetDir() + "/commands"
	os.Remove(commandsDir + "/supervisor.md")
	os.Remove(commandsDir + "/supervisoroff.md")

	stateDir := config.GetDir() + "/ccc"
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "supervisor-") && (strings.HasSuffix(name, ".json") || strings.HasSuffix(name, ".log")) {
			os.Remove(stateDir + "/" + name)
		}
	}
}

// envMapToPairs converts a map[string]interface{} to []EnvPair.
// It expands environment variable references like ${VAR}.
func envMapToPairs(envMap map[string]interface{}) []EnvPair {
	if envMap == nil {
		return nil
	}

	pairs := make([]EnvPair, 0, len(envMap))
	for k, v := range envMap {
		value := fmt.Sprintf("%v", v)
		// Expand environment variable references
		value = os.ExpandEnv(value)
		pairs = append(pairs, EnvPair{Key: k, Value: value})
	}
	return pairs
}

// EnvPairsToStrings converts []EnvPair to []string in "KEY=value" format.
func EnvPairsToStrings(pairs []EnvPair) []string {
	if pairs == nil {
		return nil
	}

	result := make([]string, len(pairs))
	for i, pair := range pairs {
		result[i] = fmt.Sprintf("%s=%s", pair.Key, pair.Value)
	}
	return result
}

// FormatProviderName formats a provider name for display.
// If the name is the current provider, adds a "(current)" suffix.
func FormatProviderName(name, currentProvider string) string {
	if name == currentProvider {
		return fmt.Sprintf("%s (current)", name)
	}
	return name
}

// ListProviders returns a list of all provider names from the config.
func ListProviders(cfg *config.Config) []string {
	if cfg == nil || len(cfg.Providers) == 0 {
		return []string{}
	}

	names := make([]string, 0, len(cfg.Providers))
	for name := range cfg.Providers {
		names = append(names, name)
	}
	return names
}

// ValidateProvider checks if a provider name exists in the config.
func ValidateProvider(cfg *config.Config, providerName string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if _, exists := cfg.Providers[providerName]; !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}
	return nil
}

// GetDefaultProvider returns the first provider name from the config.
// Returns empty string if no providers are configured.
func GetDefaultProvider(cfg *config.Config) string {
	if cfg == nil || len(cfg.Providers) == 0 {
		return ""
	}
	for name := range cfg.Providers {
		return name
	}
	return ""
}

// GetCurrentProvider returns the current provider from config.
// If current_provider is not set, returns the first available provider.
// Returns empty string if no providers are configured.
func GetCurrentProvider(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}

	// Try current_provider first
	if cfg.CurrentProvider != "" {
		if _, exists := cfg.Providers[cfg.CurrentProvider]; exists {
			return cfg.CurrentProvider
		}
	}

	// Fall back to first provider
	return GetDefaultProvider(cfg)
}

// ShortenError creates a shortened error message for display.
func ShortenError(err error, maxLength int) string {
	if err == nil {
		return ""
	}
	errMsg := err.Error()
	// Find the last colon for shorter error message
	if lastColon := strings.LastIndex(errMsg, ":"); lastColon > 0 && lastColon < len(errMsg)-2 {
		errMsg = errMsg[lastColon+2:]
	}
	// Limit length
	if len(errMsg) > maxLength {
		errMsg = errMsg[:maxLength-3] + "..."
	}
	return errMsg
}

// GetAuthToken extracts the ANTHROPIC_AUTH_TOKEN from merged settings.
// This is a convenience wrapper around config.GetAuthToken.
func GetAuthToken(settings map[string]interface{}) string {
	return config.GetAuthToken(settings)
}

// GetBaseURL extracts the ANTHROPIC_BASE_URL from merged settings.
// This is a convenience wrapper around config.GetBaseURL.
func GetBaseURL(settings map[string]interface{}) string {
	return config.GetBaseURL(settings)
}

// GetModel extracts the ANTHROPIC_MODEL from merged settings.
// This is a convenience wrapper around config.GetModel.
func GetModel(settings map[string]interface{}) string {
	return config.GetModel(settings)
}
