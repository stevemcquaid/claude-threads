package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestIsApprovalEmoji(t *testing.T) {
	assert.True(t, utils.IsApprovalEmoji("+1"))
	assert.True(t, utils.IsApprovalEmoji("thumbsup"))
	assert.False(t, utils.IsApprovalEmoji("x"))
	assert.False(t, utils.IsApprovalEmoji(""))
	for _, e := range utils.ApprovalEmojis {
		assert.True(t, utils.IsApprovalEmoji(e), "should match %s", e)
	}
}

func TestIsDenialEmoji(t *testing.T) {
	assert.True(t, utils.IsDenialEmoji("-1"))
	assert.True(t, utils.IsDenialEmoji("thumbsdown"))
	assert.False(t, utils.IsDenialEmoji("+1"))
	for _, e := range utils.DenialEmojis {
		assert.True(t, utils.IsDenialEmoji(e))
	}
}

func TestIsAllowAllEmoji(t *testing.T) {
	assert.True(t, utils.IsAllowAllEmoji("white_check_mark"))
	assert.True(t, utils.IsAllowAllEmoji("heavy_check_mark"))
	assert.False(t, utils.IsAllowAllEmoji("+1"))
	for _, e := range utils.AllowAllEmojis {
		assert.True(t, utils.IsAllowAllEmoji(e))
	}
}

func TestIsCancelEmoji(t *testing.T) {
	assert.True(t, utils.IsCancelEmoji("x"))
	assert.True(t, utils.IsCancelEmoji("octagonal_sign"))
	assert.True(t, utils.IsCancelEmoji("stop_sign"))
	assert.False(t, utils.IsCancelEmoji("+1"))
	for _, e := range utils.CancelEmojis {
		assert.True(t, utils.IsCancelEmoji(e))
	}
}

func TestIsEscapeEmoji(t *testing.T) {
	assert.True(t, utils.IsEscapeEmoji("double_vertical_bar"))
	assert.True(t, utils.IsEscapeEmoji("pause_button"))
	assert.False(t, utils.IsEscapeEmoji("x"))
	for _, e := range utils.EscapeEmojis {
		assert.True(t, utils.IsEscapeEmoji(e))
	}
}

func TestIsResumeEmoji(t *testing.T) {
	assert.True(t, utils.IsResumeEmoji("arrows_counterclockwise"))
	assert.True(t, utils.IsResumeEmoji("arrow_forward"))
	assert.True(t, utils.IsResumeEmoji("repeat"))
	assert.False(t, utils.IsResumeEmoji("x"))
	for _, e := range utils.ResumeEmojis {
		assert.True(t, utils.IsResumeEmoji(e))
	}
}

func TestIsMinimizeToggleEmoji(t *testing.T) {
	assert.True(t, utils.IsMinimizeToggleEmoji("arrow_down_small"))
	assert.True(t, utils.IsMinimizeToggleEmoji("small_red_triangle_down"))
	assert.False(t, utils.IsMinimizeToggleEmoji("x"))
	for _, e := range utils.MinimizeToggleEmojis {
		assert.True(t, utils.IsMinimizeToggleEmoji(e))
	}
}

func TestIsBugReportEmoji(t *testing.T) {
	assert.True(t, utils.IsBugReportEmoji("bug"))
	assert.True(t, utils.IsBugReportEmoji("🐛"))
	assert.False(t, utils.IsBugReportEmoji("x"))
	assert.False(t, utils.IsBugReportEmoji(""))
}

func TestGetNumberEmojiIndex(t *testing.T) {
	// Text names
	assert.Equal(t, 0, utils.GetNumberEmojiIndex("one"))
	assert.Equal(t, 1, utils.GetNumberEmojiIndex("two"))
	assert.Equal(t, 2, utils.GetNumberEmojiIndex("three"))
	assert.Equal(t, 3, utils.GetNumberEmojiIndex("four"))
	// Unicode variants
	assert.Equal(t, 0, utils.GetNumberEmojiIndex("1️⃣"))
	assert.Equal(t, 1, utils.GetNumberEmojiIndex("2️⃣"))
	assert.Equal(t, 2, utils.GetNumberEmojiIndex("3️⃣"))
	assert.Equal(t, 3, utils.GetNumberEmojiIndex("4️⃣"))
	// Non-number
	assert.Equal(t, -1, utils.GetNumberEmojiIndex("x"))
	assert.Equal(t, -1, utils.GetNumberEmojiIndex(""))
	// All NumberEmojis
	for i, e := range utils.NumberEmojis {
		assert.Equal(t, i, utils.GetNumberEmojiIndex(e))
	}
}
