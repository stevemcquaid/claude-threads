// Package platform defines platform-agnostic types and interfaces for
// multi-platform support. All platform implementations (Mattermost, Slack)
// must satisfy the interfaces defined here.
package platform

// PlatformUser is a normalized user representation across platforms.
type PlatformUser struct {
	ID          string  // Platform-specific user ID
	Username    string  // Login username (e.g., 'alice.smith')
	DisplayName string  // Human-friendly name (e.g., 'Alice Smith')
	Email       string  // Optional email
}

// PlatformFile is a normalized file attachment representation.
type PlatformFile struct {
	ID        string // Platform-specific file ID
	Name      string // Filename
	Size      int64  // File size in bytes
	MimeType  string // MIME type (e.g., 'image/png')
	Extension string // File extension
}

// PlatformPost is a normalized post/message representation.
type PlatformPost struct {
	ID         string            // Platform-specific post ID
	PlatformID string            // Which platform instance this is from
	ChannelID  string            // Channel/conversation ID
	UserID     string            // Author's user ID
	Message    string            // Message text content
	RootID     string            // Thread parent ID (empty if channel-level)
	CreateAt   int64             // Timestamp (ms since epoch)
	Files      []PlatformFile    // Attached files (may be empty)
	Metadata   map[string]any    // Platform-specific metadata
}

// PlatformReaction is a normalized reaction representation.
type PlatformReaction struct {
	UserID    string // User who reacted
	PostID    string // Post that was reacted to
	EmojiName string // Emoji name (e.g., '+1', 'white_check_mark')
	CreateAt  int64  // When the reaction was added (ms since epoch)
}

// ThreadMessage is a normalized thread message for context retrieval.
type ThreadMessage struct {
	ID       string // Message/post ID
	UserID   string // Author's user ID
	Username string // Author's username
	Message  string // Message content
	CreateAt int64  // Timestamp (ms since epoch)
}
