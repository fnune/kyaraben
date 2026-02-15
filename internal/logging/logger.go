package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fnune/kyaraben/internal/paths"
)

var (
	logger     *log.Logger
	logFile    *os.File
	outputHook func(string)
	hookMu     sync.RWMutex
)

// SetOutputHook sets a callback that receives formatted log lines.
// This is used to forward logs to the UI during operations.
// Pass nil to disable the hook.
func SetOutputHook(fn func(string)) {
	hookMu.Lock()
	outputHook = fn
	hookMu.Unlock()
}

func callHook(content string) {
	hookMu.RLock()
	hook := outputHook
	hookMu.RUnlock()
	if hook != nil {
		hook(content)
	}
}

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
	content := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[INFO] [%s] %s", l.component, content)
	}
	callHook(content)
}

func (l *Logger) Error(format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[ERROR] [%s] %s", l.component, content)
	}
	callHook(content)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[DEBUG] [%s] %s", l.component, content)
	}
	callHook(content)
}

// Info logs an informational message without component context.
// Prefer using a Logger instance from New() for better traceability.
func Info(format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[INFO] %s", content)
	}
	callHook(content)
}

// Error logs an error message without component context.
// Prefer using a Logger instance from New() for better traceability.
func Error(format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[ERROR] %s", content)
	}
	callHook(content)
}

// Debug logs a debug message without component context.
// Prefer using a Logger instance from New() for better traceability.
func Debug(format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)
	if logger != nil {
		logger.Printf("[DEBUG] %s", content)
	}
	callHook(content)
}

func Writer() io.Writer {
	if logFile != nil {
		return logFile
	}
	return io.Discard
}

// CurrentPosition returns the current byte position in the log file.
// This can be used to tail from a specific point.
func CurrentPosition() int64 {
	if logFile == nil {
		return 0
	}
	_ = logFile.Sync()
	info, err := logFile.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}
