## 项目背景

具体内容见 @README.md（必读）

## 开发流程

本项目使用 TDD 开发流程，你要根据具体需求，用 /test-driven-development 按流程完成开发任务。

### 特别说明：你要使用中文

我的母语是中文，你要用中文和我沟通，这样更方便。

## Write idiomatic Go code

Write idiomatic Go code with goroutines, channels, and interfaces. Optimizes concurrency, implements Go patterns, and ensures proper error handling. Use PROACTIVELY for Go refactoring, concurrency issues, or performance optimization.

When write code:
1. Analyze requirements and design idiomatic Go solutions
2. Implement concurrency patterns using goroutines, channels, and select
3. Create clear interfaces and struct composition patterns
4. Establish comprehensive error handling with custom error types
5. Set up testing framework with table-driven tests and benchmarks
6. Optimize performance using pprof profiling and measurements

Process:
- Prioritize simplicity first - clear is better than clever
- Apply composition over inheritance through well-designed interfaces
- Implement explicit error handling with no hidden magic
- Design concurrent systems that are safe by default
- Benchmark thoroughly before optimizing performance
- Prefer standard library solutions over external dependencies
- Follow effective Go guidelines and community best practices
- Organize code with proper module management and clear package structure

Provide:
-  Idiomatic Go code following effective Go guidelines and conventions
-  Concurrent code with proper synchronization and race condition prevention
-  Table-driven tests with subtests for comprehensive coverage
-  Benchmark functions for performance-critical code paths
-  Error handling with wrapped errors, context, and custom error types
-  Clear interfaces and struct composition patterns
-  go.mod setup with minimal, well-justified dependencies
-  Performance profiling setup and optimization recommendations

## 提交前检查

```bash
Usage: ./check.sh [OPTIONS]

Options:
  -l, --lint          Run lint checks (gofmt, go vet, shellcheck, markdownlint)
  -t, --test          Run tests with race detector
  -b, --build         Run build validation
  -h, --help          Show this help message

If no options specified, runs all checks (lint, test, build).

Examples:
  ./check.sh                          # Run all checks
  ./check.sh --lint                   # Run lint only
```

## 发版流程

### 版本号规范

遵循 [Semantic Versioning](https://semver.org/) 规范：
- **Major (X.0.0)**: 破坏性变更，用户需要修改配置或代码
- **Minor (0.X.0)**: 新功能，向后兼容
- **Patch (0.0.X)**: Bug 修复，向后兼容

### 发版前检查清单

1. **代码质量检查**
   ```bash
   ./check.sh  # 运行完整的 lint、test、build 检查
   ```

2. **CHANGELOG 更新**
   - 在 `CHANGELOG.md` 顶部添加新版本记录
   - 按照 [Keep a Changelog](https://keepachangelog.com/) 格式编写
   - 分类：Added、Changed、Deprecated、Removed、Fixed、Security
   - 包含相关 PR 编号（如 #85、#86）
   - 更新底部的版本比较链接

3. **版本号判断**
   - 分析自上个版本以来的所有提交
   - 根据变更类型决定版本号：
     - 有破坏性变更 → Major
     - 有新功能 → Minor
     - 仅 bug 修复 → Patch

### Git 发版流程

由于 main 分支设置了保护规则，必须通过 PR 合并：

1. **创建 release 分支**
   ```bash
   git checkout main
   git pull origin main
   git checkout -b release/vX.Y.Z
   ```

2. **更新 CHANGELOG**
   ```bash
   # 编辑 CHANGELOG.md 添加新版本记录
   git add CHANGELOG.md
   git commit -m "docs(changelog): release vX.Y.Z"
   ```

3. **推送并创建 PR**
   ```bash
   git push origin release/vX.Y.Z
   gh pr create --base main --head release/vX.Y.Z \
     --title "docs(changelog): release vX.Y.Z" \
     --body "Release vX.Y.Z with [简要说明主要变更]"
   ```

4. **合并 PR 后创建 tag**
   ```bash
   git checkout main
   git pull origin main
   git tag vX.Y.Z
   git push origin vX.Y.Z  # 推送 tag 自动触发发布
   ```

5. **清理 release 分支**（可选）
   ```bash
   git branch -d release/vX.Y.Z
   git push origin --delete release/vX.Y.Z
   ```

### 自动发布机制

- 推送 tag 会自动触发 GitHub Actions 发布流程
- 无需手动创建 GitHub Release
- 二进制文件会自动构建并上传到 Release

### CHANGELOG 编写规范

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- 新功能说明 (#PR_NUMBER)

### Changed
- 功能变更说明 (#PR_NUMBER)

### Fixed
- Bug 修复说明 (#PR_NUMBER)

### Removed
- 移除功能说明 (#PR_NUMBER)

[X.Y.Z]: https://github.com/guyskk/claude-code-config-switcher/compare/vA.B.C...vX.Y.Z
```

### 版本号判断示例

- 仅修复了 settings.json 环境变量冲突问题 → **Patch (0.4.1)**
- 新增了 --settings 参数自动覆盖功能 → **Minor (0.5.0)**
- 移除了旧的 env_guard 检查机制 → **Minor (0.5.0)**（向后兼容的改进）
- 修改了配置文件格式，用户需要迁移 → **Major (1.0.0)**
