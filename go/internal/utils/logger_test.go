package utils_test

import (
	"os"
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestLoggerDebug_NotLoggedWithoutDebugEnv(t *testing.T) {
	os.Unsetenv("DEBUG")
	var captured []string
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		captured = append(captured, msg)
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("test")
	log.Debug("should not appear")
	assert.Empty(t, captured)
}

func TestLoggerDebug_LoggedWhenDebugIs1(t *testing.T) {
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")

	var captured []utils.LogLevel
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		captured = append(captured, level)
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("test")
	log.Debug("debug message")
	assert.Contains(t, captured, utils.LogLevelDebug)
}

func TestLoggerInfo_AlwaysLogs(t *testing.T) {
	os.Unsetenv("DEBUG")
	var captured []string
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		captured = append(captured, msg)
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("test")
	log.Info("info message")
	assert.Len(t, captured, 1)
	assert.Equal(t, "info message", captured[0])
}

func TestLoggerWarn_AlwaysLogs(t *testing.T) {
	var captured []utils.LogLevel
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		captured = append(captured, level)
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("test")
	log.Warn("warn message")
	assert.Contains(t, captured, utils.LogLevelWarn)
}

func TestLoggerError_AlwaysLogs(t *testing.T) {
	var captured []utils.LogLevel
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		captured = append(captured, level)
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("test")
	log.Error("error message", nil)
	assert.Contains(t, captured, utils.LogLevelError)
}

func TestLoggerComponentPadding(t *testing.T) {
	var capturedComponents []string
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		capturedComponents = append(capturedComponents, component)
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("short")
	log.Info("msg")
	assert.Len(t, capturedComponents, 1)
	assert.Equal(t, 10, len(capturedComponents[0]))

	capturedComponents = nil
	log2 := utils.CreateLogger("averylongcomponentname")
	log2.Info("msg")
	assert.Equal(t, 10, len(capturedComponents[0]))
}

func TestLoggerForSession(t *testing.T) {
	var capturedSession *string
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		capturedSession = sessionID
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("test").ForSession("session-123")
	log.Info("msg")
	assert.NotNil(t, capturedSession)
	assert.Equal(t, "session-123", *capturedSession)
}

func TestLoggerSetHandlerNil_RevertToConsole(t *testing.T) {
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {})
	utils.SetLogHandler(nil)
	log := utils.CreateLogger("test")
	log.Info("should not panic")
}

func TestDebugJson_LoggedWhenDebugIs1(t *testing.T) {
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")

	var captured []string
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		captured = append(captured, msg)
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("test")
	log.DebugJSON("label", map[string]string{"key": "value"}, 0)
	assert.Len(t, captured, 1)
	assert.Contains(t, captured[0], "label")
}

func TestDebugJson_TruncatesLongJSON(t *testing.T) {
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")

	var captured []string
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		captured = append(captured, msg)
	})
	defer utils.SetLogHandler(nil)

	log := utils.CreateLogger("test")
	data := map[string]string{"key": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}
	log.DebugJSON("label", data, 60)
	assert.Len(t, captured, 1)
	assert.LessOrEqual(t, len(captured[0]), 120)
}

func TestMcpLogger_HasExpectedMethods(t *testing.T) {
	assert.NotNil(t, utils.McpLogger)
	utils.McpLogger.Debug("test")
}

func TestWsLogger_HasExpectedMethods(t *testing.T) {
	assert.NotNil(t, utils.WsLogger)
	utils.WsLogger.Debug("test")
}
