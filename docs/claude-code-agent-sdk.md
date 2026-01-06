# Agent SDK overview

> Build production AI agents with Claude Code as a library

The Claude Code SDK has been renamed to the Claude Agent SDK. If you're migrating from the old SDK, see the Migration Guide.

Build AI agents that autonomously read files, run commands, search the web, edit code, and more. The Agent SDK gives you the same tools, agent loop, and context management that power Claude Code, programmable in Python and TypeScript.

## Python quickstart

```python
import asyncio
from claude_agent_sdk import query, ClaudeAgentOptions

async def main():
    async for message in query(
        prompt="Find and fix the bug in auth.py",
        options=ClaudeAgentOptions(allowed_tools=["Read", "Edit", "Bash"])
    ):
        print(message)  # Claude reads the file, finds the bug, edits it

asyncio.run(main())
```

## TypeScript quickstart

```typescript
import { query, type ClaudeAgentOptions } from '@anthropic-ai/agent-sdk';

async function main() {
  for await (const message of query({
    prompt: 'Find and fix the bug in auth.py',
    options: {
      allowedTools: ['Read', 'Edit', 'Bash']
    } as ClaudeAgentOptions
  })) {
    console.log(message); // Claude reads the file, finds the bug, edits it
  }
}

main();
```

The Agent SDK includes built-in tools for reading files, running commands, and editing code, so your agent can start working immediately without you implementing tool execution. Dive into the quickstart or explore real agents built with the SDK:

- [Quickstart](https://docs.claude.com/en/docs/agent-sdk/quickstart) - Build a bug-fixing agent in minutes
- [Example agents](https://docs.claude.com/en/docs/agent-sdk/example-agents) - Email assistant, research agent, and more

## Capabilities

Everything that makes Claude Code powerful is available in the SDK:

### Built-in tools

The SDK includes all the tools that power Claude Code:

- **File operations**: Read, Write, Edit files on disk
- **Command execution**: Run bash commands with Bash tool
- **Search**: Find files with Glob, search content with Grep
- **Web**: Fetch web pages with WebFetch, search with WebSearch
- **Advanced**: Task delegation, todo management, user interaction

### Subagents

Create specialized subagents for different tasks, just like Claude Code:

- Define custom subagents with specific prompts and tools
- Delegate tasks to the most appropriate subagent
- Each subagent has its own context window

### Hooks

Automate workflows with event-driven hooks:

- **PreToolUse**: Run commands before a tool executes
- **PostToolUse**: Run commands after a tool completes
- **SessionStart**: Initialize environment when starting
- **UserPromptSubmit**: Process user input before sending to Claude
- **Stop**: Handle session termination

### MCP integration

Connect to external tools and data sources via MCP servers:

- Use existing MCP servers or create your own
- Extend agent capabilities with external APIs
- Access databases, issue trackers, cloud services

### Sessions

Maintain conversation state across multiple interactions:

- Resume previous conversations
- Manage multiple sessions
- Control context and memory

### Permissions

Control what the agent can do:

- Allowed tools
- Permission modes (auto-accept, ask, deny)
- Filesystem access controls

### Claude Code features

The SDK also supports Claude Code's filesystem-based configuration. To use these features, set `setting_sources=["project"]` (Python) or `settingSources: ['project']` (TypeScript) in your options.

| Feature | Description | Location |
| :--- | :--- | :--- |
| Skills | Specialized capabilities defined in Markdown | `.claude/skills/SKILL.md` |
| Slash commands | Custom commands for common tasks | `.claude/commands/*.md` |
| Memory | Project context and instructions | `CLAUDE.md` or `.claude/CLAUDE.md` |
| Plugins | Extend with custom commands, agents, and MCP servers | Programmatic via `plugins` option |

## Get started

**Ready to build?** Follow the Quickstart to create an agent that finds and fixes bugs in minutes.

## Compare the Agent SDK to other Claude tools

The Claude platform offers multiple ways to build with Claude. Here's how the Agent SDK fits in:

| Tool | Best for | Key capabilities |
| :--- | :--- | :--- |
| **Agent SDK** | Building autonomous AI agents | Full tool access, subagents, hooks, sessions, MCP |
| **Client SDK** | API requests to Claude | Simple request/response, streaming |
| **Claude Code CLI** | Interactive terminal usage | All features, interactive REPL |
| **Messages API** | Direct API integration | Low-level API access |

**When to use the Agent SDK:**

- You need autonomous agents that can take actions
- You want the same features as Claude Code in your own application
- You need complex workflows with multiple steps
- You want to integrate with external tools via MCP

**When to use the Client SDK:**

- You only need to make API requests to Claude
- You don't need tool use or autonomous behavior
- You want a simpler, lighter-weight solution

## Changelog

View the full changelog for SDK updates, bug fixes, and new features:

- **TypeScript SDK**: view [CHANGELOG.md](https://github.com/anthropics/agent-sdk/blob/main/typescript/CHANGELOG.md)
- **Python SDK**: view [CHANGELOG.md](https://github.com/anthropics/agent-sdk/blob/main/python/CHANGELOG.md)

## Reporting bugs

If you encounter bugs or issues with the Agent SDK:

- **TypeScript SDK**: report issues on [GitHub](https://github.com/anthropics/agent-sdk/issues)
- **Python SDK**: report issues on [GitHub](https://github.com/anthropics/agent-sdk/issues)

## Branding guidelines

For partners integrating the Claude Agent SDK, use of Claude branding is optional. When referencing Claude in your product:

**Allowed:**

- "Claude Agent" (preferred for dropdown menus)
- "Claude" (when within a menu already labeled "Agents")
- "{YourAgentName} Powered by Claude" (if you have an existing agent name)

**Not permitted:**

- "Claude Code" or "Claude Code Agent"
- Claude Code-branded ASCII art or visual elements that mimic Claude Code

Your product should maintain its own branding and not appear to be Claude Code or any Anthropic product. For questions about branding compliance, contact our sales team.

## License and terms

Use of the Claude Agent SDK is governed by Anthropic's Commercial Terms of Service, including when you use it to power products and services that you make available to your own customers and end users, except to the extent a specific component or dependency is covered by a different license as indicated in that component's LICENSE file.

## Installation

### 1. Install Claude Code

The SDK uses Claude Code as its runtime:

**macOS/Linux/WSL (Native):**

```bash
curl -fsSL https://claude.ai/install.sh | bash
```

**macOS (Homebrew):**

```bash
brew install --cask claude-code
```

**Linux/WSL (npm):**

```bash
npm install -g @anthropic-ai/claude-code
```

See Claude Code setup for Windows and other options.

### 2. Install the SDK

**TypeScript:**

```bash
npm install @anthropic-ai/agent-sdk
```

**Python:**

```bash
pip install claude-agent-sdk
```

### 3. Set your API key

```bash
export ANTHROPIC_API_KEY=your-api-key
```

Get your key from the [Console](https://console.anthropic.com/).

The SDK also supports authentication via third-party API providers:

- **Amazon Bedrock**: set `CLAUDE_CODE_USE_BEDROCK=1` environment variable and configure AWS credentials
- **Google Vertex AI**: set `CLAUDE_CODE_USE_VERTEX=1` environment variable and configure Google Cloud credentials
- **Microsoft Foundry**: set `CLAUDE_CODE_USE_FOUNDRY=1` environment variable and configure Azure credentials

Unless previously approved, we do not allow third party developers to offer Claude.ai login or rate limits for their products, including agents built on the Claude Agent SDK. Please use the API key authentication methods described in this document instead.

### 4. Run your first agent

This example creates an agent that lists files in your current directory using built-in tools.

**Python:**

```python
import asyncio
from claude_agent_sdk import query, ClaudeAgentOptions

async def main():
    async for message in query(
        prompt="What files are in this directory?",
        options=ClaudeAgentOptions(allowed_tools=["Bash", "Glob"])
    ):
        if hasattr(message, "result"):
            print(message.result)

asyncio.run(main())
```

**TypeScript:**

```typescript
import { query, type ClaudeAgentOptions } from '@anthropic-ai/agent-sdk';

async function main() {
  for await (const message of query({
    prompt: 'What files are in this directory?',
    options: {
      allowedTools: ['Bash', 'Glob']
    } as ClaudeAgentOptions
  })) {
    if ('result' in message) {
      console.log(message.result);
    }
  }
}

main();
```

## Agent SDK vs Client SDK

The Agent SDK and Client SDK serve different purposes:

| Aspect | Agent SDK | Client SDK |
| :--- | :--- | :--- |
| **Use case** | Build autonomous agents | Make API requests |
| **Tool use** | Built-in, automatic | Manual implementation |
| **State management** | Sessions, memory | None |
| **Subagents** | Built-in support | Not applicable |
| **Hooks** | Built-in support | Not applicable |
| **Complexity** | Higher | Lower |
| **Control** | Structured patterns | Full control |

**Choose Agent SDK when:**

- You need agents that can autonomously complete tasks
- You want built-in tools and workflows
- You need session management and memory
- You want to use subagents for delegation

**Choose Client SDK when:**

- You only need to send requests to Claude
- You want full control over the API interaction
- You don't need autonomous behavior
- You prefer a simpler, lighter-weight solution

## Agent SDK vs Claude Code CLI

The Agent SDK and Claude Code CLI share the same capabilities:

| Aspect | Agent SDK | Claude Code CLI |
| :--- | :--- | :--- |
| **Interface** | Programmatic API | Interactive terminal |
| **Use case** | Build applications | Interactive development |
| **Integration** | Embed in your app | Standalone tool |
| **Control** | Full programmatic control | Commands and flags |

**Choose Agent SDK when:**

- You're building a custom application
- You need programmatic control
- You want to embed Claude in your product
- You need custom workflows

**Choose Claude Code CLI when:**

- You want an interactive development experience
- You prefer terminal-based workflows
- You don't need programmatic access
- You want quick, ad-hoc assistance

## Example agents

Explore what's possible with the Agent SDK:

### Email assistant

An agent that can read, compose, and manage emails:

```python
from claude_agent_sdk import query, ClaudeAgentOptions

async for message in query(
    prompt="Draft a response to the email from John about the Q4 report",
    options=ClaudeAgentOptions(allowed_tools=["Read", "Write", "WebSearch"])
):
    print(message)
```

### Research agent

An agent that can search the web, read papers, and summarize findings:

```python
async for message in query(
    prompt="Research the latest developments in quantum computing and summarize the key papers",
    options=ClaudeAgentOptions(
        allowed_tools=["WebSearch", "WebFetch", "Read", "Write"],
        max_turns=10
    )
):
    print(message)
```

### Code review agent

An agent that reviews code for issues:

```python
async for message in query(
    prompt="Review the changes in this PR for security vulnerabilities",
    options=ClaudeAgentOptions(
        allowed_tools=["Bash", "Read", "Grep"],
        tools={
            "Bash": {"command": "git diff origin/main...HEAD"}
        }
    )
):
    print(message)
```

## Next steps

- [Quickstart](https://docs.claude.com/en/docs/agent-sdk/quickstart) - Build an agent that finds and fixes bugs in minutes
- [TypeScript SDK](https://docs.claude.com/en/docs/agent-sdk/typescript) - Full TypeScript API reference and examples
- [Python SDK](https://docs.claude.com/en/docs/agent-sdk/python) - Full Python API reference and examples
- [Built-in tools](https://docs.claude.com/en/docs/agent-sdk/built-in-tools) - Complete reference for all available tools
- [Hooks](https://docs.claude.com/en/docs/agent-sdk/hooks) - Automate workflows with event handlers
- [Subagents](https://docs.claude.com/en/docs/agent-sdk/subagents) - Create specialized agents for delegation
- [MCP](https://docs.claude.com/en/docs/agent-sdk/mcp) - Connect to external tools and data
- [Permissions](https://docs.claude.com/en/docs/agent-sdk/permissions) - Control agent capabilities
- [Sessions](https://docs.claude.com/en/docs/agent-sdk/sessions) - Maintain conversation state
