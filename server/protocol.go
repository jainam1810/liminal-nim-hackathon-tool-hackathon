// Package server provides a ready-to-run WebSocket server for the Nim agent.
package server

// ClientMessage is a message from the client.
type ClientMessage struct {
	Type           string `json:"type"` // "new_conversation", "resume_conversation", "message", "confirm", "cancel"
	Content        string `json:"content,omitempty"`
	ActionID       string `json:"actionId,omitempty"`
	ConversationID string `json:"conversationId,omitempty"`
}

// ServerMessage is a message to the client.
type ServerMessage struct {
	Type           string      `json:"type"` // "conversation_started", "conversation_resumed", "text", "text_chunk", "confirm_request", "complete", "error"
	Content        string      `json:"content,omitempty"`
	ActionID       string      `json:"actionId,omitempty"`
	Tool           string      `json:"tool,omitempty"`
	Summary        string      `json:"summary,omitempty"`
	ExpiresAt      string      `json:"expiresAt,omitempty"`
	ConversationID string      `json:"conversationId,omitempty"`
	Messages       interface{} `json:"messages,omitempty"`
	TokenUsage     *TokenUsage `json:"tokenUsage,omitempty"`
}

// TokenUsage tracks Claude API token consumption.
type TokenUsage struct {
	InputTokens              int `json:"inputTokens"`
	OutputTokens             int `json:"outputTokens"`
	CacheCreationInputTokens int `json:"cacheCreationInputTokens,omitempty"`
	CacheReadInputTokens     int `json:"cacheReadInputTokens,omitempty"`
	TotalTokens              int `json:"totalTokens"`
}

// Confirmation contains details about a pending action.
type Confirmation struct {
	ID        string `json:"id"`
	Tool      string `json:"tool"`
	Summary   string `json:"summary"`
	ExpiresAt int64  `json:"expiresAt"`
}
