# 讨论：env 冲突处理改进方案 — 能否用进程 env 覆盖 settings.json env

日期：2026-06-09
话题：探索是否有更好的方式处理 settings.json env 与 ccc provider env 的冲突，是否能让 ccc 进程 env 覆盖 settings.json env。
前置文档：docs/discuss-20260519-env-priority.md（已实测确认 settings.json env > 进程 env）

> 本文件 Append-only，只追加不覆盖。

---

## 第 1 轮：问题理解与现状分析

### 1.1 用户的核心问题

当前 ccc 处理 env 冲突的策略是"硬守卫"：检测到 settings.json 的 `env` 中存在冲突 key（`ANTHROPIC_*`/`CLAUDE_*` 前缀或 managed key），直接报错中止，不启动 claude。

用户想知道：
1. 能否在 ccc 中对冲突的环境变量，全部用进程环境变量进行覆盖？
2. 即 settings.json `env` 的值 vs ccc claude 进程的 `env` 值冲突时，能否让 ccc 进程的 `env` 生效？

### 1.2 已知事实（来自 2026-05-19 实测）

**确定性结论**：Claude Code 中 settings.json 的 `env` 严格覆盖 claude 进程继承的 OS 环境变量。

- 实验 A：中性变量 `CCC_PRECEDENCE_TEST`，settings.json 的值覆盖进程 env
- 实验 B：`ANTHROPIC_BASE_URL`，settings.json 的值覆盖进程 env（决定性实验）
- 实验 C：settings.json 无 env 时，进程 env 兜底生效

**当前 ccc 的 env 传递链路**：
1. `os.Environ()` 继承 ccc 进程所有环境变量
2. 过滤掉所有 `CLAUDE_*`/`ANTHROPIC_*` 前缀
3. 追加 base + provider env
4. `syscall.Exec` 替换为 claude 进程

但 Claude Code 启动后，自己读取 settings.json 的 `env` 字段，用其值覆盖进程 env → ccc 传入的 provider env 被 settings.json 中残留的同名 key 覆盖。

### 1.3 需要深入研究的方向

用户的问题本质上是：**有没有办法让进程 env 的优先级高于 settings.json env？**

可能的探索方向：
1. **Claude Code 源码分析**：Claude Code 是如何读取 settings.json env 并应用到进程的？是否有命令行参数或环境变量可以改变这个行为？
2. **Claude Code 版本变化**：之前的实测用的是 claude 2.1.144，当前最新版本是否有变化？
3. **更高优先级的配置方式**：managed-settings.json 或命令行参数是否能覆盖 settings.json 的 env？
4. **变通方案**：在启动 claude 前临时修改 settings.json（移除冲突 key），启动后再恢复？

---

## 第 2 轮：突破性发现 — `--settings` 参数的 env 覆盖能力

### 2.1 关键发现

Claude Code 提供 `--settings <file-or-json>` 命令行参数，描述为 "Path to a settings JSON file or a JSON string to load additional settings from"。

**settings 优先级（文档确认）**：
1. Enterprise (managed-settings.json) — 最高
2. **Command line arguments** — 临时会话覆盖
3. Local — 覆盖 project 和 user settings
4. Project — 覆盖 user settings
5. User (最低) — `~/.claude/settings.json` 属于此级别

`--settings` 属于 "Command line arguments" 级别，优先级**高于** User settings（`~/.claude/settings.json`）。

### 2.2 实验验证（claude 2.1.169）

**实验 1 — `--settings` env 是否生效**：
- settings.json：无 env
- `--settings '{"env":{"CCC_SETTINGS_TEST":"VALUE_FROM_SETTINGS_PARAM"}}'`
- 结果：`VALUE_FROM_SETTINGS_PARAM` ✅

**实验 2 — `--settings` env 覆盖 settings.json env**：
- settings.json：`CCC_OVERRIDE_TEST=FROM_SETTINGS_JSON`
- `--settings '{"env":{"CCC_OVERRIDE_TEST":"FROM_SETTINGS_PARAM"}}'`
- 结果：`FROM_SETTINGS_PARAM` ✅ — `--settings` 覆盖了 settings.json

**实验 3 — 合并行为（决定性）**：
- settings.json：`CCC_TEST_A=A_FROM_SETTINGS_JSON`, `CCC_TEST_B=B_FROM_SETTINGS_JSON`
- `--settings '{"env":{"CCC_TEST_B":"B_FROM_SETTINGS_PARAM","CCC_TEST_C":"C_FROM_SETTINGS_PARAM"}}'`
- 结果：
  - `CCC_TEST_A` = `A_FROM_SETTINGS_JSON` ✅（settings.json 独有 key，保留）
  - `CCC_TEST_B` = `B_FROM_SETTINGS_PARAM` ✅（同名 key，`--settings` 覆盖）
  - `CCC_TEST_C` = `C_FROM_SETTINGS_PARAM` ✅（`--settings` 新增 key）

**确定性结论**：`--settings` 的 `env` 与 `settings.json` 的 `env` 是**合并**关系，同名 key 由 `--settings` 覆盖，不同 key 各自保留。

### 2.3 对 ccc 的意义

这意味着 ccc 可以通过 `--settings` 参数传入 provider env，从而：
1. **覆盖** settings.json 中的冲突 key（如 `ANTHROPIC_BASE_URL`），不需要修改 settings.json
2. **保留** settings.json 中非冲突的用户自定义 env（如 `MY_CUSTOM_VAR`）
3. **避免报错中止**，用户体验更好

### 2.4 方案设计：用 `--settings` 参数传入 provider env

**核心思路**：在启动 claude 时，通过 `--settings` 参数传入包含 provider env 的 JSON，利用 Command line arguments 的高优先级覆盖 settings.json 中的冲突 env key。

**当前流程**（exec.go `runClaude()`）：
```
1. 确定 provider
2. checkSettingsEnvConflict() → 冲突则报错中止 ← 需要移除
3. SwitchWithHook() → 合并配置、过滤 env、写入 settings.json ← env 逻辑需调整
4. 构建 execArgs（claude 命令行参数）
5. 构建进程 env：过滤 CLAUDE_*/ANTHROPIC_* + 追加 provider env
6. syscall.Exec 启动 claude
```

**新流程**：
```
1. 确定 provider
2. SwitchWithHook() → 合并配置、写入 settings.json（env 保留用户完整 env，不过滤冲突 key）
3. 构建 execArgs：
   - 追加 --settings '{"env": {...provider env...}}'
4. 构建进程 env：过滤 CLAUDE_*/ANTHROPIC_* + 追加 provider env（保持不变）
5. syscall.Exec 启动 claude
```

**关键变化**：
1. **移除 `checkSettingsEnvConflict()`**：不再需要硬守卫，因为 `--settings` 的 env 会覆盖 settings.json 的冲突 key
2. **简化 `SwitchWithHook()` 的 env 处理**：
   - settings.json 的 env 保留用户完整 env（不再用 `FilterUserEnvForSettings` 过滤冲突 key）
   - 子进程 env 构建不变（仍然只有 base + provider，不含 user env）
3. **新增 `--settings` 参数构建**：将 provider env 序列化为 `{"env": {...}}` JSON，作为 `--settings` 参数追加到 execArgs

**预期效果**：
- settings.json 保留用户完整 env（包括 `ANTHROPIC_*`/`CLAUDE_*`）→ 用户手动配置不丢失
- `--settings` 的 provider env 覆盖 settings.json 的冲突 key → provider 切换正确生效
- settings.json 的非冲突 key（如 `MY_CUSTOM_VAR`）仍然保留 → 用户自定义 env 不受影响
- 不需要报错中止 → 用户体验更好

### 2.5 需要验证的边界情况

1. **settings.json 没有 env**：`--settings` 的 env 应该正常生效（实验 1 已验证）
2. **settings.json 有 env 但不冲突**：`--settings` 的 env 与 settings.json 的 env 合并，各自生效（实验 3 已验证）
3. **settings.json 有 env 且冲突**：`--settings` 的 env 覆盖 settings.json 的冲突 key（实验 2 已验证）
4. **provider env 为空**：不需要传 `--settings` 参数
5. **`--settings` 与其他 CLI 参数的兼容性**：需要验证 `--settings` 与 `-p`、`--model` 等参数是否兼容

### 2.6 代码修改范围（预估）

| 文件 | 修改内容 |
|------|----------|
| `internal/cli/exec.go` | 移除 `checkSettingsEnvConflict()`；在 `runClaude()` 中构建 `--settings` JSON 并追加到 execArgs |
| `internal/provider/provider.go` | `SwitchWithHook()` 中 env 处理简化：不再过滤 settings.json 的 env，保留用户完整 env |
| `internal/config/config.go` | `FilterUserEnvForSettings()` 函数可能不再需要（需确认是否有其他调用方） |
| `internal/config/env_guard.go` | `DetectSettingsEnvConflicts()` 和 `FormatEnvConflictError()` 可能不再需要 |
| 测试文件 | 更新相关测试 |

### 2.7 转义和环境变量引用验证（claude 2.1.169）

**实验 4 — 特殊字符转义**：
- `--settings '{"env":{"CCC_ESCAPE_TEST":"value with \"quotes\" and \\backslash and $HOME ref"}}'`
- 结果：`value with "quotes" and \backslash and $HOME ref` ✅
- 双引号、反斜杠正确保留，`$HOME` 未展开（字面量）

**实验 5 — 环境变量引用**：
- 进程 env：`CCC_BASE_URL=http://example.com`
- `--settings '{"env":{"CCC_EXPAND_TEST":"prefix-${CCC_BASE_URL}-suffix"}}'`
- 结果：`prefix-${CCC_BASE_URL}-suffix` ✅（未展开，字面量）
- 说明 `--settings` 的值不会被 claude 展开环境变量引用

**结论**：
1. Go 的 `json.Marshal` 生成的 JSON 可以正确传递特殊字符给 claude
2. `--settings` 的值是字面量，ccc 在构建 JSON 时需要先用 `os.ExpandEnv()` 展开环境变量引用
3. 这与当前 `envMapToPairs()` 的行为一致（它已经调用了 `os.ExpandEnv()`）

### 2.8 用户决策

1. `--settings` 是稳定的正式 API ✅
2. 转义问题需要处理好并实际验证 ✅（已验证）
3. 不需要考虑 bare 模式 ✅
4. 硬守卫不需要了 ✅（完全移除）

### 2.9 最终方案

**核心变更**：用 `--settings` 参数传入 provider env，替代硬守卫报错中止。

**代码修改**：
1. `internal/cli/exec.go`：
   - 移除 `checkSettingsEnvConflict()` 函数及其调用
   - 在 `runClaude()` 中构建 `--settings` JSON：将 provider env 序列化为 `{"env": {...}}`
   - 将 `--settings` 和 JSON 追加到 `execArgs`
2. `internal/provider/provider.go`：
   - `SwitchWithHook()` 中移除 `FilterUserEnvForSettings` 调用
   - settings.json 的 env 保留用户完整 env（不再过滤冲突 key）
   - 子进程 env 构建不变（仍然只有 base + provider，不含 user env）
3. `internal/config/config.go`：
   - `FilterUserEnvForSettings()` 函数删除（无其他调用方）
4. `internal/config/env_guard.go`：
   - `DetectSettingsEnvConflicts()` 和 `FormatEnvConflictError()` 删除
5. 测试文件：更新相关测试
6. 文档：更新 settings-merge-strategy.md、README 等

**行为变化**：
- 之前：settings.json env 冲突 → 报错中止，用户必须手动清理
- 之后：settings.json env 冲突 → `--settings` 自动覆盖冲突 key，正常启动
- settings.json 保留用户完整 env → 用户手动配置不丢失

**状态：方案已完全对齐，无遗留疑问。**

---

## 第 3 轮：用户确认，进入计划模式

用户回复"好"，确认方案。进入计划模式编写详细 Plan。

---

## 第 4 轮：实施完成

### 4.1 实施内容

按 TDD 流程完成所有代码修改：

1. **新增 `buildProviderSettingsJSON()` 函数**（exec.go）：将 provider env 序列化为 `{"env": {...}}` JSON，环境变量引用预先展开
2. **修改 `runClaude()`**（exec.go）：移除 `checkSettingsEnvConflict` 调用，新增 `--settings` 参数构建
3. **删除 `checkSettingsEnvConflict()` 函数**（exec.go）
4. **修改 `SwitchWithHook()`**（provider.go）：保留用户完整 env（不再过滤冲突 key），新增 `ProviderEnv` 字段
5. **删除 `checkValidateEnvConflict()` 和 `validateTargetProviders()`**（cli.go）
6. **删除 `env_guard.go` 和 `env_guard_test.go`**
7. **删除 `FilterUserEnvForSettings()`**（config.go）
8. **更新测试**：删除旧的冲突检测测试，新增 `buildProviderSettingsJSON` 测试，更新 provider 测试为新行为
9. **更新文档**：README.md、README-CN.md、settings-merge-strategy.md

### 4.2 验证结果

- `./check.sh` 全量通过（lint + test + build）✅
- 所有 133 个测试用例通过 ✅
- 编译通过 ✅

### 4.3 行为变化

| 之前 | 之后 |
|------|------|
| settings.json env 冲突 → 报错中止 | settings.json env 冲突 → `--settings` 自动覆盖 |
| settings.json 的冲突 key 被过滤删除 | settings.json 保留用户完整 env |
| 用户必须手动清理 settings.json | 用户无需任何操作 |

---

