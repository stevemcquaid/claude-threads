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
