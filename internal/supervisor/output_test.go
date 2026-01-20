package supervisor

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// captureStdout 捕获 stdout 输出
func captureStdout(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// createTestLogger 创建测试用的 logger
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// ============================================================================
// OutputDecision 测试
// ============================================================================

func TestOutputDecision_AllowStop(t *testing.T) {
	log := createTestLogger()

	output := captureStdout(func() {
		err := OutputDecision(log, true, "work completed")
		if err != nil {
			t.Errorf("OutputDecision() error = %v", err)
		}
	})

	// 验证输出是有效的 JSON
	var result StopHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	// 允许停止时，decision 应该为 nil
	if result.Decision != nil {
		t.Errorf("Decision = %q, want nil when allowing stop", *result.Decision)
	}

	// reason 应该包含 feedback
	if result.Reason != "work completed" {
		t.Errorf("Reason = %q, want 'work completed'", result.Reason)
	}
}

func TestOutputDecision_Block(t *testing.T) {
	log := createTestLogger()

	output := captureStdout(func() {
		err := OutputDecision(log, false, "needs more work")
		if err != nil {
			t.Errorf("OutputDecision() error = %v", err)
		}
	})

	var result StopHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	// 阻止停止时，decision 应该是 "block"
	if result.Decision == nil {
		t.Fatal("Decision is nil, want 'block'")
	}
	if *result.Decision != "block" {
		t.Errorf("Decision = %q, want 'block'", *result.Decision)
	}

	if result.Reason != "needs more work" {
		t.Errorf("Reason = %q, want 'needs more work'", result.Reason)
	}
}

func TestOutputDecision_BlockWithEmptyFeedback(t *testing.T) {
	log := createTestLogger()

	output := captureStdout(func() {
		err := OutputDecision(log, false, "")
		if err != nil {
			t.Errorf("OutputDecision() error = %v", err)
		}
	})

	var result StopHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	// 空 feedback 时应该有默认消息
	if result.Reason == "" {
		t.Error("Reason should not be empty when feedback is empty and blocking")
	}
	if result.Reason != "Please continue completing the task" {
		t.Errorf("Reason = %q, want default message", result.Reason)
	}
}

func TestOutputDecision_TrimsWhitespace(t *testing.T) {
	log := createTestLogger()

	output := captureStdout(func() {
		err := OutputDecision(log, true, "  trimmed feedback  ")
		if err != nil {
			t.Errorf("OutputDecision() error = %v", err)
		}
	})

	var result StopHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	if result.Reason != "trimmed feedback" {
		t.Errorf("Reason = %q, want 'trimmed feedback' (whitespace trimmed)", result.Reason)
	}
}

// ============================================================================
// OutputPreToolUseDecision 测试
// ============================================================================

func TestOutputPreToolUseDecision_Allow(t *testing.T) {
	log := createTestLogger()

	output := captureStdout(func() {
		err := OutputPreToolUseDecision(log, true, "问题合理")
		if err != nil {
			t.Errorf("OutputPreToolUseDecision() error = %v", err)
		}
	})

	var result PreToolUseHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	if result.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}

	if result.HookSpecificOutput.HookEventName != "PreToolUse" {
		t.Errorf("HookEventName = %q, want 'PreToolUse'", result.HookSpecificOutput.HookEventName)
	}

	if result.HookSpecificOutput.PermissionDecision != "allow" {
		t.Errorf("PermissionDecision = %q, want 'allow'", result.HookSpecificOutput.PermissionDecision)
	}

	if result.HookSpecificOutput.PermissionDecisionReason != "问题合理" {
		t.Errorf("PermissionDecisionReason = %q, want '问题合理'", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

func TestOutputPreToolUseDecision_Deny(t *testing.T) {
	log := createTestLogger()

	output := captureStdout(func() {
		err := OutputPreToolUseDecision(log, false, "需要更多上下文")
		if err != nil {
			t.Errorf("OutputPreToolUseDecision() error = %v", err)
		}
	})

	var result PreToolUseHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	if result.HookSpecificOutput == nil {
		t.Fatal("HookSpecificOutput is nil")
	}

	if result.HookSpecificOutput.PermissionDecision != "deny" {
		t.Errorf("PermissionDecision = %q, want 'deny'", result.HookSpecificOutput.PermissionDecision)
	}

	if result.HookSpecificOutput.PermissionDecisionReason != "需要更多上下文" {
		t.Errorf("PermissionDecisionReason = %q, want '需要更多上下文'", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

func TestOutputPreToolUseDecision_DenyWithEmptyFeedback(t *testing.T) {
	log := createTestLogger()

	output := captureStdout(func() {
		err := OutputPreToolUseDecision(log, false, "")
		if err != nil {
			t.Errorf("OutputPreToolUseDecision() error = %v", err)
		}
	})

	var result PreToolUseHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	// 空 feedback 时应该有默认消息
	if result.HookSpecificOutput.PermissionDecisionReason == "" {
		t.Error("PermissionDecisionReason should not be empty when feedback is empty and denying")
	}
}

func TestOutputPreToolUseDecision_TrimsWhitespace(t *testing.T) {
	log := createTestLogger()

	output := captureStdout(func() {
		err := OutputPreToolUseDecision(log, true, "  trimmed  ")
		if err != nil {
			t.Errorf("OutputPreToolUseDecision() error = %v", err)
		}
	})

	var result PreToolUseHookOutput
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("Failed to parse output JSON: %v\nOutput: %s", err, output)
	}

	if result.HookSpecificOutput.PermissionDecisionReason != "trimmed" {
		t.Errorf("PermissionDecisionReason = %q, want 'trimmed'", result.HookSpecificOutput.PermissionDecisionReason)
	}
}

// ============================================================================
// HookEventType 常量测试
// ============================================================================

func TestHookEventType_Constants(t *testing.T) {
	if EventTypeStop != "Stop" {
		t.Errorf("EventTypeStop = %q, want 'Stop'", EventTypeStop)
	}

	if EventTypePreToolUse != "PreToolUse" {
		t.Errorf("EventTypePreToolUse = %q, want 'PreToolUse'", EventTypePreToolUse)
	}
}

// ============================================================================
// 输出类型结构测试
// ============================================================================

func TestStopHookOutput_JSONFormat(t *testing.T) {
	// 测试 Stop 事件的 JSON 输出格式
	block := "block"
	output := StopHookOutput{
		Decision: &block,
		Reason:   "测试原因",
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(jsonBytes)

	if !strings.Contains(jsonStr, `"decision"`) {
		t.Error("JSON should contain 'decision' field")
	}
	if !strings.Contains(jsonStr, `"reason"`) {
		t.Error("JSON should contain 'reason' field")
	}
	if !strings.Contains(jsonStr, `"block"`) {
		t.Error("JSON should contain 'block' value")
	}
}

func TestPreToolUseHookOutput_JSONFormat(t *testing.T) {
	// 测试 PreToolUse 事件的 JSON 输出格式
	output := PreToolUseHookOutput{
		HookSpecificOutput: &PreToolUseSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "allow",
			PermissionDecisionReason: "测试原因",
		},
	}

	jsonBytes, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(jsonBytes)

	if !strings.Contains(jsonStr, `"hookSpecificOutput"`) {
		t.Error("JSON should contain 'hookSpecificOutput' field")
	}
	if !strings.Contains(jsonStr, `"hookEventName"`) {
		t.Error("JSON should contain 'hookEventName' field")
	}
	if !strings.Contains(jsonStr, `"permissionDecision"`) {
		t.Error("JSON should contain 'permissionDecision' field")
	}
	if !strings.Contains(jsonStr, `"permissionDecisionReason"`) {
		t.Error("JSON should contain 'permissionDecisionReason' field")
	}
}

func TestHookOutput_Alias(t *testing.T) {
	// 验证 HookOutput 是 StopHookOutput 的别名
	var hookOutput HookOutput
	var stopHookOutput StopHookOutput

	// 它们应该是相同的类型
	hookOutput = stopHookOutput
	_ = hookOutput
}
