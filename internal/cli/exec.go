package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/guyskk/ccc/internal/config"
	"github.com/guyskk/ccc/internal/provider"
)

// executeProcess replaces the current process with the specified command.
// This uses syscall.Exec which does not return on success.
func executeProcess(path string, args []string, env []string) error {
	return syscall.Exec(path, args, env)
}

// determineProvider determines which provider to use based on the command and config.
func determineProvider(cmd *Command, cfg *config.Config) string {
	if cmd.Provider != "" {
		// User specified a provider, check if it's valid
		if _, exists := cfg.Providers[cmd.Provider]; exists {
			return cmd.Provider
		}
		// Not a valid provider, try using current provider
		if cfg.CurrentProvider != "" {
			fmt.Printf("Unknown provider: %s\n", cmd.Provider)
			fmt.Printf("Using current provider: %s\n", cfg.CurrentProvider)
			return cfg.CurrentProvider
		}
		return ""
	}

	// No provider specified, use current or first available
	if cfg.CurrentProvider != "" {
		return cfg.CurrentProvider
	}

	// Use the first available provider
	for name := range cfg.Providers {
		return name
	}

	return ""
}

// runClaude executes the claude command for the given provider.
// This replaces the current process with claude using syscall.Exec.
// Provider env variables are passed to the claude subprocess via both
// process env and --settings CLI parameter (which has higher priority
// than settings.json env).
func runClaude(cfg *config.Config, cmd *Command) error {
	// Determine which provider to use
	providerName := determineProvider(cmd, cfg)
	if providerName == "" {
		return fmt.Errorf("no providers configured")
	}

	// Switch provider and clean up supervisor hooks
	result, err := provider.SwitchWithHook(cfg, providerName)
	if err != nil {
		return fmt.Errorf("error switching provider: %w", err)
	}
	fmt.Printf("Launching with provider: %s\n", providerName)

	// Find claude executable path
	// 优先使用 CCC_CLAUDE 环境变量（由包装脚本设置）
	var claudePath string
	if realPath := os.Getenv("CCC_CLAUDE"); realPath != "" {
		// 环境变量存在，验证路径是否有效
		// 使用 exec.LookPath 验证文件存在且可执行
		claudePath, err = exec.LookPath(realPath)
		if err != nil {
			return fmt.Errorf("CCC_CLAUDE environment variable points to invalid path: %s: %w", realPath, err)
		}
	} else {
		// 环境变量不存在，使用 LookPath 查找
		claudePath, err = exec.LookPath("claude")
		if err != nil {
			return fmt.Errorf("claude not found in PATH: %w", err)
		}
	}

	// Build arguments (argv[0] must be the program name)
	execArgs := []string{"claude"}
	if len(cfg.ClaudeArgs) > 0 {
		execArgs = append(execArgs, cfg.ClaudeArgs...)
	}
	execArgs = append(execArgs, cmd.ClaudeArgs...)

	// Pass provider env via --settings CLI parameter.
	// --settings has "Command line arguments" priority (level 2), which is higher
	// than User settings (level 5, ~/.claude/settings.json). This ensures provider
	// env overrides any conflicting keys in settings.json without modifying the file.
	// See docs/discuss-20260609-env-override.md for the empirical proof.
	if result.ProviderEnv != nil && len(result.ProviderEnv) > 0 {
		settingsJSON, err := buildProviderSettingsJSON(result.ProviderEnv)
		if err != nil {
			return fmt.Errorf("failed to build provider settings: %w", err)
		}
		execArgs = append(execArgs, "--settings", settingsJSON)
	}

	// Build environment variables
	// Start with current process environment
	env := os.Environ()

	// Remove existing environment variables to ensure provider config takes precedence
	prefixes := []string{"CLAUDE_", "ANTHROPIC_"}
	env = filterEnvVars(env, func(key string) bool {
		for _, prefix := range prefixes {
			if strings.HasPrefix(key, prefix) {
				return false
			}
		}
		return true
	})

	// Add merged provider env variables
	if result.EnvVars != nil {
		envPairs := provider.EnvPairsToStrings(result.EnvVars)
		env = append(env, envPairs...)
	}

	// Execute the process (replaces current process, does not return on success)
	return executeProcess(claudePath, execArgs, env)
}

// filterEnvVars filters environment variables based on a predicate function
func filterEnvVars(env []string, shouldKeep func(string) bool) []string {
	var filtered []string
	for _, envVar := range env {
		if parts := strings.SplitN(envVar, "=", 2); len(parts) == 2 {
			if shouldKeep(parts[0]) {
				filtered = append(filtered, envVar)
			}
		}
	}
	return filtered
}

// buildProviderSettingsJSON serializes provider env into a JSON string
// suitable for passing to claude --settings.
// Environment variable references like ${VAR} are expanded before serialization.
// Returns empty string if providerEnv is nil or empty.
func buildProviderSettingsJSON(providerEnv map[string]interface{}) (string, error) {
	if len(providerEnv) == 0 {
		return "", nil
	}

	// Expand env var references
	expanded := make(map[string]interface{}, len(providerEnv))
	for k, v := range providerEnv {
		expanded[k] = os.ExpandEnv(fmt.Sprintf("%v", v))
	}

	settings := map[string]interface{}{
		"env": expanded,
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return "", fmt.Errorf("failed to marshal provider settings: %w", err)
	}
	return string(data), nil
}
