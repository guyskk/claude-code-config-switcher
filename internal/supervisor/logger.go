// Package supervisor provides supervisor-specific logging functionality.
package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

// SupervisorLogger is a handler that outputs to both stderr and a log file.
// If supervisorID is empty, it only outputs to stderr.
// If supervisorID is non-empty, it outputs to both stderr and a log file.
type SupervisorLogger struct {
	stderrHandler slog.Handler
	fileHandler   slog.Handler
	logFile       *os.File // Track the file for proper cleanup
	mu            sync.Mutex
	closed        bool
}

// NewSupervisorLogger creates a new slog.Logger with SupervisorHandler.
//
// If supervisorID is empty, only stderr output is enabled.
// If supervisorID is non-empty, both stderr and log file output are enabled.
// The log file is created at ~/.claude/ccc/supervisor-{supervisorID}.log
//
// Errors are logged to stderr and a fallback logger is returned.
func NewSupervisorLogger(supervisorID string) *slog.Logger {
	// Create stderr handler (always enabled)
	stderrHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	if supervisorID == "" {
		return slog.New(stderrHandler)
	}

	// Create log file for supervisor session
	stateDir, err := GetStateDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get state directory: %v\n", err)
		return slog.New(stderrHandler)
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create state directory: %v\n", err)
		return slog.New(stderrHandler)
	}

	logFilePath := filepath.Join(stateDir, fmt.Sprintf("supervisor-%s.log", supervisorID))
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open supervisor log file: %v\n", err)
		return slog.New(stderrHandler)
	}

	// Create file handler with debug level and supervisor_id attribute
	fileHandler := slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}).WithAttrs([]slog.Attr{slog.String("supervisor_id", supervisorID)})

	// Create combined handler
	handler := &SupervisorLogger{
		stderrHandler: stderrHandler.WithAttrs([]slog.Attr{slog.String("supervisor_id", supervisorID)}),
		fileHandler:   fileHandler,
		logFile:       logFile, // Track for cleanup
	}

	return slog.New(handler)
}

// Enabled reports whether l handles level.
func (l *SupervisorLogger) Enabled(ctx context.Context, level slog.Level) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stderrHandler.Enabled(ctx, level)
}

// Handle handles the Record by writing to both stderr and file.
func (l *SupervisorLogger) Handle(ctx context.Context, r slog.Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Always log to stderr
	l.stderrHandler.Handle(ctx, r)

	// Log to file if enabled and not closed
	if l.fileHandler != nil && !l.closed {
		l.fileHandler.Handle(ctx, r)
	}
	return nil
}

// WithAttrs returns a new Handler with the given attributes.
func (l *SupervisorLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	l.mu.Lock()
	defer l.mu.Unlock()

	return &SupervisorLogger{
		stderrHandler: l.stderrHandler.WithAttrs(attrs),
		fileHandler:   l.withFileHandlerAttrs(attrs),
		logFile:       l.logFile,
		closed:        l.closed,
	}
}

func (l *SupervisorLogger) withFileHandlerAttrs(attrs []slog.Attr) slog.Handler {
	if l.fileHandler == nil {
		return nil
	}
	return l.fileHandler.WithAttrs(attrs)
}

// WithGroup returns a new Handler with the given group name.
func (l *SupervisorLogger) WithGroup(name string) slog.Handler {
	l.mu.Lock()
	defer l.mu.Unlock()

	return &SupervisorLogger{
		stderrHandler: l.stderrHandler.WithGroup(name),
		fileHandler:   l.withFileHandlerGroup(name),
		logFile:       l.logFile,
		closed:        l.closed,
	}
}

func (l *SupervisorLogger) withFileHandlerGroup(name string) slog.Handler {
	if l.fileHandler == nil {
		return nil
	}
	return l.fileHandler.WithGroup(name)
}

// Close closes the log file if it was opened.
// This is safe to call multiple times.
func (l *SupervisorLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true

	// Close the file handle
	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
	}

	return nil
}
