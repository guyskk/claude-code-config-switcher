# 技术研究：JSON 输出格式支持

**功能**: 添加 JSON 输出格式支持
**日期**: 2025-01-15
**阶段**: 第 0 阶段 - 技术研究

## 研究目标

调研在 ccc 项目中添加 JSON 输出格式支持的技术选项和最佳实践。

## 研究内容

### 1. Go 标准库 encoding/json 用法

**决策**: 使用 Go 标准库 `encoding/json`

**理由**:
- Go 标准库已包含完整的 JSON 编码/解码支持
- 无需引入外部依赖，符合单二进制分发原则
- 性能优秀，编码/解码速度快
- 稳定可靠，经过广泛测试

**关键 API**:
- `json.Marshal(v interface{}) ([]byte, error)` - 序列化为 JSON（压缩格式）
- `json.MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)` - 序列化为带缩进的 JSON
- `json.NewEncoder(io.Writer).Encode(v interface{})` - 直接写入输出流

**示例代码**:
```go
// 压缩格式
data, err := json.Marshal(result)
if err != nil {
    return err
}
fmt.Println(string(data))

// 格式化输出
data, err := json.MarshalIndent(result, "", "  ")
if err != nil {
    return err
}
fmt.Println(string(data))
```

### 2. JSON 美化输出实现

**决策**: 使用 `json.MarshalIndent` 实现美化输出

**理由**:
- 标准库直接支持，无需额外实现
- 格式稳定，可预测
- 性能开销可接受（<1ms）

**参数选择**:
- `prefix`: "" (无前缀)
- `indent`: "  " (两个空格缩进)

### 3. 与现有 validate 包的集成方式

**当前架构分析**:
- `ValidateProvider()` 函数返回 `*ValidationResult` 结构体
- `ValidateAllProviders()` 函数返回 `*ValidationSummary` 结构体
- `PrintResult()` 和 `PrintSummary()` 函数负责人类可读输出
- `Run()` 函数协调验证流程

**集成方案**:
1. 在 `RunOptions` 中添加 `JSONOutput` 和 `JSONPretty` 字段
2. 创建 `PrintResultJSON()` 和 `PrintSummaryJSON()` 函数
3. 在 `Run()` 函数中根据标志选择输出格式

**代码结构**:
```go
// RunOptions 添加字段
type RunOptions struct {
    Provider    string
    ValidateAll bool
    JSONOutput  bool   // 新增：JSON 输出标志
    JSONPretty  bool   // 新增：美化 JSON 标志
}

// 新增 JSON 输出函数
func PrintResultJSON(result *ValidationResult, pretty bool) error
func PrintSummaryJSON(summary *ValidationSummary, pretty bool) error
```

### 4. 测试策略

**单元测试**:
- 测试 JSON 序列化正确性
- 测试美化输出格式
- 测试错误情况下的 JSON 输出

**集成测试**:
- 测试 CLI 参数解析
- 测试完整的验证流程（JSON 输出）
- 测试向后兼容性（默认文本输出）

**测试工具**:
- `go test` - 单元测试
- `go test -race` - 竞态检测
- `CCC_CONFIG_DIR` 环境变量隔离测试环境

## 技术决策总结

| 决策项 | 选择 | 理由 |
|--------|------|------|
| JSON 库 | Go 标准库 encoding/json | 无外部依赖，性能优秀 |
| 美化输出 | json.MarshalIndent | 标准库支持，格式稳定 |
| 集成方式 | 扩展 RunOptions + 新增 JSON 输出函数 | 最小化修改，保持向后兼容 |
| 测试策略 | 单元测试 + 集成测试 | 确保功能正确性和兼容性 |

## 被拒绝的替代方案

### 方案 1: 使用第三方 JSON 库（如 jsoniter）

**拒绝理由**:
- 引入外部依赖违反单二进制分发原则
- 标准库性能已足够好
- 增加维护复杂度

### 方案 2: 手动构建 JSON 字符串

**拒绝理由**:
- 容易出错，转义字符处理复杂
- 不符合 Go 语言习惯
- 标准库已提供完善支持

## 结论

所有技术研究已完成，无未解决的技术问题。可以进入第 1 阶段架构设计。
