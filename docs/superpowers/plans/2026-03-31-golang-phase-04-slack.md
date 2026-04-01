# Phase 4: Slack Client — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Full Slack Socket Mode + Web API implementation of PlatformClient.

**Architecture:**
- Slack uses **Socket Mode**: connect via WebSocket for real-time events, Web API (REST) for operations.
- Two tokens: App token (`xapp-`) for Socket Mode WS URL; Bot token (`xoxb-`) for REST API.
- Every Socket Mode event must be **ACKed within 3 seconds** (`{"envelope_id": "..."}`).
- User ID mentions in Slack look like `<@U12345>` not `@username`.
- Slack timestamps are float strings (e.g., `"1767690059.430179"`), used as post IDs.
- SlackFormatter converts standard markdown → mrkdwn (uses `ConvertMarkdownToSlack` from `platform` package).

**Tech Stack:** Go 1.24, gorilla/websocket, net/http, Testify

**Working directory:** `/Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion/go`

**Run all go commands with:** `env -u GOROOT /opt/homebrew/bin/go`

---

## Files

| Action | Path | Description |
|--------|------|-------------|
| Create | `internal/platform/slack/types.go` | Slack-specific wire types |
| Create | `internal/platform/slack/formatter.go` | SlackFormatter implementing PlatformFormatter |
| Create | `internal/platform/slack/formatter_test.go` | Formatter tests |
| Create | `internal/platform/slack/client.go` | SlackClient implementing PlatformClient |
| Create | `internal/platform/slack/client_test.go` | Client unit tests (httptest) |
| Modify | `internal/config/config.go` | Add SlackPlatformConfig |

---

## Task 1: Slack wire types

**Files:**
- Create: `internal/platform/slack/types.go`

- [ ] **Step 1: Write types.go**

```go
// Package slack provides a Slack implementation of PlatformClient using Socket Mode.
package slack

// SocketModeEvent is a Slack Socket Mode envelope.
type SocketModeEvent struct {
	EnvelopeID             string        `json:"envelope_id"`
	Type                   string        `json:"type"` // events_api, hello, disconnect
	AcceptsResponsePayload bool          `json:"accepts_response_payload"`
	RetryAttempt           int           `json:"retry_attempt"`
	RetryReason            string        `json:"retry_reason"`
	Payload                *EventPayload `json:"payload,omitempty"`
}

// EventPayload wraps the inner event for events_api envelopes.
type EventPayload struct {
	TeamID  string      `json:"team_id"`
	Event   *SlackEvent `json:"event,omitempty"`
	Type    string      `json:"type"`
	EventID string      `json:"event_id"`
}

// SlackEvent is the inner event (message, reaction_added, etc.).
type SlackEvent struct {
	Type        string            `json:"type"`
	Subtype     string            `json:"subtype,omitempty"`
	User        string            `json:"user,omitempty"`
	Channel     string            `json:"channel,omitempty"`
	Ts          string            `json:"ts,omitempty"`
	ThreadTs    string            `json:"thread_ts,omitempty"`
	Text        string            `json:"text,omitempty"`
	Reaction    string            `json:"reaction,omitempty"`
	Item        *ReactionItem     `json:"item,omitempty"`
	ItemUser    string            `json:"item_user,omitempty"`
	BotID       string            `json:"bot_id,omitempty"`
	Files       []SlackFile       `json:"files,omitempty"`
	ChannelType string            `json:"channel_type,omitempty"`
}

// ReactionItem identifies the target of a reaction event.
type ReactionItem struct {
	Type    string `json:"type"`    // message, file, file_comment
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
}

// SlackMessage is a Slack message (from conversations.history/replies).
type SlackMessage struct {
	Type      string      `json:"type"`
	Subtype   string      `json:"subtype,omitempty"`
	Ts        string      `json:"ts"`
	User      string      `json:"user,omitempty"`
	BotID     string      `json:"bot_id,omitempty"`
	Text      string      `json:"text"`
	ThreadTs  string      `json:"thread_ts,omitempty"`
	Files     []SlackFile `json:"files,omitempty"`
}

// SlackUser is a Slack user from the API.
type SlackUser struct {
	ID      string          `json:"id"`
	TeamID  string          `json:"team_id"`
	Name    string          `json:"name"`
	Deleted bool            `json:"deleted"`
	RealName string         `json:"real_name,omitempty"`
	Profile SlackUserProfile `json:"profile"`
	IsBot   bool            `json:"is_bot,omitempty"`
}

// SlackUserProfile holds profile fields.
type SlackUserProfile struct {
	DisplayName string `json:"display_name,omitempty"`
	RealName    string `json:"real_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// SlackFile is a Slack file attachment.
type SlackFile struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Mimetype            string `json:"mimetype"`
	Filetype            string `json:"filetype"`
	Size                int64  `json:"size"`
	URLPrivate          string `json:"url_private,omitempty"`
	URLPrivateDownload  string `json:"url_private_download,omitempty"`
}

// SlackPin is a pinned item.
type SlackPin struct {
	Type    string        `json:"type"`
	Message *SlackMessage `json:"message,omitempty"`
}

// --- API Response types ---

// APIResponse is the base Slack API response envelope.
type APIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// PostMessageResponse is the response for chat.postMessage.
type PostMessageResponse struct {
	APIResponse
	Channel string       `json:"channel"`
	Ts      string       `json:"ts"`
	Message SlackMessage `json:"message"`
}

// UpdateMessageResponse is the response for chat.update.
type UpdateMessageResponse struct {
	APIResponse
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
	Text    string `json:"text"`
}

// ConversationsRepliesResponse is the response for conversations.replies.
type ConversationsRepliesResponse struct {
	APIResponse
	Messages []SlackMessage `json:"messages"`
	HasMore  bool           `json:"has_more"`
}

// ConversationsHistoryResponse is the response for conversations.history.
type ConversationsHistoryResponse struct {
	APIResponse
	Messages []SlackMessage `json:"messages"`
	HasMore  bool           `json:"has_more"`
	ResponseMetadata *struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata,omitempty"`
}

// UsersInfoResponse is the response for users.info.
type UsersInfoResponse struct {
	APIResponse
	User SlackUser `json:"user"`
}

// UsersListResponse is the response for users.list.
type UsersListResponse struct {
	APIResponse
	Members []SlackUser `json:"members"`
	ResponseMetadata *struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata,omitempty"`
}

// AuthTestResponse is the response for auth.test.
type AuthTestResponse struct {
	APIResponse
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
	BotID  string `json:"bot_id,omitempty"`
}

// AppsConnectionsOpenResponse is the response for apps.connections.open.
type AppsConnectionsOpenResponse struct {
	APIResponse
	URL string `json:"url"`
}

// PinsListResponse is the response for pins.list.
type PinsListResponse struct {
	APIResponse
	Items []SlackPin `json:"items"`
}

// FilesInfoResponse is the response for files.info.
type FilesInfoResponse struct {
	APIResponse
	File SlackFile `json:"file"`
}
```

- [ ] **Step 2: Build check**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion/go && env -u GOROOT /opt/homebrew/bin/go build ./internal/platform/slack/...
```

- [ ] **Step 3: Commit**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion && git add go/internal/platform/slack/types.go && git commit -m "feat(go): add Slack wire types (Phase 4 Task 1)"
```

---

## Task 2: SlackFormatter — tests first

**Files:**
- Create: `internal/platform/slack/formatter_test.go`
- Create: `internal/platform/slack/formatter.go`

- [ ] **Step 1: Write formatter_test.go (RED)**

```go
package slack_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/platform/slack"
	"github.com/stretchr/testify/assert"
)

func TestSlackFormatBold(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "*hello*", f.FormatBold("hello"))
	assert.Equal(t, "**", f.FormatBold(""))
}

func TestSlackFormatItalic(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "_hello_", f.FormatItalic("hello"))
}

func TestSlackFormatCode(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "`const x = 1`", f.FormatCode("const x = 1"))
}

func TestSlackFormatCodeBlock(t *testing.T) {
	f := slack.NewFormatter()
	// Slack doesn't use language hints
	assert.Equal(t, "```\nconst x = 1\n```\n", f.FormatCodeBlock("const x = 1", "javascript"))
	assert.Equal(t, "```\nconst x = 1\n```\n", f.FormatCodeBlock("const x = 1", ""))
}

func TestSlackFormatUserMention_WithID(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "<@U123456>", f.FormatUserMention("alice", "U123456"))
}

func TestSlackFormatUserMention_WithoutID(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "@alice", f.FormatUserMention("alice", ""))
}

func TestSlackFormatLink(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "<https://example.com|Click here>", f.FormatLink("Click here", "https://example.com"))
}

func TestSlackFormatListItem(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "- Item 1", f.FormatListItem("Item 1"))
}

func TestSlackFormatNumberedListItem(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "1. First item", f.FormatNumberedListItem(1, "First item"))
}

func TestSlackFormatBlockquote(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "> quoted text", f.FormatBlockquote("quoted text"))
}

func TestSlackFormatHorizontalRule(t *testing.T) {
	f := slack.NewFormatter()
	assert.Contains(t, f.FormatHorizontalRule(), "━")
}

func TestSlackFormatStrikethrough(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "~deleted~", f.FormatStrikethrough("deleted"))
}

func TestSlackFormatStrikethrough_EscapesTildes(t *testing.T) {
	f := slack.NewFormatter()
	// Tildes inside text get zero-width space inserted
	result := f.FormatStrikethrough("~/some/path")
	assert.HasPrefix(t, result, "~")
	assert.HasSuffix(t, result, "~")
	assert.Contains(t, result, "\u200B") // zero-width space after each ~
}

func TestSlackFormatHeading(t *testing.T) {
	f := slack.NewFormatter()
	// Slack renders all headings as bold
	assert.Equal(t, "*Title*", f.FormatHeading("Title", 1))
	assert.Equal(t, "*Subtitle*", f.FormatHeading("Subtitle", 2))
}

func TestSlackEscapeText(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "a &amp; b", f.EscapeText("a & b"))
	assert.Equal(t, "&lt;tag&gt;", f.EscapeText("<tag>"))
	assert.Equal(t, "hello world", f.EscapeText("hello world"))
}

func TestSlackFormatTable(t *testing.T) {
	f := slack.NewFormatter()
	headers := []string{"Name", "Age"}
	rows := [][]string{{"Alice", "30"}, {"Bob", "25"}}
	result := f.FormatTable(headers, rows)
	assert.Contains(t, result, "*Name:* Alice")
	assert.Contains(t, result, "*Age:* 30")
	assert.Contains(t, result, "*Name:* Bob")
	assert.NotContains(t, result, "|") // No pipe characters in Slack table
}

func TestSlackFormatKeyValueList(t *testing.T) {
	f := slack.NewFormatter()
	items := [][3]string{
		{"🔵", "Status", "Active"},
		{"🏷️", "Version", "1.0.0"},
	}
	result := f.FormatKeyValueList(items)
	assert.Contains(t, result, "🔵 *Status:* Active")
	assert.Contains(t, result, "🏷️ *Version:* 1.0.0")
}

func TestSlackFormatMarkdown_Bold(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "*bold*", f.FormatMarkdown("**bold**"))
}

func TestSlackFormatMarkdown_Headers(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "*My Header*", f.FormatMarkdown("## My Header"))
}

func TestSlackFormatMarkdown_Links(t *testing.T) {
	f := slack.NewFormatter()
	assert.Equal(t, "<https://example.com|click here>", f.FormatMarkdown("[click here](https://example.com)"))
}

func TestSlackFormatMarkdown_PreservesCodeBlocks(t *testing.T) {
	f := slack.NewFormatter()
	input := "```go\n**not bold**\n```"
	assert.Contains(t, f.FormatMarkdown(input), "**not bold**")
}
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion/go && env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/slack/... 2>&1 | head -20
```

- [ ] **Step 3: Write formatter.go (GREEN)**

```go
package slack

import (
	"fmt"
	"strings"

	"github.com/anneschuth/claude-threads/internal/platform"
)

// Formatter implements platform.PlatformFormatter for Slack mrkdwn syntax.
type Formatter struct{}

// NewFormatter returns a new SlackFormatter.
func NewFormatter() *Formatter { return &Formatter{} }

func (f *Formatter) FormatBold(text string) string   { return "*" + text + "*" }
func (f *Formatter) FormatItalic(text string) string { return "_" + text + "_" }
func (f *Formatter) FormatCode(text string) string   { return "`" + text + "`" }

func (f *Formatter) FormatCodeBlock(code, _ string) string {
	// Slack doesn't support language hints in code blocks
	return "```\n" + code + "\n```\n"
}

func (f *Formatter) FormatUserMention(username, userID string) string {
	if userID != "" {
		return "<@" + userID + ">"
	}
	return "@" + username
}

func (f *Formatter) FormatLink(text, url string) string {
	return fmt.Sprintf("<%s|%s>", url, text)
}

func (f *Formatter) FormatListItem(text string) string         { return "- " + text }
func (f *Formatter) FormatBlockquote(text string) string       { return "> " + text }
func (f *Formatter) FormatHorizontalRule() string              { return "━━━━━━━━━━━━━━━━━━━━" }
func (f *Formatter) FormatHeading(text string, _ int) string   { return "*" + text + "*" }

func (f *Formatter) FormatNumberedListItem(number int, text string) string {
	return fmt.Sprintf("%d. %s", number, text)
}

func (f *Formatter) FormatStrikethrough(text string) string {
	// Insert zero-width space after each ~ to prevent Slack from breaking
	// strikethrough formatting on paths like ~/some/path
	escaped := strings.ReplaceAll(text, "~", "~\u200B")
	return "~" + escaped + "~"
}

func (f *Formatter) EscapeText(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}

func (f *Formatter) FormatTable(headers []string, rows [][]string) string {
	var lines []string
	for _, row := range rows {
		var parts []string
		for i, cell := range row {
			if i < len(headers) && headers[i] != "" {
				parts = append(parts, "*"+headers[i]+":* "+cell)
			} else {
				parts = append(parts, cell)
			}
		}
		lines = append(lines, strings.Join(parts, " · "))
	}
	return strings.Join(lines, "\n")
}

func (f *Formatter) FormatKeyValueList(items [][3]string) string {
	var lines []string
	for _, item := range items {
		icon, label, value := item[0], item[1], item[2]
		lines = append(lines, fmt.Sprintf("%s *%s:* %s", icon, label, value))
	}
	return strings.Join(lines, "\n")
}

func (f *Formatter) FormatMarkdown(content string) string {
	return platform.ConvertMarkdownToSlack(content)
}
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/slack/... -v -run "TestSlack"
```

Fix any test assertion issues. Note: `assert.HasPrefix` doesn't exist in testify — use `assert.True(t, strings.HasPrefix(result, "~"))` and `assert.True(t, strings.HasSuffix(result, "~"))`.

- [ ] **Step 5: Commit**

```bash
git add go/internal/platform/slack/formatter.go go/internal/platform/slack/formatter_test.go && git commit -m "feat(go): add SlackFormatter with tests (Phase 4 Task 2)"
```

---

## Task 3: Add SlackPlatformConfig to config package

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Check if it already exists**

```bash
grep -r "SlackPlatformConfig" internal/config/
```

- [ ] **Step 2: If not found, append to config.go**

```go
// SlackPlatformConfig holds configuration for a Slack platform instance.
type SlackPlatformConfig struct {
	ID              string   `yaml:"id"`
	DisplayName     string   `yaml:"displayName"`
	BotToken        string   `yaml:"botToken"`  // xoxb-...
	AppToken        string   `yaml:"appToken"`  // xapp-...
	ChannelID       string   `yaml:"channelId"`
	BotName         string   `yaml:"botName"`
	AllowedUsers    []string `yaml:"allowedUsers"`
	SkipPermissions bool     `yaml:"skipPermissions"`
	APIURL          string   `yaml:"apiUrl,omitempty"` // override for tests
}
```

- [ ] **Step 3: Build check**

```bash
env -u GOROOT /opt/homebrew/bin/go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add go/internal/config/config.go && git commit -m "feat(go): add SlackPlatformConfig to config package (Phase 4 Task 3)"
```

---

## Task 4: SlackClient — tests first

**Files:**
- Create: `internal/platform/slack/client_test.go`
- Create: `internal/platform/slack/client.go`

- [ ] **Step 1: Write client_test.go (RED)**

```go
package slack_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anneschuth/claude-threads/internal/config"
	"github.com/anneschuth/claude-threads/internal/platform/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newSlackTestServer creates a mock Slack API server.
// handlers maps endpoint names (without leading slash) to handlers.
func newSlackTestServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	for path, h := range handlers {
		mux.HandleFunc("/"+path, h)
	}
	return httptest.NewServer(mux)
}

func testSlackConfig(apiURL string) config.SlackPlatformConfig {
	return config.SlackPlatformConfig{
		ID:           "slack-test",
		DisplayName:  "Test Slack",
		BotToken:     "xoxb-test",
		AppToken:     "xapp-test",
		ChannelID:    "C123456",
		BotName:      "testbot",
		AllowedUsers: []string{"alice", "bob"},
		APIURL:       apiURL,
	}
}

func slackOK(v interface{}) map[string]interface{} {
	b, _ := json.Marshal(v)
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	m["ok"] = true
	return m
}

func TestSlackClient_PlatformIdentity(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	assert.Equal(t, "slack-test", client.PlatformID())
	assert.Equal(t, "slack", client.PlatformType())
	assert.Equal(t, "Test Slack", client.DisplayName())
}

func TestSlackClient_IsUserAllowed(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	assert.True(t, client.IsUserAllowed("alice"))
	assert.False(t, client.IsUserAllowed("charlie"))
}

func TestSlackClient_IsUserAllowed_EmptyList(t *testing.T) {
	cfg := testSlackConfig("http://localhost")
	cfg.AllowedUsers = nil
	client := slack.NewClient(cfg)
	assert.True(t, client.IsUserAllowed("anyone"))
}

func TestSlackClient_IsBotMentioned_ByName(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	assert.True(t, client.IsBotMentioned("@testbot hello"))
	assert.False(t, client.IsBotMentioned("hello world"))
}

func TestSlackClient_ExtractPrompt(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	assert.Equal(t, "hello", client.ExtractPrompt("@testbot hello"))
}

func TestSlackClient_GetMessageLimits(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	limits := client.GetMessageLimits()
	assert.Equal(t, 12000, limits.MaxLength)
}

func TestSlackClient_CreatePost(t *testing.T) {
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer xoxb-test", r.Header.Get("Authorization"))
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "C123456", body["channel"])
			assert.Equal(t, "Hello!", body["text"])
			json.NewEncoder(w).Encode(slackOK(map[string]interface{}{
				"channel": "C123456",
				"ts":      "1234567890.000001",
				"message": map[string]interface{}{"text": "Hello!"},
			}))
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	post, err := client.CreatePost(context.Background(), "Hello!", "")
	require.NoError(t, err)
	assert.Equal(t, "1234567890.000001", post.ID)
}

func TestSlackClient_CreatePost_WithThread(t *testing.T) {
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "1234567890.000000", body["thread_ts"])
			json.NewEncoder(w).Encode(slackOK(map[string]interface{}{
				"channel": "C123456",
				"ts":      "1234567890.000002",
				"message": map[string]interface{}{"text": "Reply"},
			}))
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	post, err := client.CreatePost(context.Background(), "Reply", "1234567890.000000")
	require.NoError(t, err)
	assert.Equal(t, "1234567890.000000", post.RootID)
}

func TestSlackClient_UpdatePost(t *testing.T) {
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"chat.update": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "post-ts", body["ts"])
			json.NewEncoder(w).Encode(slackOK(map[string]interface{}{
				"channel": "C123456",
				"ts":      "post-ts",
				"text":    "Updated!",
			}))
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	post, err := client.UpdatePost(context.Background(), "post-ts", "Updated!")
	require.NoError(t, err)
	assert.Equal(t, "post-ts", post.ID)
}

func TestSlackClient_GetThreadLink_WithTeamURL(t *testing.T) {
	// Test without team URL first (fallback)
	client := slack.NewClient(testSlackConfig("http://localhost"))
	link := client.GetThreadLink("1234567890.123456", "", "")
	assert.Equal(t, "#1234567890.123456", link)
}

func TestSlackClient_AddReaction_ConvertsUnicode(t *testing.T) {
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"reactions.add": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			// 👍 should be converted to "+1"
			assert.Equal(t, "+1", body["name"])
			json.NewEncoder(w).Encode(slackOK(nil))
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	err := client.AddReaction(context.Background(), "post-ts", "👍")
	require.NoError(t, err)
}

func TestSlackClient_UserCaching(t *testing.T) {
	callCount := 0
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"users.info": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			json.NewEncoder(w).Encode(slackOK(map[string]interface{}{
				"user": SlackUser{ID: "U123", Name: "alice", Profile: SlackUserProfile{}},
			}))
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	client.GetUser(context.Background(), "U123")
	client.GetUser(context.Background(), "U123")
	assert.Equal(t, 1, callCount, "should cache user after first fetch")
}

func TestSlackClient_IsBotMentioned_ByID(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	// Simulate having botUserID set
	client.SetBotUserIDForTest("U_BOT123")
	assert.True(t, client.IsBotMentioned("<@U_BOT123> hello"))
	assert.False(t, client.IsBotMentioned("<@U_SOMEONE_ELSE> hello"))
}
```

Note: `TestSlackClient_IsBotMentioned_ByID` uses `client.SetBotUserIDForTest` — you need to add this test helper method to the Client struct.

Also `SlackUser` and `SlackUserProfile` are imported from the slack package in the test — verify the import.

- [ ] **Step 2: Run — expect FAIL**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/slack/... 2>&1 | head -20
```

- [ ] **Step 3: Write client.go (GREEN)**

```go
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
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

var log = utils.CreateLogger("slack")
var wsLog = utils.WsLogger

const (
	maxRateLimitRetries = 5
)

// Client is the Slack Socket Mode implementation of platform.PlatformClient.
type Client struct {
	platform.BasePlatformClient

	platformIDVal  string
	displayNameVal string
	botToken       string
	appToken       string
	channelID      string
	apiURL         string
	formatter      *Formatter

	mu                sync.RWMutex
	userCache         map[string]SlackUser
	usernameToIDCache map[string]string
	botUserID         string
	teamURL           string
	lastTs            string // for missed-message recovery

	// Message deduplication
	processedMessages map[string]struct{}

	// Rate limiting
	rateLimitUntil time.Time

	conn *websocket.Conn
}

// NewClient creates a new SlackClient from config.
func NewClient(cfg config.SlackPlatformConfig) *Client {
	apiURL := cfg.APIURL
	if apiURL == "" {
		apiURL = "https://slack.com/api"
	}
	c := &Client{
		platformIDVal:     cfg.ID,
		displayNameVal:    cfg.DisplayName,
		botToken:          cfg.BotToken,
		appToken:          cfg.AppToken,
		channelID:         cfg.ChannelID,
		apiURL:            apiURL,
		formatter:         NewFormatter(),
		userCache:         make(map[string]SlackUser),
		usernameToIDCache: make(map[string]string),
		processedMessages: make(map[string]struct{}),
	}
	c.BasePlatformClient.AllowedUsers = cfg.AllowedUsers
	c.BasePlatformClient.BotNameVal = cfg.BotName
	c.BasePlatformClient.InitBase(c.Connect, c.forceClose, c.recoverMissed)
	return c
}

// SetBotUserIDForTest sets the bot user ID (test helper only).
func (c *Client) SetBotUserIDForTest(id string) {
	c.mu.Lock()
	c.botUserID = id
	c.mu.Unlock()
}

// Compile-time interface check.
var _ platform.PlatformClient = (*Client)(nil)

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func (c *Client) PlatformID() string   { return c.platformIDVal }
func (c *Client) PlatformType() string { return "slack" }
func (c *Client) DisplayName() string  { return c.displayNameVal }

// ---------------------------------------------------------------------------
// REST API helper
// ---------------------------------------------------------------------------

func (c *Client) botAPI(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	return c.apiWith(ctx, c.botToken, method, endpoint, body, 0, nil)
}

func (c *Client) appAPI(ctx context.Context, endpoint string) ([]byte, error) {
	return c.apiWith(ctx, c.appToken, "POST", endpoint, nil, 0, nil)
}

func (c *Client) apiWith(ctx context.Context, token, method, endpoint string, body interface{}, retries int, expectedErrors []string) ([]byte, error) {
	// Rate limit backoff
	c.mu.RLock()
	until := c.rateLimitUntil
	c.mu.RUnlock()
	if wait := time.Until(until); wait > 0 {
		log.Debug(fmt.Sprintf("Rate limited, waiting %s", wait))
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	url := c.apiURL + "/" + endpoint
	log.Debug(method + " " + endpoint)

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 429 {
		if retries >= maxRateLimitRetries {
			return nil, fmt.Errorf("Slack rate limit exceeded after %d retries", maxRateLimitRetries)
		}
		retryAfter := 5
		if v := resp.Header.Get("Retry-After"); v != "" {
			fmt.Sscanf(v, "%d", &retryAfter)
		}
		wait := time.Duration(retryAfter) * time.Second
		log.Warn(fmt.Sprintf("Rate limited, retrying after %s (attempt %d/%d)", wait, retries+1, maxRateLimitRetries))
		c.mu.Lock()
		c.rateLimitUntil = time.Now().Add(wait)
		c.mu.Unlock()
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return c.apiWith(ctx, token, method, endpoint, body, retries+1, expectedErrors)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Slack HTTP error %d: %s", resp.StatusCode, string(data))
	}

	var base APIResponse
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}
	if !base.OK {
		for _, expected := range expectedErrors {
			if base.Error == expected {
				return nil, fmt.Errorf("Slack API error: %s", base.Error)
			}
		}
		return nil, fmt.Errorf("Slack API error: %s", base.Error)
	}

	return data, nil
}

func (c *Client) apiJSON(ctx context.Context, method, endpoint string, body interface{}, out interface{}, expectedErrors ...string) error {
	data, err := c.apiWith(ctx, c.botToken, method, endpoint, body, 0, expectedErrors)
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

func (c *Client) normUser(u SlackUser) *platform.PlatformUser {
	displayName := u.Profile.DisplayName
	if displayName == "" {
		displayName = u.Profile.RealName
	}
	if displayName == "" {
		displayName = u.RealName
	}
	if displayName == "" {
		displayName = u.Name
	}
	return &platform.PlatformUser{
		ID:          u.ID,
		Username:    u.Name,
		DisplayName: displayName,
		Email:       u.Profile.Email,
	}
}

func (c *Client) fetchBotUser(ctx context.Context) error {
	var auth AuthTestResponse
	if err := c.apiJSON(ctx, "POST", "auth.test", nil, &auth); err != nil {
		return err
	}
	c.mu.Lock()
	c.botUserID = auth.UserID
	c.teamURL = strings.TrimRight(auth.URL, "/")
	c.mu.Unlock()

	var info UsersInfoResponse
	if err := c.apiJSON(ctx, "GET", "users.info?user="+auth.UserID, nil, &info); err != nil {
		return err
	}
	c.mu.Lock()
	c.userCache[auth.UserID] = info.User
	c.mu.Unlock()
	return nil
}

func (c *Client) GetBotUser(ctx context.Context) (*platform.PlatformUser, error) {
	c.mu.RLock()
	botID := c.botUserID
	cached, ok := c.userCache[botID]
	c.mu.RUnlock()
	if ok {
		return c.normUser(cached), nil
	}
	if err := c.fetchBotUser(ctx); err != nil {
		return nil, err
	}
	c.mu.RLock()
	user := c.userCache[c.botUserID]
	c.mu.RUnlock()
	return c.normUser(user), nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*platform.PlatformUser, error) {
	if userID == "" {
		return nil, nil
	}
	c.mu.RLock()
	cached, ok := c.userCache[userID]
	c.mu.RUnlock()
	if ok {
		return c.normUser(cached), nil
	}

	var resp UsersInfoResponse
	if err := c.apiJSON(ctx, "GET", "users.info?user="+userID, nil, &resp); err != nil {
		return nil, nil // treat as not found
	}
	c.mu.Lock()
	c.userCache[userID] = resp.User
	c.usernameToIDCache[resp.User.Name] = userID
	c.mu.Unlock()
	return c.normUser(resp.User), nil
}

func (c *Client) GetUserByUsername(ctx context.Context, username string) (*platform.PlatformUser, error) {
	c.mu.RLock()
	id, ok := c.usernameToIDCache[username]
	c.mu.RUnlock()
	if ok {
		return c.GetUser(ctx, id)
	}

	// Paginate users.list to find by username
	var cursor string
	for {
		endpoint := "users.list?limit=200"
		if cursor != "" {
			endpoint += "&cursor=" + cursor
		}
		var resp UsersListResponse
		if err := c.apiJSON(ctx, "GET", endpoint, nil, &resp); err != nil {
			return nil, nil
		}
		for _, u := range resp.Members {
			c.mu.Lock()
			c.userCache[u.ID] = u
			c.usernameToIDCache[u.Name] = u.ID
			c.mu.Unlock()
			if u.Name == username {
				return c.normUser(u), nil
			}
		}
		if resp.ResponseMetadata == nil || resp.ResponseMetadata.NextCursor == "" {
			break
		}
		cursor = resp.ResponseMetadata.NextCursor
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Normalization
// ---------------------------------------------------------------------------

func (c *Client) normPost(msg SlackMessage, channelID, rootID string) *platform.PlatformPost {
	np := &platform.PlatformPost{
		ID:         msg.Ts,
		PlatformID: c.platformIDVal,
		ChannelID:  channelID,
		UserID:     msg.User,
		Message:    msg.Text,
		RootID:     rootID,
		CreateAt:   tsToMs(msg.Ts),
	}
	for _, f := range msg.Files {
		ext := f.Filetype
		if idx := strings.LastIndex(f.Name, "."); idx >= 0 {
			ext = f.Name[idx+1:]
		}
		np.Files = append(np.Files, platform.PlatformFile{
			ID:        f.ID,
			Name:      f.Name,
			Size:      f.Size,
			MimeType:  f.Mimetype,
			Extension: ext,
		})
	}
	return np
}

func tsToMs(ts string) int64 {
	var f float64
	fmt.Sscanf(ts, "%f", &f)
	return int64(f * 1000)
}

// ---------------------------------------------------------------------------
// Messaging
// ---------------------------------------------------------------------------

func (c *Client) CreatePost(ctx context.Context, message, threadID string) (*platform.PlatformPost, error) {
	// Truncate if needed
	limits := c.GetMessageLimits()
	if len(message) > limits.MaxLength {
		message = platform.TruncateMessageSafely(message, limits.MaxLength, "_... (truncated)_")
	}

	body := map[string]interface{}{
		"channel":       c.channelID,
		"text":          message,
		"unfurl_links":  threadID != "",
		"unfurl_media":  threadID != "",
	}
	if threadID != "" {
		body["thread_ts"] = threadID
	}

	var resp PostMessageResponse
	if err := c.apiJSON(ctx, "POST", "chat.postMessage", body, &resp); err != nil {
		return nil, err
	}
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	return &platform.PlatformPost{
		ID:         resp.Ts,
		PlatformID: c.platformIDVal,
		ChannelID:  resp.Channel,
		UserID:     botID,
		Message:    resp.Message.Text,
		RootID:     threadID,
		CreateAt:   tsToMs(resp.Ts),
	}, nil
}

func (c *Client) UpdatePost(ctx context.Context, postID, message string) (*platform.PlatformPost, error) {
	limits := c.GetMessageLimits()
	if len(message) > limits.MaxLength {
		message = platform.TruncateMessageSafely(message, limits.MaxLength, "_... (truncated)_")
	}

	var resp UpdateMessageResponse
	if err := c.apiJSON(ctx, "POST", "chat.update", map[string]interface{}{
		"channel": c.channelID,
		"ts":      postID,
		"text":    message,
	}, &resp); err != nil {
		return nil, err
	}
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	return &platform.PlatformPost{
		ID:         resp.Ts,
		PlatformID: c.platformIDVal,
		ChannelID:  resp.Channel,
		UserID:     botID,
		Message:    resp.Text,
		CreateAt:   tsToMs(resp.Ts),
	}, nil
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
	var resp ConversationsHistoryResponse
	endpoint := fmt.Sprintf("conversations.history?channel=%s&latest=%s&oldest=%s&inclusive=true&limit=1", c.channelID, postID, postID)
	if err := c.apiJSON(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, nil
	}
	if len(resp.Messages) == 0 {
		return nil, nil
	}
	return c.normPost(resp.Messages[0], c.channelID, ""), nil
}

func (c *Client) DeletePost(ctx context.Context, postID string) error {
	return c.apiJSON(ctx, "POST", "chat.delete", map[string]interface{}{
		"channel": c.channelID,
		"ts":      postID,
	}, nil)
}

func (c *Client) PinPost(ctx context.Context, postID string) error {
	err := c.apiJSON(ctx, "POST", "pins.add", map[string]interface{}{
		"channel":   c.channelID,
		"timestamp": postID,
	}, nil, "already_pinned")
	if err != nil && strings.Contains(err.Error(), "already_pinned") {
		return nil
	}
	return err
}

func (c *Client) UnpinPost(ctx context.Context, postID string) error {
	err := c.apiJSON(ctx, "POST", "pins.remove", map[string]interface{}{
		"channel":   c.channelID,
		"timestamp": postID,
	}, nil, "no_pin")
	if err != nil && strings.Contains(err.Error(), "no_pin") {
		return nil
	}
	return err
}

func (c *Client) GetPinnedPosts(ctx context.Context) ([]string, error) {
	var resp PinsListResponse
	if err := c.apiJSON(ctx, "GET", "pins.list?channel="+c.channelID, nil, &resp); err != nil {
		return nil, err
	}
	var ids []string
	for _, item := range resp.Items {
		if item.Message != nil {
			ids = append(ids, item.Message.Ts)
		}
	}
	return ids, nil
}

func (c *Client) GetMessageLimits() platform.MessageLimits {
	return platform.MessageLimits{MaxLength: 12000, HardThreshold: 10000}
}

func (c *Client) GetThreadHistory(ctx context.Context, threadID string, opts *platform.ThreadHistoryOptions) ([]platform.ThreadMessage, error) {
	limit := 100
	if opts != nil && opts.Limit > 0 {
		limit = opts.Limit
	}
	var resp ConversationsRepliesResponse
	endpoint := fmt.Sprintf("conversations.replies?channel=%s&ts=%s&limit=%d", c.channelID, threadID, limit)
	if err := c.apiJSON(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, nil
	}

	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()

	var msgs []platform.ThreadMessage
	for _, msg := range resp.Messages {
		if opts != nil && opts.ExcludeBotMessages && (msg.User == botID || msg.BotID != "") {
			continue
		}
		user, _ := c.GetUser(ctx, msg.User)
		username := "unknown"
		if user != nil {
			username = user.Username
		}
		msgs = append(msgs, platform.ThreadMessage{
			ID:       msg.Ts,
			UserID:   msg.User,
			Username: username,
			Message:  msg.Text,
			CreateAt: tsToMs(msg.Ts),
		})
	}
	sort.Slice(msgs, func(i, j int) bool { return msgs[i].CreateAt < msgs[j].CreateAt })
	return msgs, nil
}

// ---------------------------------------------------------------------------
// Reactions
// ---------------------------------------------------------------------------

func (c *Client) AddReaction(ctx context.Context, postID, emojiName string) error {
	name := platform.GetEmojiName(emojiName)
	return c.apiJSON(ctx, "POST", "reactions.add", map[string]interface{}{
		"channel":   c.channelID,
		"timestamp": postID,
		"name":      name,
	}, nil)
}

func (c *Client) RemoveReaction(ctx context.Context, postID, emojiName string) error {
	name := platform.GetEmojiName(emojiName)
	return c.apiJSON(ctx, "POST", "reactions.remove", map[string]interface{}{
		"channel":   c.channelID,
		"timestamp": postID,
		"name":      name,
	}, nil)
}

// ---------------------------------------------------------------------------
// Bot mentions
// ---------------------------------------------------------------------------

func (c *Client) IsBotMentioned(message string) bool {
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	if botID != "" && strings.Contains(message, "<@"+botID+">") {
		return true
	}
	botName := regexp.QuoteMeta(c.BotNameVal)
	re := regexp.MustCompile(`(?i)(^|\s)@` + botName + `\b`)
	return re.MatchString(message)
}

func (c *Client) ExtractPrompt(message string) string {
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	prompt := message
	if botID != "" {
		prompt = strings.ReplaceAll(prompt, "<@"+botID+">", "")
	}
	botName := regexp.QuoteMeta(c.BotNameVal)
	re := regexp.MustCompile(`(?i)(^|\s)@` + botName + `\b`)
	return strings.TrimSpace(re.ReplaceAllString(prompt, " "))
}

// ---------------------------------------------------------------------------
// Files
// ---------------------------------------------------------------------------

func (c *Client) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	var resp FilesInfoResponse
	if err := c.apiJSON(ctx, "GET", "files.info?file="+fileID, nil, &resp); err != nil {
		return nil, err
	}
	downloadURL := resp.File.URLPrivateDownload
	if downloadURL == "" {
		downloadURL = resp.File.URLPrivate
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("no download URL for file %s", fileID)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.botToken)
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	return io.ReadAll(httpResp.Body)
}

func (c *Client) GetFileInfo(ctx context.Context, fileID string) (*platform.PlatformFile, error) {
	var resp FilesInfoResponse
	if err := c.apiJSON(ctx, "GET", "files.info?file="+fileID, nil, &resp); err != nil {
		return nil, err
	}
	ext := resp.File.Filetype
	if idx := strings.LastIndex(resp.File.Name, "."); idx >= 0 {
		ext = resp.File.Name[idx+1:]
	}
	return &platform.PlatformFile{
		ID:        resp.File.ID,
		Name:      resp.File.Name,
		Size:      resp.File.Size,
		MimeType:  resp.File.Mimetype,
		Extension: ext,
	}, nil
}

// ---------------------------------------------------------------------------
// Platform helpers
// ---------------------------------------------------------------------------

func (c *Client) GetMcpConfig() platform.McpConfig {
	return platform.McpConfig{
		Type:         "slack",
		URL:          "https://slack.com",
		Token:        c.botToken,
		ChannelID:    c.channelID,
		AllowedUsers: c.AllowedUsers,
	}
}

func (c *Client) GetFormatter() platform.PlatformFormatter { return c.formatter }

func (c *Client) GetThreadLink(threadID, _, lastMessageTs string) string {
	c.mu.RLock()
	teamURL := c.teamURL
	c.mu.RUnlock()
	if teamURL == "" {
		return "#" + threadID
	}
	targetTs := threadID
	if lastMessageTs != "" {
		targetTs = lastMessageTs
	}
	permalinkTs := strings.ReplaceAll(targetTs, ".", "")
	if lastMessageTs != "" && lastMessageTs != threadID {
		return fmt.Sprintf("%s/archives/%s/p%s?thread_ts=%s&cid=%s", teamURL, c.channelID, permalinkTs, threadID, c.channelID)
	}
	return fmt.Sprintf("%s/archives/%s/p%s", teamURL, c.channelID, permalinkTs)
}

func (c *Client) SendTyping(_ string) {} // Slack doesn't support typing indicators for bots

// ---------------------------------------------------------------------------
// WebSocket connection (Socket Mode)
// ---------------------------------------------------------------------------

func (c *Client) Connect(ctx context.Context) error {
	if err := c.fetchBotUser(ctx); err != nil {
		return fmt.Errorf("failed to fetch bot user: %w", err)
	}

	// Get WebSocket URL from apps.connections.open (uses app token)
	data, err := c.appAPI(ctx, "apps.connections.open")
	if err != nil {
		return fmt.Errorf("apps.connections.open: %w", err)
	}
	var connResp AppsConnectionsOpenResponse
	if err := json.Unmarshal(data, &connResp); err != nil {
		return err
	}

	wsLog.Info("Socket Mode: connecting to " + connResp.URL)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, connResp.URL, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	connected := make(chan error, 1)

	go func() {
		for {
			var envelope SocketModeEvent
			if err := conn.ReadJSON(&envelope); err != nil {
				wsLog.Info("Socket Mode disconnected: " + err.Error())
				c.mu.Lock()
				c.conn = nil
				c.mu.Unlock()
				c.BasePlatformClient.OnConnectionClosed()
				return
			}
			c.BasePlatformClient.UpdateLastMessageTime()
			c.handleEnvelope(ctx, conn, envelope, connected)
		}
	}()

	select {
	case err := <-connected:
		return err
	case <-ctx.Done():
		conn.Close()
		return ctx.Err()
	case <-time.After(30 * time.Second):
		conn.Close()
		return fmt.Errorf("Socket Mode connection timeout")
	}
}

func (c *Client) handleEnvelope(ctx context.Context, conn *websocket.Conn, env SocketModeEvent, connected chan<- error) {
	// ACK immediately for events_api
	if env.EnvelopeID != "" {
		conn.WriteJSON(map[string]string{"envelope_id": env.EnvelopeID})
		wsLog.Debug("ACKed " + env.EnvelopeID)
	}

	switch env.Type {
	case "hello":
		c.BasePlatformClient.OnConnectionEstablished()
		select {
		case connected <- nil:
		default:
		}

	case "disconnect":
		wsLog.Info("Socket Mode: received disconnect, reconnecting...")
		conn.Close()

	case "events_api":
		if env.Payload != nil && env.Payload.Event != nil {
			c.handleEvent(ctx, env.Payload.Event)
		}
	}
}

func (c *Client) handleEvent(ctx context.Context, event *SlackEvent) {
	switch event.Type {
	case "message":
		if event.Subtype != "" && event.Subtype != "file_share" {
			return
		}
		c.mu.RLock()
		botID := c.botUserID
		c.mu.RUnlock()
		if event.User == botID || event.BotID != "" {
			return
		}
		if event.Channel != c.channelID {
			return
		}

		// Deduplicate
		c.mu.Lock()
		if _, seen := c.processedMessages[event.Ts]; seen {
			c.mu.Unlock()
			return
		}
		c.processedMessages[event.Ts] = struct{}{}
		if len(c.processedMessages) > 1000 {
			// Remove one old entry
			for k := range c.processedMessages {
				delete(c.processedMessages, k)
				break
			}
		}
		c.lastTs = event.Ts
		c.mu.Unlock()

		rootID := ""
		if event.ThreadTs != "" && event.ThreadTs != event.Ts {
			rootID = event.ThreadTs
		}
		msg := SlackMessage{
			Ts:   event.Ts,
			User: event.User,
			Text: event.Text,
			Files: event.Files,
		}
		np := c.normPost(msg, event.Channel, rootID)
		go func() {
			user, _ := c.GetUser(ctx, event.User)
			c.BasePlatformClient.EmitMessage(*np, user)
			if rootID == "" {
				c.BasePlatformClient.EmitChannelPost(*np, user)
			}
		}()

	case "reaction_added", "reaction_removed":
		if event.Item == nil || event.Item.Type != "message" {
			return
		}
		c.mu.RLock()
		botID := c.botUserID
		c.mu.RUnlock()
		if event.User == botID {
			return
		}
		if event.Item.Channel != c.channelID {
			return
		}
		r := platform.PlatformReaction{
			UserID:    event.User,
			PostID:    event.Item.Ts,
			EmojiName: event.Reaction,
			CreateAt:  time.Now().UnixMilli(),
		}
		go func() {
			user, _ := c.GetUser(ctx, event.User)
			if event.Type == "reaction_added" {
				c.BasePlatformClient.EmitReaction(r, user)
			} else {
				c.BasePlatformClient.EmitReactionRemoved(r, user)
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
	lastTs := c.lastTs
	botID := c.botUserID
	c.mu.RUnlock()
	if lastTs == "" {
		return nil
	}

	log.Info("Recovering missed Slack messages after " + lastTs)
	endpoint := fmt.Sprintf("conversations.history?channel=%s&oldest=%s&inclusive=false&limit=100", c.channelID, lastTs)
	var resp ConversationsHistoryResponse
	if err := c.apiJSON(context.Background(), "GET", endpoint, nil, &resp); err != nil {
		return err
	}

	msgs := resp.Messages
	sort.Slice(msgs, func(i, j int) bool {
		return parseFloat(msgs[i].Ts) < parseFloat(msgs[j].Ts)
	})

	for _, msg := range msgs {
		if msg.User == botID || msg.BotID != "" {
			continue
		}
		c.mu.Lock()
		c.lastTs = msg.Ts
		c.mu.Unlock()

		rootID := ""
		if msg.ThreadTs != "" && msg.ThreadTs != msg.Ts {
			rootID = msg.ThreadTs
		}
		np := c.normPost(msg, c.channelID, rootID)
		user, _ := c.GetUser(context.Background(), msg.User)
		c.BasePlatformClient.EmitMessage(*np, user)
		if rootID == "" {
			c.BasePlatformClient.EmitChannelPost(*np, user)
		}
	}
	return nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// unused but satisfies math import
var _ = math.Pow
```

- [ ] **Step 4: Fix math import — remove if unused**

If `math` is not actually used in the final client.go, remove it from the import block. The `_ = math.Pow` line at the bottom is just a placeholder.

- [ ] **Step 5: Run tests — expect PASS**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./internal/platform/slack/... -v
```

Fix any compilation errors or test failures before proceeding.

- [ ] **Step 6: Commit**

```bash
git add go/internal/platform/slack/ go/internal/config/config.go && git commit -m "feat(go): add SlackClient with Socket Mode (Phase 4 Task 4)"
```

---

## Task 5: Phase 4 final verification

- [ ] **Step 1: Run all tests**

```bash
env -u GOROOT /opt/homebrew/bin/go test ./... 2>&1 | tail -20
```
Expected: all packages GREEN

- [ ] **Step 2: Run go vet**

```bash
env -u GOROOT /opt/homebrew/bin/go vet ./...
```

- [ ] **Step 3: Update progress docs**

Mark Phase 4 complete in:
- `docs/superpowers/plans/2026-03-31-golang-conversion-master.md` (change `| 4 | ... | ⬜ |` to `✅`)
- `docs/plans/golang-conversion.md` (update Phase 4 status row)

- [ ] **Step 4: Commit docs**

```bash
cd /Users/stevemcquaid/Library/CloudStorage/Dropbox/Code/claude-threads/.worktrees/go-conversion && git add docs/ && git commit -m "docs: mark Phase 4 complete"
```
