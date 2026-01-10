# Change: enhance-supervisor-prompt-phase1

## Why

当前 Supervisor Prompt (`internal/cli/supervisor_prompt_default.md`) 约为 48 行，审查指令较为简单，缺少：
1. 具体的检测指南和判断标准
2. 常见陷阱（只问不做、测试循环、虚假完成等）的识别方法
3. 边缘情况的处理指导
4. 事无巨细的严格检查流程

这导致 Supervisor 往往过于宽松，允许 Agent 在未达到"无可挑剔的完美状态"时就停止工作。

## What Changes

- **完全重写** `internal/cli/supervisor_prompt_default.md`，创建一个极度严格、详尽的 Supervisor Prompt
- **参考来源**：
  - Ralph 项目的检测指南（从 `docs/supervisor-mode-improvement-proposal.md` 提取）
  - 之前尝试的 v3 版本（`tmp/supervisor-prompt-v3.md`）作为参考
  - AI-First 设计理念：Prompt 工程优于代码工程
- **核心框架**：
  1. 理解用户需求
  2. 检查实际工作（vs 只问不做）
  3. 检查常见陷阱
  4. 评估完成质量
  5. 判断完成状态
  6. 提供反馈
- **关键检测点**：
  - 只问不做：Agent 只在提问/建议/等待，没有执行工具
  - 测试循环：连续只运行测试不做实现
  - 虚假完成：声称完成但无实质性工作
  - 缺少验证：修改后不运行测试/构建验证
  - 错误放弃：第一次失败就停止

## Impact

- **受影响的规范**: `supervisor-hooks` - 需要更新 Supervisor Prompt 的内容要求
- **受影响的代码**: 无（只更新 Prompt 文件，不修改代码逻辑）
- **受影响的文件**: `internal/cli/supervisor_prompt_default.md`
- **向后兼容**: 是（Prompt 内容变更，不影响 API 或配置格式）
