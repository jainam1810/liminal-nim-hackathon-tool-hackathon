package store

import (
	"context"

	"github.com/becomeliminal/nim-go-sdk/core"
)

// Confirmations stores pending actions awaiting user approval.
// The SDK provides MemoryConfirmations for development and RistrettoConfirmations
// for production single-instance deployments. Distributed deployments (like nim/agent)
// should implement this interface with Redis or similar.
type Confirmations interface {
	// Store saves a pending action.
	Store(ctx context.Context, action *core.PendingAction) error

	// Get retrieves a pending action by ID for the given user.
	// Returns error if not found or expired.
	Get(ctx context.Context, userID, actionID string) (*core.PendingAction, error)

	// GetByIdempotency retrieves a pending action by its idempotency key.
	// Returns nil, nil if no action found (not an error).
	GetByIdempotency(ctx context.Context, userID, key string) (*core.PendingAction, error)

	// Confirm marks an action as confirmed, removes it from pending, and returns it.
	// The caller should then execute the confirmed action.
	Confirm(ctx context.Context, userID, actionID string) (*core.PendingAction, error)

	// Cancel removes a pending action without executing it.
	Cancel(ctx context.Context, userID, actionID string) error

	// Cleanup removes all expired actions. Returns count of removed actions.
	Cleanup(ctx context.Context) (int, error)
}

// Conversations stores conversation history.
// The SDK provides MemoryConversations for development.
// Production deployments should implement with PostgreSQL or similar.
type Conversations interface {
	// Create starts a new conversation for the user.
	Create(ctx context.Context, userID string) (*Conversation, error)

	// Get retrieves a conversation with all messages.
	Get(ctx context.Context, conversationID string) (*ConversationWithMessages, error)

	// Append adds a message to a conversation.
	Append(ctx context.Context, msg *AppendMessage) error

	// SetTitle updates the conversation title.
	SetTitle(ctx context.Context, conversationID, title string) error

	// List returns recent conversations for a user.
	List(ctx context.Context, userID string, limit int) ([]*Conversation, error)

	// Delete removes a conversation.
	Delete(ctx context.Context, conversationID string) error
}
