package utils_test

import (
	"testing"

	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestKeepAlive_InitialState(t *testing.T) {
	ka := utils.NewKeepAliveManager()
	assert.Equal(t, 0, ka.GetSessionCount())
	assert.True(t, ka.IsEnabled())
	assert.False(t, ka.IsActive())
}

func TestKeepAlive_Disable(t *testing.T) {
	ka := utils.NewKeepAliveManager()
	ka.SetEnabled(false)
	assert.False(t, ka.IsEnabled())
}

func TestKeepAlive_SessionCount(t *testing.T) {
	ka := utils.NewKeepAliveManager()
	ka.SetEnabled(false)
	ka.SessionStarted()
	assert.Equal(t, 1, ka.GetSessionCount())
	ka.SessionStarted()
	assert.Equal(t, 2, ka.GetSessionCount())
	ka.SessionEnded()
	assert.Equal(t, 1, ka.GetSessionCount())
}

func TestKeepAlive_NoNegativeCount(t *testing.T) {
	ka := utils.NewKeepAliveManager()
	ka.SetEnabled(false)
	ka.SessionEnded()
	assert.Equal(t, 0, ka.GetSessionCount())
}

func TestKeepAlive_ForceStop(t *testing.T) {
	ka := utils.NewKeepAliveManager()
	ka.SetEnabled(false)
	ka.SessionStarted()
	ka.SessionStarted()
	ka.ForceStop()
	assert.Equal(t, 0, ka.GetSessionCount())
}

func TestKeepAlive_IsActiveWhenDisabled(t *testing.T) {
	ka := utils.NewKeepAliveManager()
	ka.SetEnabled(false)
	ka.SessionStarted()
	assert.False(t, ka.IsActive())
}
