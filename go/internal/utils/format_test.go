package utils_test

import (
	"testing"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestExtractThreadID(t *testing.T) {
	assert.Equal(t, "thread123", utils.ExtractThreadID("platform:thread123"))
	assert.Equal(t, "abc", utils.ExtractThreadID("abc"))
	assert.Equal(t, "c", utils.ExtractThreadID("a:b:c"))
	assert.Equal(t, "", utils.ExtractThreadID(""))
}

func TestFormatShortID(t *testing.T) {
	assert.Equal(t, "abc12345", utils.FormatShortID("abc12345"))
	assert.Equal(t, "ab", utils.FormatShortID("ab"))
	assert.Equal(t, "abc12345…", utils.FormatShortID("abc123456789"))
	assert.Equal(t, "", utils.FormatShortID(""))
	// Composite ID: extracts thread part then truncates
	assert.Equal(t, "abc12345…", utils.FormatShortID("platform:abc123456789"))
}

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "5s", utils.FormatDuration(5000))
	assert.Equal(t, "45s", utils.FormatDuration(45000))
	assert.Equal(t, "1m 30s", utils.FormatDuration(90000))
	assert.Equal(t, "2m", utils.FormatDuration(120000))
	assert.Equal(t, "1h", utils.FormatDuration(3600000))
	assert.Equal(t, "1h 30m", utils.FormatDuration(5400000))
	assert.Equal(t, "2h", utils.FormatDuration(7200000))
	assert.Equal(t, "0s", utils.FormatDuration(0))
}

func TestFormatRelativeTimeShort(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "<1m ago", utils.FormatRelativeTimeShort(now.Add(-30*time.Second)))
	assert.Equal(t, "5m ago", utils.FormatRelativeTimeShort(now.Add(-5*time.Minute)))
	assert.Equal(t, "2h ago", utils.FormatRelativeTimeShort(now.Add(-2*time.Hour)))
	assert.Equal(t, "3d ago", utils.FormatRelativeTimeShort(now.Add(-3*24*time.Hour)))
}

func TestTruncateAtWord(t *testing.T) {
	// Within limit → unchanged
	assert.Equal(t, "hello", utils.TruncateAtWord("hello", 10))
	// Hard truncate: string "hello world" maxLen=8, last space at idx 5 (5/8=62.5% < 70%) → hard truncate
	result := utils.TruncateAtWord("hello world", 8)
	assert.Contains(t, result, "…")
	// Word break: "hello world extra" maxLen=14, space at 11 (11/14=78.5% > 70%) → word break
	result2 := utils.TruncateAtWord("hello world extra text here", 14)
	assert.Contains(t, result2, "…")
	assert.LessOrEqual(t, len([]rune(result2)), 15)
}
