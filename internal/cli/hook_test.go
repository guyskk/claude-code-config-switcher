package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/guyskk/ccc/internal/config"
	"github.com/guyskk/ccc/internal/supervisor"
)

// ============================================================================
// Input Types Tests
// ============================================================================

func TestHookInput_Unmarshal_Stop(t *testing.T) {
	jsonInput := `{"session_id": "test-stop-123", "stop_hook_active": true}`

	var input HookInput
	if err := json.Unmarshal([]byte(jsonInput), &input); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if input.SessionID != "test-stop-123" {
		t.Errorf("SessionID = %q, want 'test-stop-123'", input.SessionID)
	}
	if input.StopHookActive != true {
		t.Errorf("StopHookActive = %v, want true", input.StopHookActive)
	}
}

func TestHookInput_Unmarshal_PreToolUse(t *testing.T) {
	jsonInput := `{"session_id": "test-pretool-123", "hook_event_name": "PreToolUse", "tool_name": "AskUserQuestion", "tool_input": {"questions": [{"question": "选择方案?", "header": "方案"}]}, "tool_use_id": "toolu_abc123", "transcript_path": "/path/to/transcript.json", "cwd": "/workspace"}`

	var input HookInput
	if err := json.Unmarshal([]byte(jsonInput), &input); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if input.SessionID != "test-pretool-123" {
		t.Errorf("SessionID = %q, want 'test-pretool-123'", input.SessionID)
	}
	if input.HookEventName != "PreToolUse" {
		t.Errorf("HookEventName = %q, want 'PreToolUse'", input.HookEventName)
	}
	if input.ToolName != "AskUserQuestion" {
		t.Errorf("ToolName = %q, want 'AskUserQuestion'", input.ToolName)
	}
	if input.ToolUseID != "toolu_abc123" {
		t.Errorf("ToolUseID = %q, want 'toolu_abc123'", input.ToolUseID)
	}
	if input.TranscriptPath != "/path/to/transcript.json" {
		t.Errorf("TranscriptPath = %q, want '/path/to/transcript.json'", input.TranscriptPath)
	}
	if input.CWD != "/workspace" {
		t.Errorf("CWD = %q, want '/workspace'", input.CWD)
	}
	if len(input.ToolInput) == 0 {
		t.Error("ToolInput should not be empty")
	}
}

func TestHookInput_Unmarshal_StopNoEventName(t *testing.T) {
	jsonInput := `{"session_id": "test-stop-789", "stop_hook_active": false}`

	var input HookInput
	if err := json.Unmarshal([]byte(jsonInput), &input); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if input.SessionID != "test-stop-789" {
		t.Errorf("SessionID = %q, want 'test-stop-789'", input.SessionID)
	}
	if input.HookEventName != "" {
		t.Errorf("HookEventName = %q, want '' (empty)", input.HookEventName)
	}
}

// ============================================================================
// parseResultJSON 测试
// ============================================================================

func TestParseResultJSON(t *testing.T) {
	tests := []struct {
		name         string
		jsonText     string
		wantAllow    bool
		wantFeedback string
	}{
		{
			name:         "valid json - allow stop true",
			jsonText:     `{"allow_stop": true, "feedback": "work is complete"}`,
			wantAllow:    true,
			wantFeedback: "work is complete",
		},
		{
			name:         "valid json - allow stop false",
			jsonText:     `{"allow_stop": false, "feedback": "needs more work"}`,
			wantAllow:    false,
			wantFeedback: "needs more work",
		},
		{
			// llmparser can repair JSON with trailing commas
			name:         "malformed json with trailing comma (repaired by llmparser)",
			jsonText:     `{"allow_stop": true, "feedback": "test",}`,
			wantAllow:    true,
			wantFeedback: "test",
		},
		{
			// llmparser can extract JSON from markdown code blocks
			name:         "json in markdown code block",
			jsonText:     "Some text\n```json\n{\"allow_stop\": false, \"feedback\": \"keep going\"}\n```\nMore text",
			wantAllow:    false,
			wantFeedback: "keep going",
		},
		{
			// Fallback: missing required field - use original text as feedback
			name:         "missing required feedback field - fallback",
			jsonText:     `{"allow_stop": true}`,
			wantAllow:    false,
			wantFeedback: `{"allow_stop": true}`,
		},
		{
			// Fallback: missing required allow_stop field - use original text as feedback
			name:         "missing required allow_stop field - fallback",
			jsonText:     `{"feedback": "some feedback"}`,
			wantAllow:    false,
			wantFeedback: `{"feedback": "some feedback"}`,
		},
		{
			// Fallback: empty string - use default feedback
			name:         "empty string - fallback with default",
			jsonText:     "",
			wantAllow:    false,
			wantFeedback: "Please continue completing the task",
		},
		{
			// Fallback: not json - use original text as feedback
			name:         "not json - fallback",
			jsonText:     "just plain text",
			wantAllow:    false,
			wantFeedback: "just plain text",
		},
		{
			// Fallback: invalid JSON-like content
			name:         "invalid json - fallback",
			jsonText:     "{broken json",
			wantAllow:    false,
			wantFeedback: "{broken json",
		},
		{
			// Fallback: whitespace only - use default feedback
			name:         "whitespace only - fallback with default",
			jsonText:     "   \n\t  ",
			wantAllow:    false,
			wantFeedback: "Please continue completing the task",
		},
		{
			// Fallback: Chinese text feedback
			name:         "chinese feedback - fallback",
			jsonText:     "任务还没有完成，请继续",
			wantAllow:    false,
			wantFeedback: "任务还没有完成，请继续",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseResultJSON(tt.jsonText)
			if result == nil {
				t.Fatal("parseResultJSON() returned nil result")
			}
			if result.AllowStop != tt.wantAllow {
				t.Errorf("parseResultJSON() allow_stop = %v, want %v", result.AllowStop, tt.wantAllow)
			}
			if result.Feedback != tt.wantFeedback {
				t.Errorf("parseResultJSON() feedback = %q, want %q", result.Feedback, tt.wantFeedback)
			}
		})
	}
}

// ============================================================================
// getDefaultSupervisorPrompt 测试
// ============================================================================

func TestGetDefaultSupervisorPrompt(t *testing.T) {
	// Save original GetDirFunc to restore after test
	originalGetDirFunc := config.GetDirFunc
	defer func() { config.GetDirFunc = originalGetDirFunc }()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	config.GetDirFunc = func() string { return tempDir }

	t.Run("default prompt when no custom file", func(t *testing.T) {
		prompt, source := getDefaultSupervisorPrompt()
		if prompt == "" {
			t.Error("getDefaultSupervisorPrompt() returned empty string")
		}
		if source != "supervisor_prompt_default" {
			t.Errorf("getDefaultSupervisorPrompt() source = %q, want %q", source, "supervisor_prompt_default")
		}
		// Check that key parts are present (prompt is in Chinese)
		if !strings.Contains(prompt, "监督者") && !strings.Contains(prompt, "审查者") && !strings.Contains(prompt, "Supervisor") {
			t.Error("getDefaultSupervisorPrompt() missing role keyword")
		}
		if !strings.Contains(prompt, "allow_stop") {
			t.Error("getDefaultSupervisorPrompt() missing 'allow_stop'")
		}
		if !strings.Contains(prompt, "feedback") {
			t.Error("getDefaultSupervisorPrompt() missing 'feedback'")
		}
	})

	t.Run("custom prompt from SUPERVISOR.md", func(t *testing.T) {
		customPrompt := "Custom supervisor prompt for testing"
		customPath := filepath.Join(tempDir, "SUPERVISOR.md")
		if err := os.WriteFile(customPath, []byte(customPrompt), 0644); err != nil {
			t.Fatalf("failed to write custom prompt file: %v", err)
		}

		prompt, source := getDefaultSupervisorPrompt()
		if prompt != customPrompt {
			t.Errorf("getDefaultSupervisorPrompt() = %q, want %q", prompt, customPrompt)
		}
		if source != customPath {
			t.Errorf("getDefaultSupervisorPrompt() source = %q, want %q", source, customPath)
		}
	})

	t.Run("empty custom file falls back to default", func(t *testing.T) {
		customPath := filepath.Join(tempDir, "SUPERVISOR.md")
		if err := os.WriteFile(customPath, []byte("   \n\t  "), 0644); err != nil {
			t.Fatalf("failed to write empty custom prompt file: %v", err)
		}

		prompt, source := getDefaultSupervisorPrompt()
		if prompt == "" {
			t.Error("getDefaultSupervisorPrompt() returned empty string for empty custom file")
		}
		if source != "supervisor_prompt_default" {
			t.Errorf("getDefaultSupervisorPrompt() source = %q, want %q", source, "supervisor_prompt_default")
		}
	})
}

// ============================================================================
// supervisorResultSchema 测试
// ============================================================================

func TestSupervisorResultSchema(t *testing.T) {
	if supervisorResultSchema == nil {
		t.Fatal("supervisorResultSchema is nil")
	}

	schemaMap := supervisorResultSchema

	if schemaMap["type"] != "object" {
		t.Errorf("schema type = %v, want 'object'", schemaMap["type"])
	}

	properties, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("schema properties is not a map")
	}

	required, ok := schemaMap["required"].([]string)
	if !ok {
		t.Fatal("schema required is not a string slice")
	}
	if len(required) != 2 {
		t.Errorf("required fields count = %d, want 2", len(required))
	}

	if _, ok := properties["allow_stop"]; !ok {
		t.Error("schema missing 'allow_stop' property")
	}
	if _, ok := properties["feedback"]; !ok {
		t.Error("schema missing 'feedback' property")
	}
}

// ============================================================================
// 输入结构体测试
// ============================================================================

func TestRunSupervisorHook_StopEvent(t *testing.T) {
	originalGetDirFunc := config.GetDirFunc
	defer func() { config.GetDirFunc = originalGetDirFunc }()

	tempDir := t.TempDir()
	config.GetDirFunc = func() string { return tempDir }

	state := &supervisor.State{Enabled: false}
	if err := supervisor.SaveState("test-stop-event", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	oldSupervisorID := os.Getenv("CCC_SUPERVISOR_ID")
	oldSupervisorHook := os.Getenv("CCC_SUPERVISOR_HOOK")
	defer func() {
		os.Setenv("CCC_SUPERVISOR_ID", oldSupervisorID)
		os.Setenv("CCC_SUPERVISOR_HOOK", oldSupervisorHook)
	}()
	os.Setenv("CCC_SUPERVISOR_ID", "test-stop-event")
	os.Setenv("CCC_SUPERVISOR_HOOK", "")

	hookInputJSON := `{"session_id": "test-stop-event", "stop_hook_active": false}`

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte(hookInputJSON))
		w.Close()
	}()
	defer func() { os.Stdin = oldStdin }()

	opts := &SupervisorHookCommand{}
	err := RunSupervisorHook(opts)

	if err != nil {
		t.Errorf("RunSupervisorHook() error = %v, want nil (stop event)", err)
	}
}

func TestRunSupervisorHook_PreToolUseEvent(t *testing.T) {
	originalGetDirFunc := config.GetDirFunc
	defer func() { config.GetDirFunc = originalGetDirFunc }()

	tempDir := t.TempDir()
	config.GetDirFunc = func() string { return tempDir }

	state := &supervisor.State{Enabled: false}
	if err := supervisor.SaveState("test-pretool-event", state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	oldSupervisorID := os.Getenv("CCC_SUPERVISOR_ID")
	oldSupervisorHook := os.Getenv("CCC_SUPERVISOR_HOOK")
	defer func() {
		os.Setenv("CCC_SUPERVISOR_ID", oldSupervisorID)
		os.Setenv("CCC_SUPERVISOR_HOOK", oldSupervisorHook)
	}()
	os.Setenv("CCC_SUPERVISOR_ID", "test-pretool-event")
	os.Setenv("CCC_SUPERVISOR_HOOK", "")

	hookInputJSON := `{"session_id": "test-pretool-event", "hook_event_name": "PreToolUse", "tool_name": "AskUserQuestion"}`

	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r
	go func() {
		w.Write([]byte(hookInputJSON))
		w.Close()
	}()
	defer func() { os.Stdin = oldStdin }()

	opts := &SupervisorHookCommand{}
	err := RunSupervisorHook(opts)

	if err != nil {
		t.Errorf("RunSupervisorHook() error = %v, want nil (pretool event)", err)
	}
}
