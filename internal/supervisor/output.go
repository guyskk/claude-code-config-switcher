// Package supervisor provides supervisor output functionality.
package supervisor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// HookEventType 定义 hook 事件类型
type HookEventType string

const (
	// EventTypeStop 表示 Stop 事件（任务结束审查）
	EventTypeStop HookEventType = "Stop"
	// EventTypePreToolUse 表示 PreToolUse 事件（工具调用前审查）
	EventTypePreToolUse HookEventType = "PreToolUse"
)

// StopHookOutput 表示 Stop 事件的输出格式
// Reason is always set (provides context for the decision).
// Decision is "block" when not allowing stop, omitted when allowing stop.
type StopHookOutput struct {
	Decision *string `json:"decision,omitempty"` // "block" or omitted (allows stop)
	Reason   string  `json:"reason"`             // Always set
}

// PreToolUseSpecificOutput 表示 PreToolUse 事件的特定输出字段
type PreToolUseSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`            // "PreToolUse"
	PermissionDecision       string `json:"permissionDecision"`       // "allow", "deny"
	PermissionDecisionReason string `json:"permissionDecisionReason"` // 决策原因
}

// PreToolUseHookOutput 表示 PreToolUse 事件的输出格式
type PreToolUseHookOutput struct {
	HookSpecificOutput *PreToolUseSpecificOutput `json:"hookSpecificOutput"`
}

// HookOutput represents the output to stdout (for Stop events).
// Reason is always set (provides context for the decision).
// Decision is "block" when not allowing stop, omitted when allowing stop.
// Deprecated: 使用 StopHookOutput 代替，此类型保留用于向后兼容
type HookOutput = StopHookOutput

// OutputDecision outputs the supervisor's decision.
//
// Parameters:
//   - log: The logger to use
//   - allowStop: true to allow the agent to stop, false to block and require more work
//   - feedback: Feedback message explaining the decision (can be empty)
//
// The function:
// 1. Outputs JSON to stdout for Claude Code to parse
// 2. Logs the decision
func OutputDecision(log *slog.Logger, allowStop bool, feedback string) error {
	// Trim feedback
	feedback = strings.TrimSpace(feedback)

	// Build output
	output := HookOutput{Reason: feedback}
	if !allowStop {
		block := "block"
		output.Decision = &block
		if feedback == "" {
			output.Reason = "Please continue completing the task"
		}
	}

	outputJSON, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal hook output: %w", err)
	}

	// Log the decision
	if allowStop {
		log.Info("supervisor output: allow stop", "feedback", feedback)
	} else {
		log.Info("supervisor output: not allow stop", "feedback", output.Reason)
	}

	// Output JSON to stdout (for Claude Code to parse)
	fmt.Println(string(outputJSON))

	return nil
}

// OutputPreToolUseDecision 输出 PreToolUse 事件的决策
//
// 参数:
//   - log: 日志记录器
//   - allow: true 允许工具调用，false 拒绝工具调用
//   - feedback: 决策反馈信息
//
// 功能:
// 1. 输出 JSON 到 stdout 供 Claude Code 解析
// 2. 记录决策日志
func OutputPreToolUseDecision(log *slog.Logger, allow bool, feedback string) error {
	feedback = strings.TrimSpace(feedback)

	decision := "allow"
	if !allow {
		decision = "deny"
		if feedback == "" {
			feedback = "请继续完成任务后再提问"
		}
	}

	output := PreToolUseHookOutput{
		HookSpecificOutput: &PreToolUseSpecificOutput{
			HookEventName:            "PreToolUse",
			PermissionDecision:       decision,
			PermissionDecisionReason: feedback,
		},
	}

	outputJSON, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal PreToolUse hook output: %w", err)
	}

	// 记录决策日志
	if allow {
		log.Info("supervisor output: allow tool call", "feedback", feedback)
	} else {
		log.Info("supervisor output: deny tool call", "feedback", feedback)
	}

	// 输出 JSON 到 stdout
	fmt.Println(string(outputJSON))

	return nil
}
