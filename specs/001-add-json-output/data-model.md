# 数据模型：JSON 输出格式支持

**功能**: 添加 JSON 输出格式支持
**日期**: 2025-01-15
**阶段**: 第 1 阶段 - 架构设计

## 核心数据结构

### 1. JSON 输出格式（单个提供商）

**场景**: 验证单个提供商时的 JSON 输出

```json
{
  "valid": true,
  "provider": "glm"
}
```

**成功验证示例**:
```json
{
  "valid": true,
  "provider": "glm",
  "base_url": "https://open.bigmodel.cn/api/anthropic",
  "model": "glm-4.7",
  "api_status": "ok"
}
```

**验证失败示例**:
```json
{
  "valid": false,
  "provider": "glm",
  "error": "Missing required environment variable: ANTHROPIC_AUTH_TOKEN"
}
```

**API 连接失败示例**:
```json
{
  "valid": true,
  "provider": "glm",
  "base_url": "https://open.bigmodel.cn/api/anthropic",
  "model": "glm-4.7",
  "api_status": "failed: HTTP 401"
}
```

### 2. JSON 输出格式（所有提供商）

**场景**: 使用 `--all` 参数验证所有提供商时的 JSON 输出

```json
[
  {
    "valid": true,
    "provider": "glm",
    "base_url": "https://open.bigmodel.cn/api/anthropic",
    "model": "glm-4.7",
    "api_status": "ok"
  },
  {
    "valid": false,
    "provider": "kimi",
    "error": "Missing required environment variable: ANTHROPIC_AUTH_TOKEN"
  }
]
```

### 3. JSON 错误输出格式

**场景**: 配置文件不存在或格式错误时的 JSON 错误输出

```json
{
  "error": "config file not found: ~/.claude/ccc.json"
}
```

**配置文件解析错误示例**:
```json
{
  "error": "invalid JSON format in config file"
}
```

## 配置错误输出

**场景**: 无提供商配置时的 JSON 输出

```json
{
  "error": "no providers configured"
}
```

**无当前提供商设置时的 JSON 输出**:
```json
{
  "error": "no current provider set",
  "available_providers": ["glm", "kimi", "m2"]
}
```

## 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| valid | boolean | 是 | 验证是否通过 |
| provider | string | 是 | 提供商名称 |
| base_url | string | 否 | API 基础 URL（验证成功且存在时） |
| model | string | 否 | 配置的模型（验证成功且存在时） |
| api_status | string | 否 | API 连接状态（"ok" 或错误描述） |
| error | string | 否 | 错误描述（验证失败时） |
| available_providers | array | 否 | 可用提供商列表（无当前提供商时） |

## Go 结构体定义

基于现有的 `ValidationResult` 结构体，添加 JSON 序列化支持：

```go
// ValidationResult 现有结构体（已存在）
type ValidationResult struct {
    Provider  string
    Valid     bool
    Warnings  []string
    Errors    []string
    BaseURL   string
    Model     string
    APIStatus string
    APIError  error
}

// JSONValidationResult 用于 JSON 输出的结构体
type JSONValidationResult struct {
    Valid     bool   `json:"valid"`
    Provider  string `json:"provider"`
    BaseURL   string `json:"base_url,omitempty"`
    Model     string `json:"model,omitempty"`
    APIStatus string `json:"api_status,omitempty"`
    Error     string `json:"error,omitempty"`
}

// JSONError 用于错误输出的结构体
type JSONError struct {
    Error              string   `json:"error"`
    AvailableProviders []string `json:"available_providers,omitempty"`
}

// JSONValidationArray 用于 --all 输出的数组格式
type JSONValidationArray []JSONValidationResult
```

## 数据流

```
用户执行命令
    ↓
CLI 参数解析 (cli.go)
    ↓
RunOptions { JSONOutput: true, JSONPretty: false }
    ↓
validate.Run(cfg, opts)
    ↓
ValidateProvider() / ValidateAllProviders()
    ↓
PrintResultJSON() / PrintSummaryJSON()
    ↓
json.Marshal() / json.MarshalIndent()
    ↓
JSON 输出到 stdout
```

## 错误处理

所有 JSON 输出都应遵循以下规则：

1. **成功响应**: 使用 `200` 状态码概念（JSON 对象包含 `valid: true`）
2. **验证失败**: 返回 JSON 对象，`valid: false`，包含 `error` 字段
3. **系统错误**: 返回 JSON 对象，包含顶级 `error` 字段
4. **输出流**: 成功输出到 stdout，错误输出到 stderr

## 向后兼容性

- 默认行为（无 `--json` 参数）保持不变
- 现有的 `PrintResult()` 和 `PrintSummary()` 函数不变
- 新增的 JSON 输出函数不影响现有代码路径
