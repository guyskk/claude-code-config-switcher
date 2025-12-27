package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Version is set by build flags during release
var Version = "dev"

// getClaudeDirFunc is a variable that holds the function to get the Claude directory
// This allows tests to override it for testing purposes
var getClaudeDirFunc = func() string {
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

// getClaudeDir returns the Claude configuration directory
func getClaudeDir() string {
	return getClaudeDirFunc()
}

// getUserInputFunc is a variable that holds the function to get user input
// This allows tests to override it for testing purposes
var getUserInputFunc = func(prompt string) (string, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	return reader.ReadString('\n')
}

// Config represents the structure of ccc.json
type Config struct {
	Settings        map[string]interface{}            `json:"settings"`
	CurrentProvider string                            `json:"current_provider"`
	Providers       map[string]map[string]interface{} `json:"providers"`
}

// getConfigPath returns the path to ccc.json
func getConfigPath() string {
	return filepath.Join(getClaudeDir(), "ccc.json")
}

// getSettingsPath returns the path to settings-{provider}.json
func getSettingsPath(providerName string) string {
	if providerName == "" {
		return filepath.Join(getClaudeDir(), "settings.json")
	}
	return filepath.Join(getClaudeDir(), fmt.Sprintf("settings-%s.json", providerName))
}

// showHelp displays usage information
func showHelp(config *Config, configErr error) {
	help := `Usage: ccc [provider|command] [args...]

Claude Code Configuration Switcher

Commands:
  ccc                      Use the current provider (or the first provider if none is set)
  ccc <provider>           Switch to the specified provider and run Claude Code
  ccc provider <command>   Manage providers (list, add, remove, show, set)
  ccc --help               Show this help message
  ccc --version            Show version information

Provider Management:
  ccc provider list                    List all configured providers
  ccc provider add <name>              Add a new provider
  ccc provider remove <name>           Remove a provider
  ccc provider show <name>             Show provider details
  ccc provider set <name> <key> <value> Update provider configuration

Environment Variables:
  CCC_CONFIG_DIR     Override the configuration directory (default: ~/.claude/)

Examples:
  ccc kimi           # Switch to kimi provider and run Claude Code
  ccc provider list  # List all providers
  ccc provider add openai --base-url=https://api.openai.com/v1 --token=sk-xxx --model=gpt-4
`
	fmt.Print(help)

	// Display config path and status
	configPath := getConfigPath()
	if configErr != nil {
		// Extract short error message - take only the last part (most relevant)
		errMsg := configErr.Error()
		// Find the last colon for shorter error message
		if lastColon := strings.LastIndex(errMsg, ":"); lastColon > 0 && lastColon < len(errMsg)-2 {
			errMsg = errMsg[lastColon+2:]
		}
		// Still limit length
		if len(errMsg) > 40 {
			errMsg = errMsg[:37] + "..."
		}
		fmt.Printf("\nCurrent config: %s (%s)\n", configPath, errMsg)
	} else {
		fmt.Printf("\nCurrent config: %s\n", configPath)

		// Display provider list from config
		if len(config.Providers) > 0 {
			fmt.Println("\nAvailable Providers:")
			for name := range config.Providers {
				marker := ""
				if name == config.CurrentProvider {
					marker = " (current)"
				}
				fmt.Printf("  %s%s\n", name, marker)
			}
		}
	}
	fmt.Println()
}

// checkExistingSettings checks if ~/.claude/settings.json exists
func checkExistingSettings() bool {
	settingsPath := getSettingsPath("")
	_, err := os.Stat(settingsPath)
	return err == nil
}

// promptUserForMigration prompts the user to confirm migration from existing settings
func promptUserForMigration() bool {
	cccPath := getConfigPath()
	settingsPath := getSettingsPath("")

	fmt.Printf("ccc configuration not found: %s\n", cccPath)
	fmt.Printf("Found existing Claude configuration: %s\n", settingsPath)
	fmt.Println()

	input, err := getUserInputFunc("Would you like to create ccc config from existing settings? [y/N] ")
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// migrateFromSettings creates a new ccc.json from existing settings.json
func migrateFromSettings() error {
	settingsPath := getSettingsPath("")

	// Read existing settings.json
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	// Extract env from settings
	var envConfig map[string]interface{}
	if envVal, exists := settings["env"]; exists {
		if envMap, ok := envVal.(map[string]interface{}); ok {
			envConfig = envMap
		}
	}

	// Remove env from settings to create base settings
	baseSettings := make(map[string]interface{})
	for k, v := range settings {
		if k != "env" {
			baseSettings[k] = v
		}
	}

	// Create new ccc config
	config := &Config{
		Settings:        baseSettings,
		CurrentProvider: "default",
		Providers:       make(map[string]map[string]interface{}),
	}

	// Always create default provider (even if env is empty)
	defaultProvider := make(map[string]interface{})
	if envConfig != nil {
		defaultProvider["env"] = envConfig
	}
	config.Providers["default"] = defaultProvider

	// Save the new config
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save ccc config: %w", err)
	}

	cccPath := getConfigPath()
	fmt.Printf("Created ccc config with 'default' provider: %s\n", cccPath)
	return nil
}

// loadConfig reads and parses ccc.json
func loadConfig() (*Config, error) {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// saveConfig writes the config back to ccc.json
func saveConfig(config *Config) error {
	configPath := getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// saveSettings writes the merged settings to settings-{provider}.json
func saveSettings(settings map[string]interface{}, providerName string) error {
	settingsPath := getSettingsPath(providerName)

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

// deepCopy creates a deep copy of a map
func deepCopy(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return nil
	}

	copied := make(map[string]interface{})
	for k, v := range original {
		if nestedMap, ok := v.(map[string]interface{}); ok {
			copied[k] = deepCopy(nestedMap)
		} else {
			copied[k] = v
		}
	}
	return copied
}

// deepMerge recursively merges provider settings into base settings
// Provider settings override base settings for the same keys
func deepMerge(base, provider map[string]interface{}) map[string]interface{} {
	result := deepCopy(base)
	if result == nil {
		result = make(map[string]interface{})
	}

	for key, value := range provider {
		if existingVal, exists := result[key]; exists {
			// If both are maps, merge them recursively
			if existingMap, ok := existingVal.(map[string]interface{}); ok {
				if newMap, ok := value.(map[string]interface{}); ok {
					result[key] = deepMerge(existingMap, newMap)
					continue
				}
			}
		}
		// Otherwise, override with provider value
		result[key] = value
	}

	return result
}

// switchProvider switches to the specified provider by merging configurations
func switchProvider(providerName string) error {
	fmt.Printf("Launching with provider: %s\n", providerName)

	// Load the configuration
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if provider exists
	providerSettings, exists := config.Providers[providerName]
	if !exists {
		return fmt.Errorf("provider '%s' not found in configuration", providerName)
	}

	// Create the merged settings
	// Start with the base settings template
	mergedSettings := deepCopy(config.Settings)
	if mergedSettings == nil {
		mergedSettings = make(map[string]interface{})
	}

	// Merge provider-specific settings
	mergedSettings = deepMerge(mergedSettings, providerSettings)

	// Save the merged settings to settings-{provider}.json
	if err := saveSettings(mergedSettings, providerName); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Update current_provider in ccc.json
	config.CurrentProvider = providerName
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to update current provider: %w", err)
	}

	return nil
}

// runClaude executes the claude command with the settings file
func runClaude(providerName string, args []string) error {
	settingsPath := getSettingsPath(providerName)

	// Check if settings file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return fmt.Errorf("settings file not found: %s. Please run 'ccc <provider>' first", settingsPath)
	}

	// Build the claude command
	cmd := exec.Command("claude", append([]string{"--settings", settingsPath}, args...)...)

	// Set up stdin, stdout, stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run claude: %w", err)
	}

	return nil
}

// ========== Provider Management Functions ==========

// validateURL checks if a string is a valid HTTPS URL
func validateURL(urlStr string) error {
	if !strings.HasPrefix(urlStr, "https://") {
		return fmt.Errorf("URL must start with https://")
	}
	return nil
}

// validateProviderName checks if a provider name is valid
func validateProviderName(name string) error {
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	// Allow lowercase letters, numbers, hyphens, underscores
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			return fmt.Errorf("provider name must contain only lowercase letters, numbers, hyphens, and underscores")
		}
	}
	return nil
}

// maskToken masks sensitive token for display (show first 3 and last 3 chars)
func maskToken(token string) string {
	if len(token) <= 6 {
		return "***"
	}
	return token[:3] + "***" + token[len(token)-3:]
}

// deleteSettingsFile deletes the settings-{provider}.json file
func deleteSettingsFile(providerName string) error {
	settingsPath := getSettingsPath(providerName)
	if _, err := os.Stat(settingsPath); err == nil {
		if err := os.Remove(settingsPath); err != nil {
			return fmt.Errorf("failed to delete settings file: %w", err)
		}
	}
	return nil
}

// promptForProviderConfig interactively prompts user for provider configuration
func promptForProviderConfig(name string) (map[string]interface{}, error) {
	env := make(map[string]interface{})

	fmt.Printf("\nConfiguring provider '%s'\n\n", name)

	// Prompt for BASE_URL
	for {
		input, err := getUserInputFunc("ANTHROPIC_BASE_URL (https://...): ")
		if err != nil {
			return nil, err
		}
		baseURL := strings.TrimSpace(input)
		if baseURL == "" {
			fmt.Println("Error: BASE_URL is required")
			continue
		}
		if err := validateURL(baseURL); err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		env["ANTHROPIC_BASE_URL"] = baseURL
		break
	}

	// Prompt for AUTH_TOKEN
	for {
		input, err := getUserInputFunc("ANTHROPIC_AUTH_TOKEN: ")
		if err != nil {
			return nil, err
		}
		token := strings.TrimSpace(input)
		if token == "" {
			fmt.Println("Error: AUTH_TOKEN is required")
			continue
		}
		env["ANTHROPIC_AUTH_TOKEN"] = token
		break
	}

	// Prompt for MODEL
	for {
		input, err := getUserInputFunc("ANTHROPIC_MODEL: ")
		if err != nil {
			return nil, err
		}
		model := strings.TrimSpace(input)
		if model == "" {
			fmt.Println("Error: MODEL is required")
			continue
		}
		env["ANTHROPIC_MODEL"] = model
		break
	}

	// Prompt for SMALL_FAST_MODEL (optional)
	input, err := getUserInputFunc("ANTHROPIC_SMALL_FAST_MODEL (optional, press Enter to use same as MODEL): ")
	if err != nil {
		return nil, err
	}
	smallModel := strings.TrimSpace(input)
	if smallModel != "" {
		env["ANTHROPIC_SMALL_FAST_MODEL"] = smallModel
	} else {
		env["ANTHROPIC_SMALL_FAST_MODEL"] = env["ANTHROPIC_MODEL"]
	}

	return env, nil
}

// listProviders lists all configured providers
func listProviders(config *Config) error {
	if len(config.Providers) == 0 {
		fmt.Println("No providers configured")
		fmt.Println("\nUse 'ccc provider add <name>' to add a new provider")
		return nil
	}

	fmt.Println("Available Providers:")
	fmt.Println()

	for name, provider := range config.Providers {
		marker := " "
		if name == config.CurrentProvider {
			marker = "*"
		}

		fmt.Printf("%s %s\n", marker, name)

		// Show basic info from env if available
		if env, ok := provider["env"].(map[string]interface{}); ok {
			if baseURL, ok := env["ANTHROPIC_BASE_URL"].(string); ok {
				fmt.Printf("    BASE_URL: %s\n", baseURL)
			}
			if model, ok := env["ANTHROPIC_MODEL"].(string); ok {
				fmt.Printf("    MODEL: %s\n", model)
			}
		}
		fmt.Println()
	}

	fmt.Println("* = current provider")
	return nil
}

// addProvider adds a new provider to the configuration
func addProvider(config *Config, name string, flags map[string]string) error {
	// Validate provider name
	if err := validateProviderName(name); err != nil {
		return err
	}

	// Check if provider already exists
	if _, exists := config.Providers[name]; exists {
		return fmt.Errorf("provider '%s' already exists\nUse 'ccc provider set %s <key> <value>' to modify or 'ccc provider remove %s' to delete", name, name, name)
	}

	var env map[string]interface{}
	var err error

	// Check if flags provided (non-interactive mode)
	if len(flags) > 0 {
		env = make(map[string]interface{})

		// Required flags
		baseURL, hasBaseURL := flags["base-url"]
		token, hasToken := flags["token"]
		model, hasModel := flags["model"]

		if !hasBaseURL || !hasToken || !hasModel {
			missing := []string{}
			if !hasBaseURL {
				missing = append(missing, "--base-url")
			}
			if !hasToken {
				missing = append(missing, "--token")
			}
			if !hasModel {
				missing = append(missing, "--model")
			}
			return fmt.Errorf("missing required flags: %s", strings.Join(missing, ", "))
		}

		// Validate BASE_URL
		if err := validateURL(baseURL); err != nil {
			return fmt.Errorf("invalid --base-url: %w", err)
		}

		env["ANTHROPIC_BASE_URL"] = baseURL
		env["ANTHROPIC_AUTH_TOKEN"] = token
		env["ANTHROPIC_MODEL"] = model

		// Optional SMALL_FAST_MODEL
		if smallModel, hasSmall := flags["small-model"]; hasSmall && smallModel != "" {
			env["ANTHROPIC_SMALL_FAST_MODEL"] = smallModel
		} else {
			env["ANTHROPIC_SMALL_FAST_MODEL"] = model
		}
	} else {
		// Interactive mode
		env, err = promptForProviderConfig(name)
		if err != nil {
			return fmt.Errorf("configuration cancelled: %w", err)
		}
	}

	// Add provider to config
	if config.Providers == nil {
		config.Providers = make(map[string]map[string]interface{})
	}
	config.Providers[name] = map[string]interface{}{
		"env": env,
	}

	// Save config
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Provider '%s' added successfully\n", name)
	return nil
}

// removeProvider removes a provider from the configuration
func removeProvider(config *Config, name string) error {
	// Check if provider exists
	if _, exists := config.Providers[name]; !exists {
		fmt.Fprintf(os.Stderr, "Provider '%s' not found\n\n", name)
		fmt.Fprintf(os.Stderr, "Available providers:\n")
		for providerName := range config.Providers {
			fmt.Fprintf(os.Stderr, "  - %s\n", providerName)
		}
		return fmt.Errorf("provider not found")
	}

	// Check if it's the current provider
	if config.CurrentProvider == name {
		return fmt.Errorf("cannot remove the current provider '%s'\nPlease switch to another provider first", name)
	}

	// Check if it's the last provider
	if len(config.Providers) == 1 {
		return fmt.Errorf("cannot remove the last provider\nAt least one provider must be configured")
	}

	// Delete provider from config
	delete(config.Providers, name)

	// Save config
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Delete corresponding settings file
	if err := deleteSettingsFile(name); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	fmt.Printf("Provider '%s' removed successfully\n", name)
	return nil
}

// showProvider displays detailed configuration for a provider
func showProvider(config *Config, name string) error {
	// Check if provider exists
	provider, exists := config.Providers[name]
	if !exists {
		fmt.Fprintf(os.Stderr, "Provider '%s' not found\n\n", name)
		fmt.Fprintf(os.Stderr, "Available providers:\n")
		for providerName := range config.Providers {
			fmt.Fprintf(os.Stderr, "  - %s\n", providerName)
		}
		return fmt.Errorf("provider not found")
	}

	fmt.Printf("Provider: %s\n", name)
	if name == config.CurrentProvider {
		fmt.Println("Status: current")
	}
	fmt.Println()

	// Display env configuration
	if env, ok := provider["env"].(map[string]interface{}); ok {
		fmt.Println("Configuration:")
		for key, value := range env {
			valueStr := fmt.Sprintf("%v", value)
			// Mask token for security
			if strings.Contains(key, "TOKEN") || strings.Contains(key, "KEY") {
				valueStr = maskToken(valueStr)
			}
			fmt.Printf("  %s: %s\n", key, valueStr)
		}
	} else {
		fmt.Println("No configuration found")
	}

	return nil
}

// setProviderEnv sets an environment variable for a provider
func setProviderEnv(config *Config, name, key, value string) error {
	// Check if provider exists
	provider, exists := config.Providers[name]
	if !exists {
		return fmt.Errorf("provider '%s' not found", name)
	}

	// Validate specific keys
	if key == "ANTHROPIC_BASE_URL" {
		if err := validateURL(value); err != nil {
			return fmt.Errorf("invalid value for %s: %w", key, err)
		}
	}

	// Get or create env map
	var env map[string]interface{}
	if envVal, ok := provider["env"].(map[string]interface{}); ok {
		env = envVal
	} else {
		env = make(map[string]interface{})
		provider["env"] = env
	}

	// Set the value
	env[key] = value

	// Save config
	if err := saveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// If this is the current provider, regenerate settings file
	if name == config.CurrentProvider {
		if err := switchProvider(name); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to regenerate settings file: %v\n", err)
		}
	}

	fmt.Printf("Provider '%s' updated: %s=%s\n", name, key, value)
	return nil
}

// showProviderHelp displays help for provider subcommands
func showProviderHelp() {
	help := `Usage: ccc provider <command> [args...]

Manage Claude Code API providers

Commands:
  list                    List all configured providers
  add <name>              Add a new provider (interactive)
  remove <name>           Remove a provider
  show <name>             Show provider configuration details
  set <name> <key> <value> Set a provider environment variable

Examples:
  # List all providers
  ccc provider list

  # Add a new provider interactively
  ccc provider add openai

  # Add a provider non-interactively
  ccc provider add openai --base-url=https://api.openai.com/v1 --token=sk-xxx --model=gpt-4

  # Show provider details
  ccc provider show kimi

  # Update a provider's configuration
  ccc provider set kimi ANTHROPIC_MODEL kimi-k1.5

  # Remove a provider
  ccc provider remove old-provider

Flags for 'add' command:
  --base-url <url>     ANTHROPIC_BASE_URL (required in non-interactive mode)
  --token <token>      ANTHROPIC_AUTH_TOKEN (required in non-interactive mode)
  --model <model>      ANTHROPIC_MODEL (required in non-interactive mode)
  --small-model <model> ANTHROPIC_SMALL_FAST_MODEL (optional)
`
	fmt.Print(help)
}

func main() {
	// Parse command line arguments
	args := os.Args[1:]

	// Handle --version
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Printf("ccc version %s\n", Version)
		os.Exit(0)
	}

	// Handle --help (try to load config for provider list, but show help anyway if it fails)
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		config, err := loadConfig()
		showHelp(config, err)
		os.Exit(0)
	}

	// Handle 'provider' subcommand
	if len(args) > 0 && args[0] == "provider" {
		// Load configuration
		config, err := loadConfig()
		if err != nil {
			// Try migration for provider commands too
			if checkExistingSettings() && promptUserForMigration() {
				if err := migrateFromSettings(); err != nil {
					fmt.Fprintf(os.Stderr, "Error migrating from settings: %v\n", err)
					os.Exit(1)
				}
				config, err = loadConfig()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				fmt.Fprintf(os.Stderr, "Run 'ccc --help' for usage information\n")
				os.Exit(1)
			}
		}

		// Parse provider subcommand
		if len(args) == 1 || (len(args) == 2 && (args[1] == "--help" || args[1] == "-h")) {
			showProviderHelp()
			os.Exit(0)
		}

		subcommand := args[1]
		subargs := args[2:]

		var cmdErr error
		switch subcommand {
		case "list":
			cmdErr = listProviders(config)
		case "add":
			if len(subargs) == 0 {
				fmt.Fprintf(os.Stderr, "Error: provider name required\n")
				fmt.Fprintf(os.Stderr, "Usage: ccc provider add <name> [flags]\n")
				os.Exit(1)
			}
			name := subargs[0]
			flags := make(map[string]string)
			// Parse flags
			for i := 1; i < len(subargs); i++ {
				arg := subargs[i]
				if strings.HasPrefix(arg, "--base-url=") {
					flags["base-url"] = strings.TrimPrefix(arg, "--base-url=")
				} else if strings.HasPrefix(arg, "--token=") {
					flags["token"] = strings.TrimPrefix(arg, "--token=")
				} else if strings.HasPrefix(arg, "--model=") {
					flags["model"] = strings.TrimPrefix(arg, "--model=")
				} else if strings.HasPrefix(arg, "--small-model=") {
					flags["small-model"] = strings.TrimPrefix(arg, "--small-model=")
				}
			}
			cmdErr = addProvider(config, name, flags)
		case "remove":
			if len(subargs) == 0 {
				fmt.Fprintf(os.Stderr, "Error: provider name required\n")
				fmt.Fprintf(os.Stderr, "Usage: ccc provider remove <name>\n")
				os.Exit(1)
			}
			cmdErr = removeProvider(config, subargs[0])
		case "show":
			if len(subargs) == 0 {
				fmt.Fprintf(os.Stderr, "Error: provider name required\n")
				fmt.Fprintf(os.Stderr, "Usage: ccc provider show <name>\n")
				os.Exit(1)
			}
			cmdErr = showProvider(config, subargs[0])
		case "set":
			if len(subargs) < 3 {
				fmt.Fprintf(os.Stderr, "Error: requires <name> <key> <value>\n")
				fmt.Fprintf(os.Stderr, "Usage: ccc provider set <name> <key> <value>\n")
				os.Exit(1)
			}
			cmdErr = setProviderEnv(config, subargs[0], subargs[1], subargs[2])
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", subcommand)
			fmt.Fprintf(os.Stderr, "Run 'ccc provider --help' for usage information\n")
			os.Exit(1)
		}

		if cmdErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Load configuration first to check if it exists and is valid
	config, err := loadConfig()
	if err != nil {
		// Try to migrate from existing settings.json
		if checkExistingSettings() && promptUserForMigration() {
			if err := migrateFromSettings(); err != nil {
				fmt.Fprintf(os.Stderr, "Error migrating from settings: %v\n", err)
				os.Exit(1)
			}
			// Reload config after migration
			config, err = loadConfig()
			if err != nil {
				showHelp(nil, err)
				os.Exit(1)
			}
		} else {
			showHelp(nil, err)
			os.Exit(1)
		}
	}

	// Determine which provider to use
	var providerName string
	var claudeArgs []string

	if len(args) == 0 {
		// No arguments - use current provider
		if config.CurrentProvider != "" {
			providerName = config.CurrentProvider
		} else {
			// Use the first available provider
			for name := range config.Providers {
				providerName = name
				break
			}
			if providerName == "" {
				fmt.Fprintf(os.Stderr, "No providers configured\n")
				os.Exit(1)
			}
		}
		claudeArgs = []string{}
	} else {
		// First argument might be a provider name
		potentialProvider := args[0]

		// Check if it's a valid provider
		if _, exists := config.Providers[potentialProvider]; exists {
			providerName = potentialProvider
			claudeArgs = args[1:]
		} else {
			// Not a provider name, use current provider and pass all args to claude
			if config.CurrentProvider != "" {
				providerName = config.CurrentProvider
				fmt.Printf("Using current provider: %s\n", providerName)
			} else {
				fmt.Fprintf(os.Stderr, "Unknown provider: %s\n", potentialProvider)
				os.Exit(1)
			}
			claudeArgs = args
		}
	}

	// Switch to the provider (this will merge and save settings)
	if err := switchProvider(providerName); err != nil {
		fmt.Fprintf(os.Stderr, "Error switching provider: %v\n", err)
		os.Exit(1)
	}

	// Run claude with the settings file
	if err := runClaude(providerName, claudeArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error running claude: %v\n", err)
		os.Exit(1)
	}
}
