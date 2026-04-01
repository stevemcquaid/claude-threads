package persistence

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(t *testing.T, platformID, sessionID string) (ThreadLogger, string) {
	t.Helper()
	baseDir := t.TempDir()
	opts := ThreadLoggerOptions{
		BufferSize:      5,
		FlushIntervalMs: 100,
		BaseDir:         baseDir,
	}
	logger := NewThreadLogger(platformID, "thread-1", sessionID, opts)
	logPath := GetLogFilePath(platformID, sessionID, baseDir)
	return logger, logPath
}

func readJSONLFile(t *testing.T, path string) []map[string]interface{} {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err, "log file should exist: %s", path)
	defer f.Close()

	var entries []map[string]interface{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var entry map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		entries = append(entries, entry)
	}
	return entries
}

func TestNewThreadLogger_Disabled(t *testing.T) {
	opts := ThreadLoggerOptions{
		Disabled: true,
		BaseDir:  t.TempDir(),
	}
	logger := NewThreadLogger("p1", "t1", "s1", opts)
	assert.False(t, logger.IsEnabled())
	assert.NotPanics(t, func() {
		logger.LogUserMessage("user", "hello", "", false)
		logger.LogLifecycle("start", nil)
		logger.LogCommand("!stop", "", "user")
		logger.LogPermission("request", "write", "user")
		logger.LogReaction("cancel", "user", "x", "")
		logger.LogExecutor("task_list", "create", "post-1", "", nil)
		_ = logger.Flush()
		_ = logger.Close()
	})
}

func TestThreadLogger_GetLogPath(t *testing.T) {
	baseDir := t.TempDir()
	logger := NewThreadLogger("platform1", "thread1", "session-abc", ThreadLoggerOptions{
		BaseDir: baseDir,
	})
	expected := filepath.Join(baseDir, "platform1", "session-abc.jsonl")
	assert.Equal(t, expected, logger.GetLogPath())
}

func TestThreadLogger_LogUserMessage(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-user")
	logger.LogUserMessage("alice", "hello world", "Alice Smith", true)
	require.NoError(t, logger.Flush())

	entries := readJSONLFile(t, logPath)
	require.Len(t, entries, 1)
	assert.Equal(t, "user_message", entries[0]["type"])
	assert.Equal(t, "alice", entries[0]["username"])
	assert.Equal(t, "hello world", entries[0]["message"])
	assert.Equal(t, "Alice Smith", entries[0]["displayName"])
	assert.Equal(t, true, entries[0]["hasFiles"])
	assert.Equal(t, "sess-user", entries[0]["sessionId"])
	assert.NotEmpty(t, entries[0]["ts"])
}

func TestThreadLogger_LogLifecycle(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-lc")
	logger.LogLifecycle("start", map[string]interface{}{
		"workingDir": "/tmp/work",
		"username":   "bob",
	})
	require.NoError(t, logger.Flush())

	entries := readJSONLFile(t, logPath)
	require.Len(t, entries, 1)
	assert.Equal(t, "lifecycle", entries[0]["type"])
	assert.Equal(t, "start", entries[0]["action"])
	assert.Equal(t, "/tmp/work", entries[0]["workingDir"])
	assert.Equal(t, "bob", entries[0]["username"])
}

func TestThreadLogger_LogCommand(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-cmd")
	logger.LogCommand("!cd", "/tmp/newdir", "alice")
	require.NoError(t, logger.Flush())

	entries := readJSONLFile(t, logPath)
	require.Len(t, entries, 1)
	assert.Equal(t, "command", entries[0]["type"])
	assert.Equal(t, "!cd", entries[0]["command"])
	assert.Equal(t, "/tmp/newdir", entries[0]["args"])
	assert.Equal(t, "alice", entries[0]["username"])
}

func TestThreadLogger_LogPermission(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-perm")
	logger.LogPermission("approve", "write /tmp/file.txt", "alice")
	require.NoError(t, logger.Flush())

	entries := readJSONLFile(t, logPath)
	require.Len(t, entries, 1)
	assert.Equal(t, "permission", entries[0]["type"])
	assert.Equal(t, "approve", entries[0]["action"])
	assert.Equal(t, "write /tmp/file.txt", entries[0]["permission"])
	assert.Equal(t, "alice", entries[0]["username"])
}

func TestThreadLogger_LogReaction(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-react")
	logger.LogReaction("plan_approve", "alice", "thumbsup", "")
	require.NoError(t, logger.Flush())

	entries := readJSONLFile(t, logPath)
	require.Len(t, entries, 1)
	assert.Equal(t, "reaction", entries[0]["type"])
	assert.Equal(t, "plan_approve", entries[0]["action"])
	assert.Equal(t, "alice", entries[0]["username"])
	assert.Equal(t, "thumbsup", entries[0]["emoji"])
}

func TestThreadLogger_LogExecutor(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-exec")
	logger.LogExecutor("task_list", "create", "post-123", "CreatePost", map[string]interface{}{
		"count": 3,
	})
	require.NoError(t, logger.Flush())

	entries := readJSONLFile(t, logPath)
	require.Len(t, entries, 1)
	assert.Equal(t, "executor", entries[0]["type"])
	assert.Equal(t, "task_list", entries[0]["executor"])
	assert.Equal(t, "create", entries[0]["operation"])
	assert.Equal(t, "post-123", entries[0]["postId"])
	assert.Equal(t, "CreatePost", entries[0]["method"])
	details, ok := entries[0]["details"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(3), details["count"])
}

func TestThreadLogger_BufferAutoFlush(t *testing.T) {
	baseDir := t.TempDir()
	// Buffer size = 3, so 4th write should trigger flush
	opts := ThreadLoggerOptions{
		BufferSize:      3,
		FlushIntervalMs: 5000, // long interval, should flush on buffer full
		BaseDir:         baseDir,
	}
	logger := NewThreadLogger("p1", "t1", "sess-buf", opts)
	logPath := GetLogFilePath("p1", "sess-buf", baseDir)

	for i := 0; i < 4; i++ {
		logger.LogCommand("!test", "", "user")
	}

	// Give auto-flush a moment to trigger
	time.Sleep(50 * time.Millisecond)

	entries := readJSONLFile(t, logPath)
	assert.GreaterOrEqual(t, len(entries), 3, "buffer overflow should have triggered a flush")

	require.NoError(t, logger.Close())
}

func TestThreadLogger_Close(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-close")
	logger.LogUserMessage("alice", "msg1", "", false)
	logger.LogUserMessage("alice", "msg2", "", false)

	// Close without explicit flush - Close should flush
	require.NoError(t, logger.Close())

	entries := readJSONLFile(t, logPath)
	assert.Len(t, entries, 2)
}

func TestCleanupOldLogs(t *testing.T) {
	baseDir := t.TempDir()

	// Create platform dir and some log files
	platformDir := filepath.Join(baseDir, "platform1")
	require.NoError(t, os.MkdirAll(platformDir, 0700))

	// Old file (modify time set to past)
	oldFile := filepath.Join(platformDir, "old-session.jsonl")
	require.NoError(t, os.WriteFile(oldFile, []byte(`{"ts":1}`+"\n"), 0600))
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	require.NoError(t, os.Chtimes(oldFile, oldTime, oldTime))

	// Recent file
	recentFile := filepath.Join(platformDir, "recent-session.jsonl")
	require.NoError(t, os.WriteFile(recentFile, []byte(`{"ts":2}`+"\n"), 0600))

	removed, err := CleanupOldLogs(30, baseDir)
	require.NoError(t, err)
	assert.Equal(t, 1, removed)

	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err), "old log file should be deleted")

	_, err = os.Stat(recentFile)
	assert.NoError(t, err, "recent log file should remain")
}

func TestCleanupOldLogs_RemovesEmptyDir(t *testing.T) {
	baseDir := t.TempDir()
	platformDir := filepath.Join(baseDir, "platform1")
	require.NoError(t, os.MkdirAll(platformDir, 0700))

	// Single old file
	oldFile := filepath.Join(platformDir, "old.jsonl")
	require.NoError(t, os.WriteFile(oldFile, []byte{}, 0600))
	oldTime := time.Now().Add(-40 * 24 * time.Hour)
	require.NoError(t, os.Chtimes(oldFile, oldTime, oldTime))

	_, err := CleanupOldLogs(30, baseDir)
	require.NoError(t, err)

	_, err = os.Stat(platformDir)
	assert.True(t, os.IsNotExist(err), "empty platform dir should be removed")
}

func TestReadRecentLogEntries(t *testing.T) {
	baseDir := t.TempDir()
	platformDir := filepath.Join(baseDir, "platform1")
	require.NoError(t, os.MkdirAll(platformDir, 0700))
	logFile := filepath.Join(platformDir, "sess-read.jsonl")

	// Write 10 JSONL entries
	var content string
	for i := 0; i < 10; i++ {
		line := map[string]interface{}{"ts": i, "type": "test", "sessionId": "sess-read"}
		data, _ := json.Marshal(line)
		content += string(data) + "\n"
	}
	require.NoError(t, os.WriteFile(logFile, []byte(content), 0600))

	// Read last 5
	entries, err := ReadRecentLogEntries("platform1", "sess-read", 5, baseDir)
	require.NoError(t, err)
	assert.Len(t, entries, 5)

	// Verify content — last 5 entries should have ts 5..9
	for i, raw := range entries {
		var entry map[string]interface{}
		require.NoError(t, json.Unmarshal(raw, &entry))
		assert.Equal(t, float64(i+5), entry["ts"])
	}
}

func TestReadRecentLogEntries_FewerThanMax(t *testing.T) {
	baseDir := t.TempDir()
	platformDir := filepath.Join(baseDir, "platform2")
	require.NoError(t, os.MkdirAll(platformDir, 0700))
	logFile := filepath.Join(platformDir, "sess-few.jsonl")

	// Write 3 entries but request last 10
	var content string
	for i := 0; i < 3; i++ {
		line := map[string]interface{}{"ts": i}
		data, _ := json.Marshal(line)
		content += string(data) + "\n"
	}
	require.NoError(t, os.WriteFile(logFile, []byte(content), 0600))

	entries, err := ReadRecentLogEntries("platform2", "sess-few", 10, baseDir)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
}

func TestThreadLogger_LogEvent(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-event")
	rawEvent := json.RawMessage(`{"tool":"write","path":"/tmp/x"}`)
	logger.LogEvent("tool_use", rawEvent)
	require.NoError(t, logger.Flush())

	entries := readJSONLFile(t, logPath)
	require.Len(t, entries, 1)
	assert.Equal(t, "claude_event", entries[0]["type"])
	assert.Equal(t, "tool_use", entries[0]["eventType"])
	// event field should be raw JSON
	assert.NotNil(t, entries[0]["event"])
}

func TestThreadLogger_FilePermissions(t *testing.T) {
	logger, logPath := newTestLogger(t, "p1", "sess-perm-check")
	logger.LogUserMessage("user", "test", "", false)
	require.NoError(t, logger.Close())

	info, err := os.Stat(logPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}
