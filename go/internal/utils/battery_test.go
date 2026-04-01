package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetBatteryStatus_DoesNotPanic(t *testing.T) {
	status, err := utils.GetBatteryStatus()
	assert.NoError(t, err)
	if status != nil {
		assert.GreaterOrEqual(t, status.Percentage, 0)
		assert.LessOrEqual(t, status.Percentage, 100)
	}
}

func TestFormatBatteryStatus_DoesNotPanic(t *testing.T) {
	result, err := utils.FormatBatteryStatus()
	assert.NoError(t, err)
	if result != nil {
		assert.NotEmpty(t, *result)
	}
}

func TestBatteryStatusShape(t *testing.T) {
	status := &utils.BatteryStatus{Percentage: 85, Charging: false}
	assert.Equal(t, 85, status.Percentage)
	assert.False(t, status.Charging)
}
