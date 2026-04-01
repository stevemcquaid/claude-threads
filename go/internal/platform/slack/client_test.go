package slack_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anneschuth/claude-threads/internal/config"
	"github.com/anneschuth/claude-threads/internal/platform/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSlackTestServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()
	for path, h := range handlers {
		mux.HandleFunc("/"+path, h)
	}
	return httptest.NewServer(mux)
}

func testSlackConfig(apiURL string) config.SlackPlatformConfig {
	return config.SlackPlatformConfig{
		PlatformInstanceConfig: config.PlatformInstanceConfig{
			ID:          "slack-test",
			DisplayName: "Test Slack",
		},
		BotToken:     "xoxb-test",
		AppToken:     "xapp-test",
		ChannelID:    "C123456",
		BotName:      "testbot",
		AllowedUsers: []string{"alice", "bob"},
		APIURL:       apiURL,
	}
}

func TestSlackClient_PlatformIdentity(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	assert.Equal(t, "slack-test", client.PlatformID())
	assert.Equal(t, "slack", client.PlatformType())
	assert.Equal(t, "Test Slack", client.DisplayName())
}

func TestSlackClient_IsUserAllowed(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	assert.True(t, client.IsUserAllowed("alice"))
	assert.False(t, client.IsUserAllowed("charlie"))
}

func TestSlackClient_IsUserAllowed_EmptyList(t *testing.T) {
	cfg := testSlackConfig("http://localhost")
	cfg.AllowedUsers = nil
	client := slack.NewClient(cfg)
	assert.True(t, client.IsUserAllowed("anyone"))
}

func TestSlackClient_IsBotMentioned_ByName(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	assert.True(t, client.IsBotMentioned("@testbot hello"))
	assert.False(t, client.IsBotMentioned("hello world"))
}

func TestSlackClient_ExtractPrompt(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	assert.Equal(t, "hello", client.ExtractPrompt("@testbot hello"))
}

func TestSlackClient_GetMessageLimits(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	limits := client.GetMessageLimits()
	assert.Equal(t, 12000, limits.MaxLength)
}

func TestSlackClient_CreatePost(t *testing.T) {
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer xoxb-test", r.Header.Get("Authorization"))
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "C123456", body["channel"])
			assert.Equal(t, "Hello!", body["text"])
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":      true,
				"channel": "C123456",
				"ts":      "1234567890.000001",
				"message": map[string]interface{}{"text": "Hello!"},
			})
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	post, err := client.CreatePost(context.Background(), "Hello!", "")
	require.NoError(t, err)
	assert.Equal(t, "1234567890.000001", post.ID)
}

func TestSlackClient_CreatePost_WithThread(t *testing.T) {
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"chat.postMessage": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "1234567890.000000", body["thread_ts"])
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":      true,
				"channel": "C123456",
				"ts":      "1234567890.000002",
				"message": map[string]interface{}{"text": "Reply"},
			})
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	post, err := client.CreatePost(context.Background(), "Reply", "1234567890.000000")
	require.NoError(t, err)
	assert.Equal(t, "1234567890.000000", post.RootID)
}

func TestSlackClient_UpdatePost(t *testing.T) {
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"chat.update": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "post-ts", body["ts"])
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":      true,
				"channel": "C123456",
				"ts":      "post-ts",
				"text":    "Updated!",
			})
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	post, err := client.UpdatePost(context.Background(), "post-ts", "Updated!")
	require.NoError(t, err)
	assert.Equal(t, "post-ts", post.ID)
}

func TestSlackClient_GetThreadLink_WithTeamURL(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	link := client.GetThreadLink("1234567890.123456", "", "")
	assert.Equal(t, "#1234567890.123456", link)
}

func TestSlackClient_AddReaction_ConvertsUnicode(t *testing.T) {
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"reactions.add": func(w http.ResponseWriter, r *http.Request) {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "+1", body["name"])
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	err := client.AddReaction(context.Background(), "post-ts", "👍")
	require.NoError(t, err)
}

func TestSlackClient_UserCaching(t *testing.T) {
	callCount := 0
	srv := newSlackTestServer(t, map[string]http.HandlerFunc{
		"users.info": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok": true,
				"user": map[string]interface{}{
					"id": "U123", "name": "alice",
					"profile": map[string]interface{}{},
				},
			})
		},
	})
	defer srv.Close()

	client := slack.NewClient(testSlackConfig(srv.URL))
	client.GetUser(context.Background(), "U123")
	client.GetUser(context.Background(), "U123")
	assert.Equal(t, 1, callCount, "should cache user after first fetch")
}

func TestSlackClient_IsBotMentioned_ByID(t *testing.T) {
	client := slack.NewClient(testSlackConfig("http://localhost"))
	client.SetBotUserIDForTest("U_BOT123")
	assert.True(t, client.IsBotMentioned("<@U_BOT123> hello"))
	assert.False(t, client.IsBotMentioned("<@U_SOMEONE_ELSE> hello"))
}
