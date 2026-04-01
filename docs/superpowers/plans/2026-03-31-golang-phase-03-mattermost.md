# Phase 3: Mattermost Client — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Full Mattermost implementation of PlatformClient: REST API, WebSocket events, reconnection, heartbeat, formatter.

**Architecture:** `BasePlatformClient` (embedded struct with shared reconnect/heartbeat/callback logic) + `MattermostClient` (implements PlatformClient). The TypeScript `extends BasePlatformClient` pattern maps to Go struct embedding.

**Tech Stack:** Go 1.24, gorilla/websocket, net/http, Testify

**Working directory:** `/Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion/go`

**Run all go commands with:** `env -u GOROOT /opt/homebrew/bin/go`

---

## Files

| Action | Path | Description |
|--------|------|-------------|
| Create | `internal/platform/mattermost/types.go` | Mattermost-specific wire types (WebSocket events, API request/response) |
| Create | `internal/platform/mattermost/formatter.go` | MattermostFormatter implementing PlatformFormatter |
| Create | `internal/platform/mattermost/formatter_test.go` | All formatter tests |
| Create | `internal/platform/base_client.go` | BasePlatformClient embedded struct (heartbeat, reconnect, callbacks) |
| Create | `internal/platform/mattermost/client.go` | MattermostClient implementing PlatformClient |
| Create | `internal/platform/mattermost/client_test.go` | Unit tests for MattermostClient (using httptest) |

---

## Task 1: Mattermost wire types

**Files:**
- Create: `internal/platform/mattermost/types.go`

- [ ] **Step 1: Write types.go**

```go
// Package mattermost provides a Mattermost implementation of PlatformClient.
package mattermost

// WebSocketEvent is a Mattermost WebSocket event envelope.
type WebSocketEvent struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Broadcast struct {
		ChannelID string `json:"channel_id"`
		UserID    string `json:"user_id"`
		TeamID    string `json:"team_id"`
	} `json:"broadcast"`
	Seq int `json:"seq"`
}

// File is a Mattermost file attachment.
type File struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	Extension string `json:"extension"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

// Post is a Mattermost post from the API or WebSocket.
type Post struct {
	ID        string                 `json:"id"`
	CreateAt  int64                  `json:"create_at"`
	UpdateAt  int64                  `json:"update_at"`
	DeleteAt  int64                  `json:"delete_at"`
	UserID    string                 `json:"user_id"`
	ChannelID string                 `json:"channel_id"`
	RootID    string                 `json:"root_id"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	Props     map[string]interface{} `json:"props"`
	FileIDs   []string               `json:"file_ids,omitempty"`
	Metadata  *PostMetadata          `json:"metadata,omitempty"`
}

// PostMetadata holds embedded metadata for a post.
type PostMetadata struct {
	Files []File `json:"files,omitempty"`
}

// User is a Mattermost user from the API.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
}

// Reaction is a Mattermost reaction.
type Reaction struct {
	UserID    string `json:"user_id"`
	PostID    string `json:"post_id"`
	EmojiName string `json:"emoji_name"`
	CreateAt  int64  `json:"create_at"`
}

// PostedEventData is the data field for 'posted' WebSocket events.
type PostedEventData struct {
	ChannelDisplayName string `json:"channel_display_name"`
	ChannelName        string `json:"channel_name"`
	ChannelType        string `json:"channel_type"`
	Post               string `json:"post"` // JSON-encoded Post
	SenderName         string `json:"sender_name"`
	TeamID             string `json:"team_id"`
}

// ReactionEventData is the data field for 'reaction_added'/'reaction_removed' events.
type ReactionEventData struct {
	Reaction string `json:"reaction"` // JSON-encoded Reaction
}

// CreatePostRequest is the body for POST /posts.
type CreatePostRequest struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
	RootID    string `json:"root_id,omitempty"`
}

// UpdatePostRequest is the body for PUT /posts/{id}.
type UpdatePostRequest struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ThreadResponse is the response for GET /posts/{id}/thread.
type ThreadResponse struct {
	Order []string        `json:"order"`
	Posts map[string]Post `json:"posts"`
}

// ChannelPostsResponse is the response for GET /channels/{id}/posts.
type ChannelPostsResponse struct {
	Order []string        `json:"order"`
	Posts map[string]Post `json:"posts"`
}

// PinnedPostsResponse is the response for GET /channels/{id}/pinned.
type PinnedPostsResponse struct {
	Order []string        `json:"order"`
	Posts map[string]Post `json:"posts"`
}
```

- [ ] **Step 2: Build check**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion/go && env -u GOROOT /opt/homebrew/bin/go build ./internal/platform/mattermost/...
```

- [ ] **Step 3: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion && git add go/internal/platform/mattermost/types.go && git commit -m "feat(go): add Mattermost wire types (Phase 3 Task 1)"
```

---

## Task 2: MattermostFormatter — tests first

**Files:**
- Create: `internal/platform/mattermost/formatter_test.go`
- Create: `internal/platform/mattermost/formatter.go`

- [ ] **Step 1: Write formatter_test.go (RED)**

```go
package mattermost_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/platform/mattermost"
	"github.com/stretchr/testify/assert"
)

func TestFormatBold(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "**hello**", f.FormatBold("hello"))
	assert.Equal(t, "****", f.FormatBold(""))
}

func TestFormatItalic(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "_hello_", f.FormatItalic("hello"))
}

func TestFormatCode(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "`const x = 1`", f.FormatCode("const x = 1"))
}

func TestFormatCodeBlock(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "```javascript\nconst x = 1\n```\n", f.FormatCodeBlock("const x = 1", "javascript"))
	assert.Equal(t, "```\nconst x = 1\n```\n", f.FormatCodeBlock("const x = 1", ""))
}

func TestFormatUserMention(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "@johndoe", f.FormatUserMention("johndoe", ""))
}

func TestFormatLink(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "[Click here](https://example.com)", f.FormatLink("Click here", "https://example.com"))
}

func TestFormatListItem(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "- Item 1", f.FormatListItem("Item 1"))
}

func TestFormatNumberedListItem(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "1. First item", f.FormatNumberedListItem(1, "First item"))
	assert.Equal(t, "10. Tenth item", f.FormatNumberedListItem(10, "Tenth item"))
}

func TestFormatBlockquote(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "> quoted text", f.FormatBlockquote("quoted text"))
}

func TestFormatStrikethrough(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "~~deleted~~", f.FormatStrikethrough("deleted"))
}

func TestFormatTable(t *testing.T) {
	f := mattermost.NewFormatter()
	headers := []string{"Name", "Age", "City"}
	rows := [][]string{
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}
	result := f.FormatTable(headers, rows)
	assert.Contains(t, result, "| Name | Age | City |")
	assert.Contains(t, result, "| --- | --- | --- |")
	assert.Contains(t, result, "| Alice | 30 | NYC |")
	assert.Contains(t, result, "| Bob | 25 | LA |")
}

func TestFormatTable_Empty(t *testing.T) {
	f := mattermost.NewFormatter()
	result := f.FormatTable([]string{"A", "B"}, nil)
	assert.Contains(t, result, "| A | B |")
	assert.Contains(t, result, "| --- | --- |")
}

func TestFormatTable_EscapesPipes(t *testing.T) {
	f := mattermost.NewFormatter()
	result := f.FormatTable(
		[]string{"Command", "Description"},
		[][]string{{"` + "`" + `!permissions interactive|skip` + "`" + `", "Toggle"}},
	)
	assert.Contains(t, result, `interactive\|skip`)
}

func TestFormatKeyValueList(t *testing.T) {
	f := mattermost.NewFormatter()
	items := [][3]string{
		{"🔵", "Status", "Active"},
		{"🏷️", "Version", "1.0.0"},
	}
	result := f.FormatKeyValueList(items)
	assert.Contains(t, result, "| 🔵 **Status** | Active |")
	assert.Contains(t, result, "| 🏷️ **Version** | 1.0.0 |")
}

func TestFormatHorizontalRule(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "---", f.FormatHorizontalRule())
}

func TestFormatHeading(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "# Title", f.FormatHeading("Title", 1))
	assert.Equal(t, "## Subtitle", f.FormatHeading("Subtitle", 2))
	assert.Equal(t, "### Section", f.FormatHeading("Section", 3))
	// Clamp: level 0 → 1, level 7 → 6
	assert.Equal(t, "# Title", f.FormatHeading("Title", 0))
	assert.Equal(t, "###### Title", f.FormatHeading("Title", 7))
}

func TestEscapeText(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, `\*bold\* and \_italic\_`, f.EscapeText("*bold* and _italic_"))
	assert.Equal(t, "use \\`code\\` here", f.EscapeText("use `code` here"))
	assert.Equal(t, `\[link\]\(url\)`, f.EscapeText("[link](url)"))
	assert.Equal(t, "hello world", f.EscapeText("hello world"))
}

func TestFormatMarkdown_NormalizesNewlines(t *testing.T) {
	f := mattermost.NewFormatter()
	assert.Equal(t, "Line 1\n\nLine 2", f.FormatMarkdown("Line 1\n\n\n\nLine 2"))
}

func TestFormatMarkdown_PreservesStandardMarkdown(t *testing.T) {
	f := mattermost.NewFormatter()
	input := "**bold** and [link](url) and ## Header"
	assert.Equal(t, input, f.FormatMarkdown(input))
}

func TestFormatMarkdown_PreservesCodeBlocks(t *testing.T) {
	f := mattermost.NewFormatter()
	input := "```javascript\nconst x = 1;\n```"
	assert.Equal(t, input, f.FormatMarkdown(input))
}

func TestFormatMarkdown_AddsNewlineAfterCodeBlock(t *testing.T) {
	f := mattermost.NewFormatter()
	input := "```javascript\nconst x = 1;\n```More text here"
	assert.Equal(t, "```javascript\nconst x = 1;\n```\nMore text here", f.FormatMarkdown(input))
}

func TestFormatMarkdown_DoesNotAddExtraNewline(t *testing.T) {
	f := mattermost.NewFormatter()
	input := "```javascript\nconst x = 1;\n```\nMore text here"
	assert.Equal(t, input, f.FormatMarkdown(input))
}

func TestFormatMarkdown_MultipleCodeBlocks(t *testing.T) {
	f := mattermost.NewFormatter()
	input := "```js\ncode1\n```Text1\n```js\ncode2\n```Text2"
	assert.Equal(t, "```js\ncode1\n```\nText1\n```js\ncode2\n```\nText2", f.FormatMarkdown(input))
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion/go && env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/mattermost/... 2>&1 | head -20
```
Expected: compile error (formatter not yet defined)

- [ ] **Step 3: Write formatter.go (GREEN)**

```go
package mattermost

import (
	"fmt"
	"regexp"
	"strings"
)

// Formatter implements platform.PlatformFormatter for Mattermost.
// Mattermost uses standard markdown, so most methods are pass-through.
type Formatter struct{}

// NewFormatter returns a new MattermostFormatter.
func NewFormatter() *Formatter { return &Formatter{} }

func (f *Formatter) FormatBold(text string) string { return "**" + text + "**" }

func (f *Formatter) FormatItalic(text string) string { return "_" + text + "_" }

func (f *Formatter) FormatCode(text string) string { return "`" + text + "`" }

func (f *Formatter) FormatCodeBlock(code, language string) string {
	return "```" + language + "\n" + code + "\n```\n"
}

func (f *Formatter) FormatUserMention(username, _ string) string { return "@" + username }

func (f *Formatter) FormatLink(text, url string) string {
	return fmt.Sprintf("[%s](%s)", text, url)
}

func (f *Formatter) FormatListItem(text string) string { return "- " + text }

func (f *Formatter) FormatNumberedListItem(number int, text string) string {
	return fmt.Sprintf("%d. %s", number, text)
}

func (f *Formatter) FormatBlockquote(text string) string { return "> " + text }

func (f *Formatter) FormatHorizontalRule() string { return "---" }

func (f *Formatter) FormatStrikethrough(text string) string { return "~~" + text + "~~" }

func (f *Formatter) FormatHeading(text string, level int) string {
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	return strings.Repeat("#", level) + " " + text
}

func (f *Formatter) EscapeText(text string) string {
	re := regexp.MustCompile(`([*_` + "`" + `\[\]()+\-.!#])`)
	return re.ReplaceAllString(text, `\$1`)
}

func (f *Formatter) FormatTable(headers []string, rows [][]string) string {
	escape := func(s string) string { return strings.ReplaceAll(s, "|", `\|`) }

	var parts []string
	// Header row
	headerCells := make([]string, len(headers))
	for i, h := range headers {
		headerCells[i] = escape(h)
	}
	parts = append(parts, "| "+strings.Join(headerCells, " | ")+" |")
	// Separator
	seps := make([]string, len(headers))
	for i := range seps {
		seps[i] = "---"
	}
	parts = append(parts, "| "+strings.Join(seps, " | ")+" |")
	// Data rows
	for _, row := range rows {
		cells := make([]string, len(row))
		for i, c := range row {
			cells[i] = escape(c)
		}
		parts = append(parts, "| "+strings.Join(cells, " | ")+" |")
	}
	return strings.Join(parts, "\n")
}

func (f *Formatter) FormatKeyValueList(items [][3]string) string {
	escape := func(s string) string { return strings.ReplaceAll(s, "|", `\|`) }
	rows := []string{"| | |", "|---|---|"}
	for _, item := range items {
		icon, label, value := item[0], item[1], item[2]
		rows = append(rows, fmt.Sprintf("| %s **%s** | %s |", icon, escape(label), escape(value)))
	}
	return strings.Join(rows, "\n")
}

var (
	excessNewlinesRe = regexp.MustCompile(`\n{3,}`)
	// Match ``` preceded by newline (closing marker), followed by non-whitespace
	// that isn't a language identifier (opening marker pattern).
	closingCodeBlockRe = regexp.MustCompile("(?m)(?<=\n)```(?=\\S)(?![a-zA-Z]*\n)")
)

func (f *Formatter) FormatMarkdown(content string) string {
	// Fix code blocks with text immediately after closing ```
	// Go's regexp doesn't support lookbehind, so we use a different approach.
	// We find ```<nonspace> patterns that appear after a newline.
	result := fixCodeBlockNewlines(content)
	// Normalize excessive newlines
	result = excessNewlinesRe.ReplaceAllString(result, "\n\n")
	return result
}

// fixCodeBlockNewlines adds a newline between ``` and immediately following text.
// This handles the case where closing ``` is not followed by a newline.
// Example: "```\ncode\n```Text" → "```\ncode\n```\nText"
func fixCodeBlockNewlines(content string) string {
	// Split into lines to process line by line
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// If a line starts with ``` and has more text after (closing marker with text),
		// and it's not an opening marker (which has a language identifier)
		if strings.HasPrefix(line, "```") && len(line) > 3 {
			rest := line[3:]
			// Check if it looks like a language identifier (opening ```lang) — skip those
			// A language identifier only has letters/digits/hyphens
			isLangId := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(rest)
			if !isLangId {
				// This is a closing ``` with text after it — split
				lines[i] = "```"
				lines = append(lines[:i+1], append([]string{rest}, lines[i+1:]...)...)
			}
		}
	}
	return strings.Join(lines, "\n")
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/mattermost/... -v -run "TestFormat"
```

- [ ] **Step 5: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion && git add go/internal/platform/mattermost/formatter.go go/internal/platform/mattermost/formatter_test.go && git commit -m "feat(go): add MattermostFormatter with tests (Phase 3 Task 2)"
```

---

## Task 3: BasePlatformClient (shared reconnect/heartbeat/callbacks)

**Files:**
- Create: `internal/platform/base_client.go`

This provides shared implementation embedded by MattermostClient and (later) SlackClient.

- [ ] **Step 1: Write base_client.go**

```go
package platform

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
)

var baseLog = utils.CreateLogger("base-client")
var wsLog = utils.WsLogger

// BasePlatformClient provides shared connection management for platform clients.
// Embed this struct and call its methods from Connect/Disconnect implementations.
type BasePlatformClient struct {
	mu sync.RWMutex

	AllowedUsers []string
	BotNameVal   string

	// Connection state
	intentionalDisconnect bool
	isReconnecting        bool

	// Heartbeat
	heartbeatStop   chan struct{}
	heartbeatDone   chan struct{}
	lastMessageAt   time.Time
	HeartbeatInterval time.Duration // default 30s
	HeartbeatTimeout  time.Duration // default 60s

	// Reconnect
	reconnectAttempts    int
	MaxReconnectAttempts int // default 10
	ReconnectBaseDelay   time.Duration // default 1s

	// Callbacks
	onConnected       func()
	onDisconnected    func()
	onReconnecting    func(int)
	onError           func(error)
	onMessage         func(PlatformPost, *PlatformUser)
	onReaction        func(PlatformReaction, *PlatformUser)
	onReactionRemoved func(PlatformReaction, *PlatformUser)
	onChannelPost     func(PlatformPost, *PlatformUser)

	// connectFn is set by the concrete client so BasePlatformClient can
	// call Connect() during reconnection without knowing the concrete type.
	connectFn          func(ctx context.Context) error
	forceCloseFn       func()
	recoverMissedFn    func() error
}

// InitBase sets defaults. Call from concrete client's constructor.
func (b *BasePlatformClient) InitBase(
	connectFn func(ctx context.Context) error,
	forceCloseFn func(),
	recoverMissedFn func() error,
) {
	b.connectFn = connectFn
	b.forceCloseFn = forceCloseFn
	b.recoverMissedFn = recoverMissedFn
	b.lastMessageAt = time.Now()
	if b.HeartbeatInterval == 0 {
		b.HeartbeatInterval = 30 * time.Second
	}
	if b.HeartbeatTimeout == 0 {
		b.HeartbeatTimeout = 60 * time.Second
	}
	if b.MaxReconnectAttempts == 0 {
		b.MaxReconnectAttempts = 10
	}
	if b.ReconnectBaseDelay == 0 {
		b.ReconnectBaseDelay = time.Second
	}
}

// IsUserAllowed reports whether username is allowed (empty list = allow all).
func (b *BasePlatformClient) IsUserAllowed(username string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.AllowedUsers) == 0 {
		return true
	}
	for _, u := range b.AllowedUsers {
		if u == username {
			return true
		}
	}
	return false
}

// GetBotName returns the bot's mention name.
func (b *BasePlatformClient) GetBotName() string { return b.BotNameVal }

// Disconnect marks disconnect as intentional, stops heartbeat, closes connection.
func (b *BasePlatformClient) Disconnect() {
	wsLog.Info("Disconnecting (intentional)")
	b.mu.Lock()
	b.intentionalDisconnect = true
	b.mu.Unlock()
	b.stopHeartbeat()
	if b.forceCloseFn != nil {
		b.forceCloseFn()
	}
}

// PrepareForReconnect resets state so Connect() can be called again.
func (b *BasePlatformClient) PrepareForReconnect() {
	b.mu.Lock()
	b.intentionalDisconnect = false
	b.reconnectAttempts = 0
	b.mu.Unlock()
}

// UpdateLastMessageTime records activity. Call on every incoming WS message.
func (b *BasePlatformClient) UpdateLastMessageTime() {
	b.mu.Lock()
	b.lastMessageAt = time.Now()
	b.mu.Unlock()
}

// StartHeartbeat begins monitoring the connection for inactivity.
func (b *BasePlatformClient) StartHeartbeat() {
	b.stopHeartbeat()
	b.mu.Lock()
	b.lastMessageAt = time.Now()
	stop := make(chan struct{})
	done := make(chan struct{})
	b.heartbeatStop = stop
	b.heartbeatDone = done
	b.mu.Unlock()

	go func() {
		defer close(done)
		ticker := time.NewTicker(b.HeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				b.mu.RLock()
				silent := time.Since(b.lastMessageAt)
				b.mu.RUnlock()
				if silent > b.HeartbeatTimeout {
					baseLog.Warn("Connection dead (" + silent.String() + " silent), reconnecting...")
					b.stopHeartbeat()
					b.scheduleReconnect()
					return
				}
				wsLog.Debug("Heartbeat ok (" + silent.String() + " ago)")
			}
		}
	}()
}

func (b *BasePlatformClient) stopHeartbeat() {
	b.mu.Lock()
	stop := b.heartbeatStop
	done := b.heartbeatDone
	b.heartbeatStop = nil
	b.heartbeatDone = nil
	b.mu.Unlock()

	if stop != nil {
		select {
		case <-stop:
			// already closed
		default:
			close(stop)
		}
		if done != nil {
			<-done
		}
	}
}

func (b *BasePlatformClient) scheduleReconnect() {
	b.mu.RLock()
	intentional := b.intentionalDisconnect
	attempts := b.reconnectAttempts
	b.mu.RUnlock()

	if intentional {
		wsLog.Debug("Skipping reconnect: intentional disconnect")
		return
	}
	if attempts >= b.MaxReconnectAttempts {
		baseLog.Error("Max reconnection attempts reached")
		return
	}

	if b.forceCloseFn != nil {
		b.forceCloseFn()
	}

	b.mu.Lock()
	b.isReconnecting = true
	b.reconnectAttempts++
	attempt := b.reconnectAttempts
	b.mu.Unlock()

	delay := time.Duration(float64(b.ReconnectBaseDelay) * math.Pow(2, float64(attempt-1)))
	wsLog.Info("Reconnecting in " + delay.String() + " (attempt " + fmt.Sprintf("%d/%d", attempt, b.MaxReconnectAttempts) + ")")
	b.EmitReconnecting(attempt)

	go func() {
		time.Sleep(delay)
		b.mu.RLock()
		intentional := b.intentionalDisconnect
		b.mu.RUnlock()
		if intentional {
			wsLog.Debug("Skipping reconnect: intentional disconnect was called")
			return
		}
		if b.connectFn != nil {
			if err := b.connectFn(context.Background()); err != nil {
				wsLog.Error("Reconnection failed: " + err.Error())
				b.scheduleReconnect()
			}
		}
	}()
}

// OnConnectionEstablished resets reconnect counter, starts heartbeat, emits connected.
// Call from Connect() after authentication succeeds.
func (b *BasePlatformClient) OnConnectionEstablished() {
	b.mu.Lock()
	wasReconnecting := b.isReconnecting
	b.reconnectAttempts = 0
	b.isReconnecting = false
	b.mu.Unlock()

	b.StartHeartbeat()
	b.EmitConnected()

	if wasReconnecting && b.recoverMissedFn != nil {
		go func() {
			if err := b.recoverMissedFn(); err != nil {
				baseLog.Warn("Failed to recover missed messages: " + err.Error())
			}
		}()
	}
}

// OnConnectionClosed stops heartbeat and schedules reconnect if not intentional.
// Call from WebSocket onclose handler.
func (b *BasePlatformClient) OnConnectionClosed() {
	b.stopHeartbeat()
	b.EmitDisconnected()

	b.mu.RLock()
	intentional := b.intentionalDisconnect
	b.mu.RUnlock()

	if !intentional {
		b.scheduleReconnect()
	}
}

// ---------------------------------------------------------------------------
// Callback registration (implements PlatformClient On* methods)
// ---------------------------------------------------------------------------

func (b *BasePlatformClient) OnConnected(f func()) {
	b.mu.Lock()
	b.onConnected = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnDisconnected(f func()) {
	b.mu.Lock()
	b.onDisconnected = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnReconnecting(f func(int)) {
	b.mu.Lock()
	b.onReconnecting = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnError(f func(error)) {
	b.mu.Lock()
	b.onError = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnMessage(f func(PlatformPost, *PlatformUser)) {
	b.mu.Lock()
	b.onMessage = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnReaction(f func(PlatformReaction, *PlatformUser)) {
	b.mu.Lock()
	b.onReaction = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnReactionRemoved(f func(PlatformReaction, *PlatformUser)) {
	b.mu.Lock()
	b.onReactionRemoved = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnChannelPost(f func(PlatformPost, *PlatformUser)) {
	b.mu.Lock()
	b.onChannelPost = f
	b.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Emit helpers (thread-safe)
// ---------------------------------------------------------------------------

func (b *BasePlatformClient) EmitConnected() {
	b.mu.RLock()
	f := b.onConnected
	b.mu.RUnlock()
	if f != nil {
		f()
	}
}
func (b *BasePlatformClient) EmitDisconnected() {
	b.mu.RLock()
	f := b.onDisconnected
	b.mu.RUnlock()
	if f != nil {
		f()
	}
}
func (b *BasePlatformClient) EmitReconnecting(attempt int) {
	b.mu.RLock()
	f := b.onReconnecting
	b.mu.RUnlock()
	if f != nil {
		f(attempt)
	}
}
func (b *BasePlatformClient) EmitError(err error) {
	b.mu.RLock()
	f := b.onError
	b.mu.RUnlock()
	if f != nil {
		f(err)
	}
}
func (b *BasePlatformClient) EmitMessage(post PlatformPost, user *PlatformUser) {
	b.mu.RLock()
	f := b.onMessage
	b.mu.RUnlock()
	if f != nil {
		f(post, user)
	}
}
func (b *BasePlatformClient) EmitReaction(reaction PlatformReaction, user *PlatformUser) {
	b.mu.RLock()
	f := b.onReaction
	b.mu.RUnlock()
	if f != nil {
		f(reaction, user)
	}
}
func (b *BasePlatformClient) EmitReactionRemoved(reaction PlatformReaction, user *PlatformUser) {
	b.mu.RLock()
	f := b.onReactionRemoved
	b.mu.RUnlock()
	if f != nil {
		f(reaction, user)
	}
}
func (b *BasePlatformClient) EmitChannelPost(post PlatformPost, user *PlatformUser) {
	b.mu.RLock()
	f := b.onChannelPost
	b.mu.RUnlock()
	if f != nil {
		f(post, user)
	}
}
```

Note: You need `"fmt"` in the import for the Sprintf in scheduleReconnect. The full import block:
```go
import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
)
```

- [ ] **Step 2: Build check**

```bash
env -u GOROOT /opt/homebrew/bin/go build ./internal/platform/...
```

- [ ] **Step 3: Commit**

```bash
git add go/internal/platform/base_client.go && git commit -m "feat(go): add BasePlatformClient with heartbeat and callbacks (Phase 3 Task 3)"
```

---

## Task 4: MattermostClient — tests first

**Files:**
- Create: `internal/platform/mattermost/client_test.go`
- Create: `internal/platform/mattermost/client.go`

- [ ] **Step 1: Write client_test.go (RED)**

```go
package mattermost_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anneschuth/claude-threads/internal/config"
	"github.com/anneschuth/claude-threads/internal/platform/mattermost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServer creates an httptest.Server that handles the Mattermost REST API.
// handlers is a map from path to handler func.
func newTestServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}
	return httptest.NewServer(mux)
}

func testConfig(serverURL string) config.MattermostPlatformConfig {
	return config.MattermostPlatformConfig{
		ID:           "mm-test",
		DisplayName:  "Test Mattermost",
		URL:          serverURL,
		Token:        "test-token",
		ChannelID:    "channel123",
		BotName:      "testbot",
		AllowedUsers: []string{"alice", "bob"},
	}
}

func TestMattermostClient_GetBotUser(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/users/me": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			json.NewEncoder(w).Encode(User{
				ID:       "bot-id",
				Username: "testbot",
			})
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	user, err := client.GetBotUser(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "bot-id", user.ID)
	assert.Equal(t, "testbot", user.Username)
}

func TestMattermostClient_GetUser(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/users/user123": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(User{
				ID:       "user123",
				Username: "alice",
			})
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	user, err := client.GetUser(context.Background(), "user123")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "alice", user.Username)
}

func TestMattermostClient_GetUser_NotFound(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/users/missing": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	user, err := client.GetUser(context.Background(), "missing")
	// Returns nil, nil for not-found (same behavior as TypeScript)
	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestMattermostClient_CreatePost(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/posts": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			var req CreatePostRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "channel123", req.ChannelID)
			assert.Equal(t, "Hello!", req.Message)
			json.NewEncoder(w).Encode(Post{
				ID:        "post-abc",
				ChannelID: "channel123",
				Message:   "Hello!",
			})
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	post, err := client.CreatePost(context.Background(), "Hello!", "")
	require.NoError(t, err)
	assert.Equal(t, "post-abc", post.ID)
	assert.Equal(t, "Hello!", post.Message)
}

func TestMattermostClient_IsUserAllowed(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	assert.True(t, client.IsUserAllowed("alice"))
	assert.True(t, client.IsUserAllowed("bob"))
	assert.False(t, client.IsUserAllowed("charlie"))
}

func TestMattermostClient_IsUserAllowed_EmptyList(t *testing.T) {
	cfg := testConfig("http://localhost")
	cfg.AllowedUsers = nil
	client := mattermost.NewClient(cfg)
	assert.True(t, client.IsUserAllowed("anyone"))
}

func TestMattermostClient_IsBotMentioned(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	assert.True(t, client.IsBotMentioned("@testbot hello"))
	assert.True(t, client.IsBotMentioned("hey @testbot"))
	assert.False(t, client.IsBotMentioned("hello world"))
}

func TestMattermostClient_ExtractPrompt(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	assert.Equal(t, "hello", client.ExtractPrompt("@testbot hello"))
	assert.Equal(t, "hey", client.ExtractPrompt("hey @testbot"))
}

func TestMattermostClient_GetThreadLink(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://mm.example.com"))
	link := client.GetThreadLink("thread123", "", "")
	assert.Equal(t, "http://mm.example.com/_redirect/pl/thread123", link)
	link2 := client.GetThreadLink("thread123", "msg456", "")
	assert.Equal(t, "http://mm.example.com/_redirect/pl/msg456", link2)
}

func TestMattermostClient_GetMessageLimits(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	limits := client.GetMessageLimits()
	assert.Equal(t, 16000, limits.MaxLength)
	assert.Equal(t, 14000, limits.HardThreshold)
}

func TestMattermostClient_PlatformIdentity(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	assert.Equal(t, "mm-test", client.PlatformID())
	assert.Equal(t, "mattermost", client.PlatformType())
	assert.Equal(t, "Test Mattermost", client.DisplayName())
}

func TestMattermostClient_UserCaching(t *testing.T) {
	callCount := 0
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/users/user123": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			json.NewEncoder(w).Encode(User{ID: "user123", Username: "alice"})
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	// Call twice — should only hit API once
	client.GetUser(context.Background(), "user123")
	client.GetUser(context.Background(), "user123")
	assert.Equal(t, 1, callCount, "should cache user after first fetch")
}
```

Note: The test file imports `User`, `Post`, `CreatePostRequest` from the mattermost package — these are already defined in types.go.

- [ ] **Step 2: Run — expect FAIL**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/mattermost/... 2>&1 | head -20
```

- [ ] **Step 3: Write client.go (GREEN)**

```go
package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anneschuth/claude-threads/internal/config"
	"github.com/anneschuth/claude-threads/internal/platform"
	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/gorilla/websocket"
)

var log = utils.CreateLogger("mattermost")
var wsLog = utils.WsLogger

const (
	maxRetries      = 3
	retryDelayBase  = 500 * time.Millisecond
)

// Client is the Mattermost implementation of platform.PlatformClient.
type Client struct {
	platform.BasePlatformClient

	platformIDVal  string
	displayNameVal string
	url            string
	token          string
	channelID      string
	formatter      *Formatter

	mu          sync.RWMutex
	userCache   map[string]User
	botUserID   string
	lastPostID  string // for missed-message recovery

	conn *websocket.Conn
}

// NewClient creates a new MattermostClient from config.
func NewClient(cfg config.MattermostPlatformConfig) *Client {
	c := &Client{
		platformIDVal:  cfg.ID,
		displayNameVal: cfg.DisplayName,
		url:            cfg.URL,
		token:          cfg.Token,
		channelID:      cfg.ChannelID,
		formatter:      NewFormatter(),
		userCache:      make(map[string]User),
	}
	c.BasePlatformClient.AllowedUsers = cfg.AllowedUsers
	c.BasePlatformClient.BotNameVal = cfg.BotName
	c.BasePlatformClient.InitBase(c.Connect, c.forceClose, c.recoverMissed)
	return c
}

// Compile-time interface check.
var _ platform.PlatformClient = (*Client)(nil)

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func (c *Client) PlatformID() string   { return c.platformIDVal }
func (c *Client) PlatformType() string { return "mattermost" }
func (c *Client) DisplayName() string  { return c.displayNameVal }

// ---------------------------------------------------------------------------
// REST API helper
// ---------------------------------------------------------------------------

func (c *Client) api(ctx context.Context, method, path string, body interface{}, retries int) ([]byte, int, error) {
	url := c.url + "/api/v4" + path
	log.Debug(method + " " + path)

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		// Retry on 500 with exponential backoff
		if resp.StatusCode == 500 && retries < maxRetries {
			delay := retryDelayBase * time.Duration(1<<uint(retries))
			log.Warn(fmt.Sprintf("%s %s failed 500, retrying in %s (attempt %d/%d)", method, path, delay, retries+1, maxRetries))
			time.Sleep(delay)
			return c.api(ctx, method, path, body, retries+1)
		}
		return nil, resp.StatusCode, fmt.Errorf("Mattermost API error %d: %s", resp.StatusCode, string(data))
	}

	log.Debug(fmt.Sprintf("%s %s → %d", method, path, resp.StatusCode))
	return data, resp.StatusCode, nil
}

func (c *Client) apiJSON(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	data, _, err := c.api(ctx, method, path, body, 0)
	if err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

// ---------------------------------------------------------------------------
// User management
// ---------------------------------------------------------------------------

func (c *Client) normUser(u User) *platform.PlatformUser {
	displayName := u.FirstName
	if displayName == "" {
		displayName = u.Nickname
	}
	if displayName == "" {
		displayName = u.Username
	}
	return &platform.PlatformUser{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: displayName,
		Email:       u.Email,
	}
}

func (c *Client) GetBotUser(ctx context.Context) (*platform.PlatformUser, error) {
	var u User
	if err := c.apiJSON(ctx, "GET", "/users/me", nil, &u); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.botUserID = u.ID
	c.mu.Unlock()
	return c.normUser(u), nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*platform.PlatformUser, error) {
	c.mu.RLock()
	cached, ok := c.userCache[userID]
	c.mu.RUnlock()
	if ok {
		return c.normUser(cached), nil
	}

	data, status, err := c.api(ctx, "GET", "/users/"+userID, nil, 0)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.userCache[userID] = u
	c.mu.Unlock()
	return c.normUser(u), nil
}

func (c *Client) GetUserByUsername(ctx context.Context, username string) (*platform.PlatformUser, error) {
	var u User
	if err := c.apiJSON(ctx, "GET", "/users/username/"+username, nil, &u); err != nil {
		return nil, nil // not found
	}
	c.mu.Lock()
	c.userCache[u.ID] = u
	c.mu.Unlock()
	return c.normUser(u), nil
}

// ---------------------------------------------------------------------------
// Normalization helpers
// ---------------------------------------------------------------------------

func (c *Client) normPost(p Post) *platform.PlatformPost {
	np := &platform.PlatformPost{
		ID:         p.ID,
		PlatformID: c.platformIDVal,
		ChannelID:  p.ChannelID,
		UserID:     p.UserID,
		Message:    p.Message,
		RootID:     p.RootID,
		CreateAt:   p.CreateAt,
	}
	if p.Metadata != nil {
		for _, f := range p.Metadata.Files {
			np.Files = append(np.Files, platform.PlatformFile{
				ID:        f.ID,
				Name:      f.Name,
				Size:      f.Size,
				MimeType:  f.MimeType,
				Extension: f.Extension,
			})
		}
	}
	return np
}

// ---------------------------------------------------------------------------
// Messaging
// ---------------------------------------------------------------------------

func (c *Client) CreatePost(ctx context.Context, message, threadID string) (*platform.PlatformPost, error) {
	req := CreatePostRequest{
		ChannelID: c.channelID,
		Message:   message,
		RootID:    threadID,
	}
	var p Post
	if err := c.apiJSON(ctx, "POST", "/posts", req, &p); err != nil {
		return nil, err
	}
	return c.normPost(p), nil
}

func (c *Client) UpdatePost(ctx context.Context, postID, message string) (*platform.PlatformPost, error) {
	req := UpdatePostRequest{ID: postID, Message: message}
	var p Post
	if err := c.apiJSON(ctx, "PUT", "/posts/"+postID, req, &p); err != nil {
		return nil, err
	}
	return c.normPost(p), nil
}

func (c *Client) CreateInteractivePost(ctx context.Context, message string, reactions []string, threadID string) (*platform.PlatformPost, error) {
	post, err := c.CreatePost(ctx, message, threadID)
	if err != nil {
		return nil, err
	}
	for _, emoji := range reactions {
		if err := c.AddReaction(ctx, post.ID, emoji); err != nil {
			log.Warn("Failed to add reaction " + emoji + ": " + err.Error())
		}
	}
	return post, nil
}

func (c *Client) GetPost(ctx context.Context, postID string) (*platform.PlatformPost, error) {
	data, status, err := c.api(ctx, "GET", "/posts/"+postID, nil, 0)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	var p Post
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return c.normPost(p), nil
}

func (c *Client) DeletePost(ctx context.Context, postID string) error {
	_, _, err := c.api(ctx, "DELETE", "/posts/"+postID, nil, 0)
	return err
}

func (c *Client) PinPost(ctx context.Context, postID string) error {
	_, _, err := c.api(ctx, "POST", "/posts/"+postID+"/pin", nil, 0)
	return err
}

func (c *Client) UnpinPost(ctx context.Context, postID string) error {
	_, status, err := c.api(ctx, "POST", "/posts/"+postID+"/unpin", nil, 0)
	if err != nil && (status == 403 || status == 404) {
		return nil // expected failures
	}
	return err
}

func (c *Client) GetPinnedPosts(ctx context.Context) ([]string, error) {
	var resp PinnedPostsResponse
	if err := c.apiJSON(ctx, "GET", "/channels/"+c.channelID+"/pinned", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Order, nil
}

func (c *Client) GetMessageLimits() platform.MessageLimits {
	return platform.MessageLimits{MaxLength: 16000, HardThreshold: 14000}
}

func (c *Client) GetThreadHistory(ctx context.Context, threadID string, opts *platform.ThreadHistoryOptions) ([]platform.ThreadMessage, error) {
	var resp ThreadResponse
	if err := c.apiJSON(ctx, "GET", "/posts/"+threadID+"/thread", nil, &resp); err != nil {
		log.Warn("Failed to get thread history: " + err.Error())
		return nil, nil
	}

	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()

	var msgs []platform.ThreadMessage
	for _, postID := range resp.Order {
		p, ok := resp.Posts[postID]
		if !ok {
			continue
		}
		if opts != nil && opts.ExcludeBotMessages && p.UserID == botID {
			continue
		}
		user, _ := c.GetUser(ctx, p.UserID)
		username := "unknown"
		if user != nil {
			username = user.Username
		}
		msgs = append(msgs, platform.ThreadMessage{
			ID:       p.ID,
			UserID:   p.UserID,
			Username: username,
			Message:  p.Message,
			CreateAt: p.CreateAt,
		})
	}

	sort.Slice(msgs, func(i, j int) bool { return msgs[i].CreateAt < msgs[j].CreateAt })

	if opts != nil && opts.Limit > 0 && len(msgs) > opts.Limit {
		msgs = msgs[len(msgs)-opts.Limit:]
	}
	return msgs, nil
}

// ---------------------------------------------------------------------------
// Reactions
// ---------------------------------------------------------------------------

func (c *Client) AddReaction(ctx context.Context, postID, emojiName string) error {
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	_, _, err := c.api(ctx, "POST", "/reactions", map[string]string{
		"user_id":    botID,
		"post_id":    postID,
		"emoji_name": emojiName,
	}, 0)
	return err
}

func (c *Client) RemoveReaction(ctx context.Context, postID, emojiName string) error {
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	_, _, err := c.api(ctx, "DELETE", fmt.Sprintf("/users/%s/posts/%s/reactions/%s", botID, postID, emojiName), nil, 0)
	return err
}

// ---------------------------------------------------------------------------
// Bot mentions
// ---------------------------------------------------------------------------

func (c *Client) IsBotMentioned(message string) bool {
	botName := regexp.QuoteMeta(c.BotNameVal)
	re := regexp.MustCompile(`(?i)(^|\s)@` + botName + `\b`)
	return re.MatchString(message)
}

func (c *Client) ExtractPrompt(message string) string {
	botName := regexp.QuoteMeta(c.BotNameVal)
	re := regexp.MustCompile(`(?i)(^|\s)@` + botName + `\b`)
	return strings.TrimSpace(re.ReplaceAllString(message, " "))
}

// ---------------------------------------------------------------------------
// Files
// ---------------------------------------------------------------------------

func (c *Client) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	data, _, err := c.api(ctx, "GET", "/files/"+fileID, nil, 0)
	return data, err
}

func (c *Client) GetFileInfo(ctx context.Context, fileID string) (*platform.PlatformFile, error) {
	var f File
	if err := c.apiJSON(ctx, "GET", "/files/"+fileID+"/info", nil, &f); err != nil {
		return nil, err
	}
	return &platform.PlatformFile{
		ID:        f.ID,
		Name:      f.Name,
		Size:      f.Size,
		MimeType:  f.MimeType,
		Extension: f.Extension,
	}, nil
}

// ---------------------------------------------------------------------------
// Platform helpers
// ---------------------------------------------------------------------------

func (c *Client) GetMcpConfig() platform.McpConfig {
	return platform.McpConfig{
		Type:         "mattermost",
		URL:          c.url,
		Token:        c.token,
		ChannelID:    c.channelID,
		AllowedUsers: c.AllowedUsers,
	}
}

func (c *Client) GetFormatter() platform.PlatformFormatter { return c.formatter }

func (c *Client) GetThreadLink(threadID, lastMessageID, _ string) string {
	target := threadID
	if lastMessageID != "" {
		target = lastMessageID
	}
	return c.url + "/_redirect/pl/" + target
}

func (c *Client) SendTyping(threadID string) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return
	}
	conn.WriteJSON(map[string]interface{}{
		"action": "user_typing",
		"seq":    time.Now().UnixMilli(),
		"data": map[string]string{
			"channel_id": c.channelID,
			"parent_id":  threadID,
		},
	})
}

// ---------------------------------------------------------------------------
// WebSocket connection
// ---------------------------------------------------------------------------

func (c *Client) Connect(ctx context.Context) error {
	// Fetch bot user first
	if _, err := c.GetBotUser(ctx); err != nil {
		return fmt.Errorf("failed to get bot user: %w", err)
	}

	wsURL := strings.Replace(c.url, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL += "/api/v4/websocket"

	wsLog.Info("Connecting to " + wsURL)

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, wsURL, http.Header{})
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// Authenticate
	if err := conn.WriteJSON(map[string]interface{}{
		"seq":    1,
		"action": "authentication_challenge",
		"data":   map[string]string{"token": c.token},
	}); err != nil {
		conn.Close()
		return fmt.Errorf("auth challenge: %w", err)
	}

	connected := make(chan error, 1)

	go func() {
		for {
			var event WebSocketEvent
			if err := conn.ReadJSON(&event); err != nil {
				wsLog.Info("WebSocket closed: " + err.Error())
				c.mu.Lock()
				c.conn = nil
				c.mu.Unlock()
				c.BasePlatformClient.OnConnectionClosed()
				return
			}
			c.BasePlatformClient.UpdateLastMessageTime()
			c.handleEvent(ctx, event, connected)
		}
	}()

	select {
	case err := <-connected:
		return err
	case <-ctx.Done():
		conn.Close()
		return ctx.Err()
	case <-time.After(10 * time.Second):
		conn.Close()
		return fmt.Errorf("connection timeout")
	}
}

func (c *Client) handleEvent(ctx context.Context, event WebSocketEvent, connected chan<- error) {
	switch event.Event {
	case "hello":
		c.BasePlatformClient.OnConnectionEstablished()
		select {
		case connected <- nil:
		default:
		}

	case "posted":
		postJSON, _ := event.Data["post"].(string)
		if postJSON == "" {
			return
		}
		var p Post
		if err := json.Unmarshal([]byte(postJSON), &p); err != nil {
			wsLog.Warn("Failed to parse post: " + err.Error())
			return
		}

		c.mu.RLock()
		botID := c.botUserID
		c.mu.RUnlock()

		if p.UserID == botID || p.ChannelID != c.channelID {
			return
		}

		c.mu.Lock()
		c.lastPostID = p.ID
		c.mu.Unlock()

		go func() {
			// Enrich with file metadata if needed
			if len(p.FileIDs) > 0 && (p.Metadata == nil || len(p.Metadata.Files) == 0) {
				var files []File
				for _, fid := range p.FileIDs {
					var f File
					if err := c.apiJSON(ctx, "GET", "/files/"+fid+"/info", nil, &f); err == nil {
						files = append(files, f)
					}
				}
				if p.Metadata == nil {
					p.Metadata = &PostMetadata{}
				}
				p.Metadata.Files = files
			}

			np := c.normPost(p)
			user, _ := c.GetUser(ctx, p.UserID)
			c.BasePlatformClient.EmitMessage(*np, user)
			if p.RootID == "" {
				c.BasePlatformClient.EmitChannelPost(*np, user)
			}
		}()

	case "reaction_added", "reaction_removed":
		reactionJSON, _ := event.Data["reaction"].(string)
		if reactionJSON == "" {
			return
		}
		var r Reaction
		if err := json.Unmarshal([]byte(reactionJSON), &r); err != nil {
			wsLog.Warn("Failed to parse reaction: " + err.Error())
			return
		}

		c.mu.RLock()
		botID := c.botUserID
		c.mu.RUnlock()
		if r.UserID == botID {
			return
		}

		go func() {
			user, _ := c.GetUser(ctx, r.UserID)
			nr := platform.PlatformReaction{
				UserID:    r.UserID,
				PostID:    r.PostID,
				EmojiName: r.EmojiName,
				CreateAt:  r.CreateAt,
			}
			if event.Event == "reaction_added" {
				c.BasePlatformClient.EmitReaction(nr, user)
			} else {
				c.BasePlatformClient.EmitReactionRemoved(nr, user)
			}
		}()
	}
}

func (c *Client) forceClose() {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()
	if conn != nil {
		conn.Close()
	}
}

func (c *Client) recoverMissed() error {
	c.mu.RLock()
	lastID := c.lastPostID
	c.mu.RUnlock()
	if lastID == "" {
		return nil
	}
	log.Info("Recovering missed messages after " + lastID)

	var resp ChannelPostsResponse
	if err := c.apiJSON(context.Background(), "GET",
		"/channels/"+c.channelID+"/posts?after="+lastID+"&per_page=100",
		nil, &resp); err != nil {
		return err
	}

	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()

	var posts []platform.PlatformPost
	for _, postID := range resp.Order {
		p, ok := resp.Posts[postID]
		if !ok || p.UserID == botID {
			continue
		}
		posts = append(posts, *c.normPost(p))
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].CreateAt < posts[j].CreateAt })

	for _, np := range posts {
		c.mu.Lock()
		c.lastPostID = np.ID
		c.mu.Unlock()
		user, _ := c.GetUser(context.Background(), np.UserID)
		c.BasePlatformClient.EmitMessage(np, user)
		if np.RootID == "" {
			c.BasePlatformClient.EmitChannelPost(np, user)
		}
	}
	if len(posts) > 0 {
		log.Info(fmt.Sprintf("Recovered %d missed message(s)", len(posts)))
	}
	return nil
}
```

- [ ] **Step 4: Add MattermostPlatformConfig to config package**

The client imports `config.MattermostPlatformConfig`. Check if it exists:

```bash
grep -r "MattermostPlatformConfig" /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion/go/internal/config/
```

If not found, add it to `internal/config/config.go`:

```go
// MattermostPlatformConfig holds configuration for a Mattermost platform instance.
type MattermostPlatformConfig struct {
	ID           string   `yaml:"id"`
	DisplayName  string   `yaml:"displayName"`
	URL          string   `yaml:"url"`
	Token        string   `yaml:"token"`
	ChannelID    string   `yaml:"channelId"`
	BotName      string   `yaml:"botName"`
	AllowedUsers []string `yaml:"allowedUsers"`
	SkipPermissions bool  `yaml:"skipPermissions"`
}
```

- [ ] **Step 5: Run tests — expect PASS**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/mattermost/... -v
```

- [ ] **Step 6: Commit**

```bash
git add go/internal/platform/mattermost/client.go go/internal/platform/mattermost/client_test.go go/internal/config/config.go && git commit -m "feat(go): add MattermostClient with tests (Phase 3 Task 4)"
```

---

## Task 5: Phase 3 final verification

- [ ] **Step 1: Run all tests**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./... -v 2>&1 | tail -30
```
Expected: all packages GREEN

- [ ] **Step 2: Run go vet**

```bash
env -u GOROOT /opt/homebrew/bin/go vet ./...
```

- [ ] **Step 3: Update progress tables**

Mark Phase 3 complete in:
- `docs/superpowers/plans/2026-03-31-golang-conversion-master.md`
- `docs/plans/golang-conversion.md`

- [ ] **Step 4: Commit**

```bash
git add docs/ && git commit -m "docs: mark Phase 3 complete"
```
