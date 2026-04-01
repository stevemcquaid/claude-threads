package slack_test

import (
	"strings"
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
	result := f.FormatStrikethrough("~/some/path")
	assert.True(t, strings.HasPrefix(result, "~"), "should start with ~")
	assert.True(t, strings.HasSuffix(result, "~"), "should end with ~")
	assert.Contains(t, result, "\u200B") // zero-width space after each ~
}

func TestSlackFormatHeading(t *testing.T) {
	f := slack.NewFormatter()
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
	assert.NotContains(t, result, "|")
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
