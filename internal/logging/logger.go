package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/paths"
)

var (
	logger  *log.Logger
	logFile *os.File
)

// Init initializes logging to the kyaraben log file.
// Call Close() when done to flush and close the file.
func Init() error {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return fmt.Errorf("getting state directory: %w", err)
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	logPath := filepath.Join(stateDir, "kyaraben.log")
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)
	return nil
}

// Close closes the log file.
func Close() {
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// LogPath returns the path to the log file.
func LogPath() (string, error) {
	stateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "kyaraben.log"), nil
}

// Logger provides component-scoped logging.
// Create one per package or component using New().
type Logger struct {
	component string
}

// New creates a logger for a specific component.
// The component name appears in log entries to identify the source.
func New(component string) *Logger {
	return &Logger{component: component}
}

func (l *Logger) Info(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] [%s] "+format, append([]interface{}{l.component}, args...)...)
	}
}

func (l *Logger) Error(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] [%s] "+format, append([]interface{}{l.component}, args...)...)
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[DEBUG] [%s] "+format, append([]interface{}{l.component}, args...)...)
	}
}

// Info logs an informational message without component context.
// Prefer using a Logger instance from New() for better traceability.
func Info(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[INFO] "+format, args...)
	}
}

// Error logs an error message without component context.
// Prefer using a Logger instance from New() for better traceability.
func Error(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, args...)
	}
}

// Debug logs a debug message without component context.
// Prefer using a Logger instance from New() for better traceability.
func Debug(format string, args ...interface{}) {
	if logger != nil {
		logger.Printf("[DEBUG] "+format, args...)
	}
}

// Writer returns an io.Writer that writes to the log.
func Writer() io.Writer {
	if logFile != nil {
		return logFile
	}
	return io.Discard
}
