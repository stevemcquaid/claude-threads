// Package mattermost provides a Mattermost implementation of PlatformClient.
package mattermost

// WebSocketEvent is a Mattermost WebSocket event envelope.
type WebSocketEvent struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Broadcast struct {
		ChannelID string `json:"channel_id"`
		UserID    string `json:"user_id"`
		TeamID    string `json:"team_id"`
	} `json:"broadcast"`
	Seq int `json:"seq"`
}

// File is a Mattermost file attachment.
type File struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	Extension string `json:"extension"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

// Post is a Mattermost post from the API or WebSocket.
type Post struct {
	ID        string                 `json:"id"`
	CreateAt  int64                  `json:"create_at"`
	UpdateAt  int64                  `json:"update_at"`
	DeleteAt  int64                  `json:"delete_at"`
	UserID    string                 `json:"user_id"`
	ChannelID string                 `json:"channel_id"`
	RootID    string                 `json:"root_id"`
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	Props     map[string]interface{} `json:"props"`
	FileIDs   []string               `json:"file_ids,omitempty"`
	Metadata  *PostMetadata          `json:"metadata,omitempty"`
}

// PostMetadata holds embedded metadata for a post.
type PostMetadata struct {
	Files []File `json:"files,omitempty"`
}

// User is a Mattermost user from the API.
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Nickname  string `json:"nickname"`
}

// Reaction is a Mattermost reaction.
type Reaction struct {
	UserID    string `json:"user_id"`
	PostID    string `json:"post_id"`
	EmojiName string `json:"emoji_name"`
	CreateAt  int64  `json:"create_at"`
}

// PostedEventData is the data field for 'posted' WebSocket events.
type PostedEventData struct {
	ChannelDisplayName string `json:"channel_display_name"`
	ChannelName        string `json:"channel_name"`
	ChannelType        string `json:"channel_type"`
	Post               string `json:"post"` // JSON-encoded Post
	SenderName         string `json:"sender_name"`
	TeamID             string `json:"team_id"`
}

// ReactionEventData is the data field for 'reaction_added'/'reaction_removed' events.
type ReactionEventData struct {
	Reaction string `json:"reaction"` // JSON-encoded Reaction
}

// CreatePostRequest is the body for POST /posts.
type CreatePostRequest struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
	RootID    string `json:"root_id,omitempty"`
}

// UpdatePostRequest is the body for PUT /posts/{id}.
type UpdatePostRequest struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// ThreadResponse is the response for GET /posts/{id}/thread.
type ThreadResponse struct {
	Order []string        `json:"order"`
	Posts map[string]Post `json:"posts"`
}

// ChannelPostsResponse is the response for GET /channels/{id}/posts.
type ChannelPostsResponse struct {
	Order []string        `json:"order"`
	Posts map[string]Post `json:"posts"`
}

// PinnedPostsResponse is the response for GET /channels/{id}/pinned.
type PinnedPostsResponse struct {
	Order []string        `json:"order"`
	Posts map[string]Post `json:"posts"`
}
