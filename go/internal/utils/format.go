package utils

import (
	"fmt"
	"strings"
	"time"
)

// ExtractThreadID extracts the thread ID from a composite session ID.
// Composite IDs have the format "platformId:threadId".
// Returns the original string if no colon is present.
func ExtractThreadID(sessionID string) string {
	idx := strings.LastIndex(sessionID, ":")
	if idx < 0 {
		return sessionID
	}
	return sessionID[idx+1:]
}

// FormatShortID formats an ID for display, showing at most 8 characters
// with an ellipsis. For composite IDs (platformId:threadId), extracts
// the thread portion first.
func FormatShortID(id string) string {
	threadID := ExtractThreadID(id)
	if len(threadID) <= 8 {
		return threadID
	}
	return threadID[:8] + "…"
}

// FormatDuration formats a duration in milliseconds to a human-readable string.
// Examples: "5s", "1m 30s", "2h", "1h 30m"
func FormatDuration(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		if m > 0 {
			return fmt.Sprintf("%dh %dm", h, m)
		}
		return fmt.Sprintf("%dh", h)
	}
	if m > 0 {
		if s > 0 {
			return fmt.Sprintf("%dm %ds", m, s)
		}
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", s)
}

// FormatRelativeTimeShort formats a time relative to now in short format.
// Examples: "<1m ago", "5m ago", "2h ago", "3d ago"
func FormatRelativeTimeShort(t time.Time) string {
	d := time.Since(t)
	minutes := int(d.Minutes())
	hours := int(d.Hours())
	days := int(d.Hours() / 24)

	if minutes < 1 {
		return "<1m ago"
	}
	if hours < 1 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	if days < 1 {
		return fmt.Sprintf("%dh ago", hours)
	}
	return fmt.Sprintf("%dd ago", days)
}

// TruncateAtWord truncates a string to maxLength characters with an ellipsis.
// Breaks at a word boundary if the last space is past 70% of maxLength;
// otherwise hard-truncates.
func TruncateAtWord(s string, maxLength int) string {
	runes := []rune(s)
	if len(runes) <= maxLength {
		return s
	}
	sub := string(runes[:maxLength])
	lastSpace := strings.LastIndex(sub, " ")
	threshold := int(float64(maxLength) * 0.7)
	if lastSpace > threshold {
		return string(runes[:lastSpace]) + "…"
	}
	return sub + "…"
}
