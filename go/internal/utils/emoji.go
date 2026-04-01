package utils

// ApprovalEmojis are reactions indicating approval/yes.
var ApprovalEmojis = []string{"+1", "thumbsup"}

// DenialEmojis are reactions indicating denial/no.
var DenialEmojis = []string{"-1", "thumbsdown"}

// AllowAllEmojis are reactions indicating "allow all permissions".
var AllowAllEmojis = []string{"white_check_mark", "heavy_check_mark"}

// NumberEmojis are reactions used for numbered choices (0-indexed).
var NumberEmojis = []string{"one", "two", "three", "four"}

// CancelEmojis are reactions that cancel/stop a session.
var CancelEmojis = []string{"x", "octagonal_sign", "stop_sign", "stop"}

// EscapeEmojis are reactions that interrupt/pause a session.
var EscapeEmojis = []string{"double_vertical_bar", "pause_button", "pause"}

// ResumeEmojis are reactions that resume a paused session.
var ResumeEmojis = []string{"arrows_counterclockwise", "arrow_forward", "repeat"}

// MinimizeToggleEmojis are reactions that toggle minimize/expand.
var MinimizeToggleEmojis = []string{"arrow_down_small", "small_red_triangle_down"}

// BugReportEmoji is the reaction that triggers a bug report.
const BugReportEmoji = "bug"

// unicodeNumberEmojis maps unicode number emojis to 0-based indices.
var unicodeNumberEmojis = []string{"1️⃣", "2️⃣", "3️⃣", "4️⃣"}

func containsEmoji(list []string, emoji string) bool {
	for _, e := range list {
		if e == emoji {
			return true
		}
	}
	return false
}

// IsApprovalEmoji returns true if the emoji indicates approval.
func IsApprovalEmoji(emoji string) bool { return containsEmoji(ApprovalEmojis, emoji) }

// IsDenialEmoji returns true if the emoji indicates denial.
func IsDenialEmoji(emoji string) bool { return containsEmoji(DenialEmojis, emoji) }

// IsAllowAllEmoji returns true if the emoji means "allow all permissions".
func IsAllowAllEmoji(emoji string) bool { return containsEmoji(AllowAllEmojis, emoji) }

// IsCancelEmoji returns true if the emoji cancels a session.
func IsCancelEmoji(emoji string) bool { return containsEmoji(CancelEmojis, emoji) }

// IsEscapeEmoji returns true if the emoji pauses/interrupts a session.
func IsEscapeEmoji(emoji string) bool { return containsEmoji(EscapeEmojis, emoji) }

// IsResumeEmoji returns true if the emoji resumes a session.
func IsResumeEmoji(emoji string) bool { return containsEmoji(ResumeEmojis, emoji) }

// IsMinimizeToggleEmoji returns true if the emoji toggles minimize/expand.
func IsMinimizeToggleEmoji(emoji string) bool { return containsEmoji(MinimizeToggleEmojis, emoji) }

// IsBugReportEmoji returns true if the emoji triggers a bug report.
// Handles both the text name "bug" and the unicode variant "🐛".
func IsBugReportEmoji(emoji string) bool {
	return emoji == BugReportEmoji || emoji == "🐛"
}

// GetNumberEmojiIndex returns the 0-based index for a number emoji,
// or -1 if the emoji is not a number emoji.
// Handles both text names ("one", "two") and unicode variants ("1️⃣", "2️⃣").
func GetNumberEmojiIndex(emoji string) int {
	for i, e := range NumberEmojis {
		if e == emoji {
			return i
		}
	}
	for i, e := range unicodeNumberEmojis {
		if e == emoji {
			return i
		}
	}
	return -1
}
