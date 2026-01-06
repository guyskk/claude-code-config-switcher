# Proposal: add-web-service

## 概述

将 ccc 转换为 Web 服务，提供 REST API 和 WebSocket 接口，支持远程管理 Claude Code 进程。

## 动机

当前 ccc 只能通过 CLI 使用，限制了使用场景。Web 服务可以实现：

1. **远程控制**: 从任何地方控制 Claude Code
2. **多用户支持**: 多个用户同时使用不同的 agent
3. **可视化界面**: Web UI 提供更好的体验
4. **API 集成**: 与其他系统集成
5. **中心化部署**: 为中心服务器架构做准备

## 变更内容

### 1. Web 服务器

创建 Web 服务器（`ccc serve` 命令）：

```go
// internal/web/server.go
type Server struct {
    config    *ServerConfig
    agents    *AgentManager
    api       *APIServer
    ws        *WebSocketServer
    logger    Logger
}

type ServerConfig struct {
    ListenAddr   string      // 监听地址，如 ":8080"
    StaticDir    string      // 静态文件目录
    AuthMode     AuthMode    // 认证模式
    MaxAgents    int         // 最大并发 agent 数
}

type AgentManager struct {
    agents map[string]*AgentInstance
    mu     sync.RWMutex
}

type AgentInstance struct {
    ID        string
    Process   *exec.Cmd
    SessionID string
    Provider  string
    Stdin     io.WriteCloser
    Stdout    io.ReadCloser
    Stderr    io.ReadCloser
    CreatedAt time.Time
}
```

### 2. REST API

提供 RESTful API：

```
POST   /api/v1/agents              创建新的 agent 实例
GET    /api/v1/agents              列出所有 agent
GET    /api/v1/agents/{id}         获取 agent 详情
DELETE /api/v1/agents/{id}         终止 agent
POST   /api/v1/agents/{id}/stdin    发送输入到 agent
GET    /api/v1/agents/{id}/stdout   获取 agent 输出（流式）

POST   /api/v1/sessions             创建新的会话
GET    /api/v1/sessions             列出会话
GET    /api/v1/sessions/{id}        获取会话详情
POST   /api/v1/sessions/{id}/chat   发送消息到会话

GET    /api/v1/providers            列出提供商
GET    /api/v1/providers/{id}       获取提供商详情
POST   /api/v1/sessions/{id}/validate  验证配置

GET    /api/v1/health               健康检查
GET    /api/v1/status               服务状态
```

### 3. WebSocket 支持

WebSocket 端点用于实时通信：

```
WS /api/v1/agents/{id}/ws   双向通信通道

消息格式：
{
  "type": "input|output|error|status",
  "data": "...",
  "timestamp": "..."
}
```

### 4. 配置文件

在 `ccc.json` 中添加 `web` 配置段：

```json
{
  "web": {
    "enabled": true,
    "listen_addr": ":8080",
    "static_dir": "./web/dist",
    "auth_mode": "none",
    "max_agents": 10,
    "cors_origins": ["*"]
  }
}
```

### 5. 认证和授权

支持多种认证模式：

- **none**: 无认证（开发模式）
- **token**: 简单 token 认证
- **oauth**: OAuth 2.0（未来）

## 影响范围

- **新增命令**: `ccc serve`
- **新增包**: `internal/web/`
- **新增依赖**:
  - `github.com/gin-gonic/gin` - Web 框架
  - `github.com/gorilla/websocket` - WebSocket
  - `github.com/rs/cors` - CORS

## 安全考虑

1. **API 访问控制**: 认证和授权
2. **命令注入**: 防止恶意输入
3. **资源限制**: 限制并发 agent 数量
4. **文件访问**: 沙箱机制

## 实施计划

### Phase 1: 基础 Web 服务
1. Gin 框架集成
2. 基础 API 端点
3. Agent 进程管理

### Phase 2: WebSocket 支持
1. WebSocket 连接管理
2. 实时消息转发
3. 断线重连

### Phase 3: 高级功能
1. 会话持久化
2. 多用户支持
3. 认证授权

## 风险

| 风险 | 影响 | 缓解 |
|------|------|------|
| 进程管理复杂 | 高 | 使用 claude_agent_sdk 封装 |
| 并发安全 | 高 | 充分的测试和锁保护 |
| 资源泄漏 | 中 | 严格的资源清理 |
| 安全漏洞 | 高 | 输入验证和沙箱 |

## 开放问题

1. 是否需要支持会话持久化到数据库？
2. 是否需要支持 agent 进程的超时自动清理？
3. 是否需要支持日志流式查询？

## 与其他功能的关系

此功能为以下功能提供基础：
- **网页版 UI**: 通过 API 与后端通信
- **中心服务器**: 作为客户端连接到中心服务器
- **多 agent 管理**: 统一的 agent 管理
