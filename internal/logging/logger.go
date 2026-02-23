package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/fnune/kyaraben/internal/paths"
)

var (
	slogger    *slog.Logger
	logFile    *os.File
	uiCallback func(LogEntry)
	logMu      sync.RWMutex
	fileWriter io.Writer
)

func SetUICallback(fn func(LogEntry)) {
	logMu.Lock()
	defer logMu.Unlock()
	uiCallback = fn
	rebuildLoggerLocked()
}

func rebuildLoggerLocked() {
	var handlers []slog.Handler

	if fileWriter != nil {
		fileHandler := slog.NewTextHandler(fileWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		handlers = append(handlers, fileHandler)
	}

	if uiCallback != nil {
		uiHandler := NewUIHandler(uiCallback, slog.LevelDebug)
		handlers = append(handlers, uiHandler)
	}

	if len(handlers) == 0 {
		slogger = slog.New(slog.NewTextHandler(io.Discard, nil))
		return
	}

	if len(handlers) == 1 {
		slogger = slog.New(handlers[0])
		return
	}

	slogger = slog.New(NewMultiHandler(handlers...))
}

func Init() error {
	return InitWithPaths(paths.DefaultPaths())
}

func InitWithPaths(p *paths.Paths) error {
	stateDir, err := p.StateDir()
	if err != nil {
		return fmt.Errorf("getting state directory: %w", err)
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	logPath := filepath.Join(stateDir, "kyaraben.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}

	logMu.Lock()
	logFile = f
	fileWriter = logFile
	rebuildLoggerLocked()
	logMu.Unlock()
	return nil
}

func Close() {
	logMu.Lock()
	defer logMu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
		fileWriter = nil
	}
}

func LogPath() (string, error) {
	return LogPathWithPaths(paths.DefaultPaths())
}

func LogPathWithPaths(p *paths.Paths) (string, error) {
	stateDir, err := p.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "kyaraben.log"), nil
}

type Logger struct {
	component string
	prefix    string
}

func New(component string) *Logger {
	return &Logger{component: component}
}

func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		component: l.component,
		prefix:    prefix,
	}
}

func (l *Logger) formatMessage(format string, args ...interface{}) string {
	content := fmt.Sprintf(format, args...)
	if l.prefix != "" {
		return l.prefix + " " + content
	}
	return content
}

func (l *Logger) slog() *slog.Logger {
	logMu.RLock()
	s := slogger
	logMu.RUnlock()
	if s == nil {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return s.WithGroup(l.component)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.slog().Info(l.formatMessage(format, args...))
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.slog().Error(l.formatMessage(format, args...))
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.slog().Warn(l.formatMessage(format, args...))
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.slog().Debug(l.formatMessage(format, args...))
}

func getDefaultLogger() *slog.Logger {
	logMu.RLock()
	s := slogger
	logMu.RUnlock()
	if s == nil {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return s
}

func Info(format string, args ...interface{}) {
	getDefaultLogger().Info(fmt.Sprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	getDefaultLogger().Error(fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	getDefaultLogger().Warn(fmt.Sprintf(format, args...))
}

func Debug(format string, args ...interface{}) {
	getDefaultLogger().Debug(fmt.Sprintf(format, args...))
}

func Writer() io.Writer {
	logMu.RLock()
	f := logFile
	logMu.RUnlock()
	if f != nil {
		return f
	}
	return io.Discard
}

func CurrentPosition() int64 {
	logMu.RLock()
	f := logFile
	logMu.RUnlock()
	if f == nil {
		return 0
	}
	_ = f.Sync()
	info, err := f.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}

// SetOutputHook is deprecated. Use SetUICallback instead.
func SetOutputHook(fn func(string)) {
	if fn == nil {
		SetUICallback(nil)
		return
	}
	SetUICallback(func(entry LogEntry) {
		fn(entry.Message)
	})
}
