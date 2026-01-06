# Common workflows

> Learn about common workflows with Claude Code.

Each task in this document includes clear instructions, example commands, and best practices to help you get the most from Claude Code.

## Understand new codebases

### Get a quick codebase overview

Suppose you've just joined a new project and need to understand its structure quickly.

**Ask Claude for a project overview:**

```
> Give me a high-level overview of this codebase - what are the main components and how do they relate to each other?
```

Claude will analyze the project structure and provide a summary of:
- Key directories and their purposes
- Main entry points
- Important modules and their relationships
- Technologies and frameworks used

### Find relevant code

Suppose you need to locate code related to a specific feature or functionality.

```
> Where is the user authentication logic implemented?
```

```
> Find all files that handle payment processing
```

```
> Show me where error handling is configured
```

Claude uses Glob and Grep tools to search for:
- File names matching patterns
- Code containing specific keywords
- Function/class definitions
- Import/export statements

---

## Fix bugs efficiently

Suppose you've encountered an error message and need to find and fix its source.

**Paste the error and ask for help:**

```
> I'm getting this error when I try to submit the form:
  "TypeError: Cannot read property 'map' of undefined"
  Can you find and fix the issue?
```

Claude will:

1. Search for the relevant code
2. Analyze the error context
3. Identify the root cause
4. Propose and implement a fix
5. Test the solution if possible

**Debug with console logs:**

```
> Run the app in development mode and check the console
  for any errors when the page loads
```

With Chrome integration, Claude can read console output directly and diagnose issues.

---

## Refactor code

Suppose you need to update old code to use modern patterns and practices.

```
> Refactor the authentication module to use async/await
  instead of callbacks
```

```
> Update the component to use React Hooks instead of
  class components
```

Claude understands:
- Design patterns and best practices
- Language-specific idioms
- Framework conventions
- Backward compatibility concerns

---

## Use specialized subagents

Suppose you want to use specialized AI subagents to handle specific tasks more effectively.

Claude Code includes built-in subagents:

- **Explore**: Fast codebase search and analysis (read-only)
- **General-purpose**: Complex multi-step tasks that require both exploration and modification
- **Plan**: Research and planning in plan mode

You can also create custom subagents for specific workflows.

---

## Use Plan Mode for safe code analysis

Plan Mode instructs Claude to create a plan by analyzing the codebase with read-only operations, perfect for exploring codebases, planning complex changes, or reviewing code safely.

### When to use Plan Mode

- **Multi-step implementation**: When your feature requires making edits to many files
- **Code exploration**: When you want to research the codebase thoroughly before changing anything
- **Interactive development**: When you want to iterate on the direction with Claude

### How to use Plan Mode

**Turn on Plan Mode during a session**

You can switch into Plan Mode during a session using **Shift+Tab** to cycle through permission modes.

If you are in Normal Mode, **Shift+Tab** first switches into Auto-Accept Mode, indicated by `⏵⏵ accept edits on` at the bottom of the terminal. A subsequent **Shift+Tab** will switch into Plan Mode, indicated by `⏸ plan mode on`.

**Start a new session in Plan Mode**

To start a new session in Plan Mode, use the `--permission-mode plan` flag:

```bash
claude --permission-mode plan
```

**Run "headless" queries in Plan Mode**

You can also run a query in Plan Mode directly with `-p` (that is, in "headless mode"):

```bash
claude --permission-mode plan -p "Analyze the authentication system and suggest improvements"
```

### Example: Planning a complex refactor

```bash
claude --permission-mode plan
```

```
> I need to refactor our authentication system to use OAuth2. Create a detailed migration plan.
```

Claude analyzes the current implementation and create a comprehensive plan. Refine with follow-ups:

```
> What about backward compatibility?
> How should we handle database migration?
```

### Configure Plan Mode as default

```json
// .claude/settings.json
{
  "permissions": {
    "defaultMode": "plan"
  }
}
```

See [settings documentation](https://code.claude.com/docs/en/settings) for more configuration options.

---

## Work with tests

Suppose you need to add tests for uncovered code.

Claude can generate tests that follow your project's existing patterns and conventions. When asking for tests, be specific about what behavior you want to verify. Claude examines your existing test files to match the style, frameworks, and assertion patterns already in use.

For comprehensive coverage, ask Claude to identify edge cases you might have missed. Claude can analyze your code paths and suggest tests for error conditions, boundary values, and unexpected inputs that are easy to overlook.

---

## Create pull requests

Suppose you need to create a well-documented pull request for your changes.

Claude can help you:

1. Review your changes with `git diff`
2. Write a clear description of what was changed and why
3. Generate a title that summarizes the changes
4. Create the PR using `gh pr create`

```
> Create a pull request for my changes with a clear description
```

---

## Handle documentation

Suppose you need to add or update documentation for your code.

```
> Update the README to include the new installation steps
```

```
> Add documentation for the new API endpoint in docs/api.md
```

```
> Create a getting started guide for new contributors
```

Claude can:
- Update existing documentation
- Create new documentation files
- Ensure consistency across docs
- Include examples and usage instructions

---

## Work with images

Suppose you need to work with images in your codebase, and you want Claude's help analyzing image content.

With the `Read` tool, Claude can view images:
- Screenshots and UI mockups
- Diagrams and architecture docs
- Design files and wireframes

```
> Review this screenshot and identify any UI issues
```

```
> Compare these two designs and tell me which is better
```

```
> Extract the text from this screenshot
```

---

## Reference files and directories

Use `@` to quickly include files or directories without waiting for Claude to read them.

```
> Review @src/components/Button.tsx and suggest improvements
```

```
> Check @tests/auth.test.js for coverage of edge cases
```

```
> Analyze @docs/api.md for consistency
```

---

## Use extended thinking (thinking mode)

Extended thinking reserves a portion of the total output token budget for Claude to reason through complex problems step-by-step. This reasoning is visible in verbose mode, which you can toggle on with `Ctrl+O`.

Extended thinking is particularly valuable for complex architectural decisions, challenging bugs, multi-step implementation planning, and evaluating tradeoffs between different approaches. It provides more space for exploring multiple solutions, analyzing edge cases, and self-correcting mistakes.

You can configure thinking mode for Claude Code in two ways:

| Scope | How to enable | Details |
| :--- | :--- | :--- |
| **Global default** | Use `/config` to toggle thinking mode on | Sets your default across all projects. Saved as `alwaysThinkingEnabled` in `~/.claude/settings.json` |
| **Environment variable override** | Set `MAX_THINKING_TOKENS` environment variable | When set, applies a custom token budget to all requests, overriding your thinking mode configuration. Example: `export MAX_THINKING_TOKENS=1024` |

### Per-request thinking with `ultrathink`

You can include `ultrathink` as a keyword in your message to enable thinking for a single request:

```
> ultrathink: design a caching layer for our API
```

Note that `ultrathink` both allocates the thinking budget AND semantically signals to Claude to reason more thoroughly, which may result in deeper thinking than necessary for your task.

The `ultrathink` keyword only works when `MAX_THINKING_TOKENS` is not set. When `MAX_THINKING_TOKENS` is configured, it takes priority and controls the thinking budget for all requests.

Other phrases like "think", "think hard", and "think more" are interpreted as regular prompt instructions and don't allocate thinking tokens.

To view Claude's thinking process, press `Ctrl+O` to toggle verbose mode and see the internal reasoning displayed as gray italic text.

See the token budget section below for detailed budget information and cost implications.

### How extended thinking token budgets work

Extended thinking uses a **token budget** that controls how much internal reasoning Claude can perform before responding.

A larger thinking token budget provides:

- More space to explore multiple solution approaches step-by-step
- Room to analyze edge cases and evaluate tradeoffs thoroughly
- Ability to revise reasoning and self-correct mistakes

Token budgets for thinking mode:

- When thinking is **enabled** (via `/config` or `ultrathink`), Claude can use up to **31,999 tokens** from your output budget for internal reasoning
- When thinking is **disabled**, Claude uses **0 tokens** for thinking

**Custom token budgets:**

- You can set a custom thinking token budget using the `MAX_THINKING_TOKENS` environment variable
- This takes highest priority and overrides the default 31,999 token budget
- See the extended thinking documentation for valid token ranges

---

## Resume previous conversations

When starting Claude Code, you can resume a previous session:

- `claude --continue` continues the most recent conversation in the current directory
- `claude --resume` opens a conversation picker or resumes by name

From inside an active session, use `/resume` to switch to a different conversation.

Sessions are stored per project directory. The `/resume` picker shows sessions from the same git repository, including worktrees.

### Name your sessions

Give sessions descriptive names to find them later. This is a best practice when working on multiple tasks or features.

```
> /rename "Add user authentication"
```

### Use the session picker

The `/resume` command (or `claude --resume` without arguments) opens an interactive session picker with these features:

**Keyboard shortcuts in the picker:**

| Shortcut | Action |
| :--- | :--- |
| `↑` / `↓` | Navigate between sessions |
| `→` / `←` | Expand or collapse grouped sessions |
| `Enter` | Select and resume the highlighted session |
| `P` | Preview the session content |
| `R` | Rename the highlighted session |
| `/` | Search to filter sessions |
| `A` | Toggle between current directory and all projects |
| `B` | Filter to sessions from your current git branch |
| `Esc` | Exit the picker or search mode |

**Session organization:**

The picker displays sessions with helpful metadata:

- Session name or initial prompt
- Time elapsed since last activity
- Message count
- Git branch (if applicable)

Forked sessions (created with `/rewind` or `--fork-session`) are grouped together under their root session, making it easier to find related conversations.

---

## Run parallel Claude Code sessions with Git worktrees

Suppose you need to work on multiple tasks simultaneously with complete code isolation between Claude Code instances.

Git worktrees allow you to have multiple working directories for the same repository, each with its own branch. Combined with Claude Code's per-directory session storage, you can:

```
# Create worktrees for different features
git worktree add ../feature-a feature-a
git worktree add ../feature-b feature-b

# Start Claude Code in each directory
cd ../feature-a && claude
cd ../feature-b && claude
```

Each Claude Code session maintains:
- Separate conversation history
- Independent context and state
- Isolated tool execution

---

## Use Claude as a unix-style utility

### Add Claude to your verification process

Suppose you want to use Claude Code as a linter or code reviewer.

**Add Claude to your build script:**

```json
// package.json
{
    ...
    "scripts": {
        ...
        "lint:claude": "claude -p 'you are a linter. please look at the changes vs. main and report any issues related to typos. report the filename and line number on one line, and a description of the issue on the second line. do not return any other text.'"
    }
}
```

### Pipe in, pipe out

Suppose you want to pipe data into Claude, and get back data in a structured format.

**Pipe data through Claude:**

```bash
cat build-error.txt | claude -p 'concisely explain the root cause of this build error' > output.txt
```

### Control output format

Suppose you need Claude's output in a specific format, especially when integrating Claude Code into scripts or other tools.

```bash
# JSON output for scripting
claude -p --output-format json "list all TypeScript files" | jq '.[]'

# Stream JSON for real-time processing
claude -p --output-format stream-json "analyze this code" | while read -r line; do
  # Process each message
done
```

---

## Create custom slash commands

Claude Code supports custom slash commands that you can create to quickly execute specific prompts or tasks.

For more details, see the [Slash commands reference page](https://code.claude.com/docs/en/slash-commands).

### Create project-specific commands

Suppose you want to create reusable slash commands for your project that all team members can use.

**Create a project command:**

```bash
mkdir -p .claude/commands
cat > .claude/commands/review.md << 'EOF'
Review this code for:
- Security vulnerabilities
- Performance issues
- Code style violations
EOF
```

Now anyone on the project can use `/review`.

### Add command arguments with $ARGUMENTS

Suppose you want to create flexible slash commands that can accept additional input from users.

```bash
cat > .claude/commands/fix-issue.md << 'EOF'
Fix issue #$ARGUMENTS following our coding standards
EOF
```

Usage:

```
> /fix-issue 123 high-priority
```

### Create personal slash commands

Suppose you want to create personal slash commands that work across all your projects.

```bash
mkdir -p ~/.claude/commands
cat > ~/.claude/commands/security-review.md << 'EOF'
Review this code for security vulnerabilities
EOF
```

---

## Ask Claude about its capabilities

Claude has built-in access to its documentation and can answer questions about its own features and limitations.

### Example questions

```
> can Claude Code create pull requests?
```

```
> how does Claude Code handle permissions?
```

```
> what slash commands are available?
```

```
> how do I use MCP with Claude Code?
```

```
> how do I configure Claude Code for Amazon Bedrock?
```

```
> what are the limitations of Claude Code?
```

---

## Next steps

- [Hooks](https://code.claude.com/docs/en/hooks) - Automate workflows with event handlers
- [Subagents](https://code.claude.com/docs/en/sub-agents) - Create specialized AI assistants
- [MCP](https://code.claude.com/docs/en/mcp) - Connect to external tools and data
- [Chrome integration](https://code.claude.com/docs/en/chrome) - Browser automation

## Claude Code reference implementation

Clone our development container reference implementation.
