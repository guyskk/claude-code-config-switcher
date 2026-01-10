# supervisor-hooks Specification Delta

## MODIFIED Requirements

### Requirement: Supervisor Claude 调用

系统 SHALL 使用指定参数调用 Supervisor claude。

#### Scenario: Supervisor 命令构建
- **GIVEN** session_id 为 "abc123"
- **AND** 增强的 Supervisor Prompt 存在于 `internal/cli/supervisor_prompt_default.md`
- **WHEN** 构建 Supervisor 命令
- **THEN** 命令应当包含：
  - `claude`
  - `--print`
  - `--resume abc123`
  - `--verbose`
  - `--output-format stream-json`
  - `--json-schema` （包含 allow_stop 和 feedback 字段）
  - `--system-prompt` （增强的 Supervisor Prompt 内容，约 400-500 行）
- **AND** Prompt SHALL 包含以下核心部分：
  - 角色定义：严格的任务审查者
  - 六步审查框架：理解需求 → 检查工作 → 检查陷阱 → 评估质量 → 判断状态 → 提供反馈
  - 常见陷阱检测：只问不做、测试循环、虚假完成、缺少验证、错误放弃
  - 判断标准：allow_stop=true 的 5 个条件，allow_stop=false 的 7 个场景
  - Feedback 模板：具体的、可执行的、包含足够细节的指令格式
  - 场景示例：至少 10 个常见情况的判断示例
- **AND** 环境变量应当包含 `CCC_SUPERVISOR_HOOK=1`

#### Scenario: 增强的 Prompt 内容结构
- **GIVEN** 增强的 Supervisor Prompt 已加载
- **WHEN** 检查 Prompt 内容
- **THEN** SHALL 包含以下部分：
  - **第一步：理解用户需求** - 提取用户的原始需求，识别明确和隐含要求
  - **第二步：检查实际工作** - 检测 Agent 是否在执行工具而非只提问
  - **第三步：检查常见陷阱** - 识别只问不做、测试循环、虚假完成、缺少验证、错误放弃
  - **第四步：评估完成质量** - 检查代码质量、任务完整性、可交付性
  - **第五步：判断完成状态** - allow_stop=true 的 5 个必须条件，allow_stop=false 的 7 个场景
  - **第六步：提供反馈** - Feedback 的质量要求、模板格式、禁止模式
- **AND** SHALL 包含至少 10 个场景示例
- **AND** SHALL 包含快速检查清单

#### Scenario: 只问不做检测
- **GIVEN** Agent 的最后几轮对话只有提问/建议/等待
- **AND** 没有工具调用（Bash/Edit/Write）
- **WHEN** Supervisor 执行审查
- **THEN** SHALL 设置 `allow_stop: false`
- **AND** Feedback SHALL 指令 Agent 直接执行具体操作

#### Scenario: 缺少验证检测
- **GIVEN** Agent 修改了代码但没有运行测试/构建验证
- **WHEN** Supervisor 执行审查
- **THEN** SHALL 设置 `allow_stop: false`
- **AND** Feedback SHALL 指令 Agent 运行相应的验证命令

#### Scenario: 测试失败检测
- **GIVEN** Agent 运行了测试但有失败
- **AND** Agent 没有修复尝试或停止在第一次失败
- **WHEN** Supervisor 执行审查
- **THEN** SHALL 设置 `allow_stop: false`
- **AND** Feedback SHALL 指令 Agent 修复错误并重新验证

#### Scenario: 完美状态允许停止
- **GIVEN** Agent 完成了实际工作（有工具调用）
- **AND** 工作质量达标（无明显 bug/TODO）
- **AND** 用户需求全部满足
- **AND** 测试已运行且通过（如适用）
- **AND** 结果可以直接交付
- **WHEN** Supervisor 执行审查
- **THEN** SHALL 设置 `allow_stop: true`
- **AND** Feedback SHALL 为空字符串

### Requirement: 结构化输出处理

系统 SHALL 解析 Supervisor 的 stream-json 输出。

#### Scenario: 结果 JSON Schema（更新）
- **GIVEN** Supervisor 被要求返回结构化结果
- **WHEN** Supervisor 返回结果
- **THEN** 结果 SHALL 符合以下 schema：
```json
{
  "type": "object",
  "properties": {
    "allow_stop": {
      "type": "boolean",
      "description": "是否允许 Agent 停止工作（true = 工作达到无可挑剔状态，false = 需要继续改进）"
    },
    "feedback": {
      "type": "string",
      "description": "当 allow_stop 为 false 时，提供具体的、可执行的、包含足够细节的改进指令"
    }
  },
  "required": ["allow_stop", "feedback"]
}
```
