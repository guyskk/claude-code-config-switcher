# Claude Code Config Switcher

[阅读中文文档](README-CN.md)

**Switch between multiple Claude Code providers (Kimi, GLM, MiniMax, etc.) with a single command.**

## Overview

`ccc` is a CLI tool that lets you seamlessly switch between different Claude Code API provider configurations. No more manually editing config files—just run `ccc <provider>` and you're done.

## Features

- One-command switching between providers (Kimi, GLM, MiniMax, and more)
- Automatic provider configuration merging
- Pass-through of all Claude Code arguments
- Debug mode with custom config directories
- Clean, intuitive CLI interface

## Installation

### Quick Install

```bash
# Clone the repository
git clone https://github.com/guyskk/claude-code-config-switcher.git
cd claude-code-config-switcher

# Build for your platform
./build.sh

# Install system-wide (optional)
sudo cp dist/ccc-$(uname -s)-$(uname -m) /usr/local/bin/ccc
```

### Build Options

```bash
# Build for all platforms
./build.sh --all

# Build for specific platforms
./build.sh -p darwin-arm64,linux-amd64

# Custom output directory
./build.sh -o ./bin
```

**Supported platforms:** `darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm64`, `windows-amd64`

## Configuration

Create `~/.claude/ccc.json`:

```json
{
  "settings": {
    "permissions": {
      "allow": ["Edit", "MultiEdit", "Write", "WebFetch", "WebSearch"],
      "defaultMode": "acceptEdits"
    },
    "alwaysThinkingEnabled": true,
    "env": {
      "API_TIMEOUT_MS": "300000",
      "DISABLE_TELEMETRY": "1",
      "DISABLE_ERROR_REPORTING": "1",
      "DISABLE_NON_ESSENTIAL_MODEL_CALLS": "1",
      "DISABLE_BUG_COMMAND": "1",
      "DISABLE_COST_WARNINGS": "1"
    }
  },
  "current_provider": "kimi",
  "providers": {
    "kimi": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://api.moonshot.cn/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "kimi-k2-thinking",
        "ANTHROPIC_SMALL_FAST_MODEL": "kimi-k2-0905-preview"
      }
    },
    "glm": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "glm-4.7",
        "ANTHROPIC_SMALL_FAST_MODEL": "glm-4.7"
      }
    },
    "m2": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://api.minimaxi.com/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "MiniMax-M2.1",
        "ANTHROPIC_SMALL_FAST_MODEL": "MiniMax-M2.1"
      }
    }
  }
}
```

**Config structure:**
- `settings` — Base template shared by all providers
- `current_provider` — Last used provider (auto-updated)
- `providers` — Provider-specific overrides

**How it works:** When switching providers, `ccc` deep-merges the provider's config with the base template, then saves it to `~/.claude/settings-{provider}.json`.

See `./tmp/example/` for more examples.

## Usage

```bash
# Show available providers
ccc --help

# Run with current provider
ccc

# Switch to a provider
ccc kimi

# Pass arguments to Claude Code
ccc kimi --help
ccc kimi /path/to/project
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `CCC_CONFIG_DIR` | Override config directory (default: `~/.claude/`) |

```bash
# Debug with custom config
CCC_CONFIG_DIR=./tmp ccc kimi
```
