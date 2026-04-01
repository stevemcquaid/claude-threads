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

func (f *Formatter) FormatListItem(text string) string       { return "- " + text }
func (f *Formatter) FormatBlockquote(text string) string     { return "> " + text }
func (f *Formatter) FormatHorizontalRule() string            { return "━━━━━━━━━━━━━━━━━━━━" }
func (f *Formatter) FormatHeading(text string, _ int) string { return "*" + text + "*" }

func (f *Formatter) FormatNumberedListItem(number int, text string) string {
	return fmt.Sprintf("%d. %s", number, text)
}

func (f *Formatter) FormatStrikethrough(text string) string {
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
