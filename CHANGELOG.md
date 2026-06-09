# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.0] - 2026-06-09

### Fixed

- Override stale `ANTHROPIC_*`/`CLAUDE_*` keys in `settings.json` via `--settings` (#86)
  - When switching providers, stale env vars from previous provider no longer leak into session

### Changed

- Use `--settings` CLI param to automatically override `settings.json` env conflicts (#85)
  - Replace hard guard (refuse-to-start) with automatic override via `--settings` (higher priority)
  - Remove `env_guard.go` and related conflict detection code
  - Behavior: settings.json env conflict → auto-override (was: error and abort)

## [0.4.0] - 2026-05-20

### Removed

- **BREAKING**: Supervisor mode (replaced by Claude Code's built-in goal and loop functionality)
  - Removed all supervisor-related code, tests, config, and dependencies
  - Removed `supervisor-hook` and `supervisor-mode` CLI commands
  - Removed slash command files (`supervisor.md`, `supervisoroff.md`)
  - Project rebranded as "Claude Code Configuration Switcher"
  - Repository renamed back to `claude-code-config-switcher` (GitHub redirects the old URL)

### Added

- **Hard guard for settings.json env conflicts**: detect and reject conflicts between
  provider env vars and user-defined env in settings.json to prevent silent overrides

### Changed

- Streamlined README/README-CN: trimmed the Configuration Merge Strategy section
  and removed leftover Supervisor references now that supervisor mode is gone

## [0.3.3] - 2026-03-26

### Fixed

- Preserve user-defined env in `settings.json` when switching providers (#76)

## [0.3.2] - 2026-02-24

### Added

- **Settings merge strategy**: preserve user-defined config in `settings.json`
  when switching providers instead of overwriting (#73)
- Documentation of configuration merge strategy in README and README-CN

### Changed

- Reverted PreToolUse hook integration to keep CLI behavior simple (#69)
- Simplified README-CN for better user experience
- Updated `CLAUDE.md` with simplified development process

### Removed

- SpecKit-related files and configuration (#70)
- PreToolUse hook related code, tests, and configuration

### Fixed

- Handle slices in `deepCopy` and add a timeout constant in config merging

## [0.3.1] - 2026-01-19

### Fixed

- Make `supervisor-mode` command silent when setting the mode (#65)
- Environment variable precedence issue (#63)
- Add beta headers for `validate` command to support all providers (#64)

## [0.3.0] - 2026-01-16

### Added

- **Patch command**: Replace system `claude` command with `ccc` (`sudo ccc patch`)
  - Enables any tool calling `claude` to use ccc with configured providers
  - Supports `--reset` flag to restore original claude command
  - Uses `CCC_CLAUDE` environment variable to avoid recursive calls
- **supervisor-mode query mode**: Query current status without arguments
  - `ccc supervisor-mode` outputs "on" or "off" to stdout
  - Enables use in statusline scripts
- **SpecKit development workflow**: Migrated from OpenSpec to SpecKit
  - Chinese localization for all development documents
  - Project constitution with 6 core principles
  - Comprehensive spec-driven development guide
- **Claude Agent SDK test tool**: Auto-compact verification tool
- Comprehensive tests: integration and E2E tests for patch command

### Changed

- **Supervisor prompt enhancements**:
  - Added mandatory tool verification step
  - Enhanced with independent validation framework
  - Improved feedback quality with stricter review criteria
- Simplified README documentation (removed excessive technical details)
- Reorganized README-CN.md for better readability

### Fixed

- Language inconsistency in README.md Statusline section (Chinese → English)

## [0.2.1] - 2025-01-14

### Added

- Unit tests for CLI commands (`internal/cli/cli_test.go`)
- Unit tests for Supervisor mode (`internal/cli/supervisor_mode_test.go`)
- Unit tests for pretty JSON formatting (`internal/prettyjson/`)

### Changed

- **Supervisor mode activation**: Changed from environment variable to slash command (`/supervisor`)
  - Use `/supervisor` to enable supervisor mode
  - Use `/supervisor-off` to disable supervisor mode
- Renamed command file `supervisor-off.md` → `supervisoroff.md`

### Removed

- Obsolete and incomplete tests
- Tests for non-existent `Enabled` field and deprecated environment variables
- Dead code in integration tests

### Fixed

- E2E tests for supervisor-hook command

## [0.2.0] - 2025-01-13

### Added

- Supervisor Mode with automatic task review
- Support for custom supervisor prompt via `~/.claude/SUPERVISOR.md`
- Structured logging and unified error handling
- MIT License

### Changed

- Repositioned project as "Claude Code Supervisor"
- Repository renamed from `claude-code-config-switcher` to `claude-code-supervisor`

[0.5.0]: https://github.com/guyskk/claude-code-config-switcher/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/guyskk/claude-code-config-switcher/compare/v0.3.3...v0.4.0
[0.3.3]: https://github.com/guyskk/claude-code-config-switcher/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/guyskk/claude-code-config-switcher/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/guyskk/claude-code-config-switcher/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/guyskk/claude-code-config-switcher/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/guyskk/claude-code-config-switcher/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/guyskk/claude-code-config-switcher/releases/tag/v0.2.0
