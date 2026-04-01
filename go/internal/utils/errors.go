package utils

import "fmt"

// ErrorSeverity determines how an error is handled.
type ErrorSeverity string

const (
	// ErrorSeverityRecoverable: log and continue, do not re-throw.
	ErrorSeverityRecoverable ErrorSeverity = "recoverable"
	// ErrorSeveritySessionFatal: log, notify user, and re-throw.
	ErrorSeveritySessionFatal ErrorSeverity = "session-fatal"
	// ErrorSeveritySystemFatal: log and re-throw (system-level failure).
	ErrorSeveritySystemFatal ErrorSeverity = "system-fatal"
)

var errorLog = CreateLogger("error")

// LogSilentError logs an error at debug level without propagating it.
// Use instead of empty catch blocks to preserve visibility when DEBUG=1.
func LogSilentError(context string, err any) {
	var msg string
	switch e := err.(type) {
	case error:
		msg = e.Error()
	default:
		msg = fmt.Sprintf("%v", e)
	}
	errorLog.Debug("[%s] Silently caught: %s", context, msg)
}
