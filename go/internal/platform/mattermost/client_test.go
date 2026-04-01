package mattermost_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/anneschuth/claude-threads/internal/config"
	"github.com/anneschuth/claude-threads/internal/platform/mattermost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServer creates an httptest.Server that handles the Mattermost REST API.
// handlers is a map from path to handler func.
func newTestServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}
	return httptest.NewServer(mux)
}

func testConfig(serverURL string) config.MattermostPlatformConfig {
	return config.MattermostPlatformConfig{
		PlatformInstanceConfig: config.PlatformInstanceConfig{
			ID:          "mm-test",
			DisplayName: "Test Mattermost",
		},
		URL:          serverURL,
		Token:        "test-token",
		ChannelID:    "channel123",
		BotName:      "testbot",
		AllowedUsers: []string{"alice", "bob"},
	}
}

func TestMattermostClient_GetBotUser(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/users/me": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			json.NewEncoder(w).Encode(mattermost.User{
				ID:       "bot-id",
				Username: "testbot",
			})
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	user, err := client.GetBotUser(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "bot-id", user.ID)
	assert.Equal(t, "testbot", user.Username)
}

func TestMattermostClient_GetUser(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/users/user123": func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(mattermost.User{
				ID:       "user123",
				Username: "alice",
			})
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	user, err := client.GetUser(context.Background(), "user123")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "alice", user.Username)
}

func TestMattermostClient_GetUser_NotFound(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/users/missing": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	user, err := client.GetUser(context.Background(), "missing")
	// Returns nil, nil for not-found (same behavior as TypeScript)
	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestMattermostClient_CreatePost(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/posts": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			var req mattermost.CreatePostRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "channel123", req.ChannelID)
			assert.Equal(t, "Hello!", req.Message)
			json.NewEncoder(w).Encode(mattermost.Post{
				ID:        "post-abc",
				ChannelID: "channel123",
				Message:   "Hello!",
			})
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	post, err := client.CreatePost(context.Background(), "Hello!", "")
	require.NoError(t, err)
	assert.Equal(t, "post-abc", post.ID)
	assert.Equal(t, "Hello!", post.Message)
}

func TestMattermostClient_IsUserAllowed(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	assert.True(t, client.IsUserAllowed("alice"))
	assert.True(t, client.IsUserAllowed("bob"))
	assert.False(t, client.IsUserAllowed("charlie"))
}

func TestMattermostClient_IsUserAllowed_EmptyList(t *testing.T) {
	cfg := testConfig("http://localhost")
	cfg.AllowedUsers = nil
	client := mattermost.NewClient(cfg)
	assert.True(t, client.IsUserAllowed("anyone"))
}

func TestMattermostClient_IsBotMentioned(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	assert.True(t, client.IsBotMentioned("@testbot hello"))
	assert.True(t, client.IsBotMentioned("hey @testbot"))
	assert.False(t, client.IsBotMentioned("hello world"))
}

func TestMattermostClient_ExtractPrompt(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	assert.Equal(t, "hello", client.ExtractPrompt("@testbot hello"))
	assert.Equal(t, "hey", client.ExtractPrompt("hey @testbot"))
}

func TestMattermostClient_GetThreadLink(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://mm.example.com"))
	link := client.GetThreadLink("thread123", "", "")
	assert.Equal(t, "http://mm.example.com/_redirect/pl/thread123", link)
	link2 := client.GetThreadLink("thread123", "msg456", "")
	assert.Equal(t, "http://mm.example.com/_redirect/pl/msg456", link2)
}

func TestMattermostClient_GetMessageLimits(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	limits := client.GetMessageLimits()
	assert.Equal(t, 16000, limits.MaxLength)
	assert.Equal(t, 14000, limits.HardThreshold)
}

func TestMattermostClient_PlatformIdentity(t *testing.T) {
	client := mattermost.NewClient(testConfig("http://localhost"))
	assert.Equal(t, "mm-test", client.PlatformID())
	assert.Equal(t, "mattermost", client.PlatformType())
	assert.Equal(t, "Test Mattermost", client.DisplayName())
}

func TestMattermostClient_UserCaching(t *testing.T) {
	callCount := 0
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/api/v4/users/user123": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			json.NewEncoder(w).Encode(mattermost.User{ID: "user123", Username: "alice"})
		},
	})
	defer srv.Close()

	client := mattermost.NewClient(testConfig(srv.URL))
	// Call twice — should only hit API once
	client.GetUser(context.Background(), "user123")
	client.GetUser(context.Background(), "user123")
	assert.Equal(t, 1, callCount, "should cache user after first fetch")
}

// Ensure time import is used.
var _ = time.Second
