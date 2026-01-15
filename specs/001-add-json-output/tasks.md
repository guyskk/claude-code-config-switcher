# 任务清单：添加 JSON 输出格式支持

**输入**: 来自 `/specs/001-add-json-output/` 的设计文档
**前置条件**: plan.md（已完成）、spec.md（用户故事已定义）、research.md、data-model.md、contracts/

**测试**: 本功能不包含测试任务（规格中未明确要求 TDD）

**组织方式**: 任务按用户故事分组，以便独立实现和测试每个故事。

## 格式：`[ID] [P?] [故事] 描述`

- **[P]**: 可并行运行（不同文件，无依赖）
- **[故事]**: 此任务属于哪个用户故事（例如：US1、US2、US3）
- 在描述中包含确切的文件路径

## 路径约定

本功能修改现有 Go 项目，主要修改文件位于 `internal/cli/` 和 `internal/validate/`。

---

## 第 1 阶段：基础设施（共享）

**目的**: 项目初始化和基本结构

本功能使用现有项目结构，无需额外的初始化工作。所有基础设施已就绪。

---

## 第 2 阶段：基础（阻塞前置条件）

**目的**: 在任何用户故事实现前必须完成的核心基础设施

**⚠️ 关键**: 在此阶段完成前不能开始任何用户故事工作

- [ ] T001 在 `internal/validate/validate.go` 的 `RunOptions` 结构体中添加 `JSONOutput bool` 字段
- [ ] T002 在 `internal/validate/validate.go` 的 `RunOptions` 结构体中添加 `JSONPretty bool` 字段

**检查点**: 基础设施就绪 - 可以开始用户故事实现

---

## 第 3 阶段：用户故事 1 - 以 JSON 格式获取验证结果 (优先级: P1) 🎯 MVP

**目标**: 实现单个提供商验证的 JSON 输出格式

**独立测试**: 运行 `ccc validate --json` 验证输出为有效 JSON 格式，包含 `valid` 和 `provider` 字段

### 用户故事 1 的实现

- [ ] T003 [P] [US1] 在 `internal/validate/validate.go` 中创建 `JSONValidationResult` 结构体，包含 JSON 序列化所需的字段和标签
- [ ] T004 [P] [US1] 在 `internal/validate/validate.go` 中创建 `JSONError` 结构体，用于错误输出的 JSON 序列化
- [ ] T005 [US1] 在 `internal/validate/validate.go` 中实现 `printResultJSON(result *ValidationResult, pretty bool) error` 函数，输出单个验证结果的 JSON
- [ ] T006 [US1] 在 `internal/validate/validate.go` 中实现 `printErrorJSON(err error) error` 函数，输出错误信息的 JSON
- [ ] T007 [US1] 在 `internal/validate/validate.go` 的 `Run()` 函数中添加 JSON 输出分支，检查 `opts.JSONOutput` 标志并调用相应的 JSON 输出函数
- [ ] T008 [US1] 在 `internal/cli/cli.go` 的 `ValidateCommand` 结构体中添加 `JSONOutput bool` 字段
- [ ] T009 [US1] 在 `internal/cli/cli.go` 的 `parseValidateArgs()` 函数中添加 `--json` 参数解析逻辑
- [ ] T010 [US1] 在 `internal/cli/cli.go` 的 `runValidate()` 函数中将 `JSONOutput` 字段传递给 `validate.RunOptions`
- [ ] T011 [US1] 验证向后兼容性：运行 `ccc validate`（无 `--json`）确认输出仍为人类可读格式

**检查点**: 此时用户故事 1 应完全功能化且可独立测试

---

## 第 4 阶段：用户故事 2 - 验证所有提供商 (优先级: P2)

**目标**: 实现所有提供商验证的 JSON 数组输出格式

**独立测试**: 运行 `ccc validate --all --json` 验证输出为 JSON 数组，每个元素包含一个提供商的验证结果

### 用户故事 2 的实现

- [ ] T012 [US2] 在 `internal/validate/validate.go` 中实现 `printSummaryJSON(summary *ValidationSummary, pretty bool) error` 函数，输出所有提供商验证结果的 JSON 数组
- [ ] T013 [US2] 在 `internal/validate/validate.go` 的 `Run()` 函数中扩展 JSON 输出分支，处理 `opts.ValidateAll && opts.JSONOutput` 组合情况
- [ ] T014 [US2] 验证 `--all --json` 组合：运行 `ccc validate --all --json` 确认输出为 JSON 数组格式

**检查点**: 此时用户故事 1 和 2 都应独立工作

---

## 第 5 阶段：用户故事 3 - 美化 JSON 输出 (优先级: P3)

**目标**: 实现格式化的 JSON 输出功能

**独立测试**: 运行 `ccc validate --json --pretty` 验证输出为格式化的、易读的 JSON

### 用户故事 3 的实现

- [ ] T015 [P] [US3] 在 `internal/validate/validate.go` 中的 `printResultJSON()` 函数中实现美化输出逻辑（使用 `json.MarshalIndent`）
- [ ] T016 [P] [US3] 在 `internal/validate/validate.go` 中的 `printSummaryJSON()` 函数中实现美化输出逻辑（使用 `json.MarshalIndent`）
- [ ] T017 [US3] 在 `internal/validate/validate.go` 中的 `printErrorJSON()` 函数中实现美化输出逻辑（使用 `json.MarshalIndent`）
- [ ] T018 [US3] 在 `internal/cli/cli.go` 的 `ValidateCommand` 结构体中添加 `JSONPretty bool` 字段
- [ ] T019 [US3] 在 `internal/cli/cli.go` 的 `parseValidateArgs()` 函数中添加 `--pretty` 参数解析逻辑
- [ ] T020 [US3] 在 `internal/cli/cli.go` 的 `runValidate()` 函数中将 `JSONPretty` 字段传递给 `validate.RunOptions`
- [ ] T021 [US3] 验证美化输出：运行 `ccc validate --json --pretty` 确认输出为格式化的 JSON（两个空格缩进）

**检查点**: 所有用户故事现在都应独立功能化

---

## 第 6 阶段：完善与横切关注点

**目的**: 影响多个用户故事的改进

- [ ] T022 运行 `./check.sh --lint` 验证代码符合 gofmt 和 go vet 标准
- [ ] T023 运行 `./check.sh --build` 验证在所有平台上构建成功
- [ ] T024 运行 `quickstart.md` 中的关键验证场景
- [ ] T025 更新 `README.md` 中的 validate 命令用法说明，添加 `--json` 和 `--pretty` 参数描述

---

## 依赖关系与执行顺序

### 阶段依赖

- **基础（第 2 阶段）**: 无依赖 - 可立即开始
- **用户故事 1 (第 3 阶段)**: 依赖基础阶段完成（T001、T002）
- **用户故事 2 (第 4 阶段)**: 依赖基础阶段完成和用户故事 1 的 JSON 输出函数
- **用户故事 3 (第 5 阶段)**: 依赖基础阶段完成和用户故事 1、2 的 JSON 输出函数
- **完善（第 6 阶段）**: 依赖所有用户故事完成

### 用户故事依赖

- **用户故事 1 (P1)**: 基础完成后可开始 - 无其他故事依赖
- **用户故事 2 (P2)**: 基础完成后可开始 - 复用 US1 的 `JSONValidationResult` 结构体
- **用户故事 3 (P3)**: 基础完成后可开始 - 扩展 US1 和 US2 的 JSON 输出函数

### 每个用户故事内

- 结构体定义在输出函数之前
- 输出函数在 CLI 集成之前
- CLI 参数解析在传递给 validate 包之前
- 验证向后兼容性在完成功能实现后

### 并行机会

- **用户故事 1**: T003 和 T004 可并行（不同结构体）
- **用户故事 3**: T015、T016、T017 可并行（不同函数，不同文件内）

---

## 并行示例：用户故事 1

```bash
# 一起启动用户故事 1 的结构体定义：
Task: "T003 [P] [US1] 在 internal/validate/validate.go 中创建 JSONValidationResult 结构体"
Task: "T004 [P] [US1] 在 internal/validate/validate.go 中创建 JSONError 结构体"
```

---

## 实施策略

### MVP 优先（仅用户故事 1）

1. 完成第 2 阶段：基础（T001、T002）
2. 完成第 3 阶段：用户故事 1（T003-T011）
3. **停止并验证**: 运行 `ccc validate --json` 验证 JSON 输出
4. 如准备就绪则部署/演示

### 增量交付

1. 完成基础 → 基础就绪
2. 添加用户故事 1 → 验证 JSON 输出 → 部署/演示（MVP！）
3. 添加用户故事 2 → 验证 `--all --json` → 部署/演示
4. 添加用户故事 3 → 验证 `--json --pretty` → 部署/演示
5. 完成完善阶段 → 最终部署

### 并行团队策略

有多个开发者时：

1. 开发者完成基础阶段（T001、T002）
2. 基础完成后：
   - 开发者 A：用户故事 1（JSON 输出）
   - 开发者 B：用户故事 2（数组输出）
3. 用户故事 1、2 完成后：
   - 开发者 A 或 B：用户故事 3（美化输出）

---

## 注意事项

- [P] 任务 = 不同结构体/函数，无依赖
- [故事] 标签将任务映射到特定用户故事以实现可追溯性
- 每个用户故事应可独立完成和测试
- 每个任务或逻辑组后提交
- 在任何检查点停止以独立验证故事
- 避免：模糊任务、同文件冲突、破坏独立性的跨故事依赖

## 特别说明：使用中文

**本文档必须使用中文编写。**

1. 所有任务描述必须使用中文。
2. 文件路径使用英文，但说明文字使用中文。
3. 变量名、函数名等标识符使用英文。
