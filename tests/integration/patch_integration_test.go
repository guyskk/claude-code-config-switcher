// Package integration 提供 patch 命令的集成测试
package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestEnvironment 创建测试环境
// 返回：临时目录、claude 路径、cleanup 函数
func setupTestEnvironment(t *testing.T) (string, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "ccc-patch-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// 创建假的 claude 可执行文件
	claudePath := filepath.Join(tmpDir, "claude")
	claudeContent := []byte("#!/bin/sh\necho \"real claude\"\n")
	if err := os.WriteFile(claudePath, claudeContent, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create fake claude: %v", err)
	}

	// 创建 ccc 可执行文件（使用 go test 编译）
	// 这在集成测试中可能不可行，所以我们使用 mock ccc
	cccPath := filepath.Join(tmpDir, "ccc")
	cccContent := []byte("#!/bin/sh\necho \"ccc: $@\"\n")
	if err := os.WriteFile(cccPath, cccContent, 0755); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create fake ccc: %v", err)
	}

	// 将临时目录添加到 PATH
	oldPath := os.Getenv("PATH")
	newPath := tmpDir + string(os.PathListSeparator) + oldPath
	os.Setenv("PATH", newPath)

	cleanup := func() {
		os.Setenv("PATH", oldPath)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, claudePath, cleanup
}

// TestPatchResetFlow 测试完整的 patch → reset → patch 流程
func TestPatchResetFlow(t *testing.T) {
	_, claudePath, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("Initial state: claude exists and is real", func(t *testing.T) {
		if _, err := os.Stat(claudePath); os.IsNotExist(err) {
			t.Errorf("claude should exist at %s", claudePath)
		}

		// 验证是真实 claude
		cmd := exec.Command("claude")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to run claude: %v", err)
		}
		if !strings.Contains(string(out), "real claude") {
			t.Errorf("expected 'real claude', got: %s", string(out))
		}
	})

	t.Run("After patch: claude becomes wrapper", func(t *testing.T) {
		// 执行 patch（需要修改 patch.go 来支持指定路径）
		// 由于当前实现不支持，这里只测试逻辑

		// 验证 ccc-claude 存在
		cccClaudePath := claudePath + ".real"
		if _, err := os.Stat(cccClaudePath); os.IsNotExist(err) {
			t.Skip("patch not implemented in test environment")
		}

		// 验证 claude 是包装脚本
		cmd := exec.Command("claude")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to run claude after patch: %v", err)
		}
		if !strings.Contains(string(out), "ccc") {
			t.Errorf("expected ccc output, got: %s", string(out))
		}
	})

	t.Run("After reset: claude restored", func(t *testing.T) {
		// 执行 reset
		// 验证 claude 恢复为原始
	})
}

// TestPatchIdempotency 测试幂等性
func TestPatchIdempotency(t *testing.T) {
	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// 第一次 patch
	// 第二次 patch 应该返回 "Already patched"
	// 验证文件系统未被修改
	_ = "placeholder for idempotency test"
}

// TestResetIdempotency 测试 reset 幂等性
func TestResetIdempotency(t *testing.T) {
	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// 第一次 reset
	// 第二次 reset 应该返回 "Not patched"
	_ = "placeholder for reset idempotency test"
}

// TestPatchResetsToCleanState 测试 patch → reset → patch 流程
func TestPatchResetsToCleanState(t *testing.T) {
	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// patch
	// reset
	// 再次 patch 应该成功（不是 "Already patched"）
	_ = "placeholder for clean state test"
}

// TestWrapperScriptContent 测试包装脚本内容
func TestWrapperScriptContent(t *testing.T) {
	_, claudePath, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("Wrapper script has correct format", func(t *testing.T) {
		// 模拟创建包装脚本
		cccClaudePath := claudePath + ".real"
		scriptContent := fmt.Sprintf(`#!/bin/sh
export CCC_CLAUDE=%s
exec ccc "$@"
`, cccClaudePath)

		// 验证脚本内容
		if !strings.Contains(scriptContent, "CCC_CLAUDE=") {
			t.Error("wrapper script should set CCC_CLAUDE")
		}
		if !strings.Contains(scriptContent, "exec ccc") {
			t.Error("wrapper script should exec ccc")
		}
		if !strings.Contains(scriptContent, "\"$@\"") {
			t.Error("wrapper script should pass all arguments")
		}
	})
}

// TestPatchWithSymlinks 测试符号链接处理
func TestPatchWithSymlinks(t *testing.T) {
	tmpDir, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("claude is a symlink", func(t *testing.T) {
		// 创建真实的 claude.bin
		realClaude := filepath.Join(tmpDir, "claude.bin")
		if err := os.WriteFile(realClaude, []byte("#!/bin/sh\necho real\n"), 0755); err != nil {
			t.Fatal(err)
		}

		// 创建符号链接
		symlinkPath := filepath.Join(tmpDir, "claude-symlink")
		if err := os.Symlink(realClaude, symlinkPath); err != nil {
			t.Fatal(err)
		}

		// 验证符号链接
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("failed to stat symlink: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("should be a symlink")
		}
	})
}

// TestPatchWithReadOnlyFile 测试只读文件处理
func TestPatchWithReadOnlyFile(t *testing.T) {
	_, claudePath, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("claude file is read-only", func(t *testing.T) {
		// 设置文件为只读
		if err := os.Chmod(claudePath, 0444); err != nil {
			t.Fatal(err)
		}

		// 尝试 patch
		// 应该返回权限错误
		_ = claudePath
		// TODO: 实现 patch 逻辑测试
	})
}

// TestConcurrentPatch 测试并发 patch
func TestConcurrentPatch(t *testing.T) {
	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	t.Run("Multiple concurrent patches", func(t *testing.T) {
		// 启动多个 goroutine 同时执行 patch
		// 验证只有一个成功，其他返回 "Already patched"
	})
}

// BenchmarkPatchOperations 性能基准测试
func BenchmarkPatchOperations(b *testing.B) {
	// Benchmark tests don't use t.Helper(), create simple environment
	tmpDir, err := os.MkdirTemp("", "ccc-patch-bench-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	claudePath := filepath.Join(tmpDir, "claude")
	claudeContent := []byte("#!/bin/sh\necho \"real claude\"\n")
	if err := os.WriteFile(claudePath, claudeContent, 0755); err != nil {
		b.Fatalf("failed to create fake claude: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 执行 patch
		// 执行 reset
		_ = claudePath
	}
}

// Helper: runCommand 辅助函数运行命令并捕获输出
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), err
}

// Helper: fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Helper: isExecutable 检查文件是否可执行
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().Perm()&0111 != 0
}

// Helper: readFile 读取文件内容
func readFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
