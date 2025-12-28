// Package migration handles migrating from old Claude settings to ccc format.
package migration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/user/ccc/config"
)

// GetUserInputFunc is a function that gets user input from stdin.
// This variable allows tests to override the default behavior.
var GetUserInputFunc = func(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	return reader.ReadString('\n')
}

// CheckExisting checks if ~/.claude/settings.json exists.
func CheckExisting() bool {
	settingsPath := config.GetSettingsPath("")
	_, err := os.Stat(settingsPath)
	return err == nil
}

// PromptUser prompts the user to confirm migration from existing settings.
func PromptUser() bool {
	cccPath := config.GetConfigPath()
	settingsPath := config.GetSettingsPath("")

	fmt.Printf("ccc configuration not found: %s\n", cccPath)
	fmt.Printf("Found existing Claude configuration: %s\n", settingsPath)
	fmt.Println()

	input, err := GetUserInputFunc("Would you like to create ccc config from existing settings? [y/N] ")
	if err != nil {
		return false
	}

	input = trimToLower(input)
	return input == "y" || input == "yes"
}

// MigrateFromSettings creates a new ccc.json from existing settings.json.
func MigrateFromSettings() error {
	settingsPath := config.GetSettingsPath("")

	// Read existing settings.json
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var oldSettings map[string]interface{}
	if err := json.Unmarshal(data, &oldSettings); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	// Build new config structure
	cfg := &config.Config{
		Settings:        config.Settings{},
		CurrentProvider: "default",
		Providers:       make(map[string]config.ProviderConfig),
	}

	// Extract env from settings
	var envConfig config.Env
	if envVal, exists := oldSettings["env"]; exists {
		if envMap, ok := envVal.(map[string]interface{}); ok {
			envConfig = make(config.Env)
			for k, v := range envMap {
				if str, ok := v.(string); ok {
					envConfig[k] = str
				}
			}
		}
	}

	// Copy non-env fields to settings
	for k, v := range oldSettings {
		if k != "env" {
			switch k {
			case "permissions":
				if p, ok := v.(map[string]interface{}); ok {
					cfg.Settings.Permissions = &config.Permissions{}
					if allow, ok := p["allow"].([]interface{}); ok {
						for _, a := range allow {
							if str, ok := a.(string); ok {
								cfg.Settings.Permissions.Allow = append(cfg.Settings.Permissions.Allow, str)
							}
						}
					}
					if mode, ok := p["defaultMode"].(string); ok {
						cfg.Settings.Permissions.DefaultMode = mode
					}
				}
			case "alwaysThinkingEnabled":
				if b, ok := v.(bool); ok {
					cfg.Settings.AlwaysThinkingEnabled = b
				}
			default:
				// Copy other fields as-is (they'll be serialized back)
				// For strict typing, we ignore unknown fields
			}
		}
	}

	// Always create default provider (even if env is empty)
	defaultProvider := config.ProviderConfig{}
	if envConfig != nil {
		defaultProvider.Env = envConfig
	}
	cfg.Providers["default"] = defaultProvider

	// Save the new config
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save ccc config: %w", err)
	}

	cccPath := config.GetConfigPath()
	fmt.Printf("Created ccc config with 'default' provider: %s\n", cccPath)
	return nil
}

// trimToLower trims whitespace and converts to lowercase.
func trimToLower(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
