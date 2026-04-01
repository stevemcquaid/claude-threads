// Package slack provides a Slack implementation of PlatformClient using Socket Mode.
package slack

// SocketModeEvent is a Slack Socket Mode envelope.
type SocketModeEvent struct {
	EnvelopeID             string        `json:"envelope_id"`
	Type                   string        `json:"type"` // events_api, hello, disconnect
	AcceptsResponsePayload bool          `json:"accepts_response_payload"`
	RetryAttempt           int           `json:"retry_attempt"`
	RetryReason            string        `json:"retry_reason"`
	Payload                *EventPayload `json:"payload,omitempty"`
}

// EventPayload wraps the inner event for events_api envelopes.
type EventPayload struct {
	TeamID  string      `json:"team_id"`
	Event   *SlackEvent `json:"event,omitempty"`
	Type    string      `json:"type"`
	EventID string      `json:"event_id"`
}

// SlackEvent is the inner event (message, reaction_added, etc.).
type SlackEvent struct {
	Type        string        `json:"type"`
	Subtype     string        `json:"subtype,omitempty"`
	User        string        `json:"user,omitempty"`
	Channel     string        `json:"channel,omitempty"`
	Ts          string        `json:"ts,omitempty"`
	ThreadTs    string        `json:"thread_ts,omitempty"`
	Text        string        `json:"text,omitempty"`
	Reaction    string        `json:"reaction,omitempty"`
	Item        *ReactionItem `json:"item,omitempty"`
	ItemUser    string        `json:"item_user,omitempty"`
	BotID       string        `json:"bot_id,omitempty"`
	Files       []SlackFile   `json:"files,omitempty"`
	ChannelType string        `json:"channel_type,omitempty"`
}

// ReactionItem identifies the target of a reaction event.
type ReactionItem struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
}

// SlackMessage is a Slack message (from conversations.history/replies).
type SlackMessage struct {
	Type     string      `json:"type"`
	Subtype  string      `json:"subtype,omitempty"`
	Ts       string      `json:"ts"`
	User     string      `json:"user,omitempty"`
	BotID    string      `json:"bot_id,omitempty"`
	Text     string      `json:"text"`
	ThreadTs string      `json:"thread_ts,omitempty"`
	Files    []SlackFile `json:"files,omitempty"`
}

// SlackUser is a Slack user from the API.
type SlackUser struct {
	ID       string           `json:"id"`
	TeamID   string           `json:"team_id"`
	Name     string           `json:"name"`
	Deleted  bool             `json:"deleted"`
	RealName string           `json:"real_name,omitempty"`
	Profile  SlackUserProfile `json:"profile"`
	IsBot    bool             `json:"is_bot,omitempty"`
}

// SlackUserProfile holds profile fields.
type SlackUserProfile struct {
	DisplayName string `json:"display_name,omitempty"`
	RealName    string `json:"real_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// SlackFile is a Slack file attachment.
type SlackFile struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Mimetype           string `json:"mimetype"`
	Filetype           string `json:"filetype"`
	Size               int64  `json:"size"`
	URLPrivate         string `json:"url_private,omitempty"`
	URLPrivateDownload string `json:"url_private_download,omitempty"`
}

// SlackPin is a pinned item.
type SlackPin struct {
	Type    string        `json:"type"`
	Message *SlackMessage `json:"message,omitempty"`
}

// --- API Response types ---

// APIResponse is the base Slack API response envelope.
type APIResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// PostMessageResponse is the response for chat.postMessage.
type PostMessageResponse struct {
	APIResponse
	Channel string       `json:"channel"`
	Ts      string       `json:"ts"`
	Message SlackMessage `json:"message"`
}

// UpdateMessageResponse is the response for chat.update.
type UpdateMessageResponse struct {
	APIResponse
	Channel string `json:"channel"`
	Ts      string `json:"ts"`
	Text    string `json:"text"`
}

// ConversationsRepliesResponse is the response for conversations.replies.
type ConversationsRepliesResponse struct {
	APIResponse
	Messages []SlackMessage `json:"messages"`
	HasMore  bool           `json:"has_more"`
}

// ConversationsHistoryResponse is the response for conversations.history.
type ConversationsHistoryResponse struct {
	APIResponse
	Messages         []SlackMessage `json:"messages"`
	HasMore          bool           `json:"has_more"`
	ResponseMetadata *struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata,omitempty"`
}

// UsersInfoResponse is the response for users.info.
type UsersInfoResponse struct {
	APIResponse
	User SlackUser `json:"user"`
}

// UsersListResponse is the response for users.list.
type UsersListResponse struct {
	APIResponse
	Members          []SlackUser `json:"members"`
	ResponseMetadata *struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata,omitempty"`
}

// AuthTestResponse is the response for auth.test.
type AuthTestResponse struct {
	APIResponse
	URL    string `json:"url"`
	Team   string `json:"team"`
	User   string `json:"user"`
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
	BotID  string `json:"bot_id,omitempty"`
}

// AppsConnectionsOpenResponse is the response for apps.connections.open.
type AppsConnectionsOpenResponse struct {
	APIResponse
	URL string `json:"url"`
}

// PinsListResponse is the response for pins.list.
type PinsListResponse struct {
	APIResponse
	Items []SlackPin `json:"items"`
}

// FilesInfoResponse is the response for files.info.
type FilesInfoResponse struct {
	APIResponse
	File SlackFile `json:"file"`
}
