# ccc - Claude Code Configuration Switcher

[English](README.md) | [中文文档](README-CN.md)

## Why ccc?

`ccc` is a CLI tool that provides seamless provider switching for Claude Code. Switch between Kimi, GLM, MiniMax, and other providers with one command.

## Quick Start

### 1. Install

#### Option A: One-line install (Linux / macOS)

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]'); ARCH=$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/'); curl -LO "https://github.com/guyskk/claude-code-supervisor/releases/latest/download/ccc-${OS}-${ARCH}" && sudo install -m 755 "ccc-${OS}-${ARCH}" /usr/local/bin/ccc && rm "ccc-${OS}-${ARCH}" && ccc --version
```

#### Option B: Download from [Releases](https://github.com/guyskk/claude-code-supervisor/releases)

Download the binary for your platform (`ccc-darwin-arm64`, `ccc-linux-amd64`, etc.) and install to `/usr/local/bin/`.

### 2. Configure

If you already have `~/.claude/settings.json`, the first time you run `ccc` it will prompt to migrate and automatically generate the ccc config at `~/.claude/ccc.json`.

You can also create the config file manually:

```json
{
  "settings": {
    "permissions": {
      "defaultMode": "bypassPermissions"
    }
  },
  "providers": {
    "glm": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "glm-4.7"
      }
    },
    "kimi": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://api.moonshot.cn/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "kimi-k2-thinking"
      }
    }
  }
}
```

> **Security Warning**: `bypassPermissions` allows Claude Code to execute tools without confirmation. Only use this in trusted environments.

### 3. Use

```bash
# Show help
ccc --help

# Switch to a provider and run Claude Code
ccc glm

# Run with current provider
ccc

# Pass any Claude Code arguments
ccc glm -p
```

### 4. Validate (Optional)

Verify your provider configuration:

```bash
# Validate current provider
ccc validate

# Validate all providers
ccc validate --all
```

## Patch Command: Replace `claude` with `ccc`

Make `ccc` your default Claude Code by replacing the system `claude` command.

```bash
# Replace claude command with ccc (requires sudo)
sudo ccc patch

# After patching, `claude` command now uses ccc
claude --help    # Shows ccc help

# Restore original claude command
sudo ccc patch --reset
```

## Configuration

Config file location, default: `~/.claude/ccc.json`

### Configuration Merge Strategy

**User settings are preserved.** ccc follows these principles when merging configuration:

1. **Priority**: User's `settings.json` configuration has the highest priority
2. **Provider settings**: Provider-specific configuration from ccc.json overrides base settings
3. **Base settings**: The `settings` field in ccc.json serves as a shared template

#### What Gets Preserved

- ✅ User-installed plugins (`enabledPlugins`) - never overwritten
- ✅ Manual `settings.json` edits (permissions, sandbox, etc.) - fully preserved
- ✅ User-configured hooks (PreToolUse, SessionStart, etc.) - respected
- ✅ Environment variables you manually set in settings.json - ccc never edits them

#### What Gets Managed

- 🧹 Supervisor Stop hook - automatically cleaned up if present from previous versions

#### Environment Variable Conflicts (Hard Guard)

Claude Code's `settings.json` `env` field **overrides** environment variables passed by ccc when launching claude (empirically verified). If `settings.json` contains keys that would shadow the provider env, switching providers silently fails (wrong base_url / token / model).

**ccc refuses to start claude — and refuses to run `ccc validate` — when it detects such conflicts.** Instead of silently editing your file, ccc prints the offending keys (without their values, to avoid leaking secrets) and tells you how to fix it. ccc never modifies the `env` field of your `settings.json`.

A key is considered conflicting if it:
- starts with `ANTHROPIC_` or `CLAUDE_`, **or**
- collides with any key defined in ccc.json's base / provider `env`.

**How to fix:** remove those keys from `~/.claude/settings.json`'s `env` and move provider-related configuration into `providers.<name>.env` in `~/.claude/ccc.json`.

#### How It Works

When you run `ccc`:
1. Your existing `settings.json` is read (if it exists)
2. Configuration is merged with priority: `user > provider > base`
3. Environment variables from provider are passed via command line (not written to settings.json)
4. Leftover Supervisor hooks are cleaned up

This ensures your manual configuration is never lost!

```json
{
  "settings": {
    "permissions": {
      "defaultMode": "bypassPermissions"
    },
    "alwaysThinkingEnabled": true
  },
  "claude_args": ["--verbose"],
  "current_provider": "glm",
  "providers": {
    "glm": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "glm-4.7"
      }
    },
    "kimi": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://api.moonshot.cn/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "kimi-k2-thinking",
        "ANTHROPIC_SMALL_FAST_MODEL": "kimi-k2-0905-preview"
      }
    }
  }
}
```

### Config Fields

| Field               | Description                                  |
| ------------------- | -------------------------------------------- |
| `settings`          | Shared Claude Code config template for all providers |
| `claude_args`       | Fixed arguments to pass to Claude Code (optional) |
| `current_provider`  | Currently used provider (auto-managed by ccc) |
| `providers.{name}`  | Provider-specific Claude Code configuration  |

### Provider Configuration

Each provider only needs to specify the fields it wants to override. Common fields:

| Field                             | Description                    |
| --------------------------------- | ------------------------------ |
| `env.ANTHROPIC_BASE_URL`          | API endpoint URL               |
| `env.ANTHROPIC_AUTH_TOKEN`        | API key/token                  |
| `env.ANTHROPIC_MODEL`             | Main model to use              |
| `env.ANTHROPIC_SMALL_FAST_MODEL`  | Fast model for quick tasks     |

**How merging works**: Provider settings are deep-merged with the base template. Provider `env` takes precedence over `settings.env`.

### Environment Variables

| Variable           | Description                                        |
| ------------------ | -------------------------------------------------- |
| `CCC_CONFIG_DIR`   | Override config directory (default: `~/.claude/`)   |

```bash
# Debug with custom config directory
CCC_CONFIG_DIR=./tmp ccc glm
```

## Building from Source

```bash
# Build for all platforms
./build.sh --all

# Build for specific platforms
./build.sh -p darwin-arm64,linux-amd64

# Custom output directory
./build.sh -o ./bin
```

**Supported platforms:** `darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm64`

## License

MIT License - see LICENSE file for details.
