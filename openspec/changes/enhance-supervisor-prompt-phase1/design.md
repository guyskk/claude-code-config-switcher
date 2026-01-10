# Design Document

## Context

### 问题背景

Supervisor Mode 是 ccc 项目的核心功能之一，通过 Stop hook 在每次 Agent 停止时自动进行审查。然而，当前的 Supervisor Prompt 过于简单，导致：

1. **审查过于宽松**：Agent 往往在未达到完美状态时就被允许停止
2. **缺少具体指导**：Supervisor 没有明确的检测标准和判断流程
3. **常见陷阱未覆盖**：如"只问不做"、"测试循环"等模式未被识别
4. **反馈质量不高**：feedback 往往不够具体，Agent 需要再次询问

### 相关约束

- **不修改代码逻辑**：只更新 Prompt 文件
- **字段名保持 `allow_stop`**：与 `hook.go` 中的结构体一致
- **极度严格**：用户要求"无可挑剔的完美状态"
- **详尽优先**：宁可详细也不要省略

## Goals / Non-Goals

### Goals
- 创建一个极度严格、详尽的 Supervisor Prompt
- 覆盖常见陷阱和边缘情况
- 提供明确的判断标准和反馈模板
- 确保 Agent 必须完成实际工作且达到无可挑剔的状态

### Non-Goals
- 不修改代码逻辑（不涉及 Go 代码变更）
- 不修改字段名（保持 `allow_stop`）
- 不添加客观信息收集（git status 等，这是 Phase 2）
- 不实现实时监控（这是 Phase 3）

## Decisions

### 决策 1：采用六步审查框架

**选择**：采用明确的六步审查流程

**原因**：
- 提供清晰的执行路径
- 每个步骤有明确的检查点
- 便于理解和维护

**框架内容**：
1. 理解用户需求
2. 检查实际工作（vs 只问不做）
3. 检查常见陷阱
4. 评估完成质量
5. 判断完成状态
6. 提供反馈

### 决策 2：使用场景驱动的示例

**选择**：包含大量具体的判断示例

**原因**：
- Supervisor AI 通过示例学习效果更好
- 覆盖常见和边缘情况
- 提供明确的判断参考

**示例类型**：
- 只问不做（必须拒绝）
- 缺少验证（必须拒绝）
- 测试失败（必须拒绝）
- 代码未完成（必须拒绝）
- 错误处理不当（必须拒绝）
- 任务完成（可以接受）
- 可接受的失败（可以接受）

### 决策 3：Feedback 使用模板化格式

**选择**：定义明确的 feedback 模板

**原因**：
- 确保 feedback 具体可执行
- 避免模糊的指令
- 让 Agent 无需再问问题

**模板格式**：`[动词] [具体对象] [具体方式] [验证要求]`

**示例**：
- "运行 `go test ./...` 验证所有测试通过"
- "在 `auth.go` 中添加 `ValidateToken` 函数"

### 决策 4：保持极度严格的标准

**选择**：只允许"无可挑剔的完美状态"才能停止

**原因**：
- 用户明确要求极度严格
- 宁可多迭代一次也不要过早停止
- 符合"事无巨细"的要求

**`allow_stop = true` 的条件**：
1. 完成了实际工作（不是只问/计划）
2. 工作质量达标（无明显 bug/TODO）
3. 用户需求全部满足
4. 如果需要测试，测试已运行
5. 结果可以直接交付，不需要用户补做

## Risks / Trade-offs

### Risk 1: Prompt 过长可能导致 AI 注意力分散

**缓解措施**：
- 使用清晰的标题和结构
- 关键信息放在前面
- 使用格式化（加粗、列表）突出重点

### Risk 2: 极度严格可能导致无限循环

**缓解措施**：
- 保留现有的迭代次数限制（代码层面）
- "可接受的失败"场景：已尝试 ≥3 次，且错误是外部问题

### Risk 3: 新 Prompt 可能有未覆盖的边缘情况

**缓解措施**：
- Phase 1 后进行充分测试
- 根据实际使用情况迭代优化
- 用户可以自定义 `~/.claude/SUPERVISOR.md`

## Migration Plan

### 迁移步骤

1. **更新 Prompt 文件**：替换 `internal/cli/supervisor_prompt_default.md` 的内容
2. **重新构建**：运行 `go build` 重新编译 ccc
3. **测试验证**：使用 Supervisor Mode 执行一些任务，观察效果
4. **回滚方案**：如有问题，可以恢复旧版本 Prompt 或使用自定义 `~/.claude/SUPERVISOR.md`

### 兼容性

- **向后兼容**：Prompt 内容变更不影响 API 或配置格式
- **用户覆盖**：用户可以通过 `~/.claude/SUPERVISOR.md` 自定义 Prompt

## Open Questions

- [ ] Prompt 的最终长度目标是多少？（v3 版本约 390 行）
- [ ] 是否需要支持多语言 Prompt？（当前只有中文）
- [ ] 如何验证新 Prompt 的效果？（需要实际测试）

## References

- Ralph 项目: https://github.com/anthropics/ralph-claude-code
- 改进提案: `docs/supervisor-mode-improvement-proposal.md`
- 当前 Prompt: `internal/cli/supervisor_prompt_default.md`
- 参考版本: `tmp/supervisor-prompt-v3.md`
