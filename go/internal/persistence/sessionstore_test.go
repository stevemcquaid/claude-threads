package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStore(t *testing.T) *SessionStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sessions.json")
	return NewSessionStore(path)
}

func sampleSession(platformID, threadID string) PersistedSession {
	now := time.Now().UTC().Format(time.RFC3339)
	return PersistedSession{
		PlatformID:          platformID,
		ThreadID:            threadID,
		ClaudeSessionID:     "claude-123",
		StartedBy:           "alice",
		StartedAt:           now,
		SessionNumber:       1,
		WorkingDir:          "/tmp/work",
		SessionAllowedUsers: []string{"alice"},
		LastActivityAt:      now,
	}
}

func TestNewSessionStore_DefaultPath(t *testing.T) {
	// With empty path, constructor should not panic and should use default
	store := NewSessionStore("")
	assert.NotNil(t, store)
	assert.NotEmpty(t, store.sessionsFile)
}

func TestSessionStore_SaveAndLoad(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("platform1", "thread1")

	err := store.Save("platform1:thread1", sess)
	require.NoError(t, err)

	loaded := store.Load()
	assert.Len(t, loaded, 1)
	got, ok := loaded["platform1:thread1"]
	assert.True(t, ok)
	assert.Equal(t, "alice", got.StartedBy)
	assert.Equal(t, "thread1", got.ThreadID)
}

func TestSessionStore_LoadExcludesSoftDeleted(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("platform1", "thread1")

	err := store.Save("platform1:thread1", sess)
	require.NoError(t, err)

	err = store.SoftDelete("platform1:thread1")
	require.NoError(t, err)

	loaded := store.Load()
	assert.Empty(t, loaded, "soft-deleted sessions must not appear in Load()")
}

func TestSessionStore_Remove(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("p1", "t1")

	require.NoError(t, store.Save("p1:t1", sess))
	require.NoError(t, store.Remove("p1:t1"))

	loaded := store.Load()
	assert.Empty(t, loaded)
}

func TestSessionStore_SoftDelete(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("p1", "t1")

	require.NoError(t, store.Save("p1:t1", sess))
	require.NoError(t, store.SoftDelete("p1:t1"))

	// Should not be in active sessions
	loaded := store.Load()
	assert.Empty(t, loaded)

	// But CleanedAt should be set in raw data
	raw := store.loadRaw()
	s, ok := raw.Sessions["p1:t1"]
	assert.True(t, ok)
	assert.NotNil(t, s.CleanedAt, "CleanedAt should be set after soft delete")
	assert.NotEmpty(t, *s.CleanedAt)
}

func TestSessionStore_CleanStale(t *testing.T) {
	store := newTestStore(t)

	// Old session: activity 2 hours ago
	old := sampleSession("p1", "old")
	old.LastActivityAt = time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	require.NoError(t, store.Save("p1:old", old))

	// Recent session: activity 1 minute ago
	recent := sampleSession("p1", "recent")
	require.NoError(t, store.Save("p1:recent", recent))

	// Clean sessions older than 1 hour
	removed, err := store.CleanStale(int64(time.Hour.Milliseconds()))
	require.NoError(t, err)
	assert.Contains(t, removed, "p1:old")
	assert.NotContains(t, removed, "p1:recent")

	loaded := store.Load()
	assert.NotContains(t, loaded, "p1:old")
	assert.Contains(t, loaded, "p1:recent")
}

func TestSessionStore_CleanHistory(t *testing.T) {
	store := newTestStore(t)

	// Soft-deleted session with old cleanedAt
	sess := sampleSession("p1", "old")
	require.NoError(t, store.Save("p1:old", sess))
	require.NoError(t, store.SoftDelete("p1:old"))

	// Manually backdate the CleanedAt
	raw := store.loadRaw()
	s := raw.Sessions["p1:old"]
	oldTime := time.Now().UTC().Add(-10 * 24 * time.Hour).Format(time.RFC3339)
	s.CleanedAt = &oldTime
	raw.Sessions["p1:old"] = s
	require.NoError(t, store.writeRaw(raw))

	// Recent soft-deleted session
	sess2 := sampleSession("p1", "recent")
	require.NoError(t, store.Save("p1:recent", sess2))
	require.NoError(t, store.SoftDelete("p1:recent"))

	// Clean with 3-day retention
	removed, err := store.CleanHistory(int64(3 * 24 * time.Hour.Milliseconds()))
	require.NoError(t, err)
	assert.Equal(t, 1, removed)

	raw2 := store.loadRaw()
	_, oldExists := raw2.Sessions["p1:old"]
	assert.False(t, oldExists, "old soft-deleted session should be permanently removed")
	_, recentExists := raw2.Sessions["p1:recent"]
	assert.True(t, recentExists, "recent soft-deleted session should remain")
}

func TestSessionStore_GetHistory(t *testing.T) {
	store := newTestStore(t)

	// Active session for p1
	active := sampleSession("p1", "active")
	require.NoError(t, store.Save("p1:active", active))

	// Soft-deleted session for p1 (should appear in history)
	deleted := sampleSession("p1", "deleted")
	deleted.LastActivityAt = time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)
	require.NoError(t, store.Save("p1:deleted", deleted))
	require.NoError(t, store.SoftDelete("p1:deleted"))

	// Session for a different platform (should not appear)
	other := sampleSession("p2", "other")
	require.NoError(t, store.Save("p2:other", other))

	activeSessions := map[string]bool{"p1:active": true}
	history := store.GetHistory("p1", activeSessions)

	// Should include soft-deleted sessions; active session excluded
	ids := make([]string, 0, len(history))
	for _, h := range history {
		ids = append(ids, h.PlatformID+":"+h.ThreadID)
	}
	assert.Contains(t, ids, "p1:deleted")
	assert.NotContains(t, ids, "p2:other")
}

func TestSessionStore_StickyPostIDs(t *testing.T) {
	store := newTestStore(t)

	require.NoError(t, store.SaveStickyPostID("platform1", "post-abc"))
	ids := store.GetStickyPostIDs()
	assert.Equal(t, "post-abc", ids["platform1"])

	require.NoError(t, store.RemoveStickyPostID("platform1"))
	ids2 := store.GetStickyPostIDs()
	assert.NotContains(t, ids2, "platform1")
}

func TestSessionStore_PlatformEnabled(t *testing.T) {
	store := newTestStore(t)

	// Default: enabled when not set
	assert.True(t, store.IsPlatformEnabled("platform1"), "should default to true")

	// Set to disabled
	require.NoError(t, store.SetPlatformEnabled("platform1", false))
	assert.False(t, store.IsPlatformEnabled("platform1"))

	// Re-enable
	require.NoError(t, store.SetPlatformEnabled("platform1", true))
	assert.True(t, store.IsPlatformEnabled("platform1"))

	// GetPlatformEnabledState
	state := store.GetPlatformEnabledState()
	enabled, exists := state["platform1"]
	assert.True(t, exists)
	assert.True(t, enabled)
}

func TestSessionStore_FindByThread(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("platform1", "thread-xyz")
	require.NoError(t, store.Save("platform1:thread-xyz", sess))

	found := store.FindByThread("platform1", "thread-xyz")
	require.NotNil(t, found)
	assert.Equal(t, "thread-xyz", found.ThreadID)

	notFound := store.FindByThread("platform1", "nonexistent")
	assert.Nil(t, notFound)
}

func TestSessionStore_FindByPostID(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("p1", "t1")
	postID := "post-999"
	sess.SessionStartPostID = &postID
	require.NoError(t, store.Save("p1:t1", sess))

	found := store.FindByPostID("p1", "post-999")
	require.NotNil(t, found)
	assert.Equal(t, "t1", found.ThreadID)

	notFound := store.FindByPostID("p1", "no-such-post")
	assert.Nil(t, notFound)
}

func TestSessionStore_FindByPostID_LifecyclePostID(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("p1", "t2")
	lifecyclePostID := "lifecycle-post-42"
	sess.LifecyclePostID = &lifecyclePostID
	require.NoError(t, store.Save("p1:t2", sess))

	found := store.FindByPostID("p1", "lifecycle-post-42")
	require.NotNil(t, found)
	assert.Equal(t, "t2", found.ThreadID)
}

func TestSessionStore_MigrationV1toV2(t *testing.T) {
	tmpDir := t.TempDir()
	sessFile := filepath.Join(tmpDir, "sessions.json")

	// Write a v1 JSON file (sessions keyed by threadId, no platformId field)
	v1Data := map[string]interface{}{
		"version": 1,
		"sessions": map[string]interface{}{
			"thread-abc": map[string]interface{}{
				"threadId":        "thread-abc",
				"claudeSessionId": "cs-001",
				"startedBy":       "bob",
				"startedAt":       "2024-01-01T00:00:00Z",
				"sessionNumber":   1,
				"workingDir":      "/tmp",
				"sessionAllowedUsers": []string{"bob"},
				"lastActivityAt":  "2024-01-01T00:00:00Z",
			},
		},
		"stickyPostIds":        map[string]interface{}{},
		"platformEnabledState": map[string]interface{}{},
	}
	data, err := json.Marshal(v1Data)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(sessFile, data, 0600))

	store := NewSessionStore(sessFile)
	loaded := store.Load()

	// After migration, session key should be "default:thread-abc"
	_, ok := loaded["default:thread-abc"]
	assert.True(t, ok, "v1 session should be migrated to 'default:thread-abc'")

	// PlatformID should be set to "default"
	sess := loaded["default:thread-abc"]
	assert.Equal(t, "default", sess.PlatformID)
}

func TestSessionStore_AtomicWrite(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("p1", "t1")
	require.NoError(t, store.Save("p1:t1", sess))

	// After save, no .tmp file should remain
	tmpFile := store.sessionsFile + ".tmp"
	_, err := os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err), "tmp file should not exist after successful save")

	// The main file should exist
	_, err = os.Stat(store.sessionsFile)
	assert.NoError(t, err, "sessions file should exist")

	// Verify file permissions (0600)
	info, err := os.Stat(store.sessionsFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestSessionStore_Clear(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("p1", "t1")
	require.NoError(t, store.Save("p1:t1", sess))
	require.NoError(t, store.SaveStickyPostID("p1", "sticky-1"))

	require.NoError(t, store.Clear())

	loaded := store.Load()
	assert.Empty(t, loaded)

	ids := store.GetStickyPostIDs()
	assert.Empty(t, ids)
}

// Helper: verify session key normalization
func TestSessionStore_KeyNormalization(t *testing.T) {
	store := newTestStore(t)
	sess := sampleSession("myplatform", "mythread")
	require.NoError(t, store.Save("myplatform:mythread", sess))

	loaded := store.Load()
	assert.Contains(t, loaded, "myplatform:mythread")

	// FindByThread should work
	found := store.FindByThread("myplatform", "mythread")
	require.NotNil(t, found)

	// Keys with colons in threadID should survive round-trip
	sess2 := sampleSession("p", "thread:with:colons")
	require.NoError(t, store.Save("p:thread:with:colons", sess2))
	found2 := store.FindByThread("p", "thread:with:colons")
	require.NotNil(t, found2)
	assert.Equal(t, "thread:with:colons", found2.ThreadID)
	_ = strings.Contains // suppress import
}
