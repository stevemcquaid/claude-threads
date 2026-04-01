package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

// WorktreeMode controls git worktree behavior for sessions.
type WorktreeMode string

const (
	WorktreeModeOff     WorktreeMode = "off"
	WorktreeModePrompt  WorktreeMode = "prompt"
	WorktreeModeRequire WorktreeMode = "require"
)

// ThreadLogsConfig controls thread log retention.
type ThreadLogsConfig struct {
	Enabled       *bool `yaml:"enabled,omitempty"`
	RetentionDays *int  `yaml:"retentionDays,omitempty"`
}

// LimitsConfig controls resource limits. All fields are optional (nil = use default).
type LimitsConfig struct {
	MaxSessions              *int  `yaml:"maxSessions,omitempty"`
	SessionTimeoutMinutes    *int  `yaml:"sessionTimeoutMinutes,omitempty"`
	SessionWarningMinutes    *int  `yaml:"sessionWarningMinutes,omitempty"`
	CleanupIntervalMinutes   *int  `yaml:"cleanupIntervalMinutes,omitempty"`
	MaxWorktreeAgeHours      *int  `yaml:"maxWorktreeAgeHours,omitempty"`
	CleanupWorktrees         *bool `yaml:"cleanupWorktrees,omitempty"`
	PermissionTimeoutSeconds *int  `yaml:"permissionTimeoutSeconds,omitempty"`
}

// ResolvedLimits contains all limits with defaults applied.
type ResolvedLimits struct {
	MaxSessions              int
	SessionTimeoutMinutes    int
	SessionWarningMinutes    int
	CleanupIntervalMinutes   int
	MaxWorktreeAgeHours      int
	CleanupWorktrees         bool
	PermissionTimeoutSeconds int
}

// ResolveLimits merges LimitsConfig with defaults.
// Also reads legacy environment variables: MAX_SESSIONS, SESSION_TIMEOUT_MS.
func ResolveLimits(limits *LimitsConfig) ResolvedLimits {
	r := ResolvedLimits{
		MaxSessions:              5,
		SessionTimeoutMinutes:    30,
		SessionWarningMinutes:    5,
		CleanupIntervalMinutes:   60,
		MaxWorktreeAgeHours:      24,
		CleanupWorktrees:         true,
		PermissionTimeoutSeconds: 120,
	}

	if v := os.Getenv("MAX_SESSIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			r.MaxSessions = n
		}
	}
	if v := os.Getenv("SESSION_TIMEOUT_MS"); v != "" {
		if ms, err := strconv.ParseInt(v, 10, 64); err == nil {
			r.SessionTimeoutMinutes = int(ms / 60000)
		}
	}

	if limits == nil {
		return r
	}
	if limits.MaxSessions != nil {
		r.MaxSessions = *limits.MaxSessions
	}
	if limits.SessionTimeoutMinutes != nil {
		r.SessionTimeoutMinutes = *limits.SessionTimeoutMinutes
	}
	if limits.SessionWarningMinutes != nil {
		r.SessionWarningMinutes = *limits.SessionWarningMinutes
	}
	if limits.CleanupIntervalMinutes != nil {
		r.CleanupIntervalMinutes = *limits.CleanupIntervalMinutes
	}
	if limits.MaxWorktreeAgeHours != nil {
		r.MaxWorktreeAgeHours = *limits.MaxWorktreeAgeHours
	}
	if limits.CleanupWorktrees != nil {
		r.CleanupWorktrees = *limits.CleanupWorktrees
	}
	if limits.PermissionTimeoutSeconds != nil {
		r.PermissionTimeoutSeconds = *limits.PermissionTimeoutSeconds
	}
	return r
}

// StickyMessageCustomization customizes the sticky channel message.
type StickyMessageCustomization struct {
	Description string `yaml:"description,omitempty"`
	Footer      string `yaml:"footer,omitempty"`
}

// AutoUpdateConfig controls auto-update behavior.
type AutoUpdateConfig struct {
	Enabled         *bool  `yaml:"enabled,omitempty"`
	Channel         string `yaml:"channel,omitempty"`
	CheckIntervalMs *int64 `yaml:"checkIntervalMs,omitempty"`
}

// PlatformInstanceConfig is the base config for any platform instance.
type PlatformInstanceConfig struct {
	ID          string `yaml:"id"`
	Type        string `yaml:"type"`
	DisplayName string `yaml:"displayName"`
}

// MattermostPlatformConfig is config for a Mattermost instance.
type MattermostPlatformConfig struct {
	PlatformInstanceConfig `yaml:",inline"`
	URL                    string   `yaml:"url"`
	Token                  string   `yaml:"token"`
	ChannelID              string   `yaml:"channelId"`
	BotName                string   `yaml:"botName"`
	AllowedUsers           []string `yaml:"allowedUsers"`
	SkipPermissions        bool     `yaml:"skipPermissions"`
}

// SlackPlatformConfig is config for a Slack instance.
type SlackPlatformConfig struct {
	PlatformInstanceConfig `yaml:",inline"`
	BotToken               string   `yaml:"botToken"`
	AppToken               string   `yaml:"appToken"`
	ChannelID              string   `yaml:"channelId"`
	BotName                string   `yaml:"botName"`
	AllowedUsers           []string `yaml:"allowedUsers"`
	SkipPermissions        bool     `yaml:"skipPermissions"`
	APIURL                 string   `yaml:"apiUrl,omitempty"`
}

// Config is the top-level configuration. Maps exactly to config.yaml.
type Config struct {
	Version       int                         `yaml:"version"`
	WorkingDir    string                      `yaml:"workingDir"`
	Chrome        bool                        `yaml:"chrome"`
	WorktreeMode  WorktreeMode                `yaml:"worktreeMode"`
	KeepAlive     *bool                       `yaml:"keepAlive,omitempty"`
	AutoUpdate    *AutoUpdateConfig           `yaml:"autoUpdate,omitempty"`
	ThreadLogs    *ThreadLogsConfig           `yaml:"threadLogs,omitempty"`
	Limits        *LimitsConfig               `yaml:"limits,omitempty"`
	StickyMessage *StickyMessageCustomization `yaml:"stickyMessage,omitempty"`
	Platforms     []PlatformInstanceConfig    `yaml:"platforms"`
}

// ConfigPath is the default path to the config file (~/.config/claude-threads/config.yaml).
var ConfigPath = filepath.Join(mustHomeDir(), ".config", "claude-threads", "config.yaml")

func mustHomeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return h
}

// LoadConfig loads a Config from the given YAML file path.
// Returns nil (no error) if the file does not exist.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes cfg as YAML to path with 0600 permissions.
// Creates parent directories with 0700 permissions if needed.
func SaveConfig(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	if info, err := os.Stat(dir); err == nil && info.Mode().Perm() != 0o700 {
		_ = os.Chmod(dir, 0o700)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// ConfigExistsAt returns true if a config file exists at the given path.
func ConfigExistsAt(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ConfigExists returns true if the default config file exists.
func ConfigExists() bool {
	return ConfigExistsAt(ConfigPath)
}

// LoadDefaultConfig loads the config from the default path.
func LoadDefaultConfig() (*Config, error) {
	return LoadConfig(ConfigPath)
}
