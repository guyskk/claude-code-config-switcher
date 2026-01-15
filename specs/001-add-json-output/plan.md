# 实现方案：添加 JSON 输出格式支持

**分支**: `001-add-json-output` | **日期**: 2025-01-15 | **规格**: [spec.md](./spec.md)
**输入**: 来自 `/specs/001-add-json-output/spec.md` 的功能规格

**注意**: 本模板由 `/speckit.plan` 命令填充。详见 `.specify/templates/commands/plan.md` 中的执行工作流。

## 特别说明：使用中文

**本文档必须使用中文编写。**

1. 所有技术描述、架构决策、实现细节必须使用中文。
2. 代码示例中的注释必须使用中文。
3. 变量名、函数名等标识符使用英文，但说明文字使用中文。

## 摘要

本功能为 `ccc validate` 命令添加 JSON 输出格式支持。主要需求包括：
- 添加 `--json` 参数输出 JSON 格式验证结果
- 添加 `--all` 参数支持验证所有提供商（已存在，需扩展 JSON 支持）
- 添加 `--pretty` 参数支持格式化 JSON 输出
- 保持向后兼容，默认输出人类可读文本格式

技术方案摘要：使用 Go 标准库 `encoding/json` 实现 JSON 序列化，无需引入外部依赖。在 `internal/validate` 包中添加 JSON 输出方法，在 `internal/cli` 包中扩展命令行参数解析。

## 技术上下文

**语言/版本**: Go 1.21
**主要依赖**: Go 标准库（encoding/json、fmt）
**存储**: 不适用（仅输出格式变更）
**测试**: go test with -race detector
**目标平台**: darwin-amd64、darwin-arm64、linux-amd64、linux-arm64
**项目类型**: single（单一 Go 可执行文件）
**性能目标**: JSON 序列化开销 <1ms，不影响现有性能
**约束**: 单二进制、静态链接、跨平台
**规模/范围**: 预计修改 2 个文件，新增约 100 行代码

## 宪章检查

*门禁：必须在第 0 阶段研究前通过。第 1 阶段设计后再次检查。*

### ccc 项目宪章合规检查

- [x] **原则一：单二进制分发** - 最终产物是单一静态链接二进制文件
  - 使用 Go 标准库 `encoding/json`，无外部依赖
- [x] **原则二：代码质量标准** - 符合 gofmt、go vet 要求
  - 新代码将遵循现有代码风格
- [x] **原则三：测试规范** - 包含单元测试和竞态检测
  - 将为新增的 JSON 输出功能添加测试
- [x] **原则四：向后兼容** - 配置格式变更保持兼容
  - 仅添加新的命令行参数，不修改配置文件格式
- [x] **原则五：跨平台支持** - 支持 darwin/linux, amd64/arm64
  - 使用平台无关的标准库实现
- [x] **原则六：错误处理与可观测性** - 错误明确且可操作
  - JSON 错误输出将包含明确的错误描述

### 复杂度跟踪

> 无违规项，无需填写

## 项目结构

### 文档组织（本功能）

```text
specs/001-add-json-output/
├── spec.md              # 功能规格（已完成）
├── plan.md              # 本文件
├── research.md          # 第 0 阶段输出
├── data-model.md        # 第 1 阶段输出
├── quickstart.md        # 第 1 阶段输出
├── contracts/           # 第 1 阶段输出
│   └── json-output.md   # JSON 输出格式契约
└── tasks.md             # 第 2 阶段输出（由 /speckit.tasks 生成）
```

### 源代码组织（仓库根目录）

```text
cmd/test-sdk/           # 测试 SDK（不修改）
main.go                 # 主入口（不修改）

internal/               # 私有应用代码
├── cli/                # CLI 命令处理
│   ├── cli.go          # [修改] 添加 --json 和 --pretty 参数
│   └── cli_test.go     # [修改] 添加参数解析测试
└── validate/           # 验证逻辑
    ├── validate.go     # [修改] 添加 JSON 输出方法
    └── validate_test.go # [修改] 添加 JSON 输出测试
```

**结构决策**: 使用现有的单一 Go 项目结构。修改 `internal/cli/cli.go` 添加命令行参数，修改 `internal/validate/validate.go` 添加 JSON 输出功能。

## 实现阶段

### 第 -1 阶段：预实现门禁

> **重要：在开始任何实现工作前必须通过此阶段**

#### 宪章合规门禁

- [x] 所有 6 条核心原则已检查
- [x] 如有违规，已在"复杂度跟踪"表中记录理由

#### 技术决策门禁

- [x] 技术栈已确定（Go 1.21 + 标准库）
- [x] 项目结构已定义（修改现有文件）
- [x] 数据模型已设计（参见 data-model.md）
- [x] API 契约已定义（参见 contracts/json-output.md）

---

### 第 0 阶段：技术研究

**目标**: 调研技术选项，收集实现所需信息

**输出**: `research.md`

**研究内容**:

- [x] Go 标准库 encoding/json 用法
- [x] JSON 美化输出（json.MarshalIndent）
- [x] 与现有 validate 包的集成方式
- [x] 测试策略（单元测试和集成测试）

---

### 第 1 阶段：架构设计

**目标**: 定义数据模型、API 契约和实现细节

**输出**: `data-model.md`、`contracts/`、`quickstart.md`

**数据模型** (data-model.md):
- [x] 定义 JSON 输出数据结构
- [x] 定义错误类型

**API 契约** (contracts/):
- [x] CLI 命令接口
- [x] JSON 输出格式规范

**快速入门** (quickstart.md):
- [x] 关键验证场景
- [x] 测试检查清单

---

### 第 2 阶段：任务分解

**目标**: 将设计转化为可执行的任务列表

**输出**: `tasks.md` (由 `/speckit.tasks` 命令生成)

> **注意**: 第 2 阶段不在此方案中完成，由独立的 `/speckit.tasks` 命令处理

---

## 实施文件创建顺序

> **重要：按照此顺序创建文件以确保质量**

1. **contracts/json-output.md** - 定义 JSON 输出格式契约
2. **测试文件** - 按以下顺序创建：
   - internal/validate/json_test.go - JSON 输出单元测试
   - internal/cli/json_cli_test.go - CLI 参数解析测试
3. **源代码文件** - 创建使测试通过的实现：
   - internal/validate/json.go - JSON 输出实现
   - internal/cli/cli.go - CLI 参数扩展

**理由**: 测试先行确保 API 设计可用，实现符合需求。
