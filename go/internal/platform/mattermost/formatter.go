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

var excessNewlinesRe = regexp.MustCompile(`\n{3,}`)

func (f *Formatter) FormatMarkdown(content string) string {
	// Fix code blocks with text immediately after closing ```
	// Go's regexp doesn't support lookbehind, so we use a line-based approach.
	result := fixCodeBlockNewlines(content)
	// Normalize excessive newlines
	result = excessNewlinesRe.ReplaceAllString(result, "\n\n")
	return result
}

// fixCodeBlockNewlines adds a newline between ``` and immediately following text.
// This handles the case where closing ``` is not followed by a newline.
// Example: "```\ncode\n```Text" → "```\ncode\n```\nText"
func fixCodeBlockNewlines(content string) string {
	// Track whether we are inside a code block to distinguish opening from closing markers.
	lines := strings.Split(content, "\n")
	insideCodeBlock := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "```") {
			if !insideCodeBlock {
				// This is an opening marker (``` or ```lang) — enter the code block.
				insideCodeBlock = true
			} else {
				// We are inside a code block — this is a closing marker.
				if len(line) > 3 {
					// Closing ``` has text immediately after it — split into two lines.
					rest := line[3:]
					lines[i] = "```"
					newLines := make([]string, 0, len(lines)+1)
					newLines = append(newLines, lines[:i+1]...)
					newLines = append(newLines, rest)
					newLines = append(newLines, lines[i+1:]...)
					lines = newLines
				}
				insideCodeBlock = false
			}
		}
	}
	return strings.Join(lines, "\n")
}
