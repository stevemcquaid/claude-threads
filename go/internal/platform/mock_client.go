package platform

import (
	"context"
	"sync"
)

// MockPlatformClient is a test double for PlatformClient.
// All methods are no-ops by default. Override fields to inject test behavior.
type MockPlatformClient struct {
	mu sync.RWMutex

	// Override these to control behavior
	PlatformIDVal    string
	PlatformTypeVal  string
	DisplayNameVal   string
	BotNameVal       string
	FormatterVal     PlatformFormatter
	MessageLimitsVal MessageLimits

	// Recorded calls (for assertions)
	CreatedPosts   []string
	UpdatedPosts   []struct{ ID, Message string }
	AddedReactions []struct{ PostID, Emoji string }
	DeletedPosts   []string

	// Errors to return
	CreatePostErr error
	UpdatePostErr error

	// Callbacks registered by the system under test
	onConnected       func()
	onDisconnected    func()
	onReconnecting    func(int)
	onError           func(error)
	onMessage         func(PlatformPost, *PlatformUser)
	onReaction        func(PlatformReaction, *PlatformUser)
	onReactionRemoved func(PlatformReaction, *PlatformUser)
	onChannelPost     func(PlatformPost, *PlatformUser)
}

func (m *MockPlatformClient) PlatformID() string   { return m.PlatformIDVal }
func (m *MockPlatformClient) PlatformType() string { return m.PlatformTypeVal }
func (m *MockPlatformClient) DisplayName() string  { return m.DisplayNameVal }

func (m *MockPlatformClient) Connect(_ context.Context) error { return nil }
func (m *MockPlatformClient) Disconnect()                     {}
func (m *MockPlatformClient) PrepareForReconnect()            {}

func (m *MockPlatformClient) GetBotUser(_ context.Context) (*PlatformUser, error) {
	return &PlatformUser{ID: "bot-id", Username: m.BotNameVal}, nil
}
func (m *MockPlatformClient) GetUser(_ context.Context, _ string) (*PlatformUser, error) {
	return nil, nil
}
func (m *MockPlatformClient) GetUserByUsername(_ context.Context, _ string) (*PlatformUser, error) {
	return nil, nil
}
func (m *MockPlatformClient) IsUserAllowed(_ string) bool { return true }
func (m *MockPlatformClient) GetBotName() string          { return m.BotNameVal }
func (m *MockPlatformClient) GetMcpConfig() McpConfig     { return McpConfig{} }
func (m *MockPlatformClient) GetFormatter() PlatformFormatter {
	if m.FormatterVal != nil {
		return m.FormatterVal
	}
	return nil
}
func (m *MockPlatformClient) GetThreadLink(_, _, _ string) string { return "" }

func (m *MockPlatformClient) CreatePost(_ context.Context, message, _ string) (*PlatformPost, error) {
	if m.CreatePostErr != nil {
		return nil, m.CreatePostErr
	}
	m.mu.Lock()
	m.CreatedPosts = append(m.CreatedPosts, message)
	m.mu.Unlock()
	return &PlatformPost{ID: "post-id", Message: message}, nil
}
func (m *MockPlatformClient) UpdatePost(_ context.Context, postID, message string) (*PlatformPost, error) {
	if m.UpdatePostErr != nil {
		return nil, m.UpdatePostErr
	}
	m.mu.Lock()
	m.UpdatedPosts = append(m.UpdatedPosts, struct{ ID, Message string }{postID, message})
	m.mu.Unlock()
	return &PlatformPost{ID: postID, Message: message}, nil
}
func (m *MockPlatformClient) CreateInteractivePost(_ context.Context, message string, _ []string, _ string) (*PlatformPost, error) {
	return &PlatformPost{ID: "interactive-id", Message: message}, nil
}
func (m *MockPlatformClient) GetPost(_ context.Context, _ string) (*PlatformPost, error) {
	return nil, nil
}
func (m *MockPlatformClient) DeletePost(_ context.Context, postID string) error {
	m.mu.Lock()
	m.DeletedPosts = append(m.DeletedPosts, postID)
	m.mu.Unlock()
	return nil
}
func (m *MockPlatformClient) PinPost(_ context.Context, _ string) error   { return nil }
func (m *MockPlatformClient) UnpinPost(_ context.Context, _ string) error { return nil }
func (m *MockPlatformClient) GetPinnedPosts(_ context.Context) ([]string, error) {
	return nil, nil
}
func (m *MockPlatformClient) GetMessageLimits() MessageLimits {
	if m.MessageLimitsVal.MaxLength == 0 {
		return MessageLimits{MaxLength: 16000, HardThreshold: 15000}
	}
	return m.MessageLimitsVal
}
func (m *MockPlatformClient) GetThreadHistory(_ context.Context, _ string, _ *ThreadHistoryOptions) ([]ThreadMessage, error) {
	return nil, nil
}

func (m *MockPlatformClient) AddReaction(_ context.Context, postID, emoji string) error {
	m.mu.Lock()
	m.AddedReactions = append(m.AddedReactions, struct{ PostID, Emoji string }{postID, emoji})
	m.mu.Unlock()
	return nil
}
func (m *MockPlatformClient) RemoveReaction(_ context.Context, _, _ string) error { return nil }

func (m *MockPlatformClient) IsBotMentioned(_ string) bool    { return false }
func (m *MockPlatformClient) ExtractPrompt(msg string) string { return msg }
func (m *MockPlatformClient) SendTyping(_ string)              {}

func (m *MockPlatformClient) DownloadFile(_ context.Context, _ string) ([]byte, error) {
	return nil, nil
}
func (m *MockPlatformClient) GetFileInfo(_ context.Context, _ string) (*PlatformFile, error) {
	return nil, nil
}

func (m *MockPlatformClient) OnConnected(f func())       { m.onConnected = f }
func (m *MockPlatformClient) OnDisconnected(f func())    { m.onDisconnected = f }
func (m *MockPlatformClient) OnReconnecting(f func(int)) { m.onReconnecting = f }
func (m *MockPlatformClient) OnError(f func(error))      { m.onError = f }
func (m *MockPlatformClient) OnMessage(f func(PlatformPost, *PlatformUser)) {
	m.onMessage = f
}
func (m *MockPlatformClient) OnReaction(f func(PlatformReaction, *PlatformUser)) {
	m.onReaction = f
}
func (m *MockPlatformClient) OnReactionRemoved(f func(PlatformReaction, *PlatformUser)) {
	m.onReactionRemoved = f
}
func (m *MockPlatformClient) OnChannelPost(f func(PlatformPost, *PlatformUser)) {
	m.onChannelPost = f
}

// SimulateMessage triggers the OnMessage callback with the given post and user.
// Use this in tests to simulate incoming messages.
func (m *MockPlatformClient) SimulateMessage(post PlatformPost, user *PlatformUser) {
	if m.onMessage != nil {
		m.onMessage(post, user)
	}
}

// SimulateReaction triggers the OnReaction callback.
func (m *MockPlatformClient) SimulateReaction(reaction PlatformReaction, user *PlatformUser) {
	if m.onReaction != nil {
		m.onReaction(reaction, user)
	}
}

// Compile-time interface check.
var _ PlatformClient = (*MockPlatformClient)(nil)
