package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/anneschuth/claude-threads/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPath_Defined(t *testing.T) {
	assert.NotEmpty(t, config.ConfigPath)
	assert.Contains(t, config.ConfigPath, "config.yaml")
}

func TestResolveLimits_Defaults(t *testing.T) {
	limits := config.ResolveLimits(nil)
	assert.Equal(t, 5, limits.MaxSessions)
	assert.Equal(t, 30, limits.SessionTimeoutMinutes)
	assert.Equal(t, 5, limits.SessionWarningMinutes)
	assert.Equal(t, 60, limits.CleanupIntervalMinutes)
	assert.Equal(t, 24, limits.MaxWorktreeAgeHours)
	assert.True(t, limits.CleanupWorktrees)
	assert.Equal(t, 120, limits.PermissionTimeoutSeconds)
}

func TestResolveLimits_MergesOverrides(t *testing.T) {
	maxSessions := 10
	limits := config.ResolveLimits(&config.LimitsConfig{MaxSessions: &maxSessions})
	assert.Equal(t, 10, limits.MaxSessions)
	assert.Equal(t, 30, limits.SessionTimeoutMinutes)
}

func TestResolveLimits_LegacyEnvMaxSessions(t *testing.T) {
	os.Setenv("MAX_SESSIONS", "15")
	defer os.Unsetenv("MAX_SESSIONS")
	limits := config.ResolveLimits(nil)
	assert.Equal(t, 15, limits.MaxSessions)
}

func TestSaveConfig_CreatesFileWithCorrectPermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &config.Config{
		Version:    1,
		WorkingDir: "/tmp",
		Chrome:     false,
		Platforms:  []config.PlatformInstanceConfig{},
	}

	err := config.SaveConfig(cfg, path)
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestSaveConfig_WritesValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &config.Config{
		Version:    1,
		WorkingDir: "/home/user",
		Chrome:     true,
		Platforms:  []config.PlatformInstanceConfig{},
	}

	require.NoError(t, config.SaveConfig(cfg, path))

	loaded, err := config.LoadConfig(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	assert.Equal(t, "/home/user", loaded.WorkingDir)
	assert.True(t, loaded.Chrome)
}

func TestLoadConfig_ReturnsNilForMissingFile(t *testing.T) {
	loaded, err := config.LoadConfig("/nonexistent/path/config.yaml")
	assert.NoError(t, err)
	assert.Nil(t, loaded)
}

func TestConfigExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	assert.False(t, config.ConfigExistsAt(path))

	require.NoError(t, os.WriteFile(path, []byte("version: 1\n"), 0o600))
	assert.True(t, config.ConfigExistsAt(path))
}
