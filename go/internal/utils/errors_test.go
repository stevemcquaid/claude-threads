package utils_test

import (
	"errors"
	"os"
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestLogSilentError_DoesNotPanic(t *testing.T) {
	utils.LogSilentError("test-context", errors.New("something went wrong"))
	utils.LogSilentError("test-context", "string error")
}

func TestLogSilentError_LogsWhenDebug(t *testing.T) {
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")

	var captured []string
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		captured = append(captured, msg)
	})
	defer utils.SetLogHandler(nil)

	utils.LogSilentError("remove-reaction", errors.New("network timeout"))
	assert.Len(t, captured, 1)
	assert.Contains(t, captured[0], "remove-reaction")
	assert.Contains(t, captured[0], "network timeout")
}

func TestErrorSeverityConstants(t *testing.T) {
	assert.Equal(t, utils.ErrorSeverityRecoverable, utils.ErrorSeverity("recoverable"))
	assert.Equal(t, utils.ErrorSeveritySessionFatal, utils.ErrorSeverity("session-fatal"))
	assert.Equal(t, utils.ErrorSeveritySystemFatal, utils.ErrorSeverity("system-fatal"))
}
