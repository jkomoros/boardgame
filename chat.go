package boardgame

import (
	"log"
	"time"
)

// ChatMessage represents a single message in the chat side-channel. Chat
// messages are stored separately from game state — they do not affect game
// logic, are not included in golden tests, and are not part of the move
// pipeline. See the ChatStorageManager interface for storage, and
// GameDelegate.ChatConfig / ChatPolicyForPlayer for access control.
type ChatMessage struct {
	// ID is a unique identifier for this message, assigned by the storage layer.
	ID string

	// GameID is the game this message belongs to.
	GameID string

	// Version is the game state version when this message was sent. Used for
	// replay correlation (showing chat alongside game state at a point in time).
	Version int

	// Sender is the player who sent the message. AdminPlayerIndex (-2) is used
	// for system messages (player joined, phase changed, etc.).
	Sender PlayerIndex

	// Channel identifies the chat channel. Convention:
	//   "all"           — visible to all players
	//   "team/Red"      — visible to players on team "Red"
	//   "dm/userA/userB" — private DM between two users (IDs sorted lexicographically)
	//   "system"        — framework-generated system messages
	// The "/" separator is used because it is illegal in enum value names.
	Channel string

	// Body is the message text, or for pre-baked messages, a key into the
	// allowed messages set defined by ChatConfig.
	Body string

	// Timestamp is when the message was sent.
	Timestamp time.Time
}

// ChatStorageManager is an optional interface that storage backends can
// implement to support the chat side-channel. The server discovers this
// capability via type assertion at startup. If the storage backend does not
// implement ChatStorageManager, chat is disabled gracefully.
//
// This follows the same pattern as server/api.ServerStorageManager — an
// optional extension of the base StorageManager interface.
type ChatStorageManager interface {
	// SaveChatMessage persists a chat message. The storage layer assigns the
	// ID field if empty.
	SaveChatMessage(msg *ChatMessage) error

	// ChatMessages returns messages for the given game and channel, ordered by
	// timestamp. If sinceID is non-empty, only messages after that ID are
	// returned. Limit caps the number of messages returned (0 = no limit).
	ChatMessages(gameID string, channel string, sinceID string, limit int) ([]*ChatMessage, error)
}

// ChatConfig describes the game-level chat configuration. It defines the
// ceiling of what chat features are available — ChatPolicyForPlayer can only
// restrict within this ceiling, never expand beyond it.
//
// Use DefaultChatConfig() to start with all features enabled, then chain
// builder methods to disable what you don't want:
//
//	DefaultChatConfig()                              // all-chat + team + DMs
//	DefaultChatConfig().WithoutDMs()                 // no private messaging
//	DefaultChatConfig().WithoutTeamChat()             // no team channels
//	DefaultChatConfig().Disabled()                    // no chat at all
//	DefaultChatConfig().PrebakedOnly("👍", "😂", "gg") // reactions only
type ChatConfig struct {
	enabled     bool
	allChat     bool
	teamChat    bool
	dmChat      bool
	prebaked    bool
	allowedMsgs []string
}

// DefaultChatConfig returns a ChatConfig with all features enabled: all-chat,
// team chat (auto-detected from PlayerTeam), and private DMs between all
// seated player pairs.
// DefaultChatConfig returns a ChatConfig with all-chat and team chat enabled.
// DM chat is disabled by default (not yet fully implemented — planned for a
// future release). Use WithDMs() to enable when DM support is added.
func DefaultChatConfig() ChatConfig {
	return ChatConfig{
		enabled:  true,
		allChat:  true,
		teamChat: true,
		dmChat:   false,
	}
}

// Disabled returns a copy with chat entirely disabled.
func (c ChatConfig) Disabled() ChatConfig {
	c.enabled = false
	return c
}

// WithoutDMs returns a copy with private DM channels disabled.
func (c ChatConfig) WithoutDMs() ChatConfig {
	c.dmChat = false
	return c
}

// WithoutTeamChat returns a copy with team channels disabled, even if
// playerState embeds PlayerTeam.
func (c ChatConfig) WithoutTeamChat() ChatConfig {
	c.teamChat = false
	return c
}

// WithoutAllChat returns a copy with the "all" channel disabled. Players can
// still use team or DM channels if enabled.
func (c ChatConfig) WithoutAllChat() ChatConfig {
	c.allChat = false
	return c
}

// PrebakedOnly returns a copy that restricts chat to a fixed set of allowed
// messages. The text input is replaced with a chip picker in the client.
func (c ChatConfig) PrebakedOnly(msgs ...string) ChatConfig {
	c.prebaked = true
	c.allowedMsgs = msgs
	return c
}

// IsEnabled returns whether chat is enabled.
func (c ChatConfig) IsEnabled() bool { return c.enabled }

// AllChatEnabled returns whether the "all" channel is enabled.
func (c ChatConfig) AllChatEnabled() bool { return c.enabled && c.allChat }

// TeamChatEnabled returns whether team channels are enabled.
func (c ChatConfig) TeamChatEnabled() bool { return c.enabled && c.teamChat }

// DMChatEnabled returns whether private DM channels are enabled.
func (c ChatConfig) DMChatEnabled() bool { return c.enabled && c.dmChat }

// IsPrebakedOnly returns whether chat is restricted to pre-baked messages.
func (c ChatConfig) IsPrebakedOnly() bool { return c.prebaked }

// AllowedMessages returns the set of allowed pre-baked messages, or nil if
// free-text is allowed.
func (c ChatConfig) AllowedMessages() []string { return c.allowedMsgs }

// ChatPolicy describes what a specific player can do with chat right now.
// Returned by GameDelegate.ChatPolicyForPlayer.
type ChatPolicy struct {
	// Enabled is false if this player cannot use chat at all right now.
	Enabled bool

	// SendChannels lists channels this player can send messages to.
	SendChannels []string

	// ViewChannels lists channels this player can read messages from.
	ViewChannels []string

	// PrebakedOnly is true if this player can only send pre-baked messages.
	PrebakedOnly bool

	// AllowedMessages is the set of allowed message keys when PrebakedOnly is
	// true. Ignored when PrebakedOnly is false.
	AllowedMessages []string
}

// EmitSystemMessage queues a system chat message to be saved when the current
// move is successfully committed. Call this from within a move's Apply method
// to generate framework or game-specific system messages like "Player drew a
// wild card" or "Round 3 starting".
//
// The message is buffered on the state — if the move fails or is rolled back,
// the buffered messages are discarded. System messages have
// Sender=AdminPlayerIndex and Channel="all".
//
// The server layer (or any layer with access to ChatStorageManager) retrieves
// buffered messages via PendingChatMessages() after a successful commit and
// persists them.
//
//	func (m *MyMove) Apply(state boardgame.State) error {
//	    // ... game logic ...
//	    boardgame.EmitSystemMessage(state, "Something happened!")
//	    return nil
//	}
// EmitSystemMessage queues a system chat message to be flushed after the
// current move is successfully committed. Uses AddCommittedCallback internally
// so the message is discarded if the move fails.
//
// The callback checks if the storage implements ChatStorageManager and saves
// the message. If chat storage is not available, the message is silently
// discarded.
func EmitSystemMessage(st State, body string) {
	gameID := st.Game().ID()
	version := st.Version()
	storage := st.Manager().Storage()

	st.Manager().Internals().AddCommittedCallback(st, func() {
		chatStorage, ok := storage.(ChatStorageManager)
		if !ok {
			return
		}
		msg := &ChatMessage{
			GameID:    gameID,
			Version:   version,
			Sender:    AdminPlayerIndex,
			Channel:   "all",
			Body:      body,
			Timestamp: time.Now(),
		}
		// Run async to avoid blocking the fixup chain if storage is slow
		go func() {
			if err := chatStorage.SaveChatMessage(msg); err != nil {
				// Log but don't fail — chat is best-effort
				log.Println("EmitSystemMessage: failed to save:", err)
			}
		}()
	})
}
