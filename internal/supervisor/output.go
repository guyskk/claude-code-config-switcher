// Package supervisor provides supervisor output functionality.
package supervisor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/guyskk/ccc/internal/logger"
)

// HookOutput represents the output to stdout.
// When Decision is empty, the decision field is omitted from JSON to allow stop.
type HookOutput struct {
	Decision string `json:"decision,omitempty"` // "block" or omitted (allows stop)
	Reason   string `json:"reason,omitempty"`
}

// OutputResult represents the supervisor result for output.
type OutputResult struct {
	AllowStop bool
	Feedback  string
}

// OutputDecision outputs the supervisor's decision.
//
// Parameters:
//   - log: The logger to use
//   - allowStop: true to allow the agent to stop, false to block and require more work
//   - feedback: Optional feedback message (used when allowStop=false)
//
// The function:
// 1. Outputs JSON to stdout for Claude Code to parse
// 2. Logs the decision
func OutputDecision(log logger.Logger, allowStop bool, feedback string) error {
	// Trim feedback
	feedback = strings.TrimSpace(feedback)

	// Build output
	output := HookOutput{}
	if !allowStop {
		output.Decision = "block"
		output.Reason = feedback
		if feedback == "" {
			output.Reason = "Please continue completing the task"
		}
	}

	outputJSON, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal hook output: %w", err)
	}

	// Output JSON to stdout (for Claude Code to parse)
	fmt.Println(string(outputJSON))

	// Log the decision
	if allowStop {
		log.Info("supervisor output: allow stop")
	} else {
		log.Info("supervisor output: not allow stop",
			logger.StringField("feedback", output.Reason),
		)
	}

	return nil
}

// OutputSupervisorResult outputs the supervisor result based on AllowStop.
func OutputSupervisorResult(log logger.Logger, result *OutputResult) error {
	// Build user message for stderr
	var userMessage string
	if result == nil {
		userMessage = fmt.Sprintf("\n%s\n[RESULT] No supervisor result found, allowing stop\n",
			strings.Repeat("=", 60))
		fmt.Fprintf(os.Stderr, "%s", userMessage)
		return OutputDecision(log, true, "")
	}

	if result.AllowStop {
		userMessage = fmt.Sprintf("\n%s\n[RESULT] Work satisfactory, allowing stop\n",
			strings.Repeat("=", 60))
		fmt.Fprintf(os.Stderr, "%s", userMessage)
		return OutputDecision(log, true, "")
	}

	// Block with feedback
	feedback := strings.TrimSpace(result.Feedback)
	if feedback == "" {
		feedback = "Please continue completing the task"
	}
	userMessage = fmt.Sprintf("\n%s\n[RESULT] Work not satisfactory\nFeedback: %s\nAgent will continue working based on feedback\n%s\n\n",
		strings.Repeat("=", 60),
		feedback,
		strings.Repeat("=", 60))
	fmt.Fprintf(os.Stderr, "%s", userMessage)

	return OutputDecision(log, false, result.Feedback)
}

// OutputMaxIterationsReached outputs a message when max iterations is reached.
func OutputMaxIterationsReached(log logger.Logger, count, maxIterations int) error {
	log.Warn("max iterations reached, allowing stop",
		logger.IntField("count", count),
		logger.IntField("max", maxIterations),
	)

	userMessage := fmt.Sprintf("\n%s\n[STOP] Max iterations (%d) reached, allowing stop\n%s\n\n",
		strings.Repeat("=", 60),
		count,
		strings.Repeat("=", 60))
	fmt.Fprintf(os.Stderr, "%s", userMessage)

	return OutputDecision(log, true, "")
}

// OutputIterationCount outputs the current iteration count.
func OutputIterationCount(log logger.Logger, count, maxIterations int) {
	log.Info("iteration count",
		logger.IntField("count", count),
		logger.IntField("max", maxIterations),
	)

	fmt.Fprintf(os.Stderr, "Iteration count: %d/%d\n", count, maxIterations)
}

// OutputSupervisorStart outputs a message when supervisor review starts.
func OutputSupervisorStart(log logger.Logger, logFilePath string) {
	log.Info("starting supervisor review")

	fmt.Fprintf(os.Stderr, "\n[SUPERVISOR] Reviewing work...\n")
	fmt.Fprintf(os.Stderr, "See log file for details: %s\n\n", logFilePath)
}

// OutputSupervisorCompleted outputs a message when supervisor review completes.
func OutputSupervisorCompleted(log logger.Logger) {
	log.Info("supervisor review completed")

	fmt.Fprintf(os.Stderr, "\n%s\n", strings.Repeat("=", 60))
}
