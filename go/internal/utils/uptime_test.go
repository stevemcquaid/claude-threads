package utils_test

import (
	"testing"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestFormatUptime(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "<1m", utils.FormatUptime(now.Add(-30*time.Second)))
	assert.Equal(t, "5m", utils.FormatUptime(now.Add(-5*time.Minute)))
	assert.Equal(t, "1h23m", utils.FormatUptime(now.Add(-(1*time.Hour + 23*time.Minute))))
	assert.Equal(t, "2h", utils.FormatUptime(now.Add(-2*time.Hour)))
	assert.Equal(t, "1d5h", utils.FormatUptime(now.Add(-(1*24*time.Hour + 5*time.Hour))))
	assert.Equal(t, "2d", utils.FormatUptime(now.Add(-2*24*time.Hour)))
}
