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
