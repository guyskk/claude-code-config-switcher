// Package cli handles command-line parsing and execution.
package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/guyskk/ccc/internal/config"
	"github.com/guyskk/ccc/internal/migration"
	"github.com/guyskk/ccc/internal/provider"
	"github.com/guyskk/ccc/internal/validate"
)

// Name is the project name.
const Name = "claude-code-switcher"

// Version is set by build flags during release.
var Version = "dev"

// BuildTime is set by build flags during release (ISO 8601 format).
var BuildTime = "unknown"

// Command represents a parsed CLI command.
type Command struct {
	Version      bool
	Help         bool
	Provider     string
	ClaudeArgs   []string
	Validate     bool
	ValidateOpts *ValidateCommand
	Patch        bool
	PatchOpts    *PatchCommandOptions
}

// ValidateCommand represents options for the validate command.
type ValidateCommand struct {
	Provider    string // Empty means current provider
	ValidateAll bool
}

// PatchCommandOptions represents options for the patch command.
type PatchCommandOptions struct {
	Reset bool // --reset flag, true means restore original claude
}

// Parse parses command-line arguments.
func Parse(args []string) *Command {
	cmd := &Command{}
	// 根据第一个参数判断是否是ccc的参数，其余参数透传给claude
	firstArg := ""
	if len(args) > 0 {
		firstArg = args[0]
	}
	if firstArg == "--version" || firstArg == "-v" {
		cmd.Version = true
	} else if firstArg == "--help" || firstArg == "-h" {
		cmd.Help = true
	} else if firstArg == "validate" {
		cmd.Validate = true
		cmd.ValidateOpts = parseValidateArgs(args[1:])
	} else if firstArg == "patch" {
		cmd.Patch = true
		cmd.PatchOpts = parsePatchArgs(args[1:])
	} else if !strings.HasPrefix(firstArg, "-") {
		cmd.Provider = firstArg
		if len(args) > 1 {
			cmd.ClaudeArgs = args[1:]
		}
	} else {
		cmd.ClaudeArgs = args
	}
	return cmd
}

// parseValidateArgs parses arguments for the validate command.
func parseValidateArgs(args []string) *ValidateCommand {
	opts := &ValidateCommand{}

	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.Usage = func() {} // Suppress default usage output
	all := fs.Bool("all", false, "validate all providers")

	if err := fs.Parse(args); err != nil {
		// On parse error, return options with defaults
		return opts
	}

	opts.ValidateAll = *all

	// Get remaining arguments as positional args
	remaining := fs.Args()
	if len(remaining) > 0 {
		opts.Provider = remaining[0]
	}

	return opts
}

// parsePatchArgs parses arguments for the patch command.
func parsePatchArgs(args []string) *PatchCommandOptions {
	opts := &PatchCommandOptions{}

	fs := flag.NewFlagSet("patch", flag.ContinueOnError)
	fs.Usage = func() {} // Suppress default usage output
	reset := fs.Bool("reset", false, "restore original claude command")

	if err := fs.Parse(args); err != nil {
		// On parse error, return options with defaults
		return opts
	}

	opts.Reset = *reset

	return opts
}

// ShowHelp displays usage information.
func ShowHelp(cfg *config.Config, cfgErr error) {
	help := `Usage: ccc [provider] [args...]
       ccc validate [provider] [--all]
       ccc patch [--reset]

Claude Code Configuration Switcher

Commands:
  ccc                    Use the current provider (or the first provider if none is set)
  ccc <provider>         Switch to the specified provider and run Claude Code
  ccc validate           Validate the current provider configuration
  ccc validate <provider>         Validate a specific provider configuration
  ccc validate --all              Validate all provider configurations
  ccc patch               Replace claude command with ccc (requires sudo)
  ccc patch --reset       Restore original claude command (requires sudo)
  ccc --help             Show this help message
  ccc --version          Show version information

Environment Variables:
  CCC_CONFIG_DIR         Override the configuration directory (default: ~/.claude/)
`
	fmt.Print(help)

	// Display config path and status
	configPath := config.GetConfigPath()
	if cfgErr != nil {
		errMsg := provider.ShortenError(cfgErr, 40)
		fmt.Printf("\nCurrent config: %s (%s)\n", configPath, errMsg)
	} else {
		fmt.Printf("\nCurrent config: %s\n", configPath)

		// Display provider list from config
		if cfg != nil && len(cfg.Providers) > 0 {
			fmt.Println("\nAvailable Providers:")
			for name := range cfg.Providers {
				marker := ""
				if name == cfg.CurrentProvider {
					marker = " (current)"
				}
				fmt.Printf("  %s%s\n", name, marker)
			}
		}
	}
	fmt.Println()
}

// ShowVersion displays version information.
func ShowVersion() {
	fmt.Printf("%s version %s (built at %s)\n", Name, Version, BuildTime)
}

// Run executes the CLI command.
func Run(cmd *Command) error {
	// Handle patch subcommand
	if cmd.Patch {
		return RunPatch(cmd.PatchOpts)
	}

	// Handle --version
	if cmd.Version {
		ShowVersion()
		return nil
	}

	// Handle --help
	if cmd.Help {
		cfg, err := config.Load()
		ShowHelp(cfg, err)
		return nil
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		// Try to migrate from existing settings.json
		if migration.CheckExisting() && migration.PromptUser() {
			if err := migration.MigrateFromSettings(); err != nil {
				return fmt.Errorf("error migrating from settings: %w", err)
			}
			// Reload config after migration
			cfg, err = config.Load()
			if err != nil {
				ShowHelp(nil, err)
				return err
			}
		} else {
			ShowHelp(nil, err)
			return err
		}
	}

	if cmd.Validate {
		return runValidate(cfg, cmd.ValidateOpts)
	}

	// Run claude with the provider (provider determination is inside runClaude)
	return runClaude(cfg, cmd)
}

// runValidate executes the validate command.
func runValidate(cfg *config.Config, opts *ValidateCommand) error {
	// Guard: refuse to validate when settings.json contains env keys that would
	// silently override provider env. The check mirrors the launch-path guard so
	// users get the same actionable error from both entry points.
	if err := checkValidateEnvConflict(cfg, opts); err != nil {
		return err
	}

	// Create a config adapter for the validate package
	cfgAdapter := &configAdapter{cfg: cfg}

	validateOpts := &validate.RunOptions{
		Provider:    opts.Provider,
		ValidateAll: opts.ValidateAll,
	}

	return validate.Run(cfgAdapter, validateOpts)
}

// checkValidateEnvConflict runs the same env-conflict guard used by runClaude, scoped
// to the providers about to be validated:
//   - --all: union of base env + every provider's env (from-strictest stance).
//   - --provider X (or current): base env + provider X's env.
//
// Returns nil when no conflicts, a formatted error otherwise.
func checkValidateEnvConflict(cfg *config.Config, opts *ValidateCommand) error {
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

	providerNames := validateTargetProviders(cfg, opts)
	for _, name := range providerNames {
		providerSettings, ok := cfg.Providers[name]
		if !ok {
			continue
		}
		for key := range config.GetEnv(providerSettings) {
			managedEnvKeys[key] = true
		}
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

// validateTargetProviders returns the provider names whose env keys should contribute
// to managedEnvKeys for the validate guard. For --all (or when a specific provider is
// not provided and there's no current provider), it returns every configured provider.
func validateTargetProviders(cfg *config.Config, opts *ValidateCommand) []string {
	if opts.ValidateAll {
		names := make([]string, 0, len(cfg.Providers))
		for name := range cfg.Providers {
			names = append(names, name)
		}
		return names
	}

	name := opts.Provider
	if name == "" {
		name = cfg.CurrentProvider
	}
	if name == "" {
		// No target identifiable: treat as --all to stay strict.
		names := make([]string, 0, len(cfg.Providers))
		for n := range cfg.Providers {
			names = append(names, n)
		}
		return names
	}
	return []string{name}
}

// configAdapter adapts config.Config to the validate.Config interface.
type configAdapter struct {
	cfg *config.Config
}

func (a *configAdapter) Providers() map[string]map[string]interface{} {
	return a.cfg.Providers
}

func (a *configAdapter) CurrentProvider() string {
	return a.cfg.CurrentProvider
}

// Execute is the main entry point for the CLI.
func Execute() error {
	cmd := Parse(os.Args[1:])
	return Run(cmd)
}
