# 快速入门：JSON 输出功能验证

**功能**: 添加 JSON 输出格式支持
**日期**: 2025-01-15
**目的**: 提供关键验证场景和测试检查清单

## 前置条件

1. ccc 已正确安装
2. 已配置至少一个提供商（如 glm）
3. 提供商配置有效（包含 API key）

## 关键验证场景

### 场景 1: 基本验证成功

```bash
# 默认文本输出（向后兼容）
ccc validate

# JSON 输出
ccc validate --json | jq .

# 美化 JSON 输出
ccc validate --json --pretty
```

**预期结果**:
- 默认输出人类可读格式
- JSON 输出包含 `valid: true`
- 美化输出格式正确（两个空格缩进）

### 场景 2: 验证指定提供商

```bash
# 验证 glm 提供商
ccc validate glm --json

# 验证 kimi 提供商
ccc validate kimi --json
```

**预期结果**:
- 输出指定提供商的验证结果
- `provider` 字段正确

### 场景 3: 验证所有提供商

```bash
# 验证所有提供商
ccc validate --all --json

# 验证所有提供商（美化输出）
ccc validate --all --json --pretty
```

**预期结果**:
- 输出 JSON 数组
- 每个元素包含一个提供商的验证结果

### 场景 4: 在脚本中使用

```bash
# 在 bash 脚本中使用 JSON 输出
RESULT=$(ccc validate --json)
VALID=$(echo "$RESULT" | jq -r '.valid')

if [ "$VALID" = "true" ]; then
    echo "验证成功"
else
    echo "验证失败"
    exit 1
fi
```

### 场景 5: 错误处理

```bash
# 配置错误
ccc validate invalid-provider --json

# 无提供商配置
ccc validate --json
```

**预期结果**:
- 错误情况返回有效的 JSON
- 包含 `error` 字段描述问题

## 测试检查清单

### 功能测试

- [ ] `--json` 参数正确启用 JSON 输出
- [ ] `--pretty` 参数正确启用美化输出
- [ ] 验证成功时输出正确的 JSON 格式
- [ ] 验证失败时输出包含 `error` 字段
- [ ] `--all` 参数输出 JSON 数组
- [ ] 默认输出（无 `--json`）保持人类可读格式

### 格式验证

- [ ] JSON 输出可以被 `jq` 正确解析
- [ ] 美化输出格式正确（两个空格缩进）
- [ ] 压缩输出为单行
- [ ] 所有必需字段都存在
- [ ] 字段类型正确（boolean、string）

### 向后兼容性

- [ ] 不指定 `--json` 时输出格式不变
- [ ] 退出码语义保持不变
- [ ] 现有脚本继续正常工作

### 错误处理

- [ ] 配置文件不存在时返回错误 JSON
- [ ] 提供商不存在时返回错误 JSON
- [ ] 无当前提供商时返回错误 JSON
- [ ] 所有错误消息清晰明确

### 跨平台测试

- [ ] Linux (amd64) 验证通过
- [ ] Linux (arm64) 验证通过
- [ ] macOS (amd64) 验证通过
- [ ] macOS (arm64) 验证通过

## 常见问题排查

### 问题 1: JSON 解析失败

**症状**: `jq` 报错 "parse error: Invalid literal"

**排查**:
```bash
# 检查原始输出
ccc validate --json | cat -A

# 检查是否有 ANSI 颜色码
# 应该没有颜色码（JSON 输出禁用颜色）
```

### 问题 2: 退出码不正确

**症状**: 脚本中退出码判断不正确

**排查**:
```bash
# 检查退出码
ccc validate --json
echo $?

# 应该是 0（成功）或 1（失败）
```

### 问题 3: 美化输出格式不对

**症状**: 缩进不是两个空格

**检查**:
```bash
# 检查缩进
ccc validate --json --pretty | cat -A
# 应该看到两个空格（^I 不是一个 tab）
```

## CI/CD 集成示例

### GitHub Actions

```yaml
- name: Validate ccc configuration
  run: |
    RESULT=$(ccc validate --json)
    VALID=$(echo "$RESULT" | jq -r '.valid')
    if [ "$VALID" != "true" ]; then
      echo "Validation failed:"
      echo "$RESULT" | jq '.'
      exit 1
    fi
```

### GitLab CI

```yaml
validate:ccc:
  script:
    - ccc validate --json | jq -e '.valid == true'
```

## 性能基准

预期性能指标：

| 操作 | 预期时间 |
|------|----------|
| JSON 序列化 | <1ms |
| 完整验证流程 | <5s (含 API 调用) |
| 内存开销 | <1MB |

验证命令：
```bash
# 测量 JSON 序列化性能
time ccc validate --json > /dev/null
```
