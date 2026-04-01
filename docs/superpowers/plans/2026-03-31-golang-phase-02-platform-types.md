# Phase 2: Platform Types & Interfaces — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Define the platform abstraction layer — interfaces, normalized types, and platform utilities. No concrete implementations (Mattermost/Slack come in later phases).

**Architecture:** Pure interfaces + one utility file with real logic (tested). The TypeScript `PlatformClient extends EventEmitter` pattern maps to Go callback registration methods (`OnMessage`, `OnReaction`, etc.) — type-safe and idiomatic.

**Tech Stack:** Go 1.24, Testify, standard library only (no external deps for types/utils)

**Working directory:** `/Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion/go`

**Run all go commands with:** `env -u GOROOT /opt/homebrew/bin/go`

---

## Files

| Action | Path | Description |
|--------|------|-------------|
| Create | `internal/platform/types.go` | PlatformUser, PlatformPost, PlatformReaction, PlatformFile, ThreadMessage |
| Create | `internal/platform/client.go` | PlatformClient interface (callback-based, replaces EventEmitter) |
| Create | `internal/platform/formatter.go` | PlatformFormatter interface |
| Create | `internal/platform/permissionapi.go` | PermissionApi interface + config structs |
| Create | `internal/platform/utils.go` | escapeRegExp, getPlatformIcon, truncateMessageSafely, normalizeEmojiName, getEmojiName, convertMarkdownToSlack, convertMarkdownTablesToSlack |
| Create | `internal/platform/utils_test.go` | All utils tests (translate from utils.test.ts) |
| Create | `internal/platform/mock_client.go` | MockPlatformClient struct for use in other packages' tests |

---

## Task 1: Platform types

**Files:**
- Create: `internal/platform/types.go`

- [ ] **Step 1: Write types.go**

```go
// Package platform defines platform-agnostic types and interfaces for
// multi-platform support. All platform implementations (Mattermost, Slack)
// must satisfy the interfaces defined here.
package platform

// PlatformUser is a normalized user representation across platforms.
type PlatformUser struct {
	ID          string  // Platform-specific user ID
	Username    string  // Login username (e.g., 'alice.smith')
	DisplayName string  // Human-friendly name (e.g., 'Alice Smith')
	Email       string  // Optional email
}

// PlatformFile is a normalized file attachment representation.
type PlatformFile struct {
	ID        string // Platform-specific file ID
	Name      string // Filename
	Size      int64  // File size in bytes
	MimeType  string // MIME type (e.g., 'image/png')
	Extension string // File extension
}

// PlatformPost is a normalized post/message representation.
type PlatformPost struct {
	ID         string            // Platform-specific post ID
	PlatformID string            // Which platform instance this is from
	ChannelID  string            // Channel/conversation ID
	UserID     string            // Author's user ID
	Message    string            // Message text content
	RootID     string            // Thread parent ID (empty if channel-level)
	CreateAt   int64             // Timestamp (ms since epoch)
	Files      []PlatformFile    // Attached files (may be empty)
	Metadata   map[string]any    // Platform-specific metadata
}

// PlatformReaction is a normalized reaction representation.
type PlatformReaction struct {
	UserID    string // User who reacted
	PostID    string // Post that was reacted to
	EmojiName string // Emoji name (e.g., '+1', 'white_check_mark')
	CreateAt  int64  // When the reaction was added (ms since epoch)
}

// ThreadMessage is a normalized thread message for context retrieval.
type ThreadMessage struct {
	ID       string // Message/post ID
	UserID   string // Author's user ID
	Username string // Author's username
	Message  string // Message content
	CreateAt int64  // Timestamp (ms since epoch)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
env -u GOROOT /opt/homebrew/bin/go build ./internal/platform/...
```
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/platform/types.go
git commit -m "feat(go): add platform types (Phase 2 Task 1)"
```

---

## Task 2: PlatformFormatter interface

**Files:**
- Create: `internal/platform/formatter.go`

- [ ] **Step 1: Write formatter.go**

```go
package platform

// PlatformFormatter abstracts markdown dialect differences between platforms.
// Mattermost uses standard markdown (**bold**); Slack uses mrkdwn (*bold*).
type PlatformFormatter interface {
	// FormatBold formats text as bold.
	// Mattermost: **text** / Slack: *text*
	FormatBold(text string) string

	// FormatItalic formats text as italic.
	// Both: _text_
	FormatItalic(text string) string

	// FormatCode formats text as inline code.
	// Both: `code`
	FormatCode(text string) string

	// FormatCodeBlock formats text as a fenced code block.
	// Both: ```lang\ncode\n```
	FormatCodeBlock(code, language string) string

	// FormatUserMention formats a user mention.
	// Mattermost: @username / Slack: <@U123456>
	FormatUserMention(username, userID string) string

	// FormatLink formats a hyperlink.
	// Mattermost: [text](url) / Slack: <url|text>
	FormatLink(text, url string) string

	// FormatListItem formats a bulleted list item.
	FormatListItem(text string) string

	// FormatNumberedListItem formats a numbered list item.
	FormatNumberedListItem(number int, text string) string

	// FormatBlockquote formats a blockquote.
	FormatBlockquote(text string) string

	// FormatHorizontalRule returns a horizontal rule string.
	FormatHorizontalRule() string

	// FormatHeading formats a heading at the given level (1-6).
	FormatHeading(text string, level int) string

	// FormatStrikethrough formats text as strikethrough.
	// Mattermost: ~~text~~ / Slack: ~text~
	FormatStrikethrough(text string) string

	// EscapeText escapes special characters to prevent formatting.
	EscapeText(text string) string

	// FormatTable formats a table.
	// Mattermost: standard markdown table / Slack: key-value list
	FormatTable(headers []string, rows [][]string) string

	// FormatKeyValueList formats a key-value list with icons.
	// items is a slice of [icon, label, value] triples.
	FormatKeyValueList(items [][3]string) string

	// FormatMarkdown converts standard markdown to this platform's format.
	FormatMarkdown(content string) string
}
```

- [ ] **Step 2: Build check**

```bash
env -u GOROOT /opt/homebrew/bin/go build ./internal/platform/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/platform/formatter.go
git commit -m "feat(go): add PlatformFormatter interface (Phase 2 Task 2)"
```

---

## Task 3: PermissionApi interface

**Files:**
- Create: `internal/platform/permissionapi.go`

- [ ] **Step 1: Write permissionapi.go**

```go
package platform

// ReactionEvent is a reaction received from a WebSocket event.
type ReactionEvent struct {
	PostID    string
	UserID    string
	EmojiName string
}

// PostedMessage holds the ID of a newly created post.
type PostedMessage struct {
	ID string
}

// PermissionApi is the interface used by the MCP permission server to post
// permission requests and receive user responses via reactions.
// Each platform has its own implementation.
type PermissionApi interface {
	// GetFormatter returns the markdown formatter for this platform.
	GetFormatter() PlatformFormatter

	// GetBotUserID returns the bot's user ID.
	GetBotUserID(ctx context.Context) (string, error)

	// GetUsername returns the username for a user ID, or empty string if not found.
	GetUsername(ctx context.Context, userID string) (string, error)

	// IsUserAllowed reports whether username is in the allowed users list.
	IsUserAllowed(username string) bool

	// CreateInteractivePost creates a post with reaction options.
	CreateInteractivePost(ctx context.Context, message string, reactions []string, threadID string) (PostedMessage, error)

	// UpdatePost updates an existing post's text.
	UpdatePost(ctx context.Context, postID, message string) error

	// WaitForReaction waits up to timeoutMs for a reaction on postID.
	// Returns nil if no reaction arrives before the timeout.
	WaitForReaction(ctx context.Context, postID, botUserID string, timeoutMs int64) (*ReactionEvent, error)
}

// MattermostPermissionApiConfig is the config for the Mattermost permission API.
type MattermostPermissionApiConfig struct {
	URL          string
	Token        string
	ChannelID    string
	ThreadID     string // optional
	AllowedUsers []string
	Debug        bool
}

// SlackPermissionApiConfig is the config for the Slack permission API.
type SlackPermissionApiConfig struct {
	BotToken     string // xoxb-...
	AppToken     string // xapp-...
	ChannelID    string
	ThreadTs     string // optional thread timestamp
	AllowedUsers []string
	Debug        bool
}
```

Note: `context.Context` requires `"context"` import. Add it.

Full file with import:

```go
package platform

import "context"

// ReactionEvent is a reaction received from a WebSocket event.
type ReactionEvent struct {
	PostID    string
	UserID    string
	EmojiName string
}

// PostedMessage holds the ID of a newly created post.
type PostedMessage struct {
	ID string
}

// PermissionApi is the interface used by the MCP permission server to post
// permission requests and receive user responses via reactions.
type PermissionApi interface {
	GetFormatter() PlatformFormatter
	GetBotUserID(ctx context.Context) (string, error)
	GetUsername(ctx context.Context, userID string) (string, error)
	IsUserAllowed(username string) bool
	CreateInteractivePost(ctx context.Context, message string, reactions []string, threadID string) (PostedMessage, error)
	UpdatePost(ctx context.Context, postID, message string) error
	WaitForReaction(ctx context.Context, postID, botUserID string, timeoutMs int64) (*ReactionEvent, error)
}

// MattermostPermissionApiConfig holds config for the Mattermost permission API.
type MattermostPermissionApiConfig struct {
	URL          string
	Token        string
	ChannelID    string
	ThreadID     string
	AllowedUsers []string
	Debug        bool
}

// SlackPermissionApiConfig holds config for the Slack permission API.
type SlackPermissionApiConfig struct {
	BotToken     string
	AppToken     string
	ChannelID    string
	ThreadTs     string
	AllowedUsers []string
	Debug        bool
}
```

- [ ] **Step 2: Build check**

```bash
env -u GOROOT /opt/homebrew/bin/go build ./internal/platform/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/platform/permissionapi.go
git commit -m "feat(go): add PermissionApi interface (Phase 2 Task 3)"
```

---

## Task 4: PlatformClient interface

**Files:**
- Create: `internal/platform/client.go`

**Design note:** TypeScript's `PlatformClient extends EventEmitter` uses runtime string events. In Go, we use typed callback registration methods (`OnMessage`, `OnReaction`, etc.) for compile-time safety.

- [ ] **Step 1: Write client.go**

```go
package platform

import "context"

// MessageLimits holds platform-specific message size constraints.
type MessageLimits struct {
	MaxLength     int // Absolute max characters
	HardThreshold int // When to force continuation
}

// McpConfig holds config for the MCP permission server.
type McpConfig struct {
	Type         string
	URL          string
	Token        string
	ChannelID    string
	AllowedUsers []string
}

// ThreadHistoryOptions controls how thread history is fetched.
type ThreadHistoryOptions struct {
	Limit             int
	ExcludeBotMessages bool
}

// PlatformClient is the platform-agnostic client interface.
// All platform implementations (Mattermost, Slack) must satisfy this interface.
//
// Events are delivered via callbacks registered with On* methods.
// Each On* method replaces the previous callback (not additive).
type PlatformClient interface {
	// Identity

	PlatformID() string   // e.g., 'mattermost-internal'
	PlatformType() string // e.g., 'mattermost', 'slack'
	DisplayName() string  // e.g., 'Internal Team'

	// Connection Management

	Connect(ctx context.Context) error
	Disconnect()
	// PrepareForReconnect resets internal state (intentionalDisconnect flag,
	// reconnect attempts) so that Connect() can be called again.
	PrepareForReconnect()

	// User Management

	GetBotUser(ctx context.Context) (*PlatformUser, error)
	GetUser(ctx context.Context, userID string) (*PlatformUser, error)
	GetUserByUsername(ctx context.Context, username string) (*PlatformUser, error)
	IsUserAllowed(username string) bool
	GetBotName() string
	GetMcpConfig() McpConfig
	GetFormatter() PlatformFormatter
	// GetThreadLink returns a clickable URL for the thread.
	GetThreadLink(threadID, lastMessageID, lastMessageTs string) string

	// Messaging

	CreatePost(ctx context.Context, message, threadID string) (*PlatformPost, error)
	UpdatePost(ctx context.Context, postID, message string) (*PlatformPost, error)
	CreateInteractivePost(ctx context.Context, message string, reactions []string, threadID string) (*PlatformPost, error)
	GetPost(ctx context.Context, postID string) (*PlatformPost, error)
	DeletePost(ctx context.Context, postID string) error
	PinPost(ctx context.Context, postID string) error
	UnpinPost(ctx context.Context, postID string) error
	GetPinnedPosts(ctx context.Context) ([]string, error)
	GetMessageLimits() MessageLimits
	GetThreadHistory(ctx context.Context, threadID string, opts *ThreadHistoryOptions) ([]ThreadMessage, error)

	// Reactions

	AddReaction(ctx context.Context, postID, emojiName string) error
	RemoveReaction(ctx context.Context, postID, emojiName string) error

	// Bot Mentions

	IsBotMentioned(message string) bool
	ExtractPrompt(message string) string

	// Typing Indicator

	SendTyping(threadID string)

	// Files (optional — may be nil for platforms that don't support it)

	DownloadFile(ctx context.Context, fileID string) ([]byte, error)
	GetFileInfo(ctx context.Context, fileID string) (*PlatformFile, error)

	// Event Callbacks
	// Each On* method replaces the previous callback; pass nil to remove.

	OnConnected(func())
	OnDisconnected(func())
	OnReconnecting(func(attempt int))
	OnError(func(err error))
	OnMessage(func(post PlatformPost, user *PlatformUser))
	OnReaction(func(reaction PlatformReaction, user *PlatformUser))
	OnReactionRemoved(func(reaction PlatformReaction, user *PlatformUser))
	OnChannelPost(func(post PlatformPost, user *PlatformUser))
}
```

- [ ] **Step 2: Build check**

```bash
env -u GOROOT /opt/homebrew/bin/go build ./internal/platform/...
```

- [ ] **Step 3: Commit**

```bash
git add internal/platform/client.go
git commit -m "feat(go): add PlatformClient interface (Phase 2 Task 4)"
```

---

## Task 5: Platform utilities — tests first

**Files:**
- Create: `internal/platform/utils_test.go`
- Create: `internal/platform/utils.go`

- [ ] **Step 1: Write utils_test.go (RED)**

```go
package platform_test

import (
	"strings"
	"testing"

	"github.com/anneschuth/claude-threads/internal/platform"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GetPlatformIcon
// ---------------------------------------------------------------------------

func TestGetPlatformIcon_Slack(t *testing.T) {
	assert.Equal(t, "🆂 ", platform.GetPlatformIcon("slack"))
}

func TestGetPlatformIcon_Mattermost(t *testing.T) {
	assert.Equal(t, "𝓜 ", platform.GetPlatformIcon("mattermost"))
}

func TestGetPlatformIcon_Default(t *testing.T) {
	assert.Equal(t, "💬 ", platform.GetPlatformIcon("unknown"))
	assert.Equal(t, "💬 ", platform.GetPlatformIcon(""))
}

// ---------------------------------------------------------------------------
// TruncateMessageSafely
// ---------------------------------------------------------------------------

func TestTruncateMessageSafely_WithinLimit(t *testing.T) {
	assert.Equal(t, "hello", platform.TruncateMessageSafely("hello", 100, ""))
}

func TestTruncateMessageSafely_DefaultIndicator(t *testing.T) {
	result := platform.TruncateMessageSafely(strings.Repeat("a", 200), 100, "")
	assert.Contains(t, result, "... (truncated)")
	assert.LessOrEqual(t, len(result), 100)
}

func TestTruncateMessageSafely_CustomIndicator(t *testing.T) {
	result := platform.TruncateMessageSafely(strings.Repeat("a", 200), 100, "_truncated_")
	assert.Contains(t, result, "_truncated_")
	assert.LessOrEqual(t, len(result), 100)
}

func TestTruncateMessageSafely_ClosesOpenCodeBlock(t *testing.T) {
	content := "```javascript\nconst x = 1;\nconst y = 2;\n" + strings.Repeat("a", 200)
	result := platform.TruncateMessageSafely(content, 100, "")
	markers := strings.Count(result, "```")
	assert.Equal(t, 0, markers%2, "code block markers should be even (properly closed)")
	assert.Contains(t, result, "... (truncated)")
}

func TestTruncateMessageSafely_AlreadyClosedCodeBlock(t *testing.T) {
	content := "```javascript\nconst x = 1;\n```\n\nSome text after\n" + strings.Repeat("a", 200)
	result := platform.TruncateMessageSafely(content, 100, "")
	markers := strings.Count(result, "```")
	assert.Equal(t, 0, markers%2, "code block markers should be even")
}

func TestTruncateMessageSafely_MultipleCodeBlocksLastOpen(t *testing.T) {
	content := "```js\ncode1\n```\n\nText\n\n```python\ncode2\n" + strings.Repeat("a", 200)
	result := platform.TruncateMessageSafely(content, 120, "")
	markers := strings.Count(result, "```")
	assert.Equal(t, 0, markers%2, "code block markers should be even")
}

func TestTruncateMessageSafely_NoCodeBlocks(t *testing.T) {
	content := "Just plain text without any code blocks " + strings.Repeat("a", 200)
	result := platform.TruncateMessageSafely(content, 100, "")
	assert.NotContains(t, result, "```")
	assert.Contains(t, result, "... (truncated)")
}

// ---------------------------------------------------------------------------
// NormalizeEmojiName
// ---------------------------------------------------------------------------

func TestNormalizeEmojiName_Aliases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"thumbsup", "+1"},
		{"thumbs_up", "+1"},
		{"thumbsdown", "-1"},
		{"thumbs_down", "-1"},
		{"heavy_check_mark", "white_check_mark"},
		{"pause_button", "pause"},
		{"double_vertical_bar", "pause"},
		{"stop_button", "stop"},
		{"octagonal_sign", "stop"},
		{"1", "one"},
		{"2", "two"},
		{"3", "three"},
		{"4", "four"},
		{"5", "five"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, platform.NormalizeEmojiName(tt.input))
		})
	}
}

func TestNormalizeEmojiName_RemovesColons(t *testing.T) {
	assert.Equal(t, "+1", platform.NormalizeEmojiName(":thumbsup:"))
	assert.Equal(t, "white_check_mark", platform.NormalizeEmojiName(":white_check_mark:"))
}

func TestNormalizeEmojiName_PassthroughUnknown(t *testing.T) {
	assert.Equal(t, "some_emoji", platform.NormalizeEmojiName("some_emoji"))
}

// ---------------------------------------------------------------------------
// GetEmojiName
// ---------------------------------------------------------------------------

func TestGetEmojiName_UnicodeMappings(t *testing.T) {
	tests := []struct {
		emoji    string
		expected string
	}{
		{"👍", "+1"},
		{"👎", "-1"},
		{"✅", "white_check_mark"},
		{"❌", "x"},
		{"🛑", "stop"},
		{"⏸️", "pause"},
		{"1️⃣", "one"},
		{"2️⃣", "two"},
	}
	for _, tt := range tests {
		t.Run(tt.emoji, func(t *testing.T) {
			assert.Equal(t, tt.expected, platform.GetEmojiName(tt.emoji))
		})
	}
}

func TestGetEmojiName_PassthroughShortcodes(t *testing.T) {
	assert.Equal(t, "+1", platform.GetEmojiName("+1"))
	assert.Equal(t, "white_check_mark", platform.GetEmojiName("white_check_mark"))
}

// ---------------------------------------------------------------------------
// ConvertMarkdownTablesToSlack
// ---------------------------------------------------------------------------

func TestConvertMarkdownTablesToSlack_Basic(t *testing.T) {
	input := "| Header1 | Header2 |\n|---------|----------|\n| Cell1   | Cell2   |\n"
	result := platform.ConvertMarkdownTablesToSlack(input)
	assert.Contains(t, result, "*Header1:* Cell1")
	assert.Contains(t, result, "*Header2:* Cell2")
	assert.NotContains(t, result, "|")
}

func TestConvertMarkdownTablesToSlack_NoTable(t *testing.T) {
	input := "Just plain text"
	assert.Equal(t, input, platform.ConvertMarkdownTablesToSlack(input))
}

// ---------------------------------------------------------------------------
// ConvertMarkdownToSlack
// ---------------------------------------------------------------------------

func TestConvertMarkdownToSlack_Bold(t *testing.T) {
	result := platform.ConvertMarkdownToSlack("**bold text**")
	assert.Equal(t, "*bold text*", result)
}

func TestConvertMarkdownToSlack_Headers(t *testing.T) {
	result := platform.ConvertMarkdownToSlack("## My Header")
	assert.Equal(t, "*My Header*", result)
}

func TestConvertMarkdownToSlack_Links(t *testing.T) {
	result := platform.ConvertMarkdownToSlack("[click here](https://example.com)")
	assert.Equal(t, "<https://example.com|click here>", result)
}

func TestConvertMarkdownToSlack_HorizontalRule(t *testing.T) {
	result := platform.ConvertMarkdownToSlack("---")
	assert.Contains(t, result, "━")
}

func TestConvertMarkdownToSlack_PreservesCodeBlocks(t *testing.T) {
	input := "```go\n**not bold**\n```"
	result := platform.ConvertMarkdownToSlack(input)
	assert.Contains(t, result, "**not bold**")
}

func TestConvertMarkdownToSlack_PreservesInlineCode(t *testing.T) {
	input := "Use `**literal**` here"
	result := platform.ConvertMarkdownToSlack(input)
	assert.Contains(t, result, "`**literal**`")
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/... -v 2>&1 | head -30
```
Expected: compile error (functions not yet defined)

- [ ] **Step 3: Write utils.go (GREEN)**

```go
package platform

import (
	"regexp"
	"strings"
)

// escapeRegexpSpecials is used internally.
var escapeRegexpSpecials = regexp.MustCompile(`[.*+?^${}()|[\]\\]`)

// EscapeRegExp escapes special regex characters in s.
func EscapeRegExp(s string) string {
	return escapeRegexpSpecials.ReplaceAllString(s, `\$0`)
}

// GetPlatformIcon returns the display icon for a platform type.
func GetPlatformIcon(platformType string) string {
	switch platformType {
	case "slack":
		return "🆂 "
	case "mattermost":
		return "𝓜 "
	default:
		return "💬 "
	}
}

// TruncateMessageSafely truncates message to maxLength, closing any open code
// block before appending the truncation indicator.
// If indicator is empty, "... (truncated)" is used.
func TruncateMessageSafely(message string, maxLength int, indicator string) string {
	if len(message) <= maxLength {
		return message
	}
	if indicator == "" {
		indicator = "... (truncated)"
	}
	// Reserve space for possible closing ``` (4), separator "\n\n" (2), indicator.
	reserved := 4 + 2 + len(indicator)
	cutAt := maxLength - reserved
	if cutAt < 0 {
		cutAt = 0
	}
	truncated := message[:cutAt]

	// Check for unclosed code block.
	count := strings.Count(truncated, "```")
	if count%2 == 1 {
		truncated += "\n```"
	}

	return truncated + "\n\n" + indicator
}

// emojiAliases maps alternative emoji names to canonical names.
var emojiAliases = map[string]string{
	"thumbsup":                "+1",
	"thumbs_up":               "+1",
	"thumbsdown":              "-1",
	"thumbs_down":             "-1",
	"heavy_check_mark":        "white_check_mark",
	"x":                       "x",
	"cross_mark":              "x",
	"heavy_multiplication_x":  "x",
	"pause_button":            "pause",
	"double_vertical_bar":     "pause",
	"play_button":             "arrow_forward",
	"stop_button":             "stop",
	"octagonal_sign":          "stop",
	"1":                       "one",
	"2":                       "two",
	"3":                       "three",
	"4":                       "four",
	"5":                       "five",
}

// NormalizeEmojiName normalizes an emoji name, removing colons and applying
// common cross-platform aliases.
func NormalizeEmojiName(emojiName string) string {
	name := strings.Trim(emojiName, ":")
	if alias, ok := emojiAliases[strings.ToLower(name)]; ok {
		return alias
	}
	return name
}

// emojiUnicodeToName maps Unicode emoji characters to shortcode names.
var emojiUnicodeToName = map[string]string{
	"👍":  "+1",
	"👎":  "-1",
	"✅":  "white_check_mark",
	"❌":  "x",
	"⚠️": "warning",
	"🛑":  "stop",
	"⏸️": "pause",
	"▶️": "arrow_forward",
	"1️⃣": "one",
	"2️⃣": "two",
	"3️⃣": "three",
	"4️⃣": "four",
	"5️⃣": "five",
	"6️⃣": "six",
	"7️⃣": "seven",
	"8️⃣": "eight",
	"9️⃣": "nine",
	"🔟":  "keycap_ten",
	"0️⃣": "zero",
	"🤖":  "robot",
	"⚙️": "gear",
	"🔐":  "lock",
	"🔓":  "unlock",
	"📁":  "file_folder",
	"📄":  "page_facing_up",
	"📝":  "memo",
	"⏱️": "stopwatch",
	"⏳":  "hourglass",
	"🌱":  "seedling",
	"🌲":  "evergreen_tree",
	"🌳":  "deciduous_tree",
	"🧵":  "thread",
	"🔄":  "arrows_counterclockwise",
	"📦":  "package",
	"🎉":  "partying_face",
	"🌿":  "herb",
	"👤":  "bust_in_silhouette",
	"📋":  "clipboard",
	"🔽":  "small_red_triangle_down",
	"🆕":  "new",
}

// GetEmojiName converts a Unicode emoji character to its shortcode name.
// If the input is already a shortcode, it is returned as-is.
func GetEmojiName(emoji string) string {
	if name, ok := emojiUnicodeToName[emoji]; ok {
		return name
	}
	return emoji
}

// convertMarkdownTablesToSlack converts markdown tables to Slack list format.
func ConvertMarkdownTablesToSlack(content string) string {
	// Regex: | header | \n |---| \n | row |
	tableRe := regexp.MustCompile(`(?m)^\|(.+)\|\s*\n\|[-:\s|]+\|\s*\n((?:\|.+\|\s*\n?)*)`)
	return tableRe.ReplaceAllStringFunc(content, func(match string) string {
		// Split into lines
		lines := strings.Split(strings.TrimRight(match, "\n"), "\n")
		if len(lines) < 3 {
			return match
		}
		// Parse headers from line 0
		headers := splitTableRow(lines[0])
		// lines[1] is the separator — skip
		// Parse body rows from lines[2:]
		var formatted []string
		for _, rowLine := range lines[2:] {
			if strings.TrimSpace(rowLine) == "" {
				continue
			}
			cells := splitTableRow(rowLine)
			var parts []string
			for i, cell := range cells {
				if i < len(headers) && headers[i] != "" {
					parts = append(parts, "*"+headers[i]+":* "+cell)
				} else {
					parts = append(parts, cell)
				}
			}
			formatted = append(formatted, strings.Join(parts, " · "))
		}
		return strings.Join(formatted, "\n")
	})
}

func splitTableRow(row string) []string {
	parts := strings.Split(row, "|")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ConvertMarkdownToSlack converts standard markdown to Slack mrkdwn format.
func ConvertMarkdownToSlack(content string) string {
	// Extract and preserve code blocks so their content is not transformed.
	var preserved []string
	placeholder := "\x00CODE\x00"
	codeBlockRe := regexp.MustCompile("(?s)```.*?```")
	inlineCodeRe := regexp.MustCompile("`[^`\n]+`")

	result := codeBlockRe.ReplaceAllStringFunc(content, func(m string) string {
		idx := len(preserved)
		preserved = append(preserved, m)
		return placeholder + string(rune(idx+1))
	})
	result = inlineCodeRe.ReplaceAllStringFunc(result, func(m string) string {
		idx := len(preserved)
		preserved = append(preserved, m)
		return placeholder + string(rune(idx+1))
	})

	// Convert tables
	result = ConvertMarkdownTablesToSlack(result)

	// Convert headers: ## Heading → *Heading*
	headerRe := regexp.MustCompile(`(?m)^#{1,6}\s+(.+)$`)
	result = headerRe.ReplaceAllString(result, "*$1*")

	// Convert **bold** → *bold*
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	result = boldRe.ReplaceAllString(result, "*$1*")

	// Convert links [text](url) → <url|text>
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	result = linkRe.ReplaceAllString(result, "<$2|$1>")

	// Convert horizontal rules
	hrRe := regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
	result = hrRe.ReplaceAllString(result, "━━━━━━━━━━━━━━━━━━━━")

	// Restore code blocks
	for i, block := range preserved {
		result = strings.ReplaceAll(result, placeholder+string(rune(i+1)), block)
	}

	return result
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/... -v -run "TestGetPlatformIcon|TestTruncate|TestNormalize|TestGetEmoji|TestConvert"
```
Expected: all tests PASS

- [ ] **Step 5: Run full package tests**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/... -v
```

- [ ] **Step 6: Commit**

```bash
git add internal/platform/utils.go internal/platform/utils_test.go
git commit -m "feat(go): add platform utilities with tests (Phase 2 Task 5)"
```

---

## Task 6: MockPlatformClient

**Files:**
- Create: `internal/platform/mock_client.go`

This mock is used by tests in later phases (operations, session, etc.) that need a PlatformClient without a real network connection.

- [ ] **Step 1: Write mock_client.go**

```go
package platform

import (
	"context"
	"sync"
)

// MockPlatformClient is a test double for PlatformClient.
// All methods are no-ops by default. Override fields to inject test behavior.
type MockPlatformClient struct {
	mu sync.RWMutex

	// Override these to control behavior
	PlatformIDVal   string
	PlatformTypeVal string
	DisplayNameVal  string
	BotNameVal      string
	FormatterVal    PlatformFormatter
	MessageLimitsVal MessageLimits

	// Recorded calls (for assertions)
	CreatedPosts    []string
	UpdatedPosts    []struct{ ID, Message string }
	AddedReactions  []struct{ PostID, Emoji string }
	DeletedPosts    []string

	// Errors to return
	CreatePostErr error
	UpdatePostErr error

	// Callbacks registered by the system under test
	onConnected    func()
	onDisconnected func()
	onReconnecting func(int)
	onError        func(error)
	onMessage      func(PlatformPost, *PlatformUser)
	onReaction     func(PlatformReaction, *PlatformUser)
	onReactionRemoved func(PlatformReaction, *PlatformUser)
	onChannelPost  func(PlatformPost, *PlatformUser)
}

func (m *MockPlatformClient) PlatformID() string   { return m.PlatformIDVal }
func (m *MockPlatformClient) PlatformType() string { return m.PlatformTypeVal }
func (m *MockPlatformClient) DisplayName() string  { return m.DisplayNameVal }

func (m *MockPlatformClient) Connect(_ context.Context) error { return nil }
func (m *MockPlatformClient) Disconnect()                     {}
func (m *MockPlatformClient) PrepareForReconnect()            {}

func (m *MockPlatformClient) GetBotUser(_ context.Context) (*PlatformUser, error) {
	return &PlatformUser{ID: "bot-id", Username: m.BotNameVal}, nil
}
func (m *MockPlatformClient) GetUser(_ context.Context, _ string) (*PlatformUser, error) {
	return nil, nil
}
func (m *MockPlatformClient) GetUserByUsername(_ context.Context, _ string) (*PlatformUser, error) {
	return nil, nil
}
func (m *MockPlatformClient) IsUserAllowed(_ string) bool { return true }
func (m *MockPlatformClient) GetBotName() string          { return m.BotNameVal }
func (m *MockPlatformClient) GetMcpConfig() McpConfig     { return McpConfig{} }
func (m *MockPlatformClient) GetFormatter() PlatformFormatter {
	if m.FormatterVal != nil {
		return m.FormatterVal
	}
	return nil
}
func (m *MockPlatformClient) GetThreadLink(_, _, _ string) string { return "" }

func (m *MockPlatformClient) CreatePost(_ context.Context, message, _ string) (*PlatformPost, error) {
	if m.CreatePostErr != nil {
		return nil, m.CreatePostErr
	}
	m.mu.Lock()
	m.CreatedPosts = append(m.CreatedPosts, message)
	m.mu.Unlock()
	return &PlatformPost{ID: "post-id", Message: message}, nil
}
func (m *MockPlatformClient) UpdatePost(_ context.Context, postID, message string) (*PlatformPost, error) {
	if m.UpdatePostErr != nil {
		return nil, m.UpdatePostErr
	}
	m.mu.Lock()
	m.UpdatedPosts = append(m.UpdatedPosts, struct{ ID, Message string }{postID, message})
	m.mu.Unlock()
	return &PlatformPost{ID: postID, Message: message}, nil
}
func (m *MockPlatformClient) CreateInteractivePost(_ context.Context, message string, _ []string, _ string) (*PlatformPost, error) {
	return &PlatformPost{ID: "interactive-id", Message: message}, nil
}
func (m *MockPlatformClient) GetPost(_ context.Context, _ string) (*PlatformPost, error) {
	return nil, nil
}
func (m *MockPlatformClient) DeletePost(_ context.Context, postID string) error {
	m.mu.Lock()
	m.DeletedPosts = append(m.DeletedPosts, postID)
	m.mu.Unlock()
	return nil
}
func (m *MockPlatformClient) PinPost(_ context.Context, _ string) error   { return nil }
func (m *MockPlatformClient) UnpinPost(_ context.Context, _ string) error { return nil }
func (m *MockPlatformClient) GetPinnedPosts(_ context.Context) ([]string, error) {
	return nil, nil
}
func (m *MockPlatformClient) GetMessageLimits() MessageLimits {
	if m.MessageLimitsVal.MaxLength == 0 {
		return MessageLimits{MaxLength: 16000, HardThreshold: 15000}
	}
	return m.MessageLimitsVal
}
func (m *MockPlatformClient) GetThreadHistory(_ context.Context, _ string, _ *ThreadHistoryOptions) ([]ThreadMessage, error) {
	return nil, nil
}

func (m *MockPlatformClient) AddReaction(_ context.Context, postID, emoji string) error {
	m.mu.Lock()
	m.AddedReactions = append(m.AddedReactions, struct{ PostID, Emoji string }{postID, emoji})
	m.mu.Unlock()
	return nil
}
func (m *MockPlatformClient) RemoveReaction(_ context.Context, _, _ string) error { return nil }

func (m *MockPlatformClient) IsBotMentioned(_ string) bool   { return false }
func (m *MockPlatformClient) ExtractPrompt(msg string) string { return msg }
func (m *MockPlatformClient) SendTyping(_ string)             {}

func (m *MockPlatformClient) DownloadFile(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}
func (m *MockPlatformClient) GetFileInfo(_ context.Context, _ string) (*PlatformFile, error) {
	return nil, nil
}

func (m *MockPlatformClient) OnConnected(f func())          { m.onConnected = f }
func (m *MockPlatformClient) OnDisconnected(f func())       { m.onDisconnected = f }
func (m *MockPlatformClient) OnReconnecting(f func(int))    { m.onReconnecting = f }
func (m *MockPlatformClient) OnError(f func(error))         { m.onError = f }
func (m *MockPlatformClient) OnMessage(f func(PlatformPost, *PlatformUser)) {
	m.onMessage = f
}
func (m *MockPlatformClient) OnReaction(f func(PlatformReaction, *PlatformUser)) {
	m.onReaction = f
}
func (m *MockPlatformClient) OnReactionRemoved(f func(PlatformReaction, *PlatformUser)) {
	m.onReactionRemoved = f
}
func (m *MockPlatformClient) OnChannelPost(f func(PlatformPost, *PlatformUser)) {
	m.onChannelPost = f
}

// SimulateMessage triggers the OnMessage callback with the given post and user.
// Use this in tests to simulate incoming messages.
func (m *MockPlatformClient) SimulateMessage(post PlatformPost, user *PlatformUser) {
	if m.onMessage != nil {
		m.onMessage(post, user)
	}
}

// SimulateReaction triggers the OnReaction callback.
func (m *MockPlatformClient) SimulateReaction(reaction PlatformReaction, user *PlatformUser) {
	if m.onReaction != nil {
		m.onReaction(reaction, user)
	}
}
```

- [ ] **Step 2: Verify MockPlatformClient implements PlatformClient**

Add a compile-time check at the bottom of mock_client.go:

```go
// Compile-time interface check.
var _ PlatformClient = (*MockPlatformClient)(nil)
```

- [ ] **Step 3: Build and test**

```bash
env -u GOROOT /opt/homebrew/bin/go build ./internal/platform/...
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/... -v
```
Expected: all tests PASS, no build errors

- [ ] **Step 4: Commit**

```bash
git add internal/platform/mock_client.go
git commit -m "feat(go): add MockPlatformClient for tests (Phase 2 Task 6)"
```

---

## Task 7: Phase 2 final verification

- [ ] **Step 1: Run all tests in go/**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./... -v 2>&1 | tail -20
```
Expected: all packages GREEN

- [ ] **Step 2: Run go vet**

```bash
env -u GOROOT /opt/homebrew/bin/go vet ./...
```
Expected: no output

- [ ] **Step 3: Update master plan**

Mark Phase 2 complete in:
- `docs/superpowers/plans/2026-03-31-golang-conversion-master.md`
- `docs/plans/golang-conversion.md`

- [ ] **Step 4: Commit docs**

```bash
git add ../../docs/superpowers/plans/2026-03-31-golang-conversion-master.md ../../docs/plans/golang-conversion.md
git commit -m "docs: mark Phase 2 complete"
```
