package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/anneschuth/claude-threads/internal/config"
	"github.com/anneschuth/claude-threads/internal/platform"
	"github.com/anneschuth/claude-threads/internal/utils"
	"github.com/gorilla/websocket"
)

var log = utils.CreateLogger("mattermost")
var wsLog2 = utils.WsLogger

const (
	maxRetries     = 3
	retryDelayBase = 500 * time.Millisecond
)

// Client is the Mattermost implementation of platform.PlatformClient.
type Client struct {
	platform.BasePlatformClient

	platformIDVal  string
	displayNameVal string
	url            string
	token          string
	channelID      string
	formatter      *Formatter

	mu         sync.RWMutex
	userCache  map[string]User
	botUserID  string
	lastPostID string // for missed-message recovery

	conn *websocket.Conn
}

// NewClient creates a new MattermostClient from config.
func NewClient(cfg config.MattermostPlatformConfig) *Client {
	c := &Client{
		platformIDVal:  cfg.ID,
		displayNameVal: cfg.DisplayName,
		url:            cfg.URL,
		token:          cfg.Token,
		channelID:      cfg.ChannelID,
		formatter:      NewFormatter(),
		userCache:      make(map[string]User),
	}
	c.BasePlatformClient.AllowedUsers = cfg.AllowedUsers
	c.BasePlatformClient.BotNameVal = cfg.BotName
	c.BasePlatformClient.InitBase(c.Connect, c.forceClose, c.recoverMissed)
	return c
}

// Compile-time interface check.
var _ platform.PlatformClient = (*Client)(nil)

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func (c *Client) PlatformID() string   { return c.platformIDVal }
func (c *Client) PlatformType() string { return "mattermost" }
func (c *Client) DisplayName() string  { return c.displayNameVal }

// ---------------------------------------------------------------------------
// REST API helper
// ---------------------------------------------------------------------------

func (c *Client) api(ctx context.Context, method, path string, body interface{}, retries int) ([]byte, int, error) {
	url := c.url + "/api/v4" + path
	log.Debug(method + " " + path)

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		// Retry on 500 with exponential backoff
		if resp.StatusCode == 500 && retries < maxRetries {
			delay := retryDelayBase * time.Duration(1<<uint(retries))
			log.Warn(fmt.Sprintf("%s %s failed 500, retrying in %s (attempt %d/%d)", method, path, delay, retries+1, maxRetries))
			time.Sleep(delay)
			return c.api(ctx, method, path, body, retries+1)
		}
		return nil, resp.StatusCode, fmt.Errorf("Mattermost API error %d: %s", resp.StatusCode, string(data))
	}

	log.Debug(fmt.Sprintf("%s %s → %d", method, path, resp.StatusCode))
	return data, resp.StatusCode, nil
}

func (c *Client) apiJSON(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	data, _, err := c.api(ctx, method, path, body, 0)
	if err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

// ---------------------------------------------------------------------------
// User management
// ---------------------------------------------------------------------------

func (c *Client) normUser(u User) *platform.PlatformUser {
	displayName := u.FirstName
	if displayName == "" {
		displayName = u.Nickname
	}
	if displayName == "" {
		displayName = u.Username
	}
	return &platform.PlatformUser{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: displayName,
		Email:       u.Email,
	}
}

func (c *Client) GetBotUser(ctx context.Context) (*platform.PlatformUser, error) {
	var u User
	if err := c.apiJSON(ctx, "GET", "/users/me", nil, &u); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.botUserID = u.ID
	c.mu.Unlock()
	return c.normUser(u), nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*platform.PlatformUser, error) {
	c.mu.RLock()
	cached, ok := c.userCache[userID]
	c.mu.RUnlock()
	if ok {
		return c.normUser(cached), nil
	}

	data, status, err := c.api(ctx, "GET", "/users/"+userID, nil, 0)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.userCache[userID] = u
	c.mu.Unlock()
	return c.normUser(u), nil
}

func (c *Client) GetUserByUsername(ctx context.Context, username string) (*platform.PlatformUser, error) {
	var u User
	if err := c.apiJSON(ctx, "GET", "/users/username/"+username, nil, &u); err != nil {
		return nil, nil // not found
	}
	c.mu.Lock()
	c.userCache[u.ID] = u
	c.mu.Unlock()
	return c.normUser(u), nil
}

// ---------------------------------------------------------------------------
// Normalization helpers
// ---------------------------------------------------------------------------

func (c *Client) normPost(p Post) *platform.PlatformPost {
	np := &platform.PlatformPost{
		ID:         p.ID,
		PlatformID: c.platformIDVal,
		ChannelID:  p.ChannelID,
		UserID:     p.UserID,
		Message:    p.Message,
		RootID:     p.RootID,
		CreateAt:   p.CreateAt,
	}
	if p.Metadata != nil {
		for _, f := range p.Metadata.Files {
			np.Files = append(np.Files, platform.PlatformFile{
				ID:        f.ID,
				Name:      f.Name,
				Size:      f.Size,
				MimeType:  f.MimeType,
				Extension: f.Extension,
			})
		}
	}
	return np
}

// ---------------------------------------------------------------------------
// Messaging
// ---------------------------------------------------------------------------

func (c *Client) CreatePost(ctx context.Context, message, threadID string) (*platform.PlatformPost, error) {
	req := CreatePostRequest{
		ChannelID: c.channelID,
		Message:   message,
		RootID:    threadID,
	}
	var p Post
	if err := c.apiJSON(ctx, "POST", "/posts", req, &p); err != nil {
		return nil, err
	}
	return c.normPost(p), nil
}

func (c *Client) UpdatePost(ctx context.Context, postID, message string) (*platform.PlatformPost, error) {
	req := UpdatePostRequest{ID: postID, Message: message}
	var p Post
	if err := c.apiJSON(ctx, "PUT", "/posts/"+postID, req, &p); err != nil {
		return nil, err
	}
	return c.normPost(p), nil
}

func (c *Client) CreateInteractivePost(ctx context.Context, message string, reactions []string, threadID string) (*platform.PlatformPost, error) {
	post, err := c.CreatePost(ctx, message, threadID)
	if err != nil {
		return nil, err
	}
	for _, emoji := range reactions {
		if err := c.AddReaction(ctx, post.ID, emoji); err != nil {
			log.Warn("Failed to add reaction " + emoji + ": " + err.Error())
		}
	}
	return post, nil
}

func (c *Client) GetPost(ctx context.Context, postID string) (*platform.PlatformPost, error) {
	data, status, err := c.api(ctx, "GET", "/posts/"+postID, nil, 0)
	if err != nil {
		if status == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	var p Post
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return c.normPost(p), nil
}

func (c *Client) DeletePost(ctx context.Context, postID string) error {
	_, _, err := c.api(ctx, "DELETE", "/posts/"+postID, nil, 0)
	return err
}

func (c *Client) PinPost(ctx context.Context, postID string) error {
	_, _, err := c.api(ctx, "POST", "/posts/"+postID+"/pin", nil, 0)
	return err
}

func (c *Client) UnpinPost(ctx context.Context, postID string) error {
	_, status, err := c.api(ctx, "POST", "/posts/"+postID+"/unpin", nil, 0)
	if err != nil && (status == 403 || status == 404) {
		return nil // expected failures
	}
	return err
}

func (c *Client) GetPinnedPosts(ctx context.Context) ([]string, error) {
	var resp PinnedPostsResponse
	if err := c.apiJSON(ctx, "GET", "/channels/"+c.channelID+"/pinned", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Order, nil
}

func (c *Client) GetMessageLimits() platform.MessageLimits {
	return platform.MessageLimits{MaxLength: 16000, HardThreshold: 14000}
}

func (c *Client) GetThreadHistory(ctx context.Context, threadID string, opts *platform.ThreadHistoryOptions) ([]platform.ThreadMessage, error) {
	var resp ThreadResponse
	if err := c.apiJSON(ctx, "GET", "/posts/"+threadID+"/thread", nil, &resp); err != nil {
		log.Warn("Failed to get thread history: " + err.Error())
		return nil, nil
	}

	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()

	var msgs []platform.ThreadMessage
	for _, postID := range resp.Order {
		p, ok := resp.Posts[postID]
		if !ok {
			continue
		}
		if opts != nil && opts.ExcludeBotMessages && p.UserID == botID {
			continue
		}
		user, _ := c.GetUser(ctx, p.UserID)
		username := "unknown"
		if user != nil {
			username = user.Username
		}
		msgs = append(msgs, platform.ThreadMessage{
			ID:       p.ID,
			UserID:   p.UserID,
			Username: username,
			Message:  p.Message,
			CreateAt: p.CreateAt,
		})
	}

	sort.Slice(msgs, func(i, j int) bool { return msgs[i].CreateAt < msgs[j].CreateAt })

	if opts != nil && opts.Limit > 0 && len(msgs) > opts.Limit {
		msgs = msgs[len(msgs)-opts.Limit:]
	}
	return msgs, nil
}

// ---------------------------------------------------------------------------
// Reactions
// ---------------------------------------------------------------------------

func (c *Client) AddReaction(ctx context.Context, postID, emojiName string) error {
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	_, _, err := c.api(ctx, "POST", "/reactions", map[string]string{
		"user_id":    botID,
		"post_id":    postID,
		"emoji_name": emojiName,
	}, 0)
	return err
}

func (c *Client) RemoveReaction(ctx context.Context, postID, emojiName string) error {
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	_, _, err := c.api(ctx, "DELETE", fmt.Sprintf("/users/%s/posts/%s/reactions/%s", botID, postID, emojiName), nil, 0)
	return err
}

// ---------------------------------------------------------------------------
// Bot mentions
// ---------------------------------------------------------------------------

func (c *Client) IsBotMentioned(message string) bool {
	botName := regexp.QuoteMeta(c.BotNameVal)
	re := regexp.MustCompile(`(?i)(^|\s)@` + botName + `\b`)
	return re.MatchString(message)
}

func (c *Client) ExtractPrompt(message string) string {
	botName := regexp.QuoteMeta(c.BotNameVal)
	re := regexp.MustCompile(`(?i)(^|\s)@` + botName + `\b`)
	return strings.TrimSpace(re.ReplaceAllString(message, " "))
}

// ---------------------------------------------------------------------------
// Files
// ---------------------------------------------------------------------------

func (c *Client) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	data, _, err := c.api(ctx, "GET", "/files/"+fileID, nil, 0)
	return data, err
}

func (c *Client) GetFileInfo(ctx context.Context, fileID string) (*platform.PlatformFile, error) {
	var f File
	if err := c.apiJSON(ctx, "GET", "/files/"+fileID+"/info", nil, &f); err != nil {
		return nil, err
	}
	return &platform.PlatformFile{
		ID:        f.ID,
		Name:      f.Name,
		Size:      f.Size,
		MimeType:  f.MimeType,
		Extension: f.Extension,
	}, nil
}

// ---------------------------------------------------------------------------
// Platform helpers
// ---------------------------------------------------------------------------

func (c *Client) GetMcpConfig() platform.McpConfig {
	return platform.McpConfig{
		Type:         "mattermost",
		URL:          c.url,
		Token:        c.token,
		ChannelID:    c.channelID,
		AllowedUsers: c.AllowedUsers,
	}
}

func (c *Client) GetFormatter() platform.PlatformFormatter { return c.formatter }

func (c *Client) GetThreadLink(threadID, lastMessageID, _ string) string {
	target := threadID
	if lastMessageID != "" {
		target = lastMessageID
	}
	return c.url + "/_redirect/pl/" + target
}

func (c *Client) SendTyping(threadID string) {
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()
	if conn == nil {
		return
	}
	conn.WriteJSON(map[string]interface{}{
		"action": "user_typing",
		"seq":    time.Now().UnixMilli(),
		"data": map[string]string{
			"channel_id": c.channelID,
			"parent_id":  threadID,
		},
	})
}

// ---------------------------------------------------------------------------
// WebSocket connection
// ---------------------------------------------------------------------------

func (c *Client) Connect(ctx context.Context) error {
	// Fetch bot user first
	if _, err := c.GetBotUser(ctx); err != nil {
		return fmt.Errorf("failed to get bot user: %w", err)
	}

	wsURL := strings.Replace(c.url, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL += "/api/v4/websocket"

	wsLog2.Info("Connecting to " + wsURL)

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, wsURL, http.Header{})
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// Authenticate
	if err := conn.WriteJSON(map[string]interface{}{
		"seq":    1,
		"action": "authentication_challenge",
		"data":   map[string]string{"token": c.token},
	}); err != nil {
		conn.Close()
		return fmt.Errorf("auth challenge: %w", err)
	}

	connected := make(chan error, 1)

	go func() {
		for {
			var event WebSocketEvent
			if err := conn.ReadJSON(&event); err != nil {
				wsLog2.Info("WebSocket closed: " + err.Error())
				c.mu.Lock()
				c.conn = nil
				c.mu.Unlock()
				c.BasePlatformClient.OnConnectionClosed()
				return
			}
			c.BasePlatformClient.UpdateLastMessageTime()
			c.handleEvent(ctx, event, connected)
		}
	}()

	select {
	case err := <-connected:
		return err
	case <-ctx.Done():
		conn.Close()
		return ctx.Err()
	case <-time.After(10 * time.Second):
		conn.Close()
		return fmt.Errorf("connection timeout")
	}
}

func (c *Client) handleEvent(ctx context.Context, event WebSocketEvent, connected chan<- error) {
	switch event.Event {
	case "hello":
		c.BasePlatformClient.OnConnectionEstablished()
		select {
		case connected <- nil:
		default:
		}

	case "posted":
		postJSON, _ := event.Data["post"].(string)
		if postJSON == "" {
			return
		}
		var p Post
		if err := json.Unmarshal([]byte(postJSON), &p); err != nil {
			wsLog2.Warn("Failed to parse post: " + err.Error())
			return
		}

		c.mu.RLock()
		botID := c.botUserID
		c.mu.RUnlock()

		if p.UserID == botID || p.ChannelID != c.channelID {
			return
		}

		c.mu.Lock()
		c.lastPostID = p.ID
		c.mu.Unlock()

		go func() {
			// Enrich with file metadata if needed
			if len(p.FileIDs) > 0 && (p.Metadata == nil || len(p.Metadata.Files) == 0) {
				var files []File
				for _, fid := range p.FileIDs {
					var f File
					if err := c.apiJSON(ctx, "GET", "/files/"+fid+"/info", nil, &f); err == nil {
						files = append(files, f)
					}
				}
				if p.Metadata == nil {
					p.Metadata = &PostMetadata{}
				}
				p.Metadata.Files = files
			}

			np := c.normPost(p)
			user, _ := c.GetUser(ctx, p.UserID)
			c.BasePlatformClient.EmitMessage(*np, user)
			if p.RootID == "" {
				c.BasePlatformClient.EmitChannelPost(*np, user)
			}
		}()

	case "reaction_added", "reaction_removed":
		reactionJSON, _ := event.Data["reaction"].(string)
		if reactionJSON == "" {
			return
		}
		var r Reaction
		if err := json.Unmarshal([]byte(reactionJSON), &r); err != nil {
			wsLog2.Warn("Failed to parse reaction: " + err.Error())
			return
		}

		c.mu.RLock()
		botID := c.botUserID
		c.mu.RUnlock()
		if r.UserID == botID {
			return
		}

		go func() {
			user, _ := c.GetUser(ctx, r.UserID)
			nr := platform.PlatformReaction{
				UserID:    r.UserID,
				PostID:    r.PostID,
				EmojiName: r.EmojiName,
				CreateAt:  r.CreateAt,
			}
			if event.Event == "reaction_added" {
				c.BasePlatformClient.EmitReaction(nr, user)
			} else {
				c.BasePlatformClient.EmitReactionRemoved(nr, user)
			}
		}()
	}
}

func (c *Client) forceClose() {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()
	if conn != nil {
		conn.Close()
	}
}

func (c *Client) recoverMissed() error {
	c.mu.RLock()
	lastID := c.lastPostID
	c.mu.RUnlock()
	if lastID == "" {
		return nil
	}
	log.Info("Recovering missed messages after " + lastID)

	var resp ChannelPostsResponse
	if err := c.apiJSON(context.Background(), "GET",
		"/channels/"+c.channelID+"/posts?after="+lastID+"&per_page=100",
		nil, &resp); err != nil {
		return err
	}

	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()

	var posts []platform.PlatformPost
	for _, postID := range resp.Order {
		p, ok := resp.Posts[postID]
		if !ok || p.UserID == botID {
			continue
		}
		posts = append(posts, *c.normPost(p))
	}
	sort.Slice(posts, func(i, j int) bool { return posts[i].CreateAt < posts[j].CreateAt })

	for _, np := range posts {
		c.mu.Lock()
		c.lastPostID = np.ID
		c.mu.Unlock()
		user, _ := c.GetUser(context.Background(), np.UserID)
		c.BasePlatformClient.EmitMessage(np, user)
		if np.RootID == "" {
			c.BasePlatformClient.EmitChannelPost(np, user)
		}
	}
	if len(posts) > 0 {
		log.Info(fmt.Sprintf("Recovered %d missed message(s)", len(posts)))
	}
	return nil
}
