# Go Conversion Phase 1: Foundation (utils + config)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Translate all `src/utils/` and `src/config/` TypeScript modules to Go with full test coverage.

**Architecture:** Pure utility functions and configuration types. No external service dependencies. Everything else in the project builds on these packages.

**Tech Stack:** Go standard library, `gopkg.in/yaml.v3`, `github.com/stretchr/testify`

**Prerequisites:** Phase 0 complete (go module initialized)

---

### Task 1: Emoji Utilities

**Files:**
- Create: `go/internal/utils/emoji.go`
- Create: `go/internal/utils/emoji_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/emoji_test.go`:

```go
package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestIsApprovalEmoji(t *testing.T) {
	assert.True(t, utils.IsApprovalEmoji("+1"))
	assert.True(t, utils.IsApprovalEmoji("thumbsup"))
	assert.False(t, utils.IsApprovalEmoji("x"))
	assert.False(t, utils.IsApprovalEmoji(""))
	for _, e := range utils.ApprovalEmojis {
		assert.True(t, utils.IsApprovalEmoji(e), "should match %s", e)
	}
}

func TestIsDenialEmoji(t *testing.T) {
	assert.True(t, utils.IsDenialEmoji("-1"))
	assert.True(t, utils.IsDenialEmoji("thumbsdown"))
	assert.False(t, utils.IsDenialEmoji("+1"))
	for _, e := range utils.DenialEmojis {
		assert.True(t, utils.IsDenialEmoji(e))
	}
}

func TestIsAllowAllEmoji(t *testing.T) {
	assert.True(t, utils.IsAllowAllEmoji("white_check_mark"))
	assert.True(t, utils.IsAllowAllEmoji("heavy_check_mark"))
	assert.False(t, utils.IsAllowAllEmoji("+1"))
	for _, e := range utils.AllowAllEmojis {
		assert.True(t, utils.IsAllowAllEmoji(e))
	}
}

func TestIsCancelEmoji(t *testing.T) {
	assert.True(t, utils.IsCancelEmoji("x"))
	assert.True(t, utils.IsCancelEmoji("octagonal_sign"))
	assert.True(t, utils.IsCancelEmoji("stop_sign"))
	assert.False(t, utils.IsCancelEmoji("+1"))
	for _, e := range utils.CancelEmojis {
		assert.True(t, utils.IsCancelEmoji(e))
	}
}

func TestIsEscapeEmoji(t *testing.T) {
	assert.True(t, utils.IsEscapeEmoji("double_vertical_bar"))
	assert.True(t, utils.IsEscapeEmoji("pause_button"))
	assert.False(t, utils.IsEscapeEmoji("x"))
	for _, e := range utils.EscapeEmojis {
		assert.True(t, utils.IsEscapeEmoji(e))
	}
}

func TestIsResumeEmoji(t *testing.T) {
	assert.True(t, utils.IsResumeEmoji("arrows_counterclockwise"))
	assert.True(t, utils.IsResumeEmoji("arrow_forward"))
	assert.True(t, utils.IsResumeEmoji("repeat"))
	assert.False(t, utils.IsResumeEmoji("x"))
	for _, e := range utils.ResumeEmojis {
		assert.True(t, utils.IsResumeEmoji(e))
	}
}

func TestIsMinimizeToggleEmoji(t *testing.T) {
	assert.True(t, utils.IsMinimizeToggleEmoji("arrow_down_small"))
	assert.True(t, utils.IsMinimizeToggleEmoji("small_red_triangle_down"))
	assert.False(t, utils.IsMinimizeToggleEmoji("x"))
	for _, e := range utils.MinimizeToggleEmojis {
		assert.True(t, utils.IsMinimizeToggleEmoji(e))
	}
}

func TestIsBugReportEmoji(t *testing.T) {
	assert.True(t, utils.IsBugReportEmoji("bug"))
	assert.True(t, utils.IsBugReportEmoji("🐛"))
	assert.False(t, utils.IsBugReportEmoji("x"))
	assert.False(t, utils.IsBugReportEmoji(""))
}

func TestGetNumberEmojiIndex(t *testing.T) {
	// Text names
	assert.Equal(t, 0, utils.GetNumberEmojiIndex("one"))
	assert.Equal(t, 1, utils.GetNumberEmojiIndex("two"))
	assert.Equal(t, 2, utils.GetNumberEmojiIndex("three"))
	assert.Equal(t, 3, utils.GetNumberEmojiIndex("four"))
	// Unicode variants
	assert.Equal(t, 0, utils.GetNumberEmojiIndex("1️⃣"))
	assert.Equal(t, 1, utils.GetNumberEmojiIndex("2️⃣"))
	assert.Equal(t, 2, utils.GetNumberEmojiIndex("3️⃣"))
	assert.Equal(t, 3, utils.GetNumberEmojiIndex("4️⃣"))
	// Non-number
	assert.Equal(t, -1, utils.GetNumberEmojiIndex("x"))
	assert.Equal(t, -1, utils.GetNumberEmojiIndex(""))
	// All NumberEmojis
	for i, e := range utils.NumberEmojis {
		assert.Equal(t, i, utils.GetNumberEmojiIndex(e))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run TestIsApprovalEmoji -v
```

Expected: `FAIL — package not found or undefined: utils.IsApprovalEmoji`

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/emoji.go`:

```go
package utils

// ApprovalEmojis are reactions indicating approval/yes.
var ApprovalEmojis = []string{"+1", "thumbsup"}

// DenialEmojis are reactions indicating denial/no.
var DenialEmojis = []string{"-1", "thumbsdown"}

// AllowAllEmojis are reactions indicating "allow all permissions".
var AllowAllEmojis = []string{"white_check_mark", "heavy_check_mark"}

// NumberEmojis are reactions used for numbered choices (0-indexed).
var NumberEmojis = []string{"one", "two", "three", "four"}

// CancelEmojis are reactions that cancel/stop a session.
var CancelEmojis = []string{"x", "octagonal_sign", "stop_sign", "stop"}

// EscapeEmojis are reactions that interrupt/pause a session.
var EscapeEmojis = []string{"double_vertical_bar", "pause_button", "pause"}

// ResumeEmojis are reactions that resume a paused session.
var ResumeEmojis = []string{"arrows_counterclockwise", "arrow_forward", "repeat"}

// MinimizeToggleEmojis are reactions that toggle minimize/expand.
var MinimizeToggleEmojis = []string{"arrow_down_small", "small_red_triangle_down"}

// BugReportEmoji is the reaction that triggers a bug report.
const BugReportEmoji = "bug"

// unicodeNumberEmojis maps unicode number emojis to 0-based indices.
var unicodeNumberEmojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣"}

func containsEmoji(list []string, emoji string) bool {
	for _, e := range list {
		if e == emoji {
			return true
		}
	}
	return false
}

// IsApprovalEmoji returns true if the emoji indicates approval.
func IsApprovalEmoji(emoji string) bool { return containsEmoji(ApprovalEmojis, emoji) }

// IsDenialEmoji returns true if the emoji indicates denial.
func IsDenialEmoji(emoji string) bool { return containsEmoji(DenialEmojis, emoji) }

// IsAllowAllEmoji returns true if the emoji means "allow all permissions".
func IsAllowAllEmoji(emoji string) bool { return containsEmoji(AllowAllEmojis, emoji) }

// IsCancelEmoji returns true if the emoji cancels a session.
func IsCancelEmoji(emoji string) bool { return containsEmoji(CancelEmojis, emoji) }

// IsEscapeEmoji returns true if the emoji pauses/interrupts a session.
func IsEscapeEmoji(emoji string) bool { return containsEmoji(EscapeEmojis, emoji) }

// IsResumeEmoji returns true if the emoji resumes a session.
func IsResumeEmoji(emoji string) bool { return containsEmoji(ResumeEmojis, emoji) }

// IsMinimizeToggleEmoji returns true if the emoji toggles minimize/expand.
func IsMinimizeToggleEmoji(emoji string) bool { return containsEmoji(MinimizeToggleEmojis, emoji) }

// IsBugReportEmoji returns true if the emoji triggers a bug report.
// Handles both the text name "bug" and the unicode variant "🐛".
func IsBugReportEmoji(emoji string) bool {
	return emoji == BugReportEmoji || emoji == "🐛"
}

// GetNumberEmojiIndex returns the 0-based index for a number emoji,
// or -1 if the emoji is not a number emoji.
// Handles both text names ("one", "two") and unicode variants ("1️⃣", "2️⃣").
func GetNumberEmojiIndex(emoji string) int {
	for i, e := range NumberEmojis {
		if e == emoji {
			return i
		}
	}
	for i, e := range unicodeNumberEmojis {
		if e == emoji {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestIs|TestGet" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/emoji.go go/internal/utils/emoji_test.go
git commit -m "feat(go): add emoji utilities"
```

---

### Task 2: Format Utilities

**Files:**
- Create: `go/internal/utils/format.go`
- Create: `go/internal/utils/format_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/format_test.go`:

```go
package utils_test

import (
	"testing"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestExtractThreadID(t *testing.T) {
	assert.Equal(t, "thread123", utils.ExtractThreadID("platform:thread123"))
	assert.Equal(t, "abc", utils.ExtractThreadID("abc"))           // no colon → original
	assert.Equal(t, "c", utils.ExtractThreadID("a:b:c"))            // multiple colons → last segment
	assert.Equal(t, "", utils.ExtractThreadID(""))
}

func TestFormatShortID(t *testing.T) {
	// Short IDs returned as-is
	assert.Equal(t, "abc12345", utils.FormatShortID("abc12345"))    // exactly 8
	assert.Equal(t, "ab", utils.FormatShortID("ab"))                 // under 8
	// Long IDs truncated
	assert.Equal(t, "abc12345…", utils.FormatShortID("abc123456789"))
	// Empty
	assert.Equal(t, "", utils.FormatShortID(""))
	// Composite ID: extracts thread part then truncates
	assert.Equal(t, "abc12345…", utils.FormatShortID("platform:abc123456789"))
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "5s", utils.FormatDuration(5000))
	assert.Equal(t, "45s", utils.FormatDuration(45000))
	assert.Equal(t, "1m 30s", utils.FormatDuration(90000))
	assert.Equal(t, "2m", utils.FormatDuration(120000))
	assert.Equal(t, "1h", utils.FormatDuration(3600000))
	assert.Equal(t, "1h 30m", utils.FormatDuration(5400000))
	assert.Equal(t, "2h", utils.FormatDuration(7200000))
	assert.Equal(t, "0s", utils.FormatDuration(0))
}

func TestFormatRelativeTimeShort(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "<1m ago", utils.FormatRelativeTimeShort(now.Add(-30*time.Second)))
	assert.Equal(t, "5m ago", utils.FormatRelativeTimeShort(now.Add(-5*time.Minute)))
	assert.Equal(t, "2h ago", utils.FormatRelativeTimeShort(now.Add(-2*time.Hour)))
	assert.Equal(t, "3d ago", utils.FormatRelativeTimeShort(now.Add(-3*24*time.Hour)))
}

func TestTruncateAtWord(t *testing.T) {
	// Within limit → unchanged
	assert.Equal(t, "hello", utils.TruncateAtWord("hello", 10))
	// Break at word boundary when space is past 70% threshold
	assert.Equal(t, "hello…", utils.TruncateAtWord("hello world", 8))
	// Hard truncate when space is too early (before 70%)
	result := utils.TruncateAtWord("hi there world this is long text", 10)
	assert.Contains(t, result, "…")
	assert.LessOrEqual(t, len([]rune(result)), 11) // maxLength + ellipsis
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestExtractThreadID|TestFormatShortID|TestFormatDuration|TestFormatRelativeTimeShort|TestTruncateAtWord" -v
```

Expected: FAIL — `undefined: utils.ExtractThreadID`

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/format.go`:

```go
package utils

import (
	"fmt"
	"strings"
	"time"
)

// ExtractThreadID extracts the thread ID from a composite session ID.
// Composite IDs have the format "platformId:threadId".
// Returns the original string if no colon is present.
func ExtractThreadID(sessionID string) string {
	idx := strings.LastIndex(sessionID, ":")
	if idx < 0 {
		return sessionID
	}
	return sessionID[idx+1:]
}

// FormatShortID formats an ID for display, showing at most 8 characters
// with an ellipsis. For composite IDs (platformId:threadId), extracts
// the thread portion first.
func FormatShortID(id string) string {
	threadID := ExtractThreadID(id)
	if len(threadID) <= 8 {
		return threadID
	}
	return threadID[:8] + "…"
}

// FormatDuration formats a duration in milliseconds to a human-readable string.
// Examples: "5s", "1m 30s", "2h", "1h 30m"
func FormatDuration(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		if m > 0 {
			return fmt.Sprintf("%dh %dm", h, m)
		}
		return fmt.Sprintf("%dh", h)
	}
	if m > 0 {
		if s > 0 {
			return fmt.Sprintf("%dm %ds", m, s)
		}
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", s)
}

// FormatRelativeTimeShort formats a time relative to now in short format.
// Examples: "<1m ago", "5m ago", "2h ago", "3d ago"
func FormatRelativeTimeShort(t time.Time) string {
	d := time.Since(t)
	minutes := int(d.Minutes())
	hours := int(d.Hours())
	days := int(d.Hours() / 24)

	if minutes < 1 {
		return "<1m ago"
	}
	if hours < 1 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	if days < 1 {
		return fmt.Sprintf("%dh ago", hours)
	}
	return fmt.Sprintf("%dd ago", days)
}

// TruncateAtWord truncates a string to maxLength characters with an ellipsis.
// Breaks at a word boundary if the last space is past 70% of maxLength;
// otherwise hard-truncates.
func TruncateAtWord(s string, maxLength int) string {
	if len([]rune(s)) <= maxLength {
		return s
	}
	runes := []rune(s)
	sub := string(runes[:maxLength])
	lastSpace := strings.LastIndex(sub, " ")
	threshold := int(float64(maxLength) * 0.7)
	if lastSpace > threshold {
		return string(runes[:lastSpace]) + "…"
	}
	return sub + "…"
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestExtractThreadID|TestFormatShortID|TestFormatDuration|TestFormatRelativeTimeShort|TestTruncateAtWord" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/format.go go/internal/utils/format_test.go
git commit -m "feat(go): add format utilities"
```

---

### Task 3: Color Utilities

**Files:**
- Create: `go/internal/utils/colors.go`
- Create: `go/internal/utils/colors_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/colors_test.go`:

```go
package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestColors(t *testing.T) {
	// Verify ANSI codes are non-empty strings
	assert.NotEmpty(t, utils.ColorReset)
	assert.NotEmpty(t, utils.ColorBold)
	assert.NotEmpty(t, utils.ColorDim)
	assert.NotEmpty(t, utils.ColorCyan)
	assert.NotEmpty(t, utils.ColorGreen)
	assert.NotEmpty(t, utils.ColorRed)
	assert.NotEmpty(t, utils.ColorYellow)
	assert.NotEmpty(t, utils.ColorBlue)
	assert.NotEmpty(t, utils.ColorOrange)
}

func TestColorHelpers(t *testing.T) {
	text := "hello"

	dimmed := utils.Dim(text)
	assert.Contains(t, dimmed, text)
	assert.Contains(t, dimmed, utils.ColorDim)
	assert.Contains(t, dimmed, utils.ColorReset)

	bolded := utils.Bold(text)
	assert.Contains(t, bolded, text)
	assert.Contains(t, bolded, utils.ColorBold)
	assert.Contains(t, bolded, utils.ColorReset)

	greened := utils.Green(text)
	assert.Contains(t, greened, text)
	assert.Contains(t, greened, utils.ColorGreen)

	redded := utils.Red(text)
	assert.Contains(t, redded, text)
	assert.Contains(t, redded, utils.ColorRed)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestColors" -v
```

Expected: FAIL — `undefined: utils.ColorReset`

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/colors.go`:

```go
package utils

// ANSI escape codes for terminal color output.
const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[32m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"   // Claude brand blue
	ColorOrange = "\033[38;5;208m"
)

// Dim wraps text with dim ANSI styling.
func Dim(s string) string { return ColorDim + s + ColorReset }

// Bold wraps text with bold ANSI styling.
func Bold(s string) string { return ColorBold + s + ColorReset }

// Green wraps text with green ANSI color.
func Green(s string) string { return ColorGreen + s + ColorReset }

// Red wraps text with red ANSI color.
func Red(s string) string { return ColorRed + s + ColorReset }
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestColors" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/colors.go go/internal/utils/colors_test.go
git commit -m "feat(go): add color utilities"
```

---

### Task 4: Logger

**Files:**
- Create: `go/internal/utils/logger.go`
- Create: `go/internal/utils/logger_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/logger_test.go`:

```go
package utils_test

import (
	"os"
	"strings"
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
	// Component is padded to 10 chars
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
	// Set then clear — should not panic
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
	data := strings.Repeat("x", 200)
	log.DebugJSON("label", data, 60)
	assert.Len(t, captured, 1)
	assert.LessOrEqual(t, len(captured[0]), 120) // label + truncated JSON
}

func TestMcpLogger_HasExpectedMethods(t *testing.T) {
	assert.NotNil(t, utils.McpLogger)
	// Just verify it doesn't panic
	utils.McpLogger.Debug("test")
}

func TestWsLogger_HasExpectedMethods(t *testing.T) {
	assert.NotNil(t, utils.WsLogger)
	utils.WsLogger.Debug("test")
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestLogger" -v
```

Expected: FAIL — `undefined: utils.CreateLogger`

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/logger.go`:

```go
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
	component string // padded to 10 chars
	useStderr bool
	sessionID *string
}

// CreateLogger creates a new Logger for the given component.
// useStderr controls whether non-error output goes to stderr (default: stdout).
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestLogger|TestDebugJson|TestMcpLogger|TestWsLogger" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/logger.go go/internal/utils/logger_test.go
git commit -m "feat(go): add logger utility"
```

---

### Task 5: Session Log Utility

**Files:**
- Create: `go/internal/utils/sessionlog.go`
- Create: `go/internal/utils/sessionlog_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/sessionlog_test.go`:

```go
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

	// Should return the base logger without panicking
	result := logFn(nil)
	assert.NotNil(t, result)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestCreateSessionLog" -v
```

Expected: FAIL

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/sessionlog.go`:

```go
package utils

// SessionLike is the minimal interface for objects that have a session ID.
type SessionLike interface {
	GetSessionID() string
}

// CreateSessionLog returns a factory function that creates session-scoped loggers.
// If session is nil, returns the base logger unchanged.
func CreateSessionLog(base *Logger) func(session SessionLike) *Logger {
	return func(session SessionLike) *Logger {
		if session == nil {
			return base
		}
		return base.ForSession(session.GetSessionID())
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestCreateSessionLog" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/sessionlog.go go/internal/utils/sessionlog_test.go
git commit -m "feat(go): add session log utility"
```

---

### Task 6: Uptime Utility

**Files:**
- Create: `go/internal/utils/uptime.go`
- Create: `go/internal/utils/uptime_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/uptime_test.go`:

```go
package utils_test

import (
	"testing"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestFormatUptime(t *testing.T) {
	now := time.Now()

	// Less than a minute
	assert.Equal(t, "<1m", utils.FormatUptime(now.Add(-30*time.Second)))

	// Minutes only
	assert.Equal(t, "5m", utils.FormatUptime(now.Add(-5*time.Minute)))

	// Hours and minutes
	assert.Equal(t, "1h23m", utils.FormatUptime(now.Add(-(1*time.Hour + 23*time.Minute))))

	// Hours only (no minutes)
	assert.Equal(t, "2h", utils.FormatUptime(now.Add(-2*time.Hour)))

	// Days and hours
	assert.Equal(t, "1d5h", utils.FormatUptime(now.Add(-(1*24*time.Hour + 5*time.Hour))))

	// Days only
	assert.Equal(t, "2d", utils.FormatUptime(now.Add(-2*24*time.Hour)))
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestFormatUptime" -v
```

Expected: FAIL

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/uptime.go`:

```go
package utils

import (
	"fmt"
	"time"
)

// FormatUptime formats the duration since startedAt in a compact format.
// Examples: "<1m", "5m", "1h23m", "2h", "1d5h", "2d"
func FormatUptime(startedAt time.Time) string {
	d := time.Since(startedAt)
	totalMinutes := int(d.Minutes())
	totalHours := int(d.Hours())
	days := totalHours / 24
	hours := totalHours % 24
	minutes := totalMinutes % 60

	if totalMinutes < 1 {
		return "<1m"
	}
	if totalHours < 1 {
		return fmt.Sprintf("%dm", totalMinutes)
	}
	if days < 1 {
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestFormatUptime" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/uptime.go go/internal/utils/uptime_test.go
git commit -m "feat(go): add uptime utility"
```

---

### Task 7: PR Detector

**Files:**
- Create: `go/internal/utils/prdetector.go`
- Create: `go/internal/utils/prdetector_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/prdetector_test.go`:

```go
package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectPullRequests_GitHub(t *testing.T) {
	prs := utils.DetectPullRequests("See https://github.com/owner/repo/pull/123 for details")
	require.Len(t, prs, 1)
	assert.Equal(t, "https://github.com/owner/repo/pull/123", prs[0].URL)
	assert.Equal(t, "github", prs[0].Platform)
	assert.Equal(t, 123, prs[0].Number)
}

func TestDetectPullRequests_GitLab(t *testing.T) {
	prs := utils.DetectPullRequests("See https://gitlab.com/owner/repo/-/merge_requests/456")
	require.Len(t, prs, 1)
	assert.Equal(t, "gitlab", prs[0].Platform)
	assert.Equal(t, 456, prs[0].Number)
}

func TestDetectPullRequests_Bitbucket(t *testing.T) {
	prs := utils.DetectPullRequests("https://bitbucket.org/owner/repo/pull-requests/789")
	require.Len(t, prs, 1)
	assert.Equal(t, "bitbucket", prs[0].Platform)
	assert.Equal(t, 789, prs[0].Number)
}

func TestDetectPullRequests_Multiple(t *testing.T) {
	text := "PR1: https://github.com/a/b/pull/1 and PR2: https://github.com/a/b/pull/2"
	prs := utils.DetectPullRequests(text)
	assert.Len(t, prs, 2)
}

func TestDetectPullRequests_Deduplicates(t *testing.T) {
	text := "https://github.com/a/b/pull/1 and https://github.com/a/b/pull/1 again"
	prs := utils.DetectPullRequests(text)
	assert.Len(t, prs, 1)
}

func TestDetectPullRequests_Empty(t *testing.T) {
	prs := utils.DetectPullRequests("no PRs here")
	assert.Empty(t, prs)
}

func TestExtractPullRequestURL(t *testing.T) {
	url := utils.ExtractPullRequestURL("See https://github.com/a/b/pull/1 for details")
	assert.NotNil(t, url)
	assert.Equal(t, "https://github.com/a/b/pull/1", *url)
}

func TestExtractPullRequestURL_None(t *testing.T) {
	url := utils.ExtractPullRequestURL("no PRs here")
	assert.Nil(t, url)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestDetectPullRequests|TestExtractPullRequestURL" -v
```

Expected: FAIL

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/prdetector.go`:

```go
package utils

import (
	"regexp"
	"strconv"
)

// PullRequestInfo contains information about a detected pull/merge request.
type PullRequestInfo struct {
	URL      string
	Platform string // "github", "gitlab", "bitbucket", "azuredevops"
	Number   int
	Repo     string
}

var prPatterns = []struct {
	re       *regexp.Regexp
	platform string
	numGroup int
}{
	{
		re:       regexp.MustCompile(`https?://github\.com/[\w./-]+/pull/(\d+)`),
		platform: "github",
		numGroup: 1,
	},
	{
		re:       regexp.MustCompile(`https?://(?:[\w.-]+\.)?gitlab\.com/[\w./-]+/-/merge_requests/(\d+)`),
		platform: "gitlab",
		numGroup: 1,
	},
	{
		re:       regexp.MustCompile(`https?://bitbucket\.org/[\w./-]+/pull-requests/(\d+)`),
		platform: "bitbucket",
		numGroup: 1,
	},
	{
		re:       regexp.MustCompile(`https?://[\w.-]+\.visualstudio\.com/[\w./-]+/_git/[\w.-]+/pullrequest/(\d+)`),
		platform: "azuredevops",
		numGroup: 1,
	},
	{
		re:       regexp.MustCompile(`https?://dev\.azure\.com/[\w./-]+/_git/[\w.-]+/pullrequest/(\d+)`),
		platform: "azuredevops",
		numGroup: 1,
	},
}

// DetectPullRequests finds all PR/MR URLs in text. Deduplicates by URL.
func DetectPullRequests(text string) []PullRequestInfo {
	seen := map[string]bool{}
	var results []PullRequestInfo

	for _, p := range prPatterns {
		matches := p.re.FindAllStringSubmatch(text, -1)
		for _, m := range matches {
			url := m[0]
			if seen[url] {
				continue
			}
			seen[url] = true
			num, _ := strconv.Atoi(m[p.numGroup])
			results = append(results, PullRequestInfo{
				URL:      url,
				Platform: p.platform,
				Number:   num,
			})
		}
	}
	return results
}

// ExtractPullRequestURL returns the first PR URL found in text, or nil.
func ExtractPullRequestURL(text string) *string {
	prs := DetectPullRequests(text)
	if len(prs) == 0 {
		return nil
	}
	url := prs[0].URL
	return &url
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestDetectPullRequests|TestExtractPullRequestURL" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/prdetector.go go/internal/utils/prdetector_test.go
git commit -m "feat(go): add PR detector utility"
```

---

### Task 8: Keep-Alive Manager

**Files:**
- Create: `go/internal/utils/keepalive.go`
- Create: `go/internal/utils/keepalive_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/keepalive_test.go`:

```go
package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func newKeepAlive() *utils.KeepAliveManager {
	ka := utils.NewKeepAliveManager()
	return ka
}

func TestKeepAlive_InitialState(t *testing.T) {
	ka := newKeepAlive()
	assert.Equal(t, 0, ka.GetSessionCount())
	assert.True(t, ka.IsEnabled())
	assert.False(t, ka.IsActive())
}

func TestKeepAlive_Disable(t *testing.T) {
	ka := newKeepAlive()
	ka.SetEnabled(false)
	assert.False(t, ka.IsEnabled())
}

func TestKeepAlive_SessionCount(t *testing.T) {
	ka := newKeepAlive()
	ka.SetEnabled(false) // prevent actual subprocess spawning in tests
	ka.SessionStarted()
	assert.Equal(t, 1, ka.GetSessionCount())
	ka.SessionStarted()
	assert.Equal(t, 2, ka.GetSessionCount())
	ka.SessionEnded()
	assert.Equal(t, 1, ka.GetSessionCount())
}

func TestKeepAlive_NoNegativeCount(t *testing.T) {
	ka := newKeepAlive()
	ka.SetEnabled(false)
	ka.SessionEnded() // should not go below 0
	assert.Equal(t, 0, ka.GetSessionCount())
}

func TestKeepAlive_ForceStop(t *testing.T) {
	ka := newKeepAlive()
	ka.SetEnabled(false)
	ka.SessionStarted()
	ka.SessionStarted()
	ka.ForceStop()
	assert.Equal(t, 0, ka.GetSessionCount())
}

func TestKeepAlive_IsActiveWhenDisabled(t *testing.T) {
	ka := newKeepAlive()
	ka.SetEnabled(false)
	ka.SessionStarted()
	assert.False(t, ka.IsActive())
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestKeepAlive" -v
```

Expected: FAIL

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/keepalive.go`:

```go
package utils

import (
	"os"
	"os/exec"
	"runtime"
	"sync"
)

// KeepAliveManager prevents system sleep while sessions are active.
// Platform-specific: uses caffeinate (macOS) or systemd-inhibit (Linux).
type KeepAliveManager struct {
	mu           sync.Mutex
	enabled      bool
	sessionCount int
	cmd          *exec.Cmd
}

// NewKeepAliveManager creates a new manager. Enabled by default.
func NewKeepAliveManager() *KeepAliveManager {
	return &KeepAliveManager{enabled: true}
}

// SetEnabled enables or disables keep-alive. Stopping while active kills the process.
func (k *KeepAliveManager) SetEnabled(enabled bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.enabled = enabled
	if !enabled && k.cmd != nil {
		k.stopProcess()
	}
}

// IsEnabled returns whether keep-alive is enabled.
func (k *KeepAliveManager) IsEnabled() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.enabled
}

// IsActive returns true if the keep-alive subprocess is running.
func (k *KeepAliveManager) IsActive() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.cmd != nil
}

// SessionStarted increments the session count; starts keep-alive on first session.
func (k *KeepAliveManager) SessionStarted() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.sessionCount++
	if k.enabled && k.cmd == nil {
		k.startProcess()
	}
}

// SessionEnded decrements the session count; stops keep-alive when count reaches 0.
func (k *KeepAliveManager) SessionEnded() {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.sessionCount > 0 {
		k.sessionCount--
	}
	if k.sessionCount == 0 && k.cmd != nil {
		k.stopProcess()
	}
}

// ForceStop immediately stops keep-alive and resets session count.
func (k *KeepAliveManager) ForceStop() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.sessionCount = 0
	if k.cmd != nil {
		k.stopProcess()
	}
}

// GetSessionCount returns the current number of active sessions.
func (k *KeepAliveManager) GetSessionCount() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.sessionCount
}

// startProcess starts the platform-specific keep-alive subprocess.
// Must be called with k.mu held.
func (k *KeepAliveManager) startProcess() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("caffeinate", "-s", "-i")
	case "linux":
		if _, err := exec.LookPath("systemd-inhibit"); err == nil {
			cmd = exec.Command("systemd-inhibit", "--what=sleep:idle", "--who=claude-threads", "--why=Active sessions", "--mode=block", "sleep", "infinity")
		}
	}
	if cmd == nil {
		return
	}
	cmd.Stdout = os.Devnull
	cmd.Stderr = os.Devnull
	if err := cmd.Start(); err == nil {
		k.cmd = cmd
	}
}

// stopProcess kills the keep-alive subprocess.
// Must be called with k.mu held.
func (k *KeepAliveManager) stopProcess() {
	if k.cmd != nil && k.cmd.Process != nil {
		_ = k.cmd.Process.Kill()
		_ = k.cmd.Wait()
		k.cmd = nil
	}
}

// KeepAlive is the package-level singleton.
var KeepAlive = NewKeepAliveManager()
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestKeepAlive" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/keepalive.go go/internal/utils/keepalive_test.go
git commit -m "feat(go): add keep-alive manager"
```

---

### Task 9: Battery Utility

**Files:**
- Create: `go/internal/utils/battery.go`
- Create: `go/internal/utils/battery_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/battery_test.go`:

```go
package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetBatteryStatus_DoesNotPanic(t *testing.T) {
	// Should return a value or nil without panicking
	status, err := utils.GetBatteryStatus()
	assert.NoError(t, err)
	// status may be nil on CI/unsupported platforms
	if status != nil {
		assert.GreaterOrEqual(t, status.Percentage, 0)
		assert.LessOrEqual(t, status.Percentage, 100)
	}
}

func TestFormatBatteryStatus_DoesNotPanic(t *testing.T) {
	// Should return a string or nil without panicking
	result, err := utils.FormatBatteryStatus()
	assert.NoError(t, err)
	if result != nil {
		assert.NotEmpty(t, *result)
	}
}

func TestBatteryStatusShape(t *testing.T) {
	status := &utils.BatteryStatus{
		Percentage: 85,
		Charging:   false,
	}
	assert.Equal(t, 85, status.Percentage)
	assert.False(t, status.Charging)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestGetBatteryStatus|TestFormatBatteryStatus|TestBatteryStatusShape" -v
```

Expected: FAIL

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/battery.go`:

```go
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// BatteryStatus contains the current battery state.
type BatteryStatus struct {
	Percentage int
	Charging   bool
}

// GetBatteryStatus returns the current battery status, or nil if unavailable.
func GetBatteryStatus() (*BatteryStatus, error) {
	switch runtime.GOOS {
	case "darwin":
		return getMacBattery()
	case "linux":
		return getLinuxBattery()
	default:
		return nil, nil
	}
}

// FormatBatteryStatus returns a formatted battery string, or nil if unavailable.
func FormatBatteryStatus() (*string, error) {
	status, err := GetBatteryStatus()
	if err != nil || status == nil {
		return nil, err
	}
	var s string
	if status.Charging {
		s = "🔌 AC"
	} else {
		s = fmt.Sprintf("🔋 %d%%", status.Percentage)
	}
	return &s, nil
}

var macBatteryPctRe = regexp.MustCompile(`(\d+)%`)
var macChargingRe = regexp.MustCompile(`(?i)(charging|AC attached)`)

func getMacBattery() (*BatteryStatus, error) {
	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return nil, nil //nolint:nilerr // not an error if pmset unavailable
	}
	s := string(out)
	m := macBatteryPctRe.FindStringSubmatch(s)
	if m == nil {
		return nil, nil
	}
	pct, _ := strconv.Atoi(m[1])
	charging := macChargingRe.MatchString(s)
	return &BatteryStatus{Percentage: pct, Charging: charging}, nil
}

func getLinuxBattery() (*BatteryStatus, error) {
	for _, name := range []string{"BAT0", "BAT1", "battery"} {
		base := "/sys/class/power_supply/" + name
		capPath := base + "/capacity"
		statusPath := base + "/status"

		capBytes, err := os.ReadFile(capPath)
		if err != nil {
			continue
		}
		pct, err := strconv.Atoi(strings.TrimSpace(string(capBytes)))
		if err != nil {
			continue
		}
		charging := false
		if statusBytes, err := os.ReadFile(statusPath); err == nil {
			status := strings.TrimSpace(string(statusBytes))
			charging = status == "Charging" || status == "Full"
		}
		return &BatteryStatus{Percentage: pct, Charging: charging}, nil
	}
	return nil, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestGetBatteryStatus|TestFormatBatteryStatus|TestBatteryStatusShape" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/battery.go go/internal/utils/battery_test.go
git commit -m "feat(go): add battery utility"
```

---

### Task 10: Config Package

**Files:**
- Create: `go/internal/config/config.go`
- Create: `go/internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/config/config_test.go`:

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anneschuth/claude-threads/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPath_Defined(t *testing.T) {
	assert.NotEmpty(t, config.ConfigPath)
	assert.Contains(t, config.ConfigPath, "config.yaml")
}

func TestResolveLimits_Defaults(t *testing.T) {
	limits := config.ResolveLimits(nil)
	assert.Equal(t, 5, limits.MaxSessions)
	assert.Equal(t, 30, limits.SessionTimeoutMinutes)
	assert.Equal(t, 5, limits.SessionWarningMinutes)
	assert.Equal(t, 60, limits.CleanupIntervalMinutes)
	assert.Equal(t, 24, limits.MaxWorktreeAgeHours)
	assert.True(t, limits.CleanupWorktrees)
	assert.Equal(t, 120, limits.PermissionTimeoutSeconds)
}

func TestResolveLimits_MergesOverrides(t *testing.T) {
	maxSessions := 10
	limits := config.ResolveLimits(&config.LimitsConfig{MaxSessions: &maxSessions})
	assert.Equal(t, 10, limits.MaxSessions)
	assert.Equal(t, 30, limits.SessionTimeoutMinutes) // default preserved
}

func TestResolveLimits_LegacyEnvMaxSessions(t *testing.T) {
	os.Setenv("MAX_SESSIONS", "15")
	defer os.Unsetenv("MAX_SESSIONS")
	limits := config.ResolveLimits(nil)
	assert.Equal(t, 15, limits.MaxSessions)
}

func TestSaveConfig_CreatesFileWithCorrectPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &config.Config{
		Version:    1,
		WorkingDir: "/tmp",
		Chrome:     false,
		Platforms:  []config.PlatformInstanceConfig{},
	}

	err := config.SaveConfig(cfg, path)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestSaveConfig_WritesValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &config.Config{
		Version:    1,
		WorkingDir: "/home/user",
		Chrome:     true,
		Platforms:  []config.PlatformInstanceConfig{},
	}

	require.NoError(t, config.SaveConfig(cfg, path))

	loaded, err := config.LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, "/home/user", loaded.WorkingDir)
	assert.True(t, loaded.Chrome)
}

func TestLoadConfig_ReturnsNilForMissingFile(t *testing.T) {
	loaded, err := config.LoadConfig("/nonexistent/path/config.yaml")
	assert.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestConfigExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	assert.False(t, config.ConfigExistsAt(path))

	require.NoError(t, os.WriteFile(path, []byte("version: 1\n"), 0o600))
	assert.True(t, config.ConfigExistsAt(path))
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/config/ -v
```

Expected: FAIL — package not found

- [ ] **Step 3: Write implementation**

Create `go/internal/config/config.go`:

```go
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// WorktreeMode controls git worktree behavior.
type WorktreeMode string

const (
	WorktreeModeOff     WorktreeMode = "off"
	WorktreeModePrompt  WorktreeMode = "prompt"
	WorktreeModeRequire WorktreeMode = "require"
)

// ThreadLogsConfig controls thread log retention.
type ThreadLogsConfig struct {
	Enabled       *bool `yaml:"enabled,omitempty"`
	RetentionDays *int  `yaml:"retentionDays,omitempty"`
}

// LimitsConfig controls resource limits. All fields are optional (nil = use default).
type LimitsConfig struct {
	MaxSessions              *int  `yaml:"maxSessions,omitempty"`
	SessionTimeoutMinutes    *int  `yaml:"sessionTimeoutMinutes,omitempty"`
	SessionWarningMinutes    *int  `yaml:"sessionWarningMinutes,omitempty"`
	CleanupIntervalMinutes   *int  `yaml:"cleanupIntervalMinutes,omitempty"`
	MaxWorktreeAgeHours      *int  `yaml:"maxWorktreeAgeHours,omitempty"`
	CleanupWorktrees         *bool `yaml:"cleanupWorktrees,omitempty"`
	PermissionTimeoutSeconds *int  `yaml:"permissionTimeoutSeconds,omitempty"`
}

// ResolvedLimits contains all limits with defaults applied.
type ResolvedLimits struct {
	MaxSessions              int
	SessionTimeoutMinutes    int
	SessionWarningMinutes    int
	CleanupIntervalMinutes   int
	MaxWorktreeAgeHours      int
	CleanupWorktrees         bool
	PermissionTimeoutSeconds int
}

// ResolveLimits merges LimitsConfig with defaults.
// Also reads legacy environment variables: MAX_SESSIONS, SESSION_TIMEOUT_MS.
func ResolveLimits(limits *LimitsConfig) ResolvedLimits {
	r := ResolvedLimits{
		MaxSessions:              5,
		SessionTimeoutMinutes:    30,
		SessionWarningMinutes:    5,
		CleanupIntervalMinutes:   60,
		MaxWorktreeAgeHours:      24,
		CleanupWorktrees:         true,
		PermissionTimeoutSeconds: 120,
	}

	// Legacy env vars
	if v := os.Getenv("MAX_SESSIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			r.MaxSessions = n
		}
	}
	if v := os.Getenv("SESSION_TIMEOUT_MS"); v != "" {
		if ms, err := strconv.ParseInt(v, 10, 64); err == nil {
			r.SessionTimeoutMinutes = int(ms / 60000)
		}
	}

	if limits == nil {
		return r
	}
	if limits.MaxSessions != nil {
		r.MaxSessions = *limits.MaxSessions
	}
	if limits.SessionTimeoutMinutes != nil {
		r.SessionTimeoutMinutes = *limits.SessionTimeoutMinutes
	}
	if limits.SessionWarningMinutes != nil {
		r.SessionWarningMinutes = *limits.SessionWarningMinutes
	}
	if limits.CleanupIntervalMinutes != nil {
		r.CleanupIntervalMinutes = *limits.CleanupIntervalMinutes
	}
	if limits.MaxWorktreeAgeHours != nil {
		r.MaxWorktreeAgeHours = *limits.MaxWorktreeAgeHours
	}
	if limits.CleanupWorktrees != nil {
		r.CleanupWorktrees = *limits.CleanupWorktrees
	}
	if limits.PermissionTimeoutSeconds != nil {
		r.PermissionTimeoutSeconds = *limits.PermissionTimeoutSeconds
	}
	return r
}

// StickyMessageCustomization customizes the sticky channel message.
type StickyMessageCustomization struct {
	Description string `yaml:"description,omitempty"`
	Footer      string `yaml:"footer,omitempty"`
}

// AutoUpdateConfig controls auto-update behavior.
type AutoUpdateConfig struct {
	Enabled         *bool  `yaml:"enabled,omitempty"`
	Channel         string `yaml:"channel,omitempty"`
	CheckIntervalMs *int64 `yaml:"checkIntervalMs,omitempty"`
}

// PlatformInstanceConfig is the base config for any platform instance.
type PlatformInstanceConfig struct {
	ID          string `yaml:"id"`
	Type        string `yaml:"type"` // "mattermost" | "slack"
	DisplayName string `yaml:"displayName"`
}

// MattermostPlatformConfig is config for a Mattermost instance.
type MattermostPlatformConfig struct {
	PlatformInstanceConfig `yaml:",inline"`
	URL                    string   `yaml:"url"`
	Token                  string   `yaml:"token"`
	ChannelID              string   `yaml:"channelId"`
	BotName                string   `yaml:"botName"`
	AllowedUsers           []string `yaml:"allowedUsers"`
	SkipPermissions        bool     `yaml:"skipPermissions"`
}

// SlackPlatformConfig is config for a Slack instance.
type SlackPlatformConfig struct {
	PlatformInstanceConfig `yaml:",inline"`
	BotToken               string   `yaml:"botToken"`
	AppToken               string   `yaml:"appToken"`
	ChannelID              string   `yaml:"channelId"`
	BotName                string   `yaml:"botName"`
	AllowedUsers           []string `yaml:"allowedUsers"`
	SkipPermissions        bool     `yaml:"skipPermissions"`
	APIURL                 string   `yaml:"apiUrl,omitempty"`
}

// Config is the top-level configuration structure.
// Maps exactly to config.yaml format used by the TypeScript version.
type Config struct {
	Version       int                          `yaml:"version"`
	WorkingDir    string                       `yaml:"workingDir"`
	Chrome        bool                         `yaml:"chrome"`
	WorktreeMode  WorktreeMode                 `yaml:"worktreeMode"`
	KeepAlive     *bool                        `yaml:"keepAlive,omitempty"`
	AutoUpdate    *AutoUpdateConfig            `yaml:"autoUpdate,omitempty"`
	ThreadLogs    *ThreadLogsConfig            `yaml:"threadLogs,omitempty"`
	Limits        *LimitsConfig                `yaml:"limits,omitempty"`
	StickyMessage *StickyMessageCustomization  `yaml:"stickyMessage,omitempty"`
	Platforms     []PlatformInstanceConfig     `yaml:"platforms"`
}

// ConfigPath is the default path to the config file.
var ConfigPath = filepath.Join(mustHomeDir(), ".config", "claude-threads", "config.yaml")

func mustHomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return h
}

// LoadConfig loads a Config from the given YAML file path.
// Returns nil (no error) if the file does not exist.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes cfg as YAML to path with 0600 permissions.
// Creates parent directories with 0700 permissions if needed.
func SaveConfig(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// Fix directory permissions if wrong
	if info, err := os.Stat(dir); err == nil && info.Mode().Perm() != 0o700 {
		_ = os.Chmod(dir, 0o700)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	// Fix file permissions if WriteFile used umask-affected mode
	return os.Chmod(path, 0o600)
}

// ConfigExistsAt returns true if a config file exists at path.
func ConfigExistsAt(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ConfigExists returns true if the default config file exists.
func ConfigExists() bool {
	return ConfigExistsAt(ConfigPath)
}

// LoadDefaultConfig loads the config from the default path.
func LoadDefaultConfig() (*Config, error) {
	return LoadConfig(ConfigPath)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/config/ -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/config/
git commit -m "feat(go): add config package"
```

---

### Task 11: Error Utilities (Severity Types + logSilentError)

> Note: `handleError` and `withErrorHandling` depend on `Session` type (Phase 11). Only the session-agnostic parts belong here.

**Files:**
- Create: `go/internal/utils/errors.go`
- Create: `go/internal/utils/errors_test.go`

- [ ] **Step 1: Write the failing test**

Create `go/internal/utils/errors_test.go`:

```go
package utils_test

import (
	"errors"
	"os"
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestLogSilentError_DoesNotPanic(t *testing.T) {
	// logSilentError should never panic, even with nil-like errors
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestLogSilentError|TestErrorSeverity" -v
```

Expected: FAIL — `undefined: utils.LogSilentError`

- [ ] **Step 3: Write implementation**

Create `go/internal/utils/errors.go`:

```go
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
//
// Example:
//
//	if err := platform.RemoveReaction(postID, emoji); err != nil {
//	    utils.LogSilentError("remove-bot-reaction", err)
//	}
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/ -run "TestLogSilentError|TestErrorSeverity" -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add go/internal/utils/errors.go go/internal/utils/errors_test.go
git commit -m "feat(go): add error severity types and logSilentError utility"
```

> Note: `spawn.ts` is a Windows-only shim for Node.js subprocess spawning. Go's `os/exec` handles cross-platform spawning natively — no equivalent file needed.
>
> Note: `websocket.ts` and `error-handler` session-aware functions (`handleError`, `withErrorHandling`) are translated in Phase 3 (Mattermost) and Phase 11 (Session) respectively, where their dependencies exist.

---

### Task 12: Run Full Phase 1 Test Suite

**Files:** No new files

- [ ] **Step 1: Run all utils and config tests**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go test ./internal/utils/... ./internal/config/... -v
```

Expected: All tests PASS with no failures

- [ ] **Step 2: Run go vet**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/go
go vet ./internal/utils/... ./internal/config/...
```

Expected: No output (success)

- [ ] **Step 3: Update master plan progress table**

In `docs/superpowers/plans/2026-03-31-golang-conversion-master.md`, change Phase 1 row:
```
| 1 | [phase-01-foundation.md](./2026-03-31-golang-phase-01-foundation.md) | ✅ Complete | `utils/`, `config/` — no external deps |
```

- [ ] **Step 4: Update golang-conversion.md progress table**

In `docs/plans/golang-conversion.md`, change Phase 1 row:
```
| 1: Foundation | ✅ Complete | |
```

- [ ] **Step 5: Final commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads
git add docs/plans/golang-conversion.md docs/superpowers/plans/2026-03-31-golang-conversion-master.md
git commit -m "docs: mark Phase 1 (foundation) complete"
```
