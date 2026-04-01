package slack

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

var log = utils.CreateLogger("slack")
var wsLog = utils.WsLogger

const (
	maxRateLimitRetries = 5
)

// Client is the Slack Socket Mode implementation of platform.PlatformClient.
type Client struct {
	platform.BasePlatformClient

	platformIDVal  string
	displayNameVal string
	botToken       string
	appToken       string
	channelID      string
	apiURL         string
	formatter      *Formatter

	mu                sync.RWMutex
	userCache         map[string]SlackUser
	usernameToIDCache map[string]string
	botUserID         string
	teamURL           string
	lastTs            string

	processedMessages map[string]struct{}
	rateLimitUntil    time.Time

	conn *websocket.Conn
}

// NewClient creates a new SlackClient from config.
func NewClient(cfg config.SlackPlatformConfig) *Client {
	apiURL := cfg.APIURL
	if apiURL == "" {
		apiURL = "https://slack.com/api"
	}
	c := &Client{
		platformIDVal:     cfg.ID,
		displayNameVal:    cfg.DisplayName,
		botToken:          cfg.BotToken,
		appToken:          cfg.AppToken,
		channelID:         cfg.ChannelID,
		apiURL:            apiURL,
		formatter:         NewFormatter(),
		userCache:         make(map[string]SlackUser),
		usernameToIDCache: make(map[string]string),
		processedMessages: make(map[string]struct{}),
	}
	c.BasePlatformClient.AllowedUsers = cfg.AllowedUsers
	c.BasePlatformClient.BotNameVal = cfg.BotName
	c.BasePlatformClient.InitBase(c.Connect, c.forceClose, c.recoverMissed)
	return c
}

// SetBotUserIDForTest sets the bot user ID (test helper only).
func (c *Client) SetBotUserIDForTest(id string) {
	c.mu.Lock()
	c.botUserID = id
	c.mu.Unlock()
}

// Compile-time interface check.
var _ platform.PlatformClient = (*Client)(nil)

func (c *Client) PlatformID() string   { return c.platformIDVal }
func (c *Client) PlatformType() string { return "slack" }
func (c *Client) DisplayName() string  { return c.displayNameVal }

func (c *Client) botAPI(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	return c.apiWith(ctx, c.botToken, method, endpoint, body, 0, nil)
}

func (c *Client) appAPI(ctx context.Context, endpoint string) ([]byte, error) {
	return c.apiWith(ctx, c.appToken, "POST", endpoint, nil, 0, nil)
}

func (c *Client) apiWith(ctx context.Context, token, method, endpoint string, body interface{}, retries int, expectedErrors []string) ([]byte, error) {
	c.mu.RLock()
	until := c.rateLimitUntil
	c.mu.RUnlock()
	if wait := time.Until(until); wait > 0 {
		log.Debug(fmt.Sprintf("Rate limited, waiting %s", wait))
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	url := c.apiURL + "/" + endpoint
	log.Debug(method + " " + endpoint)

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 429 {
		if retries >= maxRateLimitRetries {
			return nil, fmt.Errorf("Slack rate limit exceeded after %d retries", maxRateLimitRetries)
		}
		retryAfter := 5
		if v := resp.Header.Get("Retry-After"); v != "" {
			fmt.Sscanf(v, "%d", &retryAfter)
		}
		wait := time.Duration(retryAfter) * time.Second
		log.Warn(fmt.Sprintf("Rate limited, retrying after %s (attempt %d/%d)", wait, retries+1, maxRateLimitRetries))
		c.mu.Lock()
		c.rateLimitUntil = time.Now().Add(wait)
		c.mu.Unlock()
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return c.apiWith(ctx, token, method, endpoint, body, retries+1, expectedErrors)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Slack HTTP error %d: %s", resp.StatusCode, string(data))
	}

	var base APIResponse
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}
	if !base.OK {
		return nil, fmt.Errorf("Slack API error: %s", base.Error)
	}

	return data, nil
}

func (c *Client) apiJSON(ctx context.Context, method, endpoint string, body interface{}, out interface{}, expectedErrors ...string) error {
	data, err := c.apiWith(ctx, c.botToken, method, endpoint, body, 0, expectedErrors)
	if err != nil {
		// Check if error matches an expected error - return it as-is
		for _, expected := range expectedErrors {
			if strings.Contains(err.Error(), expected) {
				return err
			}
		}
		return err
	}
	if out != nil {
		return json.Unmarshal(data, out)
	}
	return nil
}

func (c *Client) normUser(u SlackUser) *platform.PlatformUser {
	displayName := u.Profile.DisplayName
	if displayName == "" {
		displayName = u.Profile.RealName
	}
	if displayName == "" {
		displayName = u.RealName
	}
	if displayName == "" {
		displayName = u.Name
	}
	return &platform.PlatformUser{
		ID:          u.ID,
		Username:    u.Name,
		DisplayName: displayName,
		Email:       u.Profile.Email,
	}
}

func (c *Client) fetchBotUser(ctx context.Context) error {
	var auth AuthTestResponse
	if err := c.apiJSON(ctx, "POST", "auth.test", nil, &auth); err != nil {
		return err
	}
	c.mu.Lock()
	c.botUserID = auth.UserID
	c.teamURL = strings.TrimRight(auth.URL, "/")
	c.mu.Unlock()

	var info UsersInfoResponse
	if err := c.apiJSON(ctx, "GET", "users.info?user="+auth.UserID, nil, &info); err != nil {
		return err
	}
	c.mu.Lock()
	c.userCache[auth.UserID] = info.User
	c.mu.Unlock()
	return nil
}

func (c *Client) GetBotUser(ctx context.Context) (*platform.PlatformUser, error) {
	c.mu.RLock()
	botID := c.botUserID
	cached, ok := c.userCache[botID]
	c.mu.RUnlock()
	if ok {
		return c.normUser(cached), nil
	}
	if err := c.fetchBotUser(ctx); err != nil {
		return nil, err
	}
	c.mu.RLock()
	user := c.userCache[c.botUserID]
	c.mu.RUnlock()
	return c.normUser(user), nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*platform.PlatformUser, error) {
	if userID == "" {
		return nil, nil
	}
	c.mu.RLock()
	cached, ok := c.userCache[userID]
	c.mu.RUnlock()
	if ok {
		return c.normUser(cached), nil
	}

	var resp UsersInfoResponse
	if err := c.apiJSON(ctx, "GET", "users.info?user="+userID, nil, &resp); err != nil {
		return nil, nil
	}
	c.mu.Lock()
	c.userCache[userID] = resp.User
	c.usernameToIDCache[resp.User.Name] = userID
	c.mu.Unlock()
	return c.normUser(resp.User), nil
}

func (c *Client) GetUserByUsername(ctx context.Context, username string) (*platform.PlatformUser, error) {
	c.mu.RLock()
	id, ok := c.usernameToIDCache[username]
	c.mu.RUnlock()
	if ok {
		return c.GetUser(ctx, id)
	}

	var cursor string
	for {
		endpoint := "users.list?limit=200"
		if cursor != "" {
			endpoint += "&cursor=" + cursor
		}
		var resp UsersListResponse
		if err := c.apiJSON(ctx, "GET", endpoint, nil, &resp); err != nil {
			return nil, nil
		}
		for _, u := range resp.Members {
			c.mu.Lock()
			c.userCache[u.ID] = u
			c.usernameToIDCache[u.Name] = u.ID
			c.mu.Unlock()
			if u.Name == username {
				return c.normUser(u), nil
			}
		}
		if resp.ResponseMetadata == nil || resp.ResponseMetadata.NextCursor == "" {
			break
		}
		cursor = resp.ResponseMetadata.NextCursor
	}
	return nil, nil
}

func (c *Client) normPost(msg SlackMessage, channelID, rootID string) *platform.PlatformPost {
	np := &platform.PlatformPost{
		ID:         msg.Ts,
		PlatformID: c.platformIDVal,
		ChannelID:  channelID,
		UserID:     msg.User,
		Message:    msg.Text,
		RootID:     rootID,
		CreateAt:   tsToMs(msg.Ts),
	}
	for _, f := range msg.Files {
		ext := f.Filetype
		if idx := strings.LastIndex(f.Name, "."); idx >= 0 {
			ext = f.Name[idx+1:]
		}
		np.Files = append(np.Files, platform.PlatformFile{
			ID:        f.ID,
			Name:      f.Name,
			Size:      f.Size,
			MimeType:  f.Mimetype,
			Extension: ext,
		})
	}
	return np
}

func tsToMs(ts string) int64 {
	var f float64
	fmt.Sscanf(ts, "%f", &f)
	return int64(f * 1000)
}

func (c *Client) CreatePost(ctx context.Context, message, threadID string) (*platform.PlatformPost, error) {
	limits := c.GetMessageLimits()
	if len(message) > limits.MaxLength {
		message = platform.TruncateMessageSafely(message, limits.MaxLength, "_... (truncated)_")
	}

	body := map[string]interface{}{
		"channel":      c.channelID,
		"text":         message,
		"unfurl_links": threadID != "",
		"unfurl_media": threadID != "",
	}
	if threadID != "" {
		body["thread_ts"] = threadID
	}

	var resp PostMessageResponse
	if err := c.apiJSON(ctx, "POST", "chat.postMessage", body, &resp); err != nil {
		return nil, err
	}
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	return &platform.PlatformPost{
		ID:         resp.Ts,
		PlatformID: c.platformIDVal,
		ChannelID:  resp.Channel,
		UserID:     botID,
		Message:    resp.Message.Text,
		RootID:     threadID,
		CreateAt:   tsToMs(resp.Ts),
	}, nil
}

func (c *Client) UpdatePost(ctx context.Context, postID, message string) (*platform.PlatformPost, error) {
	limits := c.GetMessageLimits()
	if len(message) > limits.MaxLength {
		message = platform.TruncateMessageSafely(message, limits.MaxLength, "_... (truncated)_")
	}

	var resp UpdateMessageResponse
	if err := c.apiJSON(ctx, "POST", "chat.update", map[string]interface{}{
		"channel": c.channelID,
		"ts":      postID,
		"text":    message,
	}, &resp); err != nil {
		return nil, err
	}
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	return &platform.PlatformPost{
		ID:         resp.Ts,
		PlatformID: c.platformIDVal,
		ChannelID:  resp.Channel,
		UserID:     botID,
		Message:    resp.Text,
		CreateAt:   tsToMs(resp.Ts),
	}, nil
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
	var resp ConversationsHistoryResponse
	endpoint := fmt.Sprintf("conversations.history?channel=%s&latest=%s&oldest=%s&inclusive=true&limit=1", c.channelID, postID, postID)
	if err := c.apiJSON(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, nil
	}
	if len(resp.Messages) == 0 {
		return nil, nil
	}
	return c.normPost(resp.Messages[0], c.channelID, ""), nil
}

func (c *Client) DeletePost(ctx context.Context, postID string) error {
	return c.apiJSON(ctx, "POST", "chat.delete", map[string]interface{}{
		"channel": c.channelID,
		"ts":      postID,
	}, nil)
}

func (c *Client) PinPost(ctx context.Context, postID string) error {
	err := c.apiJSON(ctx, "POST", "pins.add", map[string]interface{}{
		"channel":   c.channelID,
		"timestamp": postID,
	}, nil, "already_pinned")
	if err != nil && strings.Contains(err.Error(), "already_pinned") {
		return nil
	}
	return err
}

func (c *Client) UnpinPost(ctx context.Context, postID string) error {
	err := c.apiJSON(ctx, "POST", "pins.remove", map[string]interface{}{
		"channel":   c.channelID,
		"timestamp": postID,
	}, nil, "no_pin")
	if err != nil && strings.Contains(err.Error(), "no_pin") {
		return nil
	}
	return err
}

func (c *Client) GetPinnedPosts(ctx context.Context) ([]string, error) {
	var resp PinsListResponse
	if err := c.apiJSON(ctx, "GET", "pins.list?channel="+c.channelID, nil, &resp); err != nil {
		return nil, err
	}
	var ids []string
	for _, item := range resp.Items {
		if item.Message != nil {
			ids = append(ids, item.Message.Ts)
		}
	}
	return ids, nil
}

func (c *Client) GetMessageLimits() platform.MessageLimits {
	return platform.MessageLimits{MaxLength: 12000, HardThreshold: 10000}
}

func (c *Client) GetThreadHistory(ctx context.Context, threadID string, opts *platform.ThreadHistoryOptions) ([]platform.ThreadMessage, error) {
	limit := 100
	if opts != nil && opts.Limit > 0 {
		limit = opts.Limit
	}
	var resp ConversationsRepliesResponse
	endpoint := fmt.Sprintf("conversations.replies?channel=%s&ts=%s&limit=%d", c.channelID, threadID, limit)
	if err := c.apiJSON(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, nil
	}

	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()

	var msgs []platform.ThreadMessage
	for _, msg := range resp.Messages {
		if opts != nil && opts.ExcludeBotMessages && (msg.User == botID || msg.BotID != "") {
			continue
		}
		user, _ := c.GetUser(ctx, msg.User)
		username := "unknown"
		if user != nil {
			username = user.Username
		}
		msgs = append(msgs, platform.ThreadMessage{
			ID:       msg.Ts,
			UserID:   msg.User,
			Username: username,
			Message:  msg.Text,
			CreateAt: tsToMs(msg.Ts),
		})
	}
	sort.Slice(msgs, func(i, j int) bool { return msgs[i].CreateAt < msgs[j].CreateAt })
	return msgs, nil
}

func (c *Client) AddReaction(ctx context.Context, postID, emojiName string) error {
	name := platform.GetEmojiName(emojiName)
	return c.apiJSON(ctx, "POST", "reactions.add", map[string]interface{}{
		"channel":   c.channelID,
		"timestamp": postID,
		"name":      name,
	}, nil)
}

func (c *Client) RemoveReaction(ctx context.Context, postID, emojiName string) error {
	name := platform.GetEmojiName(emojiName)
	return c.apiJSON(ctx, "POST", "reactions.remove", map[string]interface{}{
		"channel":   c.channelID,
		"timestamp": postID,
		"name":      name,
	}, nil)
}

func (c *Client) IsBotMentioned(message string) bool {
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	if botID != "" && strings.Contains(message, "<@"+botID+">") {
		return true
	}
	botName := regexp.QuoteMeta(c.BotNameVal)
	re := regexp.MustCompile(`(?i)(^|\s)@` + botName + `\b`)
	return re.MatchString(message)
}

func (c *Client) ExtractPrompt(message string) string {
	c.mu.RLock()
	botID := c.botUserID
	c.mu.RUnlock()
	prompt := message
	if botID != "" {
		prompt = strings.ReplaceAll(prompt, "<@"+botID+">", "")
	}
	botName := regexp.QuoteMeta(c.BotNameVal)
	re := regexp.MustCompile(`(?i)(^|\s)@` + botName + `\b`)
	return strings.TrimSpace(re.ReplaceAllString(prompt, " "))
}

func (c *Client) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	var resp FilesInfoResponse
	if err := c.apiJSON(ctx, "GET", "files.info?file="+fileID, nil, &resp); err != nil {
		return nil, err
	}
	downloadURL := resp.File.URLPrivateDownload
	if downloadURL == "" {
		downloadURL = resp.File.URLPrivate
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("no download URL for file %s", fileID)
	}
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.botToken)
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	return io.ReadAll(httpResp.Body)
}

func (c *Client) GetFileInfo(ctx context.Context, fileID string) (*platform.PlatformFile, error) {
	var resp FilesInfoResponse
	if err := c.apiJSON(ctx, "GET", "files.info?file="+fileID, nil, &resp); err != nil {
		return nil, err
	}
	ext := resp.File.Filetype
	if idx := strings.LastIndex(resp.File.Name, "."); idx >= 0 {
		ext = resp.File.Name[idx+1:]
	}
	return &platform.PlatformFile{
		ID:        resp.File.ID,
		Name:      resp.File.Name,
		Size:      resp.File.Size,
		MimeType:  resp.File.Mimetype,
		Extension: ext,
	}, nil
}

func (c *Client) GetMcpConfig() platform.McpConfig {
	return platform.McpConfig{
		Type:         "slack",
		URL:          "https://slack.com",
		Token:        c.botToken,
		ChannelID:    c.channelID,
		AllowedUsers: c.AllowedUsers,
	}
}

func (c *Client) GetFormatter() platform.PlatformFormatter { return c.formatter }

func (c *Client) GetThreadLink(threadID, _, lastMessageTs string) string {
	c.mu.RLock()
	teamURL := c.teamURL
	c.mu.RUnlock()
	if teamURL == "" {
		return "#" + threadID
	}
	targetTs := threadID
	if lastMessageTs != "" {
		targetTs = lastMessageTs
	}
	permalinkTs := strings.ReplaceAll(targetTs, ".", "")
	if lastMessageTs != "" && lastMessageTs != threadID {
		return fmt.Sprintf("%s/archives/%s/p%s?thread_ts=%s&cid=%s", teamURL, c.channelID, permalinkTs, threadID, c.channelID)
	}
	return fmt.Sprintf("%s/archives/%s/p%s", teamURL, c.channelID, permalinkTs)
}

func (c *Client) SendTyping(_ string) {}

func (c *Client) Connect(ctx context.Context) error {
	if err := c.fetchBotUser(ctx); err != nil {
		return fmt.Errorf("failed to fetch bot user: %w", err)
	}

	data, err := c.appAPI(ctx, "apps.connections.open")
	if err != nil {
		return fmt.Errorf("apps.connections.open: %w", err)
	}
	var connResp AppsConnectionsOpenResponse
	if err := json.Unmarshal(data, &connResp); err != nil {
		return err
	}

	wsLog.Info("Socket Mode: connecting to " + connResp.URL)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, connResp.URL, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	connected := make(chan error, 1)

	go func() {
		for {
			var envelope SocketModeEvent
			if err := conn.ReadJSON(&envelope); err != nil {
				wsLog.Info("Socket Mode disconnected: " + err.Error())
				c.mu.Lock()
				c.conn = nil
				c.mu.Unlock()
				c.BasePlatformClient.OnConnectionClosed()
				return
			}
			c.BasePlatformClient.UpdateLastMessageTime()
			c.handleEnvelope(ctx, conn, envelope, connected)
		}
	}()

	select {
	case err := <-connected:
		return err
	case <-ctx.Done():
		conn.Close()
		return ctx.Err()
	case <-time.After(30 * time.Second):
		conn.Close()
		return fmt.Errorf("Socket Mode connection timeout")
	}
}

func (c *Client) handleEnvelope(ctx context.Context, conn *websocket.Conn, env SocketModeEvent, connected chan<- error) {
	if env.EnvelopeID != "" {
		conn.WriteJSON(map[string]string{"envelope_id": env.EnvelopeID})
		wsLog.Debug("ACKed " + env.EnvelopeID)
	}

	switch env.Type {
	case "hello":
		c.BasePlatformClient.OnConnectionEstablished()
		select {
		case connected <- nil:
		default:
		}

	case "disconnect":
		wsLog.Info("Socket Mode: received disconnect, reconnecting...")
		conn.Close()

	case "events_api":
		if env.Payload != nil && env.Payload.Event != nil {
			c.handleEvent(ctx, env.Payload.Event)
		}
	}
}

func (c *Client) handleEvent(ctx context.Context, event *SlackEvent) {
	switch event.Type {
	case "message":
		if event.Subtype != "" && event.Subtype != "file_share" {
			return
		}
		c.mu.RLock()
		botID := c.botUserID
		c.mu.RUnlock()
		if event.User == botID || event.BotID != "" {
			return
		}
		if event.Channel != c.channelID {
			return
		}

		c.mu.Lock()
		if _, seen := c.processedMessages[event.Ts]; seen {
			c.mu.Unlock()
			return
		}
		c.processedMessages[event.Ts] = struct{}{}
		if len(c.processedMessages) > 1000 {
			for k := range c.processedMessages {
				delete(c.processedMessages, k)
				break
			}
		}
		c.lastTs = event.Ts
		c.mu.Unlock()

		rootID := ""
		if event.ThreadTs != "" && event.ThreadTs != event.Ts {
			rootID = event.ThreadTs
		}
		msg := SlackMessage{
			Ts:    event.Ts,
			User:  event.User,
			Text:  event.Text,
			Files: event.Files,
		}
		np := c.normPost(msg, event.Channel, rootID)
		go func() {
			user, _ := c.GetUser(ctx, event.User)
			c.BasePlatformClient.EmitMessage(*np, user)
			if rootID == "" {
				c.BasePlatformClient.EmitChannelPost(*np, user)
			}
		}()

	case "reaction_added", "reaction_removed":
		if event.Item == nil || event.Item.Type != "message" {
			return
		}
		c.mu.RLock()
		botID := c.botUserID
		c.mu.RUnlock()
		if event.User == botID {
			return
		}
		if event.Item.Channel != c.channelID {
			return
		}
		r := platform.PlatformReaction{
			UserID:    event.User,
			PostID:    event.Item.Ts,
			EmojiName: event.Reaction,
			CreateAt:  time.Now().UnixMilli(),
		}
		go func() {
			user, _ := c.GetUser(ctx, event.User)
			if event.Type == "reaction_added" {
				c.BasePlatformClient.EmitReaction(r, user)
			} else {
				c.BasePlatformClient.EmitReactionRemoved(r, user)
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
	lastTs := c.lastTs
	botID := c.botUserID
	c.mu.RUnlock()
	if lastTs == "" {
		return nil
	}

	log.Info("Recovering missed Slack messages after " + lastTs)
	endpoint := fmt.Sprintf("conversations.history?channel=%s&oldest=%s&inclusive=false&limit=100", c.channelID, lastTs)
	var resp ConversationsHistoryResponse
	if err := c.apiJSON(context.Background(), "GET", endpoint, nil, &resp); err != nil {
		return err
	}

	msgs := resp.Messages
	sort.Slice(msgs, func(i, j int) bool {
		return parseFloat(msgs[i].Ts) < parseFloat(msgs[j].Ts)
	})

	for _, msg := range msgs {
		if msg.User == botID || msg.BotID != "" {
			continue
		}
		c.mu.Lock()
		c.lastTs = msg.Ts
		c.mu.Unlock()

		rootID := ""
		if msg.ThreadTs != "" && msg.ThreadTs != msg.Ts {
			rootID = msg.ThreadTs
		}
		np := c.normPost(msg, c.channelID, rootID)
		user, _ := c.GetUser(context.Background(), msg.User)
		c.BasePlatformClient.EmitMessage(*np, user)
		if rootID == "" {
			c.BasePlatformClient.EmitChannelPost(*np, user)
		}
	}
	return nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
