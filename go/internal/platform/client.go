package platform

import "context"

// MessageLimits holds platform-specific message size constraints.
type MessageLimits struct {
	MaxLength     int // Absolute max characters
	HardThreshold int // When to force continuation
}

// McpConfig holds config for the MCP permission server.
type McpConfig struct {
	Type         string
	URL          string
	Token        string
	ChannelID    string
	AllowedUsers []string
}

// ThreadHistoryOptions controls how thread history is fetched.
type ThreadHistoryOptions struct {
	Limit              int
	ExcludeBotMessages bool
}

// PlatformClient is the platform-agnostic client interface.
// All platform implementations (Mattermost, Slack) must satisfy this interface.
//
// Events are delivered via callbacks registered with On* methods.
// Each On* method replaces the previous callback (not additive).
type PlatformClient interface {
	// Identity

	PlatformID() string   // e.g., 'mattermost-internal'
	PlatformType() string // e.g., 'mattermost', 'slack'
	DisplayName() string  // e.g., 'Internal Team'

	// Connection Management

	Connect(ctx context.Context) error
	Disconnect()
	// PrepareForReconnect resets internal state (intentionalDisconnect flag,
	// reconnect attempts) so that Connect() can be called again.
	PrepareForReconnect()

	// User Management

	GetBotUser(ctx context.Context) (*PlatformUser, error)
	GetUser(ctx context.Context, userID string) (*PlatformUser, error)
	GetUserByUsername(ctx context.Context, username string) (*PlatformUser, error)
	IsUserAllowed(username string) bool
	GetBotName() string
	GetMcpConfig() McpConfig
	GetFormatter() PlatformFormatter
	// GetThreadLink returns a clickable URL for the thread.
	GetThreadLink(threadID, lastMessageID, lastMessageTs string) string

	// Messaging

	CreatePost(ctx context.Context, message, threadID string) (*PlatformPost, error)
	UpdatePost(ctx context.Context, postID, message string) (*PlatformPost, error)
	CreateInteractivePost(ctx context.Context, message string, reactions []string, threadID string) (*PlatformPost, error)
	GetPost(ctx context.Context, postID string) (*PlatformPost, error)
	DeletePost(ctx context.Context, postID string) error
	PinPost(ctx context.Context, postID string) error
	UnpinPost(ctx context.Context, postID string) error
	GetPinnedPosts(ctx context.Context) ([]string, error)
	GetMessageLimits() MessageLimits
	GetThreadHistory(ctx context.Context, threadID string, opts *ThreadHistoryOptions) ([]ThreadMessage, error)

	// Reactions

	AddReaction(ctx context.Context, postID, emojiName string) error
	RemoveReaction(ctx context.Context, postID, emojiName string) error

	// Bot Mentions

	IsBotMentioned(message string) bool
	ExtractPrompt(message string) string

	// Typing Indicator

	SendTyping(threadID string)

	// Files (optional — may be nil for platforms that don't support it)

	DownloadFile(ctx context.Context, fileID string) ([]byte, error)
	GetFileInfo(ctx context.Context, fileID string) (*PlatformFile, error)

	// Event Callbacks
	// Each On* method replaces the previous callback; pass nil to remove.

	OnConnected(func())
	OnDisconnected(func())
	OnReconnecting(func(attempt int))
	OnError(func(err error))
	OnMessage(func(post PlatformPost, user *PlatformUser))
	OnReaction(func(reaction PlatformReaction, user *PlatformUser))
	OnReactionRemoved(func(reaction PlatformReaction, user *PlatformUser))
	OnChannelPost(func(post PlatformPost, user *PlatformUser))
}
