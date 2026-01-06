# Proposal: add-agent-sdk

## 概述

创建 `claude_agent_sdk` Go 包，封装 Claude Code 命令行工具为可编程 SDK。

## 动机

当前代码中直接使用 `exec.Command` 调用 claude，存在以下问题：
1. 进程管理复杂且容易出错
2. stdio 处理混乱
3. 没有超时控制
4. 难以测试和模拟

通过创建 SDK，可以：
1. 简化 claude 命令行交互
2. 统一进程管理
3. 提供清晰的 API
4. 便于测试和扩展
5. 为 Web 服务提供基础

## 变更内容

### 1. SDK 核心 API

```go
// internal/claude_agent_sdk/agent.go
package claude_agent_sdk

type Agent struct {
    claudePath string
    config     *Config
    logger     Logger
}

type Config struct {
    // 工作目录
    WorkingDir string

    // 设置文件路径（可选，默认使用 ~/.claude/settings.json）
    SettingsPath string

    // 环境变量
    Env map[string]string

    // 默认模型
    Model string

    // 允许的工具
    AllowedTools []string

    // 超时时间
    Timeout time.Duration
}

// Run 执行单个查询并返回结果
func (a *Agent) Run(ctx context.Context, prompt string) (*RunResult, error)

// RunStream 执行查询并返回流式事件
func (a *Agent) RunStream(ctx context.Context, prompt string) <-chan StreamEvent

// RunSession 在指定会话中执行查询
func (a *Agent) RunSession(ctx context.Context, sessionID, prompt string) (*RunResult, error)

// RunInteractive 运行交互式会话（连接 stdin/stdout）
func (a *Agent) RunInteractive(ctx context.Context) error
```

### 2. 运行选项

```go
type RunOptions struct {
    // 会话管理
    SessionID   string
    Resume      bool
    ForkSession bool

    // 输入
    Prompt      string

    // 输出控制
    OutputFormat string  // "stream-json", "json", "text"
    JSONSchema  string  // 结构化输出 schema

    // 模型配置
    Model       string

    // 环境
    Env []string
}
```

### 3. 结果和事件

```go
type RunResult struct {
    Output           string
    StructuredOutput map[string]interface{}
    Duration         time.Duration
    TokensUsed       int
    Success          bool
    Error            error
}

type StreamEvent struct {
    Type    EventType
    Content  string
    Error   error
    Meta    map[string]interface{}
}

type EventType int

const (
    EventStart EventType = iota
    EventMessage
    EventToolUse
    EventToolResult
    EventEnd
)
```

### 4. 进程管理

```go
type Process struct {
    cmd     *exec.Cmd
    ctx     context.Context
    cancel  context.CancelFunc

    stdin   io.WriteCloser
    stdout  io.ReadCloser
    stderr  io.ReadCloser

    timeout time.Duration
}

func (p *Process) Start() error
func (p *Process) Wait() error
func (p *Process) Kill() error
func (p *Process) Signal(sig syscall.Signal) error

// StdoutLine 返回按行输出的 channel
func (p *Process) StdoutLine() <-chan string

// StderrLine 返回按行错误输出的 channel
func (p *Process) StderrLine() <-chan string
```

### 5. 流式解析器

```go
type StreamParser struct {
    r io.Reader
}

func NewStreamParser(r io.Reader) *StreamParser
func (p *StreamParser) Parse() <-chan StreamEvent
func (p *StreamParser) ParseLine(line string) (StreamEvent, error)
```

## 设计原则

### 1. 简单性

- 最小化 API 表面
- 清晰的函数命名
- 合理的默认值

### 2. 可组合性

- 支持函数式选项模式
- 支持中间件拦截
- 支持自定义处理器

### 3. 可测试性

- 接口驱动设计
- 依赖注入
- 可模拟组件

## 使用示例

### 基础使用

```go
agent, err := claude_agent_sdk.NewAgent(&claude_agent_sdk.Config{
    WorkingDir: "/path/to/project",
    Model:      "claude-3-5-sonnet-20241022",
})

result, err := agent.Run(context.Background(), "List files in current directory")
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Output)
```

### 流式输出

```go
events := agent.RunStream(context.Background(), "Analyze the codebase")

for event := range events {
    switch event.Type {
    case claude_agent_sdk.EventMessage:
        fmt.Print(event.Content)
    case claude_agent_sdk.EventToolUse:
        fmt.Printf("Tool: %s\n", event.Meta["tool_name"])
    }
}
```

### 会话管理

```go
err := agent.RunSession(context.Background(), "session-123", "Continue the previous task")
```

## 影响范围

- **新增包**: `internal/claude_agent_sdk/`
- **修改**: `internal/cli/hook.go` - 使用 SDK
- **修改**: `internal/cli/exec.go` - 使用 SDK
- **修改**: `internal/web/` - 使用 SDK

## 与 agentsdk-go 的关系

调研发现 [agentsdk-go](https://github.com/cexll/agentsdk-go) 是一个完整的实现，但：

1. **我们的需求更简单**: 只需要封装 claude CLI，不需要完整的 agent 框架
2. **更好的集成**: 与我们的配置系统深度集成
3. **渐进式**: 可以逐步扩展功能

如果 agentsdk-go 满足需求，可以考虑：
1. 使用其流式解析逻辑
2. 参考其进程管理实现
3. 保持接口兼容

## 风险

| 风险 | 影响 | 缓解 |
|------|------|------|
| 过度设计 | 中 | 遵循 YAGNI 原则 |
| Claude CLI 变更 | 低 | 版本检测和兼容性处理 |
| 性能开销 | 低 | 最小化抽象层 |

## 开放问题

1. 是否需要支持多个 Claude 版本？
2. 是否需要支持自定义工具？
3. 是否需要支持 MCP 服务器集成？
