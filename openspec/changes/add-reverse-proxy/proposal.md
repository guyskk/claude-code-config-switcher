# Proposal: add-reverse-proxy

## 概述

实现动态反向代理功能，允许通过配置文件动态切换使用的模型提供商，无需修改 Claude Code 配置。

## 动机

当前 ccc 通过修改 settings.json 来切换提供商，存在以下问题：
1. 每次切换都需要重新生成配置文件
2. 不能动态切换（需要重启 Claude）
3. 无法同时为多个会话使用不同的提供商

通过在本地启动反向代理，可以：
1. 动态路由请求到不同的提供商
2. 支持 API Key 轮换和负载均衡
3. 统一多个提供商的访问
4. 为 Web 服务奠定基础

## 变更内容

### 1. 反向代理服务器

创建独立的反向代理服务器（`ccc proxy` 命令）：

```go
// internal/proxy/server.go
type ProxyServer struct {
    config    *ProxyConfig
    providers map[string]*ProviderConfig
    router    *Router
    logger    Logger
}

type ProxyConfig struct {
    ListenAddr   string            // 监听地址，如 ":8080"
    DefaultProvider string          // 默认提供商
    Providers     map[string]*ProviderConfig
    Routing      *RoutingConfig    // 路由规则
    RateLimit    *RateLimitConfig  // 速率限制
}

type ProviderConfig struct {
    Name      string
    BaseURL   string
    APIKeys   []string          // 支持多个 Key，轮换使用
    Timeout   time.Duration
    MaxRetries int
}

type RoutingConfig struct {
    // 按模型路由
    ByModel map[string]string  // "claude-3-5-sonnet-20241022": "kimi"

    // 按用户路由
    ByUser map[string]string    // "user-123": "glm"

    // 按路径路由
    ByPath map[string]string    // "/v1/messages/*": "m2"
}
```

### 2. 配置文件格式

在 `ccc.json` 中添加 `proxy` 配置段：

```json
{
  "proxy": {
    "enabled": true,
    "listen_addr": ":8080",
    "default_provider": "kimi",
    "providers": {
      "kimi": {
        "base_url": "https://api.moonshot.cn/anthropic",
        "api_keys": ["key1", "key2"],
        "timeout": "60s",
        "max_retries": 3
      },
      "glm": {
        "base_url": "https://open.bigmodel.cn/api/anthropic",
        "api_keys": ["key3"],
        "timeout": "60s"
      }
    },
    "routing": {
      "by_model": {
        "claude-3-5-sonnet-20241022": "kimi"
      }
    }
  }
}
```

### 3. 环境变量支持

通过环境变量动态覆盖：

- `CCC_PROXY_ENABLED`: 启用代理
- `CCC_PROXY_LISTEN_ADDR`: 监听地址
- `CCC_PROXY_DEFAULT_PROVIDER`: 默认提供商
- `CCC_PROXY_<PROVIDER>_API_KEY`: 提供商 API Key

### 4. API 兼容性

代理完全兼容 Anthropic API 规范：
- `/v1/messages` - 消息端点
- `/v1/models` - 模型列表
- 流式响应支持
- 错误响应兼容

## 影响范围

- **新增命令**: `ccc proxy`
- **新增包**: `internal/proxy/`
- **配置变更**: `ccc.json` 添加 `proxy` 段
- **新增依赖**:
  - `github.com/gorilla/mux` - HTTP 路由
  - `github.com/rs/cors` - CORS 支持

## 实施计划

### Phase 1: 基础代理
1. 实现基础 HTTP 代理服务器
2. 支持单个提供商
3. 基本请求转发

### Phase 2: 多提供商支持
1. 多提供商配置
2. 动态路由
3. API Key 轮换

### Phase 3: 高级功能
1. 速率限制
2. 请求缓存
3. 指标收集

## 风险

| 风险 | 影响 | 缓解 |
|------|------|------|
| API 兼容性问题 | 中 | 充分测试各提供商 API |
| 性能开销 | 低 | 使用 HTTP/2 连接复用 |
| 安全性 | 中 | 支持 TLS 和认证 |

## 开放问题

1. 是否需要支持 WebSocket？（流式响应当前使用 SSE）
2. 是否需要持久化请求日志？
3. 是否需要支持请求重放？
