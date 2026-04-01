package platform

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/anneschuth/claude-threads/internal/utils"
)

var baseLog = utils.CreateLogger("base-client")
var wsLog = utils.WsLogger

// BasePlatformClient provides shared connection management for platform clients.
// Embed this struct and call its methods from Connect/Disconnect implementations.
type BasePlatformClient struct {
	mu sync.RWMutex

	AllowedUsers []string
	BotNameVal   string

	// Connection state
	intentionalDisconnect bool
	isReconnecting        bool

	// Heartbeat
	heartbeatStop     chan struct{}
	heartbeatDone     chan struct{}
	lastMessageAt     time.Time
	HeartbeatInterval time.Duration // default 30s
	HeartbeatTimeout  time.Duration // default 60s

	// Reconnect
	reconnectAttempts    int
	MaxReconnectAttempts int           // default 10
	ReconnectBaseDelay   time.Duration // default 1s

	// Callbacks
	onConnected       func()
	onDisconnected    func()
	onReconnecting    func(int)
	onError           func(error)
	onMessage         func(PlatformPost, *PlatformUser)
	onReaction        func(PlatformReaction, *PlatformUser)
	onReactionRemoved func(PlatformReaction, *PlatformUser)
	onChannelPost     func(PlatformPost, *PlatformUser)

	// connectFn is set by the concrete client so BasePlatformClient can
	// call Connect() during reconnection without knowing the concrete type.
	connectFn       func(ctx context.Context) error
	forceCloseFn    func()
	recoverMissedFn func() error
}

// InitBase sets defaults. Call from concrete client's constructor.
func (b *BasePlatformClient) InitBase(
	connectFn func(ctx context.Context) error,
	forceCloseFn func(),
	recoverMissedFn func() error,
) {
	b.connectFn = connectFn
	b.forceCloseFn = forceCloseFn
	b.recoverMissedFn = recoverMissedFn
	b.lastMessageAt = time.Now()
	if b.HeartbeatInterval == 0 {
		b.HeartbeatInterval = 30 * time.Second
	}
	if b.HeartbeatTimeout == 0 {
		b.HeartbeatTimeout = 60 * time.Second
	}
	if b.MaxReconnectAttempts == 0 {
		b.MaxReconnectAttempts = 10
	}
	if b.ReconnectBaseDelay == 0 {
		b.ReconnectBaseDelay = time.Second
	}
}

// IsUserAllowed reports whether username is allowed (empty list = allow all).
func (b *BasePlatformClient) IsUserAllowed(username string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.AllowedUsers) == 0 {
		return true
	}
	for _, u := range b.AllowedUsers {
		if u == username {
			return true
		}
	}
	return false
}

// GetBotName returns the bot's mention name.
func (b *BasePlatformClient) GetBotName() string { return b.BotNameVal }

// Disconnect marks disconnect as intentional, stops heartbeat, closes connection.
func (b *BasePlatformClient) Disconnect() {
	wsLog.Info("Disconnecting (intentional)")
	b.mu.Lock()
	b.intentionalDisconnect = true
	b.mu.Unlock()
	b.stopHeartbeat()
	if b.forceCloseFn != nil {
		b.forceCloseFn()
	}
}

// PrepareForReconnect resets state so Connect() can be called again.
func (b *BasePlatformClient) PrepareForReconnect() {
	b.mu.Lock()
	b.intentionalDisconnect = false
	b.reconnectAttempts = 0
	b.mu.Unlock()
}

// UpdateLastMessageTime records activity. Call on every incoming WS message.
func (b *BasePlatformClient) UpdateLastMessageTime() {
	b.mu.Lock()
	b.lastMessageAt = time.Now()
	b.mu.Unlock()
}

// StartHeartbeat begins monitoring the connection for inactivity.
func (b *BasePlatformClient) StartHeartbeat() {
	b.stopHeartbeat()
	b.mu.Lock()
	b.lastMessageAt = time.Now()
	stop := make(chan struct{})
	done := make(chan struct{})
	b.heartbeatStop = stop
	b.heartbeatDone = done
	b.mu.Unlock()

	go func() {
		defer close(done)
		ticker := time.NewTicker(b.HeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				b.mu.RLock()
				silent := time.Since(b.lastMessageAt)
				b.mu.RUnlock()
				if silent > b.HeartbeatTimeout {
					baseLog.Warn("Connection dead (" + silent.String() + " silent), reconnecting...")
					b.stopHeartbeat()
					b.scheduleReconnect()
					return
				}
				wsLog.Debug("Heartbeat ok (" + silent.String() + " ago)")
			}
		}
	}()
}

func (b *BasePlatformClient) stopHeartbeat() {
	b.mu.Lock()
	stop := b.heartbeatStop
	done := b.heartbeatDone
	b.heartbeatStop = nil
	b.heartbeatDone = nil
	b.mu.Unlock()

	if stop != nil {
		select {
		case <-stop:
			// already closed
		default:
			close(stop)
		}
		if done != nil {
			<-done
		}
	}
}

func (b *BasePlatformClient) scheduleReconnect() {
	b.mu.RLock()
	intentional := b.intentionalDisconnect
	attempts := b.reconnectAttempts
	b.mu.RUnlock()

	if intentional {
		wsLog.Debug("Skipping reconnect: intentional disconnect")
		return
	}
	if attempts >= b.MaxReconnectAttempts {
		baseLog.Error("Max reconnection attempts reached", nil)
		return
	}

	if b.forceCloseFn != nil {
		b.forceCloseFn()
	}

	b.mu.Lock()
	b.isReconnecting = true
	b.reconnectAttempts++
	attempt := b.reconnectAttempts
	b.mu.Unlock()

	delay := time.Duration(float64(b.ReconnectBaseDelay) * math.Pow(2, float64(attempt-1)))
	wsLog.Info("Reconnecting in " + delay.String() + " (attempt " + fmt.Sprintf("%d/%d", attempt, b.MaxReconnectAttempts) + ")")
	b.EmitReconnecting(attempt)

	go func() {
		time.Sleep(delay)
		b.mu.RLock()
		intentional := b.intentionalDisconnect
		b.mu.RUnlock()
		if intentional {
			wsLog.Debug("Skipping reconnect: intentional disconnect was called")
			return
		}
		if b.connectFn != nil {
			if err := b.connectFn(context.Background()); err != nil {
				wsLog.Error("Reconnection failed", err)
				b.scheduleReconnect()
			}
		}
	}()
}

// OnConnectionEstablished resets reconnect counter, starts heartbeat, emits connected.
// Call from Connect() after authentication succeeds.
func (b *BasePlatformClient) OnConnectionEstablished() {
	b.mu.Lock()
	wasReconnecting := b.isReconnecting
	b.reconnectAttempts = 0
	b.isReconnecting = false
	b.mu.Unlock()

	b.StartHeartbeat()
	b.EmitConnected()

	if wasReconnecting && b.recoverMissedFn != nil {
		go func() {
			if err := b.recoverMissedFn(); err != nil {
				baseLog.Warn("Failed to recover missed messages: " + err.Error())
			}
		}()
	}
}

// OnConnectionClosed stops heartbeat and schedules reconnect if not intentional.
// Call from WebSocket onclose handler.
func (b *BasePlatformClient) OnConnectionClosed() {
	b.stopHeartbeat()
	b.EmitDisconnected()

	b.mu.RLock()
	intentional := b.intentionalDisconnect
	b.mu.RUnlock()

	if !intentional {
		b.scheduleReconnect()
	}
}

// ---------------------------------------------------------------------------
// Callback registration (implements PlatformClient On* methods)
// ---------------------------------------------------------------------------

func (b *BasePlatformClient) OnConnected(f func()) {
	b.mu.Lock()
	b.onConnected = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnDisconnected(f func()) {
	b.mu.Lock()
	b.onDisconnected = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnReconnecting(f func(int)) {
	b.mu.Lock()
	b.onReconnecting = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnError(f func(error)) {
	b.mu.Lock()
	b.onError = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnMessage(f func(PlatformPost, *PlatformUser)) {
	b.mu.Lock()
	b.onMessage = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnReaction(f func(PlatformReaction, *PlatformUser)) {
	b.mu.Lock()
	b.onReaction = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnReactionRemoved(f func(PlatformReaction, *PlatformUser)) {
	b.mu.Lock()
	b.onReactionRemoved = f
	b.mu.Unlock()
}
func (b *BasePlatformClient) OnChannelPost(f func(PlatformPost, *PlatformUser)) {
	b.mu.Lock()
	b.onChannelPost = f
	b.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Emit helpers (thread-safe)
// ---------------------------------------------------------------------------

func (b *BasePlatformClient) EmitConnected() {
	b.mu.RLock()
	f := b.onConnected
	b.mu.RUnlock()
	if f != nil {
		f()
	}
}
func (b *BasePlatformClient) EmitDisconnected() {
	b.mu.RLock()
	f := b.onDisconnected
	b.mu.RUnlock()
	if f != nil {
		f()
	}
}
func (b *BasePlatformClient) EmitReconnecting(attempt int) {
	b.mu.RLock()
	f := b.onReconnecting
	b.mu.RUnlock()
	if f != nil {
		f(attempt)
	}
}
func (b *BasePlatformClient) EmitError(err error) {
	b.mu.RLock()
	f := b.onError
	b.mu.RUnlock()
	if f != nil {
		f(err)
	}
}
func (b *BasePlatformClient) EmitMessage(post PlatformPost, user *PlatformUser) {
	b.mu.RLock()
	f := b.onMessage
	b.mu.RUnlock()
	if f != nil {
		f(post, user)
	}
}
func (b *BasePlatformClient) EmitReaction(reaction PlatformReaction, user *PlatformUser) {
	b.mu.RLock()
	f := b.onReaction
	b.mu.RUnlock()
	if f != nil {
		f(reaction, user)
	}
}
func (b *BasePlatformClient) EmitReactionRemoved(reaction PlatformReaction, user *PlatformUser) {
	b.mu.RLock()
	f := b.onReactionRemoved
	b.mu.RUnlock()
	if f != nil {
		f(reaction, user)
	}
}
func (b *BasePlatformClient) EmitChannelPost(post PlatformPost, user *PlatformUser) {
	b.mu.RLock()
	f := b.onChannelPost
	b.mu.RUnlock()
	if f != nil {
		f(post, user)
	}
}
