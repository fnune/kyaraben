package logging

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogEntry struct {
	Level     LogLevel       `json:"level"`
	Component string         `json:"component,omitempty"`
	Message   string         `json:"message"`
	Timestamp string         `json:"timestamp"`
	Attrs     map[string]any `json:"attrs,omitempty"`
}
