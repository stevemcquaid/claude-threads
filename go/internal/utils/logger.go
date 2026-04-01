package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// LogLevel represents the severity of a log message.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogHandler is a function that receives log events (used for UI integration).
type LogHandler func(level LogLevel, component, message string, sessionID *string)

var (
	globalHandler   LogHandler
	globalHandlerMu sync.RWMutex
)

// SetLogHandler sets a global log handler. Pass nil to revert to console output.
func SetLogHandler(h LogHandler) {
	globalHandlerMu.Lock()
	defer globalHandlerMu.Unlock()
	globalHandler = h
}

// Logger is a component-scoped logger.
type Logger struct {
	component string
	useStderr bool
	sessionID *string
}

// CreateLogger creates a new Logger for the given component.
// component is padded/truncated to 10 characters for consistent alignment.
func CreateLogger(component string, useStderr ...bool) *Logger {
	padded := component
	if len(padded) > 10 {
		padded = padded[:10]
	}
	for len(padded) < 10 {
		padded += " "
	}
	stderr := len(useStderr) > 0 && useStderr[0]
	return &Logger{component: padded, useStderr: stderr}
}

func (l *Logger) emit(level LogLevel, msg string) {
	globalHandlerMu.RLock()
	h := globalHandler
	globalHandlerMu.RUnlock()

	if h != nil {
		h(level, l.component, msg, l.sessionID)
		return
	}

	line := fmt.Sprintf("[%s] %s", l.component, msg)
	switch level {
	case LogLevelError:
		fmt.Fprintln(os.Stderr, "❌ "+line)
	case LogLevelWarn:
		fmt.Fprintln(os.Stderr, "⚠️  "+line)
	default:
		if l.useStderr {
			fmt.Fprintln(os.Stderr, line)
		} else {
			fmt.Fprintln(os.Stdout, line)
		}
	}
}

func isDebug() bool { return os.Getenv("DEBUG") == "1" }

// Debug logs only when DEBUG=1.
func (l *Logger) Debug(msg string, args ...any) {
	if !isDebug() {
		return
	}
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.emit(LogLevelDebug, msg)
}

// DebugJSON logs a labeled JSON value when DEBUG=1.
// maxLen controls truncation (0 = use default of 60).
func (l *Logger) DebugJSON(label string, data any, maxLen int) {
	if !isDebug() {
		return
	}
	if maxLen <= 0 {
		maxLen = 60
	}
	b, err := json.Marshal(data)
	if err != nil {
		l.emit(LogLevelDebug, fmt.Sprintf("%s: <marshal error: %v>", label, err))
		return
	}
	s := string(b)
	if len(s) > maxLen {
		s = s[:maxLen] + "…"
	}
	l.emit(LogLevelDebug, fmt.Sprintf("%s: %s", label, s))
}

// Info always logs.
func (l *Logger) Info(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.emit(LogLevelInfo, msg)
}

// Warn always logs as a warning.
func (l *Logger) Warn(msg string, args ...any) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	l.emit(LogLevelWarn, msg)
}

// Error always logs to stderr. Pass err=nil if no error object.
func (l *Logger) Error(msg string, err error) {
	if err != nil && isDebug() {
		msg = fmt.Sprintf("%s: %+v", msg, err)
	} else if err != nil {
		msg = fmt.Sprintf("%s: %v", msg, err)
	}
	l.emit(LogLevelError, msg)
}

// ForSession returns a new Logger scoped to the given session ID.
func (l *Logger) ForSession(sessionID string) *Logger {
	return &Logger{
		component: l.component,
		useStderr: l.useStderr,
		sessionID: &sessionID,
	}
}

// Pre-configured loggers for common components.
var (
	McpLogger = CreateLogger("mcp", true)
	WsLogger  = CreateLogger("websocket")
)
