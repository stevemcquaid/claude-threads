package persistence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
)

const storeVersion = 2

// SessionStore persists session state to disk as JSON.
type SessionStore struct {
	mu           sync.Mutex
	sessionsFile string
	log          *utils.Logger
}

// NewSessionStore creates a SessionStore. If path is empty, uses the default path
// (~/.config/claude-threads/sessions.json), overridable by CLAUDE_THREADS_SESSIONS_PATH.
func NewSessionStore(path string) *SessionStore {
	if path == "" {
		if env := os.Getenv("CLAUDE_THREADS_SESSIONS_PATH"); env != "" {
			path = env
		} else {
			home, _ := os.UserHomeDir()
			path = filepath.Join(home, ".config", "claude-threads", "sessions.json")
		}
	}
	log := utils.CreateLogger("persist")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		log.Warn("could not create config directory: "+err.Error())
	}
	return &SessionStore{sessionsFile: path, log: log}
}

// loadRaw reads the raw store data from disk. Returns an empty store if file doesn't exist.
func (s *SessionStore) loadRaw() SessionStoreData {
	data, err := os.ReadFile(s.sessionsFile)
	if os.IsNotExist(err) || len(data) == 0 {
		return SessionStoreData{
			Version:              storeVersion,
			Sessions:             make(map[string]PersistedSession),
			StickyPostIDs:        make(map[string]string),
			PlatformEnabledState: make(map[string]bool),
		}
	}
	if err != nil {
		s.log.Warn(fmt.Sprintf("failed to read sessions file: %v", err))
		return SessionStoreData{
			Version:              storeVersion,
			Sessions:             make(map[string]PersistedSession),
			StickyPostIDs:        make(map[string]string),
			PlatformEnabledState: make(map[string]bool),
		}
	}

	// Use a raw map for migration
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawData); err != nil {
		s.log.Warn(fmt.Sprintf("failed to parse sessions file: %v", err))
		return SessionStoreData{
			Version:              storeVersion,
			Sessions:             make(map[string]PersistedSession),
			StickyPostIDs:        make(map[string]string),
			PlatformEnabledState: make(map[string]bool),
		}
	}

	var version int
	if v, ok := rawData["version"]; ok {
		_ = json.Unmarshal(v, &version)
	}

	var store SessionStoreData
	if err := json.Unmarshal(data, &store); err != nil {
		s.log.Warn(fmt.Sprintf("failed to unmarshal sessions file: %v", err))
		return SessionStoreData{
			Version:              storeVersion,
			Sessions:             make(map[string]PersistedSession),
			StickyPostIDs:        make(map[string]string),
			PlatformEnabledState: make(map[string]bool),
		}
	}

	if store.Sessions == nil {
		store.Sessions = make(map[string]PersistedSession)
	}
	if store.StickyPostIDs == nil {
		store.StickyPostIDs = make(map[string]string)
	}
	if store.PlatformEnabledState == nil {
		store.PlatformEnabledState = make(map[string]bool)
	}

	// Migrate v1 → v2: add platformId='default' and rekey sessions
	if version < 2 {
		store = migrateV1toV2(store)
	}

	return store
}

// migrateV1toV2 converts v1 session keys (threadId) to v2 keys (platformId:threadId).
func migrateV1toV2(store SessionStoreData) SessionStoreData {
	newSessions := make(map[string]PersistedSession, len(store.Sessions))
	for key, sess := range store.Sessions {
		if sess.PlatformID == "" {
			sess.PlatformID = "default"
		}
		// Only rekey if the key doesn't already have platformId: prefix
		if !strings.Contains(key, ":") {
			newKey := sess.PlatformID + ":" + key
			// Set threadID from key if not set
			if sess.ThreadID == "" {
				sess.ThreadID = key
			}
			newSessions[newKey] = sess
		} else {
			newSessions[key] = sess
		}
	}
	store.Sessions = newSessions
	store.Version = storeVersion
	return store
}

// writeRaw writes the raw store data atomically: write to .tmp then rename.
func (s *SessionStore) writeRaw(store SessionStoreData) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sessions: %w", err)
	}

	tmpFile := s.sessionsFile + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("write tmp sessions file: %w", err)
	}

	if err := os.Rename(tmpFile, s.sessionsFile); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("rename sessions file: %w", err)
	}

	// Ensure correct permissions
	if err := os.Chmod(s.sessionsFile, 0600); err != nil {
		s.log.Warn(fmt.Sprintf("failed to chmod sessions file: %v", err))
	}

	return nil
}

// Load returns only active (non-soft-deleted) sessions.
func (s *SessionStore) Load() map[string]PersistedSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	result := make(map[string]PersistedSession)
	for k, sess := range raw.Sessions {
		if sess.CleanedAt == nil {
			result[k] = sess
		}
	}
	return result
}

// Save persists a session with the given ID.
func (s *SessionStore) Save(sessionID string, session PersistedSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	raw.Sessions[sessionID] = session
	return s.writeRaw(raw)
}

// Remove permanently removes a session.
func (s *SessionStore) Remove(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	delete(raw.Sessions, sessionID)
	return s.writeRaw(raw)
}

// SoftDelete marks a session as deleted by setting CleanedAt to now.
func (s *SessionStore) SoftDelete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	sess, ok := raw.Sessions[sessionID]
	if !ok {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	sess.CleanedAt = &now
	raw.Sessions[sessionID] = sess
	return s.writeRaw(raw)
}

// CleanStale soft-deletes sessions whose LastActivityAt is older than maxAgeMs milliseconds.
// Returns the session IDs that were soft-deleted.
func (s *SessionStore) CleanStale(maxAgeMs int64) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	cutoff := time.Now().UTC().Add(-time.Duration(maxAgeMs) * time.Millisecond)
	var removed []string

	now := time.Now().UTC().Format(time.RFC3339)
	for key, sess := range raw.Sessions {
		if sess.CleanedAt != nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, sess.LastActivityAt)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			sess.CleanedAt = &now
			raw.Sessions[key] = sess
			removed = append(removed, key)
		}
	}

	if len(removed) > 0 {
		if err := s.writeRaw(raw); err != nil {
			return nil, err
		}
	}
	return removed, nil
}

// CleanHistory permanently removes soft-deleted sessions older than historyRetentionMs milliseconds.
// Returns the number of sessions permanently removed.
func (s *SessionStore) CleanHistory(historyRetentionMs int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	cutoff := time.Now().UTC().Add(-time.Duration(historyRetentionMs) * time.Millisecond)
	count := 0

	for key, sess := range raw.Sessions {
		if sess.CleanedAt == nil {
			continue
		}
		t, err := time.Parse(time.RFC3339, *sess.CleanedAt)
		if err != nil {
			continue
		}
		if t.Before(cutoff) {
			delete(raw.Sessions, key)
			count++
		}
	}

	if count > 0 {
		if err := s.writeRaw(raw); err != nil {
			return 0, err
		}
	}
	return count, nil
}

// GetHistory returns soft-deleted (inactive) sessions for a platform, sorted by most recent activity.
// Active sessions (in activeSessions) are excluded.
func (s *SessionStore) GetHistory(platformID string, activeSessions map[string]bool) []PersistedSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	var history []PersistedSession

	for key, sess := range raw.Sessions {
		if sess.PlatformID != platformID {
			continue
		}
		if activeSessions[key] {
			continue
		}
		history = append(history, sess)
	}

	// Sort by LastActivityAt descending (most recent first)
	sort.Slice(history, func(i, j int) bool {
		ti, _ := time.Parse(time.RFC3339, history[i].LastActivityAt)
		tj, _ := time.Parse(time.RFC3339, history[j].LastActivityAt)
		return ti.After(tj)
	})

	return history
}

// Clear permanently removes all sessions and sticky post IDs.
func (s *SessionStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	empty := SessionStoreData{
		Version:              storeVersion,
		Sessions:             make(map[string]PersistedSession),
		StickyPostIDs:        make(map[string]string),
		PlatformEnabledState: make(map[string]bool),
	}
	return s.writeRaw(empty)
}

// SaveStickyPostID saves the sticky post ID for a platform.
func (s *SessionStore) SaveStickyPostID(platformID, postID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	raw.StickyPostIDs[platformID] = postID
	return s.writeRaw(raw)
}

// GetStickyPostIDs returns all sticky post IDs keyed by platform ID.
func (s *SessionStore) GetStickyPostIDs() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	result := make(map[string]string, len(raw.StickyPostIDs))
	for k, v := range raw.StickyPostIDs {
		result[k] = v
	}
	return result
}

// RemoveStickyPostID removes the sticky post ID for a platform.
func (s *SessionStore) RemoveStickyPostID(platformID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	delete(raw.StickyPostIDs, platformID)
	return s.writeRaw(raw)
}

// GetPlatformEnabledState returns the enabled/disabled state for all platforms.
func (s *SessionStore) GetPlatformEnabledState() map[string]bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	result := make(map[string]bool, len(raw.PlatformEnabledState))
	for k, v := range raw.PlatformEnabledState {
		result[k] = v
	}
	return result
}

// IsPlatformEnabled returns true if the platform is enabled (defaults to true if not set).
func (s *SessionStore) IsPlatformEnabled(platformID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	enabled, exists := raw.PlatformEnabledState[platformID]
	if !exists {
		return true
	}
	return enabled
}

// SetPlatformEnabled sets the enabled state for a platform.
func (s *SessionStore) SetPlatformEnabled(platformID string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	raw.PlatformEnabledState[platformID] = enabled
	return s.writeRaw(raw)
}

// FindByThread returns the active session with the given platformID and threadID, or nil.
func (s *SessionStore) FindByThread(platformID, threadID string) *PersistedSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	for _, sess := range raw.Sessions {
		if sess.CleanedAt != nil {
			continue
		}
		if sess.PlatformID == platformID && sess.ThreadID == threadID {
			copy := sess
			return &copy
		}
	}
	return nil
}

// FindByPostID returns the active session that has the given post ID as
// SessionStartPostID or LifecyclePostID, or nil.
func (s *SessionStore) FindByPostID(platformID, postID string) *PersistedSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	raw := s.loadRaw()
	for _, sess := range raw.Sessions {
		if sess.CleanedAt != nil {
			continue
		}
		if sess.PlatformID != platformID {
			continue
		}
		if sess.SessionStartPostID != nil && *sess.SessionStartPostID == postID {
			copy := sess
			return &copy
		}
		if sess.LifecyclePostID != nil && *sess.LifecyclePostID == postID {
			copy := sess
			return &copy
		}
	}
	return nil
}
