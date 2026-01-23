// Package supervisor provides supervisor output functionality.
package supervisor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// HookEventType defines the type of hook event.
type HookEventType string

const (
	// EventTypeStop represents a Stop event (end-of-task review)
	EventTypeStop HookEventType = "Stop"
	// EventTypePreToolUse represents a PreToolUse event (pre-tool-call review)
	EventTypePreToolUse HookEventType = "PreToolUse"
)

// StopHookOutput represents the output format for Stop events.
// Reason is always set (provides context for the decision).
// Decision is "block" when not allowing stop, omitted when allowing stop.
type StopHookOutput struct {
	Decision *string `json:"decision,omitempty"` // "block" or omitted (allows stop)
	Reason   string  `json:"reason"`             // Always set
}

// PreToolUseSpecificOutput contains PreToolUse event-specific output fields.
type PreToolUseSpecificOutput struct {
	HookEventName            string `json:"hookEventName"`            // "PreToolUse"
	PermissionDecision       string `json:"permissionDecision"`       // "allow", "deny"
	PermissionDecisionReason string `json:"permissionDecisionReason"` // Reason for the decision
}

// PreToolUseHookOutput represents the output format for PreToolUse events.
type PreToolUseHookOutput struct {
	HookSpecificOutput *PreToolUseSpecificOutput `json:"hookSpecificOutput"`
}

// HookOutput represents the output to stdout.
// Deprecated: Use StopHookOutput instead. This type alias is kept for backward compatibility.
// Before PreToolUse support was added, this was the only output type.
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

// OutputPreToolUseDecision outputs the decision for a PreToolUse event.
//
// Parameters:
//   - log: The logger to use
//   - allow: true to allow the tool call, false to deny it
//   - feedback: Feedback message explaining the decision
//
// The function:
// 1. Outputs JSON to stdout for Claude Code to parse
// 2. Logs the decision
func OutputPreToolUseDecision(log *slog.Logger, allow bool, feedback string) error {
	feedback = strings.TrimSpace(feedback)

	decision := "allow"
	if !allow {
		decision = "deny"
		if feedback == "" {
			feedback = "Please complete the task before asking questions"
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

	// Log the decision
	if allow {
		log.Info("supervisor output: allow tool call", "feedback", feedback)
	} else {
		log.Info("supervisor output: deny tool call", "feedback", feedback)
	}

	// Output JSON to stdout
	fmt.Println(string(outputJSON))

	return nil
}
