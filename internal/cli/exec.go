package cli

import (
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

// checkSettingsEnvConflict refuses to start claude when settings.json's env field
// contains keys that would silently override the provider env ccc passes to claude.
// See docs/discuss-20260519-env-priority.md for the empirical proof that settings.json
// env > process env in Claude Code.
//
// managedEnvKeys are derived from base env (cfg.Settings.env) plus the active
// provider's env. A non-nil error means the user must clean settings.json before ccc
// will launch claude — ccc never modifies the user's settings.json on their behalf.
func checkSettingsEnvConflict(cfg *config.Config, providerName string) error {
	providerSettings, ok := cfg.Providers[providerName]
	if !ok {
		// Provider lookup failure is surfaced later by SwitchWithHook; we just skip the guard.
		return nil
	}

	userSettings, err := config.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings.json for conflict check: %w", err)
	}
	if userSettings == nil {
		return nil
	}

	managedEnvKeys := make(map[string]bool)
	for key := range config.GetEnv(cfg.Settings) {
		managedEnvKeys[key] = true
	}
	for key := range config.GetEnv(providerSettings) {
		managedEnvKeys[key] = true
	}

	conflicts := config.DetectSettingsEnvConflicts(userSettings, managedEnvKeys)
	if len(conflicts) == 0 {
		return nil
	}

	return fmt.Errorf("%s", config.FormatEnvConflictError(
		config.GetSettingsPath(),
		config.GetConfigPath(),
		conflicts,
	))
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
// Provider env variables are passed to the claude subprocess.
func runClaude(cfg *config.Config, cmd *Command) error {
	// Determine which provider to use
	providerName := determineProvider(cmd, cfg)
	if providerName == "" {
		return fmt.Errorf("no providers configured")
	}

	// Guard: refuse to start claude when settings.json contains env keys that
	// would silently override the provider env ccc passes to the claude process.
	// Must run BEFORE SwitchWithHook so we never rewrite settings.json while leaving
	// the conflict in place. See docs/discuss-20260519-env-priority.md.
	if err := checkSettingsEnvConflict(cfg, providerName); err != nil {
		return err
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
