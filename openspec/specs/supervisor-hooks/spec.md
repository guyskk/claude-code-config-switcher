# supervisor-hooks Specification

## Purpose

定义 Supervisor Mode 使用 Claude Code Hooks 机制的行为规范。Supervisor Mode 通过 Stop hook 在每次 Agent 停止时自动进行 Supervisor 检查，根据反馈决定是否继续工作，形成自动迭代循环直到任务完成。
## Requirements
### Requirement: Supervisor Mode 启动

当 `CCC_SUPERVISOR=1` 环境变量设置时，系统 SHALL 启动 Supervisor Mode。

#### Scenario: 环境变量启用
- **GIVEN** 环境变量 `CCC_SUPERVISOR=1` 已设置
- **WHEN** 用户执行 `ccc <provider>`
- **THEN** 应当生成带 Stop hook 的 `settings.json`
- **AND** 应当启动 claude（不带 `--settings` 参数）

#### Scenario: 环境变量未设置
- **GIVEN** 环境变量 `CCC_SUPERVISOR` 未设置或不为 "1"
- **WHEN** 用户执行 `ccc <provider>`
- **THEN** 应当使用普通模式启动 claude

### Requirement: Settings 文件生成

Supervisor Mode SHALL 生成包含 Stop hook 的单一 `settings.json` 文件。

#### Scenario: 生成带 Hook 的 Settings
- **GIVEN** Supervisor Mode 启用
- **WHEN** 系统生成配置
- **THEN** 应当将配置写入 `~/.claude/settings.json`
- **AND** settings 中应当包含 `hooks.Stop` 配置
- **AND** hook 命令应当是 ccc 的绝对路径加 `supervisor-hook`（不带参数）

#### Scenario: Hook 命令格式
- **GIVEN** ccc 安装在 `/usr/local/bin/ccc`
- **WHEN** 系统生成 hook 配置
- **THEN** hook 命令应当为 `/usr/local/bin/ccc supervisor-hook`

### Requirement: supervisor-hook 子命令
系统 SHALL 提供 `supervisor-hook` 子命令处理 Stop hook 事件。

#### Scenario: 正常 Hook 调用（使用 Claude Agent SDK）
- **GIVEN** 环境变量 `CCC_SUPERVISOR_HOOK` 未设置
- **AND** stdin 包含有效的 StopHookInput JSON
- **WHEN** 执行 `ccc supervisor-hook`
- **THEN** 应当使用 `schlunsen/claude-agent-sdk-go` 创建客户端
- **AND** 使用 fork session 模式恢复当前 session（`WithForkSession(true)` 和 `WithSessionID(id)`）
- **AND** Claude 能够访问原 session 的完整上下文
- **AND** 应当根据 Supervisor 结果输出 JSON 到 stdout

#### Scenario: 解析 Supervisor 结果
- **WHEN** SDK 返回 ResultMessage
- **THEN** 应当从 Result 字段提取并解析 JSON
- **AND** 转换为 `{allow_stop: bool, feedback: string}` 格式
- **AND** 根据结果决定是否允许停止

### Requirement: 防止死循环 - 环境变量

系统 SHALL 使用环境变量防止 Supervisor 的 hook 触发死循环。

#### Scenario: 检测到环境变量跳过执行
- **GIVEN** 环境变量 `CCC_SUPERVISOR_HOOK=1` 已设置
- **WHEN** 执行 `ccc supervisor-hook`
- **THEN** 应当输出 `{"decision":"","":""}` 到 stdout
- **AND** 应当立即返回（退出码 0）

#### Scenario: Supervisor Claude 启动时设置环境变量
- **GIVEN** hook 需要调用 Supervisor claude
- **WHEN** 构建 Supervisor claude 命令
- **THEN** 应当设置 `CCC_SUPERVISOR_HOOK=1` 环境变量
- **AND** Supervisor claude 应当继承该环境变量

#### Scenario: 完整防死循环流程
- **GIVEN** Agent claude 触发 Stop hook
- **WHEN** 第一次调用 `ccc supervisor-hook`（无 `CCC_SUPERVISOR_HOOK` 环境变量）
- **THEN** 应当启动 Supervisor claude（设置 `CCC_SUPERVISOR_HOOK=1`）
- **AND** 当 Supervisor claude 停止时触发 hook
- **AND** 第二次调用 `ccc supervisor-hook`（有 `CCC_SUPERVISOR_HOOK=1`）
- **AND** 应当返回 `{"decision":"","":""}`，允许 Supervisor 停止

### Requirement: 防止死循环 - 迭代次数限制

系统 SHALL 限制迭代次数防止无限循环。

#### Scenario: 迭代次数限制
- **GIVEN** session 的迭代次数已达到 10
- **WHEN** hook 被触发
- **THEN** 应当输出空内容
- **AND** 应当允许 Agent 停止

#### Scenario: 迭代次数递增
- **GIVEN** session 当前迭代次数为 3
- **WHEN** hook 被触发
- **THEN** 应当将迭代次数更新为 4
- **AND** 应当继续执行 Supervisor 检查

### Requirement: 状态管理

系统 SHALL 使用文件管理 session 状态。

#### Scenario: 状态目录确定
- **GIVEN** 环境变量 `CCC_CONFIG_DIR` 设置为 `/custom/path`
- **WHEN** 系统确定状态目录
- **THEN** 状态目录应当为 `/custom/path/ccc`

#### Scenario: 状态目录默认值
- **GIVEN** 环境变量 `CCC_CONFIG_DIR` 未设置
- **WHEN** 系统确定状态目录
- **THEN** 状态目录应当为 `~/.claude/ccc/`

#### Scenario: 状态文件路径
- **GIVEN** session_id 为 "abc123"
- **AND** 状态目录为 `.claude/ccc`
- **WHEN** 系统访问状态文件
- **THEN** 状态文件路径应当为 `.claude/ccc/supervisor-abc123.json`

#### Scenario: 状态文件结构
- **GIVEN** session_id 为 "abc123"
- **WHEN** 系统保存状态
- **THEN** 状态文件应当包含：
  - `session_id`: "abc123"
  - `count`: 迭代次数
  - `created_at`: 创建时间（ISO 8601）
  - `updated_at`: 更新时间（ISO 8601）

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

### Requirement: Supervisor Mode 启动提示

当 Supervisor Mode 启动时，系统 SHALL 在 stderr 输出 log 文件路径信息。

#### Scenario: 显示 log 路径提示
- **GIVEN** 环境变量 `CCC_SUPERVISOR=1` 已设置
- **WHEN** 用户执行 `ccc <provider>`
- **THEN** 应当在 stderr 输出 "[Supervisor Mode] 日志文件:" 提示
- **AND** 应当输出 state 目录路径
- **AND** 应当输出 hook 调用日志路径
- **AND** 应当输出 supervisor 输出日志路径

#### Scenario: State 目录路径计算
- **GIVEN** 环境变量 `CCC_WORK_DIR` 未设置
- **WHEN** 系统计算 state 目录路径
- **THEN** 路径应当为 `~/.claude/ccc`

#### Scenario: 自定义 State 目录
- **GIVEN** 环境变量 `CCC_WORK_DIR=/tmp/test` 已设置
- **WHEN** 系统计算 state 目录路径
- **THEN** 路径应当为 `/tmp/test/ccc`

### Requirement: Hook 执行日志输出

当 Stop hook 执行时，系统 SHALL 在 stderr 输出结构化的执行进度信息。

#### Scenario: Hook 调用开始
- **GIVEN** Stop hook 被触发
- **WHEN** `ccc supervisor-hook` 开始执行
- **THEN** 应当在 stderr 输出 "[SUPERVISOR HOOK] 开始执行" 分节符
- **AND** 应当输出 session_id 和当前迭代次数

#### Scenario: Supervisor 调用中
- **GIVEN** hook 准备调用 Supervisor
- **WHEN** Supervisor claude 启动
- **THEN** 应当在 stderr 输出 "[SUPERVISOR] 正在审查工作..."
- **AND** 应当输出 "请在新窗口查看日志文件了解详情"

#### Scenario: 审查结果输出
- **GIVEN** Supervisor 返回结果
- **WHEN** `completed` 为 `false`
- **THEN** 应当在 stderr 输出 "[SUPERVISOR] 任务未完成"
- **AND** 应当输出 feedback 内容
- **AND** 应当输出 "Agent 将根据反馈继续工作"

#### Scenario: 任务完成
- **GIVEN** Supervisor 返回 `completed: true`
- **WHEN** hook 处理结果
- **THEN** 应当在 stderr 输出 "[SUPERVISOR] 任务已完成"
- **AND** 应当输出 "允许停止"

### Requirement: 日志文件格式

系统 SHALL 使用易读的格式记录日志。

#### Scenario: hook-invocation.log 格式
- **GIVEN** hook 被调用
- **WHEN** 系统记录日志到 `hook-invocation.log`
- **THEN** 每条记录应当包含 ISO 8601 时间戳
- **AND** 应当包含事件类型（如 "supervisor-hook invoked"）
- **AND** 应当包含关键参数（如 session_id, iteration count）

#### Scenario: supervisor 输出日志格式
- **GIVEN** Supervisor 输出 stream-json
- **WHEN** 系统保存输出到 `supervisor-{session}-output.jsonl`
- **THEN** 应当保留原始 stream-json 行
- **AND** 应当同时在 hook-invocation.log 中记录摘要

### Requirement: Supervisor 结果解析 Fallback

当 Supervisor 返回的结果无法解析为符合 Schema 的 JSON 时，系统 SHALL 将原始内容作为 feedback，并设置 `allow_stop=false` 让 Agent 继续工作。

#### Scenario: 解析失败时使用原始内容作为 feedback
- **GIVEN** Supervisor 返回的 result 内容无法解析为有效 JSON
- **WHEN** 系统尝试解析 Supervisor 结果
- **THEN** 应当将原始 result 内容作为 feedback
- **AND** 应当设置 `allow_stop=false`
- **AND** Agent 应当继续工作

#### Scenario: 空结果时的默认反馈
- **GIVEN** Supervisor 返回的 result 为空字符串
- **WHEN** 系统尝试解析 Supervisor 结果
- **THEN** 应当使用默认 feedback "请继续完成任务"
- **AND** 应当设置 `allow_stop=false`

### Requirement: Supervisor Mode 工作流程

系统 SHALL 实现 Agent-Supervisor 自动循环，直到任务完成或达到迭代限制。

#### Scenario: 首次启动 Supervisor Mode
- **GIVEN** 用户执行 `ccc --supervisor`
- **AND** SUPERVISOR.md 文件存在
- **WHEN** Claude Code 首次触发 Stop hook
- **THEN** supervisor-hook 应被调用
- **AND** 迭代次数应初始化为 1
- **AND** Supervisor 应检查工作质量

#### Scenario: Supervisor 反馈继续工作
- **GIVEN** Supervisor 返回 `{"completed":false,"feedback":"需要添加错误处理"}`
- **WHEN** supervisor-hook 输出反馈
- **THEN** Claude 应收到 `{"decision":"block","reason":"需要添加错误处理"}`
- **AND** Claude 应继续工作
- **AND** 下次 Stop 时迭代次数应为 2

#### Scenario: Supervisor 确认完成
- **GIVEN** Supervisor 返回 `{"completed":true,"feedback":""}`
- **WHEN** supervisor-hook 处理结果
- **THEN** 应输出空到 stdout
- **AND** Claude 应停止（不再继续）

#### Scenario: 达到迭代限制
- **GIVEN** 迭代次数已达到 10 次
- **WHEN** supervisor-hook 被调用
- **THEN** 应输出空到 stdout
- **AND** 应允许 Claude 停止
- **AND** 不应再调用 Supervisor

### Requirement: JSON Schema 输出

系统 SHALL 使用 JSON Schema 强制 Supervisor 返回结构化输出。

#### Scenario: Supervisor 返回完成
- **GIVEN** Agent 已完成所有用户要求的任务
- **WHEN** Supervisor 被调用
- **THEN** 应返回 `{"completed":true,"feedback":"任务已成功完成"}`
- **AND** completed 字段应为 true

#### Scenario: Supervisor 返回未完成
- **GIVEN** Agent 未完成用户要求的任务
- **WHEN** Supervisor 被调用
- **THEN** 应返回 `{"completed":false,"feedback":"具体的问题和改进建议"}`
- **AND** completed 字段应为 false
- **AND** feedback 字段应包含具体的反馈

#### Scenario: JSON Schema 格式要求
- **WHEN** 调用 Supervisor
- **THEN** 应使用 `--json-schema` 参数指定 schema
- **AND** schema 应要求 completed 和 feedback 字段
- **AND** schema 应定义 completed 为 boolean 类型
- **AND** schema 应定义 feedback 为 string 类型

### Requirement: Supervisor Prompt

系统 SHALL 从 SUPERVISOR.md 读取 Supervisor 提示词。

#### Scenario: 读取 Supervisor Prompt
- **GIVEN** `~/.claude/SUPERVISOR.md` 文件存在
- **WHEN** 调用 Supervisor
- **THEN** 应使用 `--system-prompt` 参数传入 SUPERVISOR.md 内容
- **AND** SUPERVISOR.md 应包含 JSON Schema 输出格式说明

#### Scenario: 默认 Supervisor Prompt
- **GIVEN** `~/.claude/SUPERVISOR.md` 文件不存在
- **WHEN** 用户首次使用 Supervisor Mode
- **THEN** 应创建默认的 SUPERVISOR.md
- **AND** 默认内容应包含角色说明和输出格式要求

### Requirement: 状态文件管理

系统 SHALL 使用 JSON 文件管理每个 session 的迭代状态。

#### Scenario: 状态文件结构
- **GIVEN** session_id 为 "abc123"
- **WHEN** 创建状态文件
- **THEN** 文件路径应为 `.claude/ccc/supervisor-abc123.json`
- **AND** 内容应包含: session_id, count, created_at, updated_at
- **AND** count 应为当前迭代次数

#### Scenario: 状态文件持久化
- **GIVEN** 状态文件已存在
- **WHEN** supervisor-hook 被调用
- **THEN** 应读取现有状态
- **AND** 应增加 count
- **AND** 应更新 updated_at 时间戳
- **AND** 应保存回文件

#### Scenario: 状态文件并发处理
- **GIVEN** 多个 hook 可能同时执行（理论上不应发生）
- **WHEN** 读写状态文件
- **THEN** 应使用文件锁或原子操作避免竞态条件

### Requirement: 输出文件管理

系统 SHALL 保存 Supervisor 的原始输出到 JSONL 文件。

#### Scenario: 输出文件创建
- **GIVEN** session_id 为 "abc123"
- **WHEN** supervisor-hook 首次被调用
- **THEN** 应创建 `.claude/ccc/supervisor-abc123-output.jsonl`
- **AND** 文件应以 append 模式写入

#### Scenario: 输出文件内容
- **GIVEN** Supervisor 输出 stream-json
- **WHEN** 处理输出
- **THEN** 每行应作为 JSON 对象写入文件
- **AND** 应保持原始格式（包括 whitespace）
- **AND** 文件应为有效的 JSONL 格式

#### Scenario: 输出文件用途
- **GIVEN** 输出文件存在
- **WHEN** 用户需要调试
- **THEN** 文件可用于查看 Supervisor 的完整输出
- **AND** 文件可用于分析 Supervisor 的决策过程

### Requirement: 错误处理

系统 SHALL 正确处理 supervisor-hook 执行中的错误。

#### Scenario: Supervisor 调用失败
- **GIVEN** claude 命令执行失败
- **WHEN** supervisor-hook 调用 Supervisor
- **THEN** 应输出错误信息到 stderr
- **AND** 应返回空到 stdout（允许停止）
- **AND** 退出码应为 0（不影响 Claude）

#### Scenario: 状态文件读写失败
- **GIVEN** 状态文件读写权限不足
- **WHEN** supervisor-hook 尝试读写状态
- **THEN** 应输出错误信息到 stderr
- **AND** 应继续执行（使用默认值或跳过状态管理）

#### Scenario: JSON 解析失败
- **GIVEN** Supervisor 返回无效的 JSON
- **WHEN** supervisor-hook 解析结果
- **THEN** 应输出错误信息到 stderr
- **AND** 应返回空到 stdout（允许停止）

### Requirement: Fork Session 使用

系统 SHALL 使用 --fork-session 避免污染主 session。

#### Scenario: Supervisor 使用 Fork Session
- **WHEN** supervisor-hook 调用 Supervisor
- **THEN** 应使用 `--fork-session` 参数
- **AND** 应使用 `--resume <session_id>` 恢复上下文
- **AND** Supervisor 的输出不应影响主 session

#### Scenario: Supervisor Settings 隔离
- **GIVEN** 主 settings 包含 Stop hook
- **WHEN** 调用 Supervisor
- **THEN** 应使用 supervisor 专用 settings
- **AND** supervisor settings 不应包含任何 hooks
- **AND** 避免 hook 递归调用

