// Package supervisor provides supervisor-specific logging functionality.
package supervisor

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSupervisorLogger_Output verifies that logs are written to both stderr and file.
func TestSupervisorLogger_Output(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a mock GetStateDir function
	origGetStateDir := getStateDirFunc
	getStateDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getStateDirFunc = origGetStateDir }()

	supervisorID := "test-session"
	log := NewSupervisorLogger(supervisorID)

	// Test logging
	log.Info("test message", "key1", "value1", "key2", 42)
	log.Debug("debug message", "debug_key", "debug_value")
	log.Warn("warning message", "warn_key", "warn_value")
	log.Error("error message", "error_key", "error_value")

	// Close the logger to flush and close the file
	if handler, ok := log.Handler().(*SupervisorLogger); ok {
		if err := handler.Close(); err != nil {
			t.Fatalf("failed to close logger: %v", err)
		}
	}

	// Verify log file was created
	logFilePath := filepath.Join(tempDir, "supervisor-"+supervisorID+".log")
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)

	// Verify all messages are in the log
	if !strings.Contains(logContent, "test message") {
		t.Error("log file missing 'test message'")
	}
	if !strings.Contains(logContent, "key1=value1") {
		t.Error("log file missing 'key1=value1'")
	}
	if !strings.Contains(logContent, "key2=42") {
		t.Error("log file missing 'key2=42'")
	}
	if !strings.Contains(logContent, "debug message") {
		t.Error("log file missing 'debug message'")
	}
	if !strings.Contains(logContent, "warning message") {
		t.Error("log file missing 'warning message'")
	}
	if !strings.Contains(logContent, "error message") {
		t.Error("log file missing 'error message'")
	}

	// Verify supervisor_id is in the log
	if !strings.Contains(logContent, "supervisor_id="+supervisorID) {
		t.Errorf("log file missing supervisor_id=%s", supervisorID)
	}

	// Verify log file contains proper number of lines (each log should have one line)
	lines := strings.Split(strings.TrimSpace(logContent), "\n")
	if len(lines) < 4 {
		t.Errorf("expected at least 4 log lines, got %d", len(lines))
	}
}

// TestSupervisorLogger_NoSupervisorID verifies that logs only go to stderr when supervisorID is empty.
func TestSupervisorLogger_NoSupervisorID(t *testing.T) {
	// Note: We can't easily redirect stderr in a test without affecting the whole process,
	// so we just verify that no file is created

	tempDir := t.TempDir()
	origGetStateDir := getStateDirFunc
	getStateDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getStateDirFunc = origGetStateDir }()

	// Create logger without supervisorID
	log := NewSupervisorLogger("")

	// Test logging - should not create any file
	log.Info("test message")

	// Verify no log file was created
	logFilePath := filepath.Join(tempDir, "supervisor-.log")
	if _, err := os.Stat(logFilePath); !os.IsNotExist(err) {
		t.Error("log file should not exist when supervisorID is empty")
	}
}

// TestSupervisorLogger_Close verifies that Close() properly closes the file.
func TestSupervisorLogger_Close(t *testing.T) {
	tempDir := t.TempDir()

	origGetStateDir := getStateDirFunc
	getStateDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getStateDirFunc = origGetStateDir }()

	supervisorID := "test-close"
	log := NewSupervisorLogger(supervisorID)

	handler, ok := log.Handler().(*SupervisorLogger)
	if !ok {
		t.Fatal("expected SupervisorLogger handler")
	}

	// Log some messages
	log.Info("before close")

	// Close the logger
	if err := handler.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	// Verify file was closed by checking logFilePath
	logFilePath := filepath.Join(tempDir, "supervisor-"+supervisorID+".log")
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "before close") {
		t.Error("log file missing 'before close' message")
	}

	// Close should be idempotent
	if err := handler.Close(); err != nil {
		t.Errorf("close should be idempotent: %v", err)
	}
}

// TestSupervisorLogger_WithAttrs verifies that WithAttrs works correctly.
func TestSupervisorLogger_WithAttrs(t *testing.T) {
	tempDir := t.TempDir()

	origGetStateDir := getStateDirFunc
	getStateDirFunc = func() (string, error) {
		return tempDir, nil
	}
	defer func() { getStateDirFunc = origGetStateDir }()

	supervisorID := "test-attrs"
	log := NewSupervisorLogger(supervisorID)

	// Add additional attributes
	logWithAttrs := log.With("extra_key", "extra_value")

	logWithAttrs.Info("test with attrs")

	// Close the logger
	if handler, ok := log.Handler().(*SupervisorLogger); ok {
		handler.Close()
	}

	// Verify log file
	logFilePath := filepath.Join(tempDir, "supervisor-"+supervisorID+".log")
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "extra_key=extra_value") {
		t.Error("log file missing extra attribute from With()")
	}
}

// TestOutputDecision_JSONFormat verifies that OutputDecision produces correct JSON.
func TestOutputDecision_JSONFormat(t *testing.T) {
	tests := []struct {
		name       string
		allowStop  bool
		feedback   string
		wantJSON   string
		wantReason string
	}{
		{
			name:      "allow stop - empty JSON",
			allowStop: true,
			feedback:  "",
			wantJSON:  "{}\n",
		},
		{
			name:      "allow stop - true with feedback",
			allowStop: true,
			feedback:  "some feedback",
			wantJSON:  "{}\n",
		},
		{
			name:       "block stop - with feedback",
			allowStop:  false,
			feedback:   "needs more work",
			wantJSON:   `{"decision":"block","reason":"needs more work"}` + "\n",
			wantReason: "needs more work",
		},
		{
			name:       "block stop - empty feedback uses default",
			allowStop:  false,
			feedback:   "",
			wantJSON:   `{"decision":"block","reason":"Please continue completing the task"}` + "\n",
			wantReason: "Please continue completing the task",
		},
		{
			name:       "block stop - whitespace feedback uses default",
			allowStop:  false,
			feedback:   "   ",
			wantJSON:   `{"decision":"block","reason":"Please continue completing the task"}` + "\n",
			wantReason: "Please continue completing the task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			var stdoutBuf bytes.Buffer

			// Redirect stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() {
				os.Stdout = oldStdout
				r.Close()
				w.Close()
			}()

			// Create logger that writes to stderr only (we don't care about logs in this test)
			log := slog.New(slog.NewTextHandler(os.Stderr, nil))

			// Call OutputDecision
			err := OutputDecision(log, tt.allowStop, tt.feedback)
			if err != nil {
				t.Fatalf("OutputDecision failed: %v", err)
			}

			// Close the write end
			w.Close()

			// Read from the pipe
			os.Stdout = oldStdout
			stdoutBuf.ReadFrom(r)

			// Verify JSON output
			gotJSON := stdoutBuf.String()
			if gotJSON != tt.wantJSON {
				t.Errorf("JSON output mismatch\nGot:      %q\nExpected: %q", gotJSON, tt.wantJSON)
			}

			// Parse JSON to verify it's valid
			var parsed map[string]interface{}
			if err := json.Unmarshal(stdoutBuf.Bytes(), &parsed); err != nil {
				t.Errorf("invalid JSON output: %v", err)
			}

			// Verify decision field
			if tt.allowStop {
				if _, exists := parsed["decision"]; exists {
					t.Error("decision field should be omitted when allowStop=true")
				}
			} else {
				if decision, exists := parsed["decision"]; !exists || decision != "block" {
					t.Errorf("decision should be 'block', got %v", decision)
				}
				if reason, exists := parsed["reason"]; !exists || reason != tt.wantReason {
					t.Errorf("reason should be %q, got %v", tt.wantReason, reason)
				}
			}
		})
	}
}
