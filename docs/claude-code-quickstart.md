# Quickstart

> Welcome to Claude Code!

This quickstart guide will have you using AI-powered coding assistance in just a few minutes. By the end, you'll understand how to use Claude Code for common development tasks.

## Before you begin

Make sure you have:

- A terminal or command prompt open
- A code project to work with
- A Claude.ai (recommended) or Claude Console account

## Step 1: Install Claude Code

To install Claude Code, use one of the following methods:

- Native Install (Recommended)
- Homebrew
- NPM

**macOS, Linux, WSL:**

```bash
curl -fsSL https://claude.ai/install.sh | bash
```

**Windows PowerShell:**

```powershell
irm https://claude.ai/install.ps1 | iex
```

**Windows CMD:**

```bash
curl -fsSL https://claude.ai/install.cmd -o install.cmd && install.cmd && del install.cmd
```

**Homebrew:**

```bash
brew install --cask claude-code
```

## Step 2: Log in to your account

Claude Code requires an account to use. When you start an interactive session with the `claude` command, you'll need to log in:

```bash
claude
# You'll be prompted to log in on first use
```

```bash
/login
# Follow the prompts to log in with your account
```

You can log in using either account type:

- Claude.ai (subscription plans - recommended)
- Claude Console (API access with pre-paid credits)

Once logged in, your credentials are stored and you won't need to log in again.

## Step 3: Start your first session

Open your terminal in any project directory and start Claude Code:

```bash
cd /path/to/your/project
claude
```

You'll see the Claude Code welcome screen with your session information, recent conversations, and latest updates. Type `/help` for available commands or `/resume` to continue a previous conversation.

## Step 4: Ask your first question

Let's start with understanding your codebase. Try one of these commands:

```
> what does this project do?
```

Claude will analyze your files and provide a summary. You can also ask more specific questions:

```
> what technologies does this project use?
```

```
> where is the main entry point?
```

```
> explain the folder structure
```

You can also ask Claude about its own capabilities:

```
> what can Claude Code do?
```

```
> how do I use slash commands in Claude Code?
```

```
> can Claude Code work with Docker?
```

## Step 5: Make your first code change

Now let's make Claude Code do some actual coding. Try a simple task:

```
> add a hello world function to the main file
```

Claude Code will:

1. Find the appropriate file
2. Show you the proposed changes
3. Ask for your approval
4. Make the edit

## Step 6: Use Git with Claude Code

Claude Code makes Git operations conversational:

```
> what files have I changed?
```

```
> commit my changes with a descriptive message
```

You can also prompt for more complex Git operations:

```
> create a new branch called feature/quickstart
```

```
> show me the last 5 commits
```

```
> help me resolve merge conflicts
```

## Step 7: Fix a bug or add a feature

Claude is proficient at debugging and feature implementation.

Describe what you want in natural language:

```
> add input validation to the user registration form
```

Or fix existing issues:

```
> there's a bug where users can submit empty forms - fix it
```

Claude Code will:

- Locate the relevant code
- Understand the context
- Implement a solution
- Run tests if available

## Step 8: Test out other common workflows

There are a number of ways to work with Claude:

**Refactor code**

```
> refactor the authentication module to use async/await instead of callbacks
```

**Write tests**

```
> write unit tests for the calculator functions
```

**Update documentation**

```
> update the README with installation instructions
```

**Code review**

```
> review my changes and suggest improvements
```

## Essential commands

Here are the most important commands for daily use:

| Command | What it does | Example |
| :--- | :--- | :--- |
| `claude` | Start interactive mode | `claude` |
| `claude "task"` | Run a one-time task | `claude "fix the build error"` |
| `claude -p "query"` | Run one-off query, then exit | `claude -p "explain this function"` |
| `claude -c` | Continue most recent conversation | `claude -c` |
| `claude -r` | Resume a previous conversation | `claude -r` |
| `claude commit` | Create a Git commit | `claude commit` |
| `/clear` | Clear conversation history | `> /clear` |
| `/help` | Show available commands | `> /help` |
| `exit` or Ctrl+C | Exit Claude Code | `> exit` |

See the [CLI reference](https://code.claude.com/docs/en/cli-reference) for a complete list of commands.

## Pro tips for beginners

- Start with clear, specific requests
- Use natural language - no need for special syntax
- Review proposed changes before approving
- Ask Claude to explain its actions
- Use `/compact` if the conversation gets too long
- Check `/context` to see token usage

## What's next?

Now that you've learned the basics, explore more advanced features:

- [Hooks](https://code.claude.com/docs/en/hooks) - Automate workflows with event handlers
- [Subagents](https://code.claude.com/docs/en/sub-agents) - Create specialized AI assistants
- [MCP](https://code.claude.com/docs/en/mcp) - Connect to external tools and data sources
- [Chrome integration](https://code.claude.com/docs/en/chrome) - Browser automation and testing
- [Common workflows](https://code.claude.com/docs/en/common-workflows) - Advanced usage patterns

## Getting help

- **In Claude Code**: Type `/help` or ask "how do I…"
- **Documentation**: You're here! Browse other guides
- **Community**: Join our Discord for tips and support
