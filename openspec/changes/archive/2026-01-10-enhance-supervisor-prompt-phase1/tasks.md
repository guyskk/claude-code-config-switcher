# Tasks

## 1. 设计并实现增强的 Supervisor Prompt

- [x] 1.1 分析 Ralph 项目的检测指南，提取核心检测模式
- [x] 1.2 分析 v3 版本的 Prompt，识别有效的结构和模式
- [x] 1.3 设计完整的审查框架（六步法）
- [x] 1.4 编写"理解用户需求"部分
- [x] 1.5 编写"检查实际工作"部分（只问不做检测）
- [x] 1.6 编写"检查常见陷阱"部分（测试循环、虚假完成等）
- [x] 1.7 编写"评估完成质量"部分（代码质量、任务完整性）
- [x] 1.8 编写"判断完成状态"部分（allow_stop 的具体条件）
- [x] 1.9 编写"提供反馈"部分（feedback 的质量要求和模板）
- [x] 1.10 添加常见场景的判断示例（至少 10 个）
- [x] 1.11 添加快速检查清单
- [x] 1.12 整合所有部分，形成完整的 Prompt 文件

## 2. 验证和测试

- [x] 2.1 使用 `openspec validate enhance-supervisor-prompt-phase1 --strict` 验证 proposal
- [x] 2.2 手动审查 Prompt 内容的完整性和准确性
- [x] 2.3 确认字段名使用 `allow_stop`（而非 `completed`）
- [x] 2.4 确认 Prompt 长度适中（详尽但不过长）

## 3. 文档和归档

- [x] 3.1 更新 proposal（如有需要）
- [x] 3.2 等待用户审批
- [x] 3.3 使用 `openspec:apply` 执行实现
- [x] 3.4 提交 PR 并 review
- [x] 3.5 合并后使用 `openspec:archive` 归档
