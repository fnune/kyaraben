package syncthing

type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Error(format string, args ...any)
}

type noopLogger struct{}

func (noopLogger) Debug(format string, args ...any) {}
func (noopLogger) Info(format string, args ...any)  {}
func (noopLogger) Error(format string, args ...any) {}

var defaultLogger Logger = noopLogger{}

func SetLogger(l Logger) {
	if l == nil {
		defaultLogger = noopLogger{}
		return
	}
	defaultLogger = l
}
