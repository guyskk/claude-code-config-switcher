# Claude Code Configuration Switcher (ccc)

A command-line tool for switching between different Claude Code configurations.

## Overview

`ccc` (Claude Code Config) allows you to easily switch between different Claude Code provider configurations (e.g., Kimi, GLM, Doubao) without manually editing configuration files.

## Features

- Switch between multiple Claude Code configurations with a single command
- Automatically updates the `current_provider` setting
- Passes through all arguments to Claude Code
- Supports debug mode with custom configuration directory
- Simple and intuitive command-line interface

## Installation

1. Build the tool:
```bash
./build.sh
```

2. Install system-wide:
```bash
sudo cp dist/ccc /usr/local/bin/ccc
```

## Configuration

Create a `~/.claude/cccli.json` configuration file:

```json
{
  "settings": {
    "permissions": {
      "allow": ["Edit", "Write", "WebFetch", "WebSearch"],
      "defaultMode": "acceptEdits"
    },
    "alwaysThinkingEnabled": true,
    "timeout": 60000
  },
  "current_provider": "kimi",
  "providers": {
    "kimi": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://api.moonshot.cn/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "kimi-k2-thinking"
      }
    },
    "glm": {
      "env": {
        "ANTHROPIC_BASE_URL": "https://open.bigmodel.cn/api/anthropic",
        "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY_HERE",
        "ANTHROPIC_MODEL": "glm-4.6"
      }
    }
  }
}
```

The configuration structure:
- `settings`: Base settings template shared by all providers
- `current_provider`: The last used provider (auto-updated)
- `providers`: Provider-specific settings that will be merged with the base

When switching to a provider, the tool:
1. Starts with the base `settings`
2. Deep merges the provider's settings on top
3. Provider settings override base settings for the same keys
4. Saves the merged result to `~/.claude/settings.json`

Example configuration files are provided in the `./tmp/example/` directory.

## Usage

### Basic Commands

```bash
# Display help information
ccc --help

# Run with current provider
ccc

# Switch to and run with a specific provider
ccc kimi

# Pass arguments to Claude Code
ccc kimi --help
ccc kimi /path/to/project
```

### Environment Variables

- `CCC_WORK_DIR`: Override the configuration directory (default: `~/.claude/`)

  Useful for debugging:
  ```bash
  CCC_WORK_DIR=./tmp ccc kimi
  ```

## How It Works

1. `ccc` reads the `~/.claude/cccli.json` configuration
2. Deep merges the selected provider's settings with the base settings template
3. Writes the merged configuration to `~/.claude/settings-{provider}.json`
4. Updates the `current_provider` field in `cccli.json`
5. Executes `claude --settings ~/.claude/settings-{provider}.json [additional-args...]`

The configuration merge is recursive, so nested objects like `env` and `permissions` are properly merged.

Each provider has its own settings file (e.g., `settings-kimi.json`, `settings-glm.json`), allowing you to easily see and manage different configurations.

## Provider Configuration Files

Each provider configuration file is a complete Claude Code settings file. Example files are provided:

- `tmp/example/settings-kimi.json` - Kimi AI configuration
- `tmp/example/settings-glm.json` - GLM-4 configuration
- `tmp/example/settings-doubao.json` - Doubao configuration
- `tmp/example/settings-vibecoding.json` - Vibe coding configuration
- etc.

## Building

The included `build.sh` script builds the tool:

```bash
./build.sh
```

This creates the binary at `dist/ccc`.

## Command Line Reference

```
Usage: ccc [provider] [args...]

Commands:
  ccc              Use the current provider (or the first provider if none is set)
  ccc <provider>   Switch to the specified provider and run Claude Code
  ccc --help       Show this help message

Environment Variables:
  CCC_WORK_DIR     Override the configuration directory (default: ~/.claude/)

Examples:
  ccc              Run Claude Code with the current provider
  ccc kimi         Switch to 'kimi' provider and run Claude Code
  ccc glm          Switch to 'glm' provider and run Claude Code
```

