package persistence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
)

// ThreadLogger writes structured JSONL log entries for a session thread.
type ThreadLogger interface {
	LogEvent(eventType string, event json.RawMessage)
	LogUserMessage(username, message string, displayName string, hasFiles bool)
	LogLifecycle(action string, details map[string]interface{})
	LogCommand(command, args, username string)
	LogPermission(action string, permission, username string)
	LogReaction(action, username, emoji, answer string)
	LogExecutor(executor, operation, postID, method string, details map[string]interface{})
	Flush() error
	Close() error
	IsEnabled() bool
	GetLogPath() string
}

// ThreadLoggerOptions configures a ThreadLogger.
type ThreadLoggerOptions struct {
	Disabled        bool   // if true, return no-op logger
	BufferSize      int    // number of entries before auto-flush (default 10)
	FlushIntervalMs int    // milliseconds between ticker flushes (default 1000)
	BaseDir         string // override base directory (default ~/.claude-threads/logs)
}

// NewThreadLogger creates a ThreadLogger for the given platform/thread/session.
// If opts.Disabled is true, returns a no-op DisabledThreadLogger.
func NewThreadLogger(platformID, threadID, sessionID string, opts ThreadLoggerOptions) ThreadLogger {
	if opts.Disabled {
		return &disabledThreadLogger{}
	}

	bufSize := opts.BufferSize
	if bufSize <= 0 {
		bufSize = 10
	}
	flushMs := opts.FlushIntervalMs
	if flushMs <= 0 {
		flushMs = 1000
	}
	baseDir := opts.BaseDir
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".claude-threads", "logs")
	}

	logPath := filepath.Join(baseDir, platformID, sessionID+".jsonl")

	impl := &threadLoggerImpl{
		logPath:    logPath,
		sessionID:  sessionID,
		bufferSize: bufSize,
		buffer:     make([]string, 0, bufSize),
		log:        utils.CreateLogger("thread-log"),
		stopCh:     make(chan struct{}),
	}

	// Start background flush ticker
	ticker := time.NewTicker(time.Duration(flushMs) * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				_ = impl.flushLocked()
			case <-impl.stopCh:
				ticker.Stop()
				return
			}
		}
	}()

	return impl
}

// GetLogFilePath returns the log file path for a given platform/session.
// If baseDir is empty, uses the default ~/.claude-threads/logs.
func GetLogFilePath(platformID, sessionID string, baseDir string) string {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".claude-threads", "logs")
	}
	return filepath.Join(baseDir, platformID, sessionID+".jsonl")
}

// CleanupOldLogs deletes .jsonl files older than retentionDays (by mtime).
// Removes empty platform directories afterwards.
// baseDir overrides the default ~/.claude-threads/logs (pass "" for default).
// Returns the number of files deleted.
func CleanupOldLogs(retentionDays int, baseDir string) (int, error) {
	if baseDir == "" {
		home, _ := os.UserHomeDir()
		baseDir = filepath.Join(home, ".claude-threads", "logs")
	}

	cutoff := time.Now().Add(-time.Duration(retentionDays) * 24 * time.Hour)
	removed := 0

	entries, err := os.ReadDir(baseDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read log base dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		platformDir := filepath.Join(baseDir, entry.Name())
		files, err := os.ReadDir(platformDir)
		if err != nil {
			continue
		}

		for _, f := range files {
			if filepath.Ext(f.Name()) != ".jsonl" {
				continue
			}
			info, err := f.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Before(cutoff) {
				_ = os.Remove(filepath.Join(platformDir, f.Name()))
				removed++
			}
		}

		// Remove empty platform dir
		remaining, _ := os.ReadDir(platformDir)
		if len(remaining) == 0 {
			_ = os.Remove(platformDir)
		}
	}

	return removed, nil
}

// ReadRecentLogEntries reads the last maxLines JSONL entries from a session log.
// baseDir overrides the default (pass "" for default).
func ReadRecentLogEntries(platformID, sessionID string, maxLines int, baseDir string) ([]json.RawMessage, error) {
	logPath := GetLogFilePath(platformID, sessionID, baseDir)

	f, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan log file: %w", err)
	}

	// Take the last maxLines
	start := len(lines) - maxLines
	if start < 0 {
		start = 0
	}
	lines = lines[start:]

	result := make([]json.RawMessage, 0, len(lines))
	for _, line := range lines {
		result = append(result, json.RawMessage(line))
	}
	return result, nil
}

// --- threadLoggerImpl ---

type threadLoggerImpl struct {
	mu         sync.Mutex
	logPath    string
	sessionID  string
	bufferSize int
	buffer     []string
	log        *utils.Logger
	stopCh     chan struct{}
	closed     bool
}

// Compile-time interface check
var _ ThreadLogger = (*threadLoggerImpl)(nil)

func (l *threadLoggerImpl) IsEnabled() bool  { return true }
func (l *threadLoggerImpl) GetLogPath() string { return l.logPath }

func (l *threadLoggerImpl) append(entry interface{}) {
	data, err := json.Marshal(entry)
	if err != nil {
		l.log.Warn(fmt.Sprintf("failed to marshal log entry: %v", err))
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.buffer = append(l.buffer, string(data))
	if len(l.buffer) >= l.bufferSize {
		_ = l.flushUnlocked()
	}
}

// flushLocked acquires lock and flushes.
func (l *threadLoggerImpl) flushLocked() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flushUnlocked()
}

// flushUnlocked writes buffer to disk WITHOUT acquiring the lock (caller must hold it).
func (l *threadLoggerImpl) flushUnlocked() error {
	if len(l.buffer) == 0 {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(l.logPath), 0700); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	f, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	for _, line := range l.buffer {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return fmt.Errorf("write log entry: %w", err)
		}
	}

	l.buffer = l.buffer[:0]
	return nil
}

func (l *threadLoggerImpl) Flush() error {
	return l.flushLocked()
}

func (l *threadLoggerImpl) Close() error {
	l.mu.Lock()
	if !l.closed {
		l.closed = true
		close(l.stopCh)
	}
	err := l.flushUnlocked()
	l.mu.Unlock()
	return err
}

// baseEntry creates a base log entry map.
func (l *threadLoggerImpl) baseEntry(typ string) map[string]interface{} {
	return map[string]interface{}{
		"ts":        time.Now().UnixMilli(),
		"sessionId": l.sessionID,
		"type":      typ,
	}
}

func (l *threadLoggerImpl) LogEvent(eventType string, event json.RawMessage) {
	entry := l.baseEntry("claude_event")
	entry["eventType"] = eventType
	entry["event"] = event
	l.append(entry)
}

func (l *threadLoggerImpl) LogUserMessage(username, message string, displayName string, hasFiles bool) {
	entry := l.baseEntry("user_message")
	entry["username"] = username
	entry["message"] = message
	if displayName != "" {
		entry["displayName"] = displayName
	}
	if hasFiles {
		entry["hasFiles"] = true
	}
	l.append(entry)
}

func (l *threadLoggerImpl) LogLifecycle(action string, details map[string]interface{}) {
	entry := l.baseEntry("lifecycle")
	entry["action"] = action
	for k, v := range details {
		entry[k] = v
	}
	l.append(entry)
}

func (l *threadLoggerImpl) LogCommand(command, args, username string) {
	entry := l.baseEntry("command")
	entry["command"] = command
	entry["username"] = username
	if args != "" {
		entry["args"] = args
	}
	l.append(entry)
}

func (l *threadLoggerImpl) LogPermission(action string, permission, username string) {
	entry := l.baseEntry("permission")
	entry["action"] = action
	if permission != "" {
		entry["permission"] = permission
	}
	if username != "" {
		entry["username"] = username
	}
	l.append(entry)
}

func (l *threadLoggerImpl) LogReaction(action, username, emoji, answer string) {
	entry := l.baseEntry("reaction")
	entry["action"] = action
	entry["username"] = username
	if emoji != "" {
		entry["emoji"] = emoji
	}
	if answer != "" {
		entry["answer"] = answer
	}
	l.append(entry)
}

func (l *threadLoggerImpl) LogExecutor(executor, operation, postID, method string, details map[string]interface{}) {
	entry := l.baseEntry("executor")
	entry["executor"] = executor
	entry["operation"] = operation
	if postID != "" {
		entry["postId"] = postID
	}
	if method != "" {
		entry["method"] = method
	}
	if len(details) > 0 {
		entry["details"] = details
	}
	l.append(entry)
}

// --- disabledThreadLogger (no-op) ---

type disabledThreadLogger struct{}

// Compile-time interface check
var _ ThreadLogger = (*disabledThreadLogger)(nil)

func (d *disabledThreadLogger) IsEnabled() bool                                                            { return false }
func (d *disabledThreadLogger) GetLogPath() string                                                         { return "" }
func (d *disabledThreadLogger) LogEvent(_ string, _ json.RawMessage)                                       {}
func (d *disabledThreadLogger) LogUserMessage(_, _ string, _ string, _ bool)                               {}
func (d *disabledThreadLogger) LogLifecycle(_ string, _ map[string]interface{})                            {}
func (d *disabledThreadLogger) LogCommand(_, _, _ string)                                                  {}
func (d *disabledThreadLogger) LogPermission(_ string, _, _ string)                                        {}
func (d *disabledThreadLogger) LogReaction(_, _, _, _ string)                                              {}
func (d *disabledThreadLogger) LogExecutor(_, _, _, _ string, _ map[string]interface{})                    {}
func (d *disabledThreadLogger) Flush() error                                                               { return nil }
func (d *disabledThreadLogger) Close() error                                                               { return nil }
