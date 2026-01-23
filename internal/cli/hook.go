// Package cli implements the supervisor-hook subcommand.
package cli

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/guyskk/ccc/internal/config"
	"github.com/guyskk/ccc/internal/llmparser"
	"github.com/guyskk/ccc/internal/prettyjson"
	"github.com/guyskk/ccc/internal/supervisor"
	"github.com/schlunsen/claude-agent-sdk-go"
	"github.com/schlunsen/claude-agent-sdk-go/types"
)

//go:embed supervisor_prompt_default.md
var defaultPromptContent []byte

// ============================================================================
// Input Types
// ============================================================================

// HookInputHeader is used for event type detection, parsing only necessary fields.
type HookInputHeader struct {
	SessionID     string `json:"session_id"`
	HookEventName string `json:"hook_event_name,omitempty"`
}

// StopHookInput represents the input from Stop event (end-of-task review).
type StopHookInput struct {
	SessionID      string `json:"session_id"`
	StopHookActive bool   `json:"stop_hook_active"`
}

// PreToolUseInput represents the input from PreToolUse event (pre-tool-call review).
// The ToolInput field is kept as RawMessage for flexible parsing of different tool inputs.
type PreToolUseInput struct {
	SessionID      string          `json:"session_id"`
	HookEventName  string          `json:"hook_event_name"`
	ToolName       string          `json:"tool_name"`
	ToolInput      json.RawMessage `json:"tool_input,omitempty"`
	ToolUseID      string          `json:"tool_use_id,omitempty"`
	TranscriptPath string          `json:"transcript_path,omitempty"`
	CWD            string          `json:"cwd,omitempty"`
}

// ============================================================================
// Supervisor Result
// ============================================================================

// SupervisorResult represents the parsed output from Supervisor.
type SupervisorResult struct {
	AllowStop bool   `json:"allow_stop"` // Whether to allow the Agent to stop
	Feedback  string `json:"feedback"`   // Feedback when AllowStop is false
}

// supervisorResultSchema is the JSON schema for parsing supervisor output.
var supervisorResultSchema = map[string]interface{}{
	"type": "object",
	"properties": map[string]interface{}{
		"allow_stop": map[string]interface{}{
			"type":        "boolean",
			"description": "Whether to allow the Agent to stop working (true = work is satisfactory, false = needs more work)",
		},
		"feedback": map[string]interface{}{
			"type":        "string",
			"description": "Specific feedback and guidance for continuing work when allow_stop is false",
		},
	},
	"required": []string{"allow_stop", "feedback"},
}

// ============================================================================
// Utility Functions
// ============================================================================

func logCurrentEnv(log *slog.Logger) {
	lines := []string{}
	prefixes := []string{"CLAUDE_", "ANTHROPIC_", "CCC_"}
	for _, env := range os.Environ() {
		for _, prefix := range prefixes {
			if strings.HasPrefix(env, prefix) {
				lines = append(lines, env)
				break
			}
		}
	}
	envStr := strings.Join(lines, "\n")
	log.Debug(fmt.Sprintf("supervisor hook environment:\n%s", envStr))
}

// detectEventType reads raw input from stdin and detects the event type.
// Returns event type, raw JSON data, and sessionID.
func detectEventType(stdin io.Reader) (supervisor.HookEventType, []byte, string, error) {
	rawInput, err := io.ReadAll(stdin)
	if err != nil {
		return "", nil, "", fmt.Errorf("failed to read stdin: %w", err)
	}

	var header HookInputHeader
	if err := json.Unmarshal(rawInput, &header); err != nil {
		return "", nil, "", fmt.Errorf("failed to parse hook input header: %w", err)
	}

	// Validate and normalize event type
	eventType := supervisor.HookEventType(header.HookEventName)
	// Default unknown event types to Stop for backward compatibility
	if eventType != supervisor.EventTypeStop && eventType != supervisor.EventTypePreToolUse {
		eventType = supervisor.EventTypeStop
	}

	// Ensure session ID is present
	if header.SessionID == "" {
		return "", nil, "", fmt.Errorf("session_id is required in hook input")
	}

	return eventType, rawInput, header.SessionID, nil
}

// ============================================================================
// Main Entry Point
// ============================================================================

// RunSupervisorHook executes the supervisor-hook subcommand.
func RunSupervisorHook(opts *SupervisorHookCommand) error {
	// Validate supervisorID
	supervisorID := os.Getenv("CCC_SUPERVISOR_ID")
	if supervisorID == "" {
		return fmt.Errorf("CCC_SUPERVISOR_ID is required from env var")
	}

	// Create logger
	log := supervisor.NewSupervisorLogger(supervisorID)
	logCurrentEnv(log)

	// Prevent recursive calls
	if os.Getenv("CCC_SUPERVISOR_HOOK") == "1" {
		return supervisor.OutputDecision(log, true, "called from supervisor hook")
	}

	// Load state to check if supervisor mode is enabled
	state, err := supervisor.LoadState(supervisorID)
	if err != nil {
		return fmt.Errorf("failed to load supervisor state: %w", err)
	}
	if !state.Enabled {
		log.Debug("supervisor mode disabled, allowing stop", "enabled", state.Enabled)
		return supervisor.OutputDecision(log, true, "supervisor mode disabled")
	}

	// Load supervisor configuration
	supervisorCfg, err := config.LoadSupervisorConfig()
	if err != nil {
		return fmt.Errorf("failed to load supervisor config: %w", err)
	}

	// Get sessionID and event type
	var sessionID string
	var eventType supervisor.HookEventType
	var rawInput []byte

	if opts != nil && opts.SessionId != "" {
		// Use sessionID from command line argument, default to Stop event
		sessionID = opts.SessionId
		eventType = supervisor.EventTypeStop
		log.Debug("using session_id from command line argument", "session_id", sessionID)
	} else {
		// Read from stdin and detect event type
		eventType, rawInput, sessionID, err = detectEventType(os.Stdin)
		if err != nil {
			return err
		}
		log.Debug("hook input", "event_type", eventType, "raw_input", string(rawInput))
	}

	// Validate sessionID
	if sessionID == "" {
		return fmt.Errorf("session_id is required (either from --session-id argument or stdin)")
	}

	// Check iteration count limit
	maxIterations := supervisorCfg.MaxIterations
	shouldContinue, count, err := supervisor.ShouldContinue(sessionID, maxIterations)
	if err != nil {
		log.Warn("failed to check supervisor state", "error", err.Error())
	}
	if !shouldContinue {
		log.Info("max iterations reached, allowing operation",
			"count", count,
			"max", maxIterations,
		)
		// When max iterations reached, allow based on event type
		return outputDecisionByEventType(log, eventType, true,
			fmt.Sprintf("max iterations (%d/%d) reached", count, maxIterations))
	}

	// Increment iteration count
	newCount, err := supervisor.IncrementCount(sessionID)
	if err != nil {
		log.Warn("failed to increment count", "error", err.Error())
	} else {
		log.Info("iteration count", "count", newCount, "max", maxIterations)
	}

	// Run supervisor review
	result, err := runSupervisorReview(sessionID, supervisorCfg, log)
	if err != nil {
		return err
	}

	// Output result
	if result == nil {
		log.Info("no supervisor result found, allowing operation")
		return outputDecisionByEventType(log, eventType, true, "no supervisor result found")
	}

	return outputDecisionByEventType(log, eventType, result.AllowStop, result.Feedback)
}

// outputDecisionByEventType outputs decision in the appropriate format based on event type.
func outputDecisionByEventType(log *slog.Logger, eventType supervisor.HookEventType, allow bool, feedback string) error {
	switch eventType {
	case supervisor.EventTypePreToolUse:
		return supervisor.OutputPreToolUseDecision(log, allow, feedback)
	case supervisor.EventTypeStop:
		fallthrough
	default:
		// Unknown event types default to Stop format for backward compatibility
		return supervisor.OutputDecision(log, allow, feedback)
	}
}

// runSupervisorReview executes the supervisor review process.
func runSupervisorReview(sessionID string, cfg *config.SupervisorConfig, log *slog.Logger) (*SupervisorResult, error) {
	// Load supervisor prompt
	supervisorPrompt, promptSource := getDefaultSupervisorPrompt()
	log.Debug("supervisor prompt loaded", "source", promptSource, "length", len(supervisorPrompt))

	log.Info("starting supervisor review", "session_id", sessionID)

	// Run supervisor using Claude Agent SDK
	result, err := runSupervisorWithSDK(context.Background(), sessionID, supervisorPrompt, cfg.Timeout(), log)
	if err != nil {
		log.Error("supervisor SDK failed", "error", err.Error())
		return nil, fmt.Errorf("supervisor SDK failed: %w", err)
	}

	log.Info("supervisor review completed")
	return result, nil
}

// runSupervisorWithSDK runs the supervisor using the Claude Agent SDK.
// The supervisor prompt is sent as a USER message, and we parse the Result field for JSON output.
func runSupervisorWithSDK(ctx context.Context, sessionID, prompt string, timeout time.Duration, log *slog.Logger) (*SupervisorResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build options for SDK
	// - ForkSession: Create a fork to review the current session state
	// - Resume: Load the session context (includes system/user/project prompts from settings)
	// - SettingSources: Load system prompts from user, project, and local settings
	// NOTE: We do NOT use WithOutputFormat - we parse the Result field directly
	// [Bug] StructuredOutput tool doesn't stop agent execution - agent continues after calling it
	// https://github.com/anthropics/claude-code/issues/17125
	opts := types.NewClaudeAgentOptions().
		WithForkSession(true).                                                                            // Fork the current session
		WithResume(sessionID).                                                                            // Resume from specific session
		WithSettingSources(types.SettingSourceUser, types.SettingSourceProject, types.SettingSourceLocal) // Load all setting sources

	// Set environment variable to avoid infinite loop
	opts.Env["CCC_SUPERVISOR_HOOK"] = "1"

	log.Debug("SDK options",
		"fork_session", "true",
		"resume", sessionID,
	)

	// Create interactive client
	log.Debug("creating SDK client")
	client, err := claude.NewClient(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDK client: %w", err)
	}
	defer client.Close(ctx)

	// Connect to Claude
	log.Debug("connecting SDK client")
	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect SDK client: %w", err)
	}

	// Send supervisor prompt as USER message
	log.Debug("sending supervisor review request as user message")
	if err := client.Query(ctx, prompt); err != nil {
		return nil, fmt.Errorf("failed to send query: %w", err)
	}

	// Process messages and get ResultMessage
	log.Debug("receiving messages from SDK")

	var resultMessage *types.ResultMessage

	for msg := range client.ReceiveResponse(ctx) {
		// Log raw message JSON for debugging (this is the ONE place where all messages are logged)
		msgJSON, err := prettyjson.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message: %w", err)
		}
		log.Debug("raw message", "json", string(msgJSON))

		switch m := msg.(type) {
		case *types.ResultMessage:
			resultMessage = m
		}
	}

	// Extract and parse result from ResultMessage
	if resultMessage == nil {
		log.Error("no result message received from SDK")
		return nil, fmt.Errorf("no result message received from SDK")
	}

	if resultMessage.Result == nil {
		log.Error("result message has no Result field")
		return nil, fmt.Errorf("result message has no Result field")
	}

	// Parse JSON from Result field using llmparser
	resultText := *resultMessage.Result
	result := parseResultJSON(resultText)

	return result, nil
}

// parseResultJSON parses JSON text into a SupervisorResult.
// It uses the llmparser package for fault-tolerant JSON parsing.
// When parsing fails, it returns a fallback result with allow_stop=false
// and the original text as feedback, allowing the agent to continue working.
func parseResultJSON(jsonText string) *SupervisorResult {
	// Use llmparser for fault-tolerant JSON parsing with schema validation
	parsed, err := llmparser.Parse(jsonText, supervisorResultSchema)
	if err != nil {
		// Fallback: parsing failed, use original text as feedback
		// This allows the agent to continue working instead of failing
		fallbackText := strings.TrimSpace(jsonText)
		if fallbackText == "" {
			fallbackText = "Please continue completing the task"
		}
		return &SupervisorResult{
			AllowStop: false,
			Feedback:  fallbackText,
		}
	}

	// Convert parsed interface{} to SupervisorResult
	outputMap, ok := parsed.(map[string]interface{})
	if !ok {
		// Fallback: wrong type
		return &SupervisorResult{
			AllowStop: false,
			Feedback:  strings.TrimSpace(jsonText),
		}
	}

	result := &SupervisorResult{}

	// Extract allow_stop field (boolean)
	if allowStop, ok := outputMap["allow_stop"].(bool); ok {
		result.AllowStop = allowStop
	} else {
		// Fallback: missing or invalid allow_stop field
		return &SupervisorResult{
			AllowStop: false,
			Feedback:  strings.TrimSpace(jsonText),
		}
	}

	// Extract feedback field (string)
	if feedback, ok := outputMap["feedback"].(string); ok {
		result.Feedback = feedback
	} else {
		// Fallback: missing or invalid feedback field
		result.Feedback = strings.TrimSpace(jsonText)
	}

	return result
}

// getDefaultSupervisorPrompt returns the supervisor prompt and its source.
// It first tries to read from ~/.claude/SUPERVISOR.md (or CCC_CONFIG_DIR/SUPERVISOR.md).
// If the custom file exists and has content, it is used; otherwise, the default embedded prompt is returned.
// The source return value indicates where the prompt came from:
// - "supervisor_prompt_default" for the embedded default prompt
// - Full file path for a custom SUPERVISOR.md file
func getDefaultSupervisorPrompt() (string, string) {
	customPromptPath := config.GetDir() + "/SUPERVISOR.md"
	data, err := os.ReadFile(customPromptPath)
	if err == nil {
		prompt := strings.TrimSpace(string(data))
		if prompt != "" {
			return prompt, customPromptPath
		}
	}
	return string(defaultPromptContent), "supervisor_prompt_default"
}
