# JSON 输出格式契约

**功能**: 添加 JSON 输出格式支持
**日期**: 2025-01-15
**类型**: API 契约

## CLI 命令接口

### 基本用法

```bash
# 验证当前提供商（默认文本输出）
ccc validate

# 验证当前提供商（JSON 输出）
ccc validate --json

# 验证当前提供商（美化 JSON 输出）
ccc validate --json --pretty

# 验证指定提供商（JSON 输出）
ccc validate glm --json

# 验证所有提供商（JSON 输出）
ccc validate --all --json

# 验证所有提供商（美化 JSON 输出）
ccc validate --all --json --pretty
```

### 命令行参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--json` | flag | false | 启用 JSON 输出格式 |
| `--pretty` | flag | false | 启用美化 JSON 输出（仅在使用 --json 时有效） |
| `--all` | flag | false | 验证所有提供商（已存在） |
| `<provider>` | string | 当前提供商 | 指定要验证的提供商名称 |

## JSON 输出格式规范

### 单个提供商验证 - 成功

**请求**:
```bash
ccc validate glm --json
```

**响应** (stdout):
```json
{
  "valid": true,
  "provider": "glm",
  "base_url": "https://open.bigmodel.cn/api/anthropic",
  "model": "glm-4.7",
  "api_status": "ok"
}
```

**HTTP 状态码概念**: 退出码 0

### 单个提供商验证 - 配置错误

**请求**:
```bash
ccc validate kimi --json
```

**响应** (stdout):
```json
{
  "valid": false,
  "provider": "kimi",
  "error": "Missing required environment variable: ANTHROPIC_AUTH_TOKEN"
}
```

**HTTP 状态码概念**: 退出码 1

### 单个提供商验证 - API 失败

**请求**:
```bash
ccc validate glm --json
```

**响应** (stdout):
```json
{
  "valid": true,
  "provider": "glm",
  "base_url": "https://open.bigmodel.cn/api/anthropic",
  "model": "glm-4.7",
  "api_status": "failed: HTTP 401"
}
```

**HTTP 状态码概念**: 退出码 1

### 所有提供商验证 - 成功

**请求**:
```bash
ccc validate --all --json
```

**响应** (stdout):
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
    "valid": true,
    "provider": "kimi",
    "base_url": "https://api.moonshot.cn/anthropic",
    "model": "kimi-k2-thinking",
    "api_status": "ok"
  }
]
```

**HTTP 状态码概念**: 退出码 0

### 所有提供商验证 - 部分失败

**请求**:
```bash
ccc validate --all --json
```

**响应** (stdout):
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

**HTTP 状态码概念**: 退出码 1

### 系统错误

**请求**:
```bash
ccc validate --json
```

**响应** (stderr):
```json
{
  "error": "no current provider set",
  "available_providers": ["glm", "kimi", "m2"]
}
```

**HTTP 状态码概念**: 退出码 1

## JSON 格式规范

### 格式化规则

1. **压缩格式** (默认 `--json`):
   - 无缩进
   - 无换行
   - 单行输出

2. **美化格式** (`--json --pretty`):
   - 两个空格缩进
   - 每个字段换行
   - 易于阅读

### 字段命名

- 使用 snake_case (例: `base_url`, `api_status`)
- 与现有 Go 结构体字段 JSON tag 保持一致

### 退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 验证成功 |
| 1 | 验证失败或系统错误 |

### 输出流

- **成功输出**: stdout
- **错误输出**: stderr（仅在严重系统错误时）

## 兼容性承诺

### 向后兼容

- 不指定 `--json` 参数时，输出格式与现有版本完全一致
- 现有脚本和工具不受影响
- 退出码语义保持不变

### 版本控制

- JSON 格式变更将遵循语义版本控制
- 破坏性变更需要主版本号升级
- 新增字段为可选（向后兼容）
