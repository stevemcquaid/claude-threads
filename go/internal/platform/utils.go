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
	"thumbsup":               "+1",
	"thumbs_up":              "+1",
	"thumbsdown":             "-1",
	"thumbs_down":            "-1",
	"heavy_check_mark":       "white_check_mark",
	"x":                      "x",
	"cross_mark":             "x",
	"heavy_multiplication_x": "x",
	"pause_button":           "pause",
	"double_vertical_bar":    "pause",
	"play_button":            "arrow_forward",
	"stop_button":            "stop",
	"octagonal_sign":         "stop",
	"1":                      "one",
	"2":                      "two",
	"3":                      "three",
	"4":                      "four",
	"5":                      "five",
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

// ConvertMarkdownTablesToSlack converts markdown tables to Slack list format.
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
