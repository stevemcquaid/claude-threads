package utils

import (
	"fmt"
	"time"
)

// FormatUptime formats the duration since startedAt in a compact format.
// Examples: "<1m", "5m", "1h23m", "2h", "1d5h", "2d"
func FormatUptime(startedAt time.Time) string {
	d := time.Since(startedAt)
	totalMinutes := int(d.Minutes())
	totalHours := int(d.Hours())
	days := totalHours / 24
	hours := totalHours % 24
	minutes := totalMinutes % 60

	if totalMinutes < 1 {
		return "<1m"
	}
	if totalHours < 1 {
		return fmt.Sprintf("%dm", totalMinutes)
	}
	if days < 1 {
		if minutes > 0 {
			return fmt.Sprintf("%dh%dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}
