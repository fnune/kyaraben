package logging

import (
	"context"
	"log/slog"
	"time"
)

type UIHandler struct {
	callback  func(LogEntry)
	component string
	attrs     []slog.Attr
	level     slog.Leveler
}

func NewUIHandler(callback func(LogEntry), level slog.Leveler) *UIHandler {
	return &UIHandler{
		callback: callback,
		level:    level,
	}
}

func (h *UIHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *UIHandler) Handle(_ context.Context, r slog.Record) error {
	if h.callback == nil {
		return nil
	}

	attrs := make(map[string]any)
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	var attrsMap map[string]any
	if len(attrs) > 0 {
		attrsMap = attrs
	}

	entry := LogEntry{
		Level:     slogLevelToLogLevel(r.Level),
		Component: h.component,
		Message:   r.Message,
		Timestamp: r.Time.Format(time.RFC3339),
		Attrs:     attrsMap,
	}

	h.callback(entry)
	return nil
}

func (h *UIHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	return &UIHandler{
		callback:  h.callback,
		component: h.component,
		attrs:     newAttrs,
		level:     h.level,
	}
}

func (h *UIHandler) WithGroup(name string) slog.Handler {
	return &UIHandler{
		callback:  h.callback,
		component: name,
		attrs:     h.attrs,
		level:     h.level,
	}
}

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: handlers}
}

func slogLevelToLogLevel(level slog.Level) LogLevel {
	switch {
	case level >= slog.LevelError:
		return LogLevelError
	case level >= slog.LevelWarn:
		return LogLevelWarn
	case level >= slog.LevelInfo:
		return LogLevelInfo
	default:
		return LogLevelDebug
	}
}
