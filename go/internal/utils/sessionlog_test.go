package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

type mockSession struct {
	id string
}

func (m *mockSession) GetSessionID() string { return m.id }

func TestCreateSessionLog_WithSession(t *testing.T) {
	var capturedSessionID *string
	utils.SetLogHandler(func(level utils.LogLevel, component, msg string, sessionID *string) {
		capturedSessionID = sessionID
	})
	defer utils.SetLogHandler(nil)

	base := utils.CreateLogger("test")
	logFn := utils.CreateSessionLog(base)

	sess := &mockSession{id: "sess-123"}
	scopedLog := logFn(sess)
	scopedLog.Info("test message")

	assert.NotNil(t, capturedSessionID)
	assert.Equal(t, "sess-123", *capturedSessionID)
}

func TestCreateSessionLog_WithNilSession(t *testing.T) {
	base := utils.CreateLogger("test")
	logFn := utils.CreateSessionLog(base)
	result := logFn(nil)
	assert.NotNil(t, result)
}
