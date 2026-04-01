package platform

import "context"

// ReactionEvent is a reaction received from a WebSocket event.
type ReactionEvent struct {
	PostID    string
	UserID    string
	EmojiName string
}

// PostedMessage holds the ID of a newly created post.
type PostedMessage struct {
	ID string
}

// PermissionApi is the interface used by the MCP permission server to post
// permission requests and receive user responses via reactions.
type PermissionApi interface {
	GetFormatter() PlatformFormatter
	GetBotUserID(ctx context.Context) (string, error)
	GetUsername(ctx context.Context, userID string) (string, error)
	IsUserAllowed(username string) bool
	CreateInteractivePost(ctx context.Context, message string, reactions []string, threadID string) (PostedMessage, error)
	UpdatePost(ctx context.Context, postID, message string) error
	WaitForReaction(ctx context.Context, postID, botUserID string, timeoutMs int64) (*ReactionEvent, error)
}

// MattermostPermissionApiConfig holds config for the Mattermost permission API.
type MattermostPermissionApiConfig struct {
	URL          string
	Token        string
	ChannelID    string
	ThreadID     string
	AllowedUsers []string
	Debug        bool
}

// SlackPermissionApiConfig holds config for the Slack permission API.
type SlackPermissionApiConfig struct {
	BotToken     string
	AppToken     string
	ChannelID    string
	ThreadTs     string
	AllowedUsers []string
	Debug        bool
}
