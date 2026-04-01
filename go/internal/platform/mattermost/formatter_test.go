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
		[][]string{{"`!permissions interactive|skip`", "Toggle"}},
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
