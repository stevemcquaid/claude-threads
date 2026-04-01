// Package persistence provides session storage and thread logging for claude-threads.
package persistence

import (
	"github.com/anneschuth/claude-threads/internal/platform"
)

// WorktreeInfo holds information about a git worktree associated with a session.
type WorktreeInfo struct {
	RepoRoot     string `json:"repoRoot"`
	WorktreePath string `json:"worktreePath"`
	Branch       string `json:"branch"`
}

// ContextPromptFile is a file reference within a context prompt.
type ContextPromptFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PersistedContextPrompt holds state for a pending context prompt.
type PersistedContextPrompt struct {
	PostID             string              `json:"postId"`
	QueuedPrompt       string              `json:"queuedPrompt,omitempty"`
	QueuedFiles        []ContextPromptFile `json:"queuedFiles,omitempty"`
	ThreadMessageCount int                 `json:"threadMessageCount"`
	CreatedAt          string              `json:"createdAt"`
	AvailableOptions   []int               `json:"availableOptions,omitempty"`
}

// PersistedSession is the full persisted state of a session.
type PersistedSession struct {
	PlatformID                     string                  `json:"platformId"`
	ThreadID                       string                  `json:"threadId"`
	ClaudeSessionID                string                  `json:"claudeSessionId"`
	StartedBy                      string                  `json:"startedBy"`
	StartedByDisplayName           string                  `json:"startedByDisplayName,omitempty"`
	StartedAt                      string                  `json:"startedAt"`
	SessionNumber                  int                     `json:"sessionNumber"`
	WorkingDir                     string                  `json:"workingDir"`
	SessionAllowedUsers            []string                `json:"sessionAllowedUsers"`
	ForceInteractivePermissions    bool                    `json:"forceInteractivePermissions"`
	SessionStartPostID             *string                 `json:"sessionStartPostId,omitempty"`
	TasksPostID                    *string                 `json:"tasksPostId,omitempty"`
	LastTasksContent               *string                 `json:"lastTasksContent,omitempty"`
	TasksCompleted                 *bool                   `json:"tasksCompleted,omitempty"`
	TasksMinimized                 *bool                   `json:"tasksMinimized,omitempty"`
	LastActivityAt                 string                  `json:"lastActivityAt"`
	PlanApproved                   bool                    `json:"planApproved"`
	WorktreeInfo                   *WorktreeInfo           `json:"worktreeInfo,omitempty"`
	IsWorktreeOwner                *bool                   `json:"isWorktreeOwner,omitempty"`
	PendingWorktreePrompt          *bool                   `json:"pendingWorktreePrompt,omitempty"`
	WorktreePromptDisabled         *bool                   `json:"worktreePromptDisabled,omitempty"`
	QueuedPrompt                   *string                 `json:"queuedPrompt,omitempty"`
	QueuedFiles                    []platform.PlatformFile `json:"queuedFiles,omitempty"`
	FirstPrompt                    *string                 `json:"firstPrompt,omitempty"`
	PendingContextPrompt           *PersistedContextPrompt `json:"pendingContextPrompt,omitempty"`
	NeedsContextPromptOnNextMessage *bool                  `json:"needsContextPromptOnNextMessage,omitempty"`
	LifecyclePostID                *string                 `json:"lifecyclePostId,omitempty"`
	IsPaused                       *bool                   `json:"isPaused,omitempty"`
	SessionTitle                   *string                 `json:"sessionTitle,omitempty"`
	SessionDescription             *string                 `json:"sessionDescription,omitempty"`
	SessionTags                    []string                `json:"sessionTags,omitempty"`
	PullRequestURL                 *string                 `json:"pullRequestUrl,omitempty"`
	MessageCount                   *int                    `json:"messageCount,omitempty"`
	ResumeFailCount                *int                    `json:"resumeFailCount,omitempty"`
	CleanedAt                      *string                 `json:"cleanedAt,omitempty"`
}

// SessionStoreData is the top-level JSON structure for sessions.json.
type SessionStoreData struct {
	Version              int                        `json:"version"`
	Sessions             map[string]PersistedSession `json:"sessions"`
	StickyPostIDs        map[string]string          `json:"stickyPostIds"`
	PlatformEnabledState map[string]bool            `json:"platformEnabledState"`
}
