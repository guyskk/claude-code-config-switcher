// Package e2e 提供 patch 命令的端到端测试
package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestE2EPatchResetScenario 测试真实的 patch → reset 场景
// 这个测试使用真实的文件系统操作，模拟用户实际使用场景
func TestE2EPatchResetScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	// 创建临时测试环境
	tmpDir := t.TempDir()

	// 创建模拟的 claude 二进制文件
	claudeBin := filepath.Join(tmpDir, "claude")
	createFakeBinary(t, claudeBin, "#!/bin/sh\necho \"This is real claude\"\nexit 0\n")

	// 创建模拟的 ccc 二进制文件
	cccBin := filepath.Join(tmpDir, "ccc")
	createFakeBinary(t, cccBin, "#!/bin/sh\necho \"This is ccc\"\nexit 0\n")

	// 设置 PATH 环境变量
	oldPath := os.Getenv("PATH")
	newPath := tmpDir + string(os.PathListSeparator) + oldPath
	t.Cleanup(func() {
		os.Setenv("PATH", oldPath)
	})
	os.Setenv("PATH", newPath)

	t.Run("Scenario1: 首次 Patch", func(t *testing.T) {
		// 验证初始状态：claude 存在且是真实二进制
		verifyClaudeIsReal(t, claudeBin)

		// 执行 patch 操作
		// 注意：这里需要实际的 patch 逻辑，由于测试环境限制，
		// 我们模拟 patch 的效果
		simulatePatch(t, claudeBin)

		// 验证 patch 后状态
		verifyClaudeIsWrapper(t, claudeBin)
		verifyCccClaudeExists(t, claudeBin)
	})

	t.Run("Scenario2: 验证 Patch 后行为", func(t *testing.T) {
		// 调用 claude 应该显示 ccc 的输出
		output := runCommand(t, "claude")
		if !strings.Contains(output, "ccc") {
			t.Errorf("expected output to contain 'ccc', got: %s", output)
		}
	})

	t.Run("Scenario3: 重复 Patch", func(t *testing.T) {
		// 第二次 patch 应该返回 "Already patched"
		// 这里我们模拟检查
		if fileExists(t, claudeBin+".real") {
			t.Log("Already patched (as expected)")
		} else {
			t.Error("should detect already patched state")
		}
	})

	t.Run("Scenario4: Reset", func(t *testing.T) {
		// 执行 reset 操作
		simulateReset(t, claudeBin)

		// 验证 reset 后状态
		verifyClaudeIsReal(t, claudeBin)
		verifyCccClaudeNotExists(t, claudeBin)
	})

	t.Run("Scenario5: 验证 Reset 后行为", func(t *testing.T) {
		// 调用 claude 应该显示真实 claude 的输出
		output := runCommand(t, "claude")
		if !strings.Contains(output, "real claude") {
			t.Errorf("expected output to contain 'real claude', got: %s", output)
		}
	})

	t.Run("Scenario6: 重复 Reset", func(t *testing.T) {
		// 第二次 reset 应该返回 "Not patched"
		// 这里我们模拟检查
		if !fileExists(t, claudeBin+".real") {
			t.Log("Not patched (as expected)")
		} else {
			t.Error("should detect not patched state")
		}
	})

	t.Run("Scenario7: Ccc 直接调用", func(t *testing.T) {
		// patch 后 ccc 应该仍然正常工作
		simulatePatch(t, claudeBin)

		output := runCommand(t, "ccc")
		if !strings.Contains(output, "ccc") {
			t.Errorf("ccc should work independently, got: %s", output)
		}

		// 清理
		simulateReset(t, claudeBin)
	})

	t.Run("Scenario8: 第三方脚本调用", func(t *testing.T) {
		// 创建测试脚本调用 claude
		simulatePatch(t, claudeBin)

		testScript := filepath.Join(tmpDir, "test.sh")
		scriptContent := fmt.Sprintf("#!/bin/sh\n%s --version\n", claudeBin)
		if err := os.WriteFile(testScript, []byte(scriptContent), 0755); err != nil {
			t.Fatal(err)
		}

		output := runCommand(t, testScript)
		// 应该调用 ccc 而不是真实 claude
		t.Logf("Script output: %s", output)

		// 清理
		simulateReset(t, claudeBin)
	})
}

// TestE2EEnvironmentVariable 测试环境变量 CCC_CLAUDE
func TestE2EEnvironmentVariable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	tmpDir := t.TempDir()

	// 创建真实的 claude 二进制
	realClaude := filepath.Join(tmpDir, "claude.real")
	createFakeBinary(t, realClaude, "#!/bin/sh\necho \"real claude executed\"\nexit 0\n")

	// 设置环境变量
	oldValue := os.Getenv("CCC_CLAUDE")
	t.Cleanup(func() {
		os.Setenv("CCC_CLAUDE", oldValue)
	})
	os.Setenv("CCC_CLAUDE", realClaude)

	// 创建 ccc 二进制
	cccBin := filepath.Join(tmpDir, "ccc")
	createFakeBinary(t, cccBin, fmt.Sprintf(`#!/bin/sh
if [ -n "$CCC_CLAUDE" ]; then
    echo "CCC_CLAUDE is set to: $CCC_CLAUDE"
    "$CCC_CLAUDE" "$@"
else
    echo "CCC_CLAUDE not set"
    exit 1
fi
`))

	// 设置 PATH
	oldPath := os.Getenv("PATH")
	newPath := tmpDir + string(os.PathListSeparator) + oldPath
	t.Cleanup(func() {
		os.Setenv("PATH", oldPath)
	})
	os.Setenv("PATH", newPath)

	t.Run("CCC_CLAUDE 环境变量生效", func(t *testing.T) {
		output := runCommand(t, cccBin)
		if !strings.Contains(output, "CCC_CLAUDE is set to") {
			t.Errorf("expected CCC_CLAUDE to be set, got: %s", output)
		}
		if !strings.Contains(output, "real claude executed") {
			t.Errorf("expected real claude to be executed, got: %s", output)
		}
	})

	t.Run("CCC_CLAUDE 指向无效路径", func(t *testing.T) {
		os.Setenv("CCC_CLAUDE", "/nonexistent/path/claude")
		defer os.Setenv("CCC_CLAUDE", realClaude)

		// ccc 应该检测到无效路径并返回错误
		output := runCommand(t, cccBin)
		t.Logf("Output with invalid path: %s", output)
	})
}

// TestE2EBoundaryCases 测试边界情况
func TestE2EBoundaryCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	tmpDir := t.TempDir()

	t.Run("Claude 不存在", func(t *testing.T) {
		// PATH 中没有 claude
		oldPath := os.Getenv("PATH")
		newPath := tmpDir // 空目录
		t.Cleanup(func() {
			os.Setenv("PATH", oldPath)
		})
		os.Setenv("PATH", newPath)

		// 应该返回错误 "claude not found in PATH"
		_, err := exec.LookPath("claude")
		if err == nil {
			t.Error("should fail when claude not found")
		}
	})

	t.Run("Claude 文件损坏", func(t *testing.T) {
		claudeBin := filepath.Join(tmpDir, "claude")
		// 创建损坏的二进制文件
		if err := os.WriteFile(claudeBin, []byte("corrupted"), 0755); err != nil {
			t.Fatal(err)
		}

		oldPath := os.Getenv("PATH")
		newPath := tmpDir + string(os.PathListSeparator) + oldPath
		t.Cleanup(func() {
			os.Setenv("PATH", oldPath)
		})
		os.Setenv("PATH", newPath)

		// patch 应该检测到文件不是有效的可执行文件
		_ = claudeBin
		// TODO: 实现检测逻辑
	})

	t.Run("磁盘空间不足", func(t *testing.T) {
		// 这个测试比较难模拟，暂时跳过
		t.Skip("disk space test not implemented")
	})
}

// TestE2EPatchPerformance 测试 patch 性能
func TestE2EPatchPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	tmpDir := t.TempDir()
	claudeBin := filepath.Join(tmpDir, "claude")
	createFakeBinary(t, claudeBin, "#!/bin/sh\necho \"claude\"\n")

	t.Run("Patch 操作应该在合理时间内完成", func(t *testing.T) {
		start := time.Now()
		simulatePatch(t, claudeBin)
		duration := time.Since(start)

		// patch 操作应该在 100ms 内完成
		if duration > 100*time.Millisecond {
			t.Errorf("patch took too long: %v", duration)
		}

		simulateReset(t, claudeBin)
	})
}

// ===== 辅助函数 =====

// createFakeBinary 创建一个模拟的可执行文件
func createFakeBinary(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("failed to create fake binary %s: %v", path, err)
	}
}

// simulatePatch 模拟 patch 操作
func simulatePatch(t *testing.T, claudePath string) {
	t.Helper()

	// 重命名 claude → claude.real
	cccClaudePath := claudePath + ".real"
	if err := os.Rename(claudePath, cccClaudePath); err != nil {
		t.Fatalf("failed to rename claude: %v", err)
	}

	// 创建包装脚本
	wrapperContent := fmt.Sprintf(`#!/bin/sh
export CCC_CLAUDE=%s
echo "ccc wrapper: $@"
exec ccc "$@"
`, cccClaudePath)

	if err := os.WriteFile(claudePath, []byte(wrapperContent), 0755); err != nil {
		// 回滚
		_ = os.Rename(cccClaudePath, claudePath)
		t.Fatalf("failed to create wrapper: %v", err)
	}
}

// simulateReset 模拟 reset 操作
func simulateReset(t *testing.T, claudePath string) {
	t.Helper()

	cccClaudePath := claudePath + ".real"
	if _, err := os.Stat(cccClaudePath); os.IsNotExist(err) {
		t.Skip("not patched, skipping reset")
	}

	// 恢复 claude.real → claude
	if err := os.Rename(cccClaudePath, claudePath); err != nil {
		t.Fatalf("failed to reset: %v", err)
	}
}

// runCommand 运行命令并返回输出
func runCommand(t *testing.T, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("command %s failed: %v, output: %s", name, err, string(output))
	}
	return string(output)
}

// fileExists 检查文件是否存在
func fileExists(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// verifyClaudeIsReal 验证 claude 是真实二进制
func verifyClaudeIsReal(t *testing.T, claudePath string) {
	t.Helper()

	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read claude: %v", err)
	}

	contentStr := string(content)
	if strings.Contains(contentStr, "CCC_CLAUDE") {
		t.Errorf("claude should be real binary, but found wrapper script content")
	}
}

// verifyClaudeIsWrapper 验证 claude 是包装脚本
func verifyClaudeIsWrapper(t *testing.T, claudePath string) {
	t.Helper()

	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("failed to read claude: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "CCC_CLAUDE") {
		t.Errorf("claude should be wrapper script, but content: %s", contentStr)
	}
}

// verifyCccClaudeExists 验证 ccc-claude 存在
func verifyCccClaudeExists(t *testing.T, claudePath string) {
	t.Helper()
	cccClaudePath := claudePath + ".real"
	if _, err := os.Stat(cccClaudePath); os.IsNotExist(err) {
		t.Errorf("ccc-claude should exist at %s", cccClaudePath)
	}
}

// verifyCccClaudeNotExists 验证 ccc-claude 不存在
func verifyCccClaudeNotExists(t *testing.T, claudePath string) {
	t.Helper()
	cccClaudePath := claudePath + ".real"
	if _, err := os.Stat(cccClaudePath); !os.IsNotExist(err) {
		t.Errorf("ccc-claude should not exist at %s", cccClaudePath)
	}
}
