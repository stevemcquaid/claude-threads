package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestColors(t *testing.T) {
	assert.NotEmpty(t, utils.ColorReset)
	assert.NotEmpty(t, utils.ColorBold)
	assert.NotEmpty(t, utils.ColorDim)
	assert.NotEmpty(t, utils.ColorCyan)
	assert.NotEmpty(t, utils.ColorGreen)
	assert.NotEmpty(t, utils.ColorRed)
	assert.NotEmpty(t, utils.ColorYellow)
	assert.NotEmpty(t, utils.ColorBlue)
	assert.NotEmpty(t, utils.ColorOrange)
}

func TestColorHelpers(t *testing.T) {
	text := "hello"

	dimmed := utils.Dim(text)
	assert.Contains(t, dimmed, text)
	assert.Contains(t, dimmed, utils.ColorDim)
	assert.Contains(t, dimmed, utils.ColorReset)

	bolded := utils.Bold(text)
	assert.Contains(t, bolded, text)
	assert.Contains(t, bolded, utils.ColorBold)
	assert.Contains(t, bolded, utils.ColorReset)

	greened := utils.Green(text)
	assert.Contains(t, greened, text)
	assert.Contains(t, greened, utils.ColorGreen)

	redded := utils.Red(text)
	assert.Contains(t, redded, text)
	assert.Contains(t, redded, utils.ColorRed)
}
