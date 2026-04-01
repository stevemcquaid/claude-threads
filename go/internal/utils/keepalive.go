package utils

import (
	"io"
	"os/exec"
	"runtime"
	"sync"
)

// KeepAliveManager prevents system sleep while sessions are active.
type KeepAliveManager struct {
	mu           sync.Mutex
	enabled      bool
	sessionCount int
	cmd          *exec.Cmd
}

// NewKeepAliveManager creates a new manager. Enabled by default.
func NewKeepAliveManager() *KeepAliveManager {
	return &KeepAliveManager{enabled: true}
}

func (k *KeepAliveManager) SetEnabled(enabled bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.enabled = enabled
	if !enabled && k.cmd != nil {
		k.stopProcess()
	}
}

func (k *KeepAliveManager) IsEnabled() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.enabled
}

func (k *KeepAliveManager) IsActive() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.cmd != nil
}

func (k *KeepAliveManager) SessionStarted() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.sessionCount++
	if k.enabled && k.cmd == nil {
		k.startProcess()
	}
}

func (k *KeepAliveManager) SessionEnded() {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.sessionCount > 0 {
		k.sessionCount--
	}
	if k.sessionCount == 0 && k.cmd != nil {
		k.stopProcess()
	}
}

func (k *KeepAliveManager) ForceStop() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.sessionCount = 0
	if k.cmd != nil {
		k.stopProcess()
	}
}

func (k *KeepAliveManager) GetSessionCount() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.sessionCount
}

func (k *KeepAliveManager) startProcess() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("caffeinate", "-s", "-i")
	case "linux":
		if _, err := exec.LookPath("systemd-inhibit"); err == nil {
			cmd = exec.Command("systemd-inhibit", "--what=sleep:idle", "--who=claude-threads", "--why=Active sessions", "--mode=block", "sleep", "infinity")
		}
	}
	if cmd == nil {
		return
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err == nil {
		k.cmd = cmd
	}
}

func (k *KeepAliveManager) stopProcess() {
	if k.cmd != nil && k.cmd.Process != nil {
		_ = k.cmd.Process.Kill()
		_ = k.cmd.Wait()
		k.cmd = nil
	}
}

// KeepAlive is the package-level singleton.
var KeepAlive = NewKeepAliveManager()
