package engine

import (
	"context"
	"encoding/json"
)

// AuditLogger logs tool executions for compliance and debugging.
// This is an interface - implementations (e.g., PostgreSQL-backed) are provided
// by the consuming application.
type AuditLogger interface {
	// Log records an audit entry for a tool execution.
	Log(ctx context.Context, entry *AuditEntry) error
}

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	// ID is the unique identifier for this audit entry.
	ID string `json:"id"`

	// UserID is the user who initiated the action.
	UserID string `json:"user_id"`

	// SessionID identifies the agent session.
	SessionID string `json:"session_id"`

	// RequestID is the unique request identifier for tracing.
	RequestID string `json:"request_id"`

	// ParentID links sub-agent entries to their parent.
	// Nil for top-level agent executions.
	ParentID *string `json:"parent_id,omitempty"`

	// AgentName identifies which agent executed the tool.
	AgentName string `json:"agent_name"`

	// ToolName is the name of the tool that was executed.
	ToolName string `json:"tool_name"`

	// ToolInput contains the tool parameters as JSON.
	ToolInput json.RawMessage `json:"tool_input"`

	// ToolOutput contains the tool result as JSON.
	ToolOutput json.RawMessage `json:"tool_output,omitempty"`

	// Error contains any error message if the tool failed.
	Error *string `json:"error,omitempty"`

	// DurationMs is the execution time in milliseconds.
	DurationMs int64 `json:"duration_ms"`

	// IsWriteOp indicates whether this was a write operation.
	IsWriteOp bool `json:"is_write_op"`

	// Timestamp is when the tool execution started (Unix timestamp).
	Timestamp int64 `json:"timestamp"`
}

// NoOpAuditLogger is an audit logger that discards all entries.
// Useful for development and testing.
type NoOpAuditLogger struct{}

// Log discards the audit entry.
func (n *NoOpAuditLogger) Log(ctx context.Context, entry *AuditEntry) error {
	return nil
}

// MemoryAuditLogger stores audit entries in memory.
// Useful for testing and debugging.
type MemoryAuditLogger struct {
	entries []*AuditEntry
}

// NewMemoryAuditLogger creates a new in-memory audit logger.
func NewMemoryAuditLogger() *MemoryAuditLogger {
	return &MemoryAuditLogger{
		entries: make([]*AuditEntry, 0),
	}
}

// Log stores the audit entry in memory.
func (m *MemoryAuditLogger) Log(ctx context.Context, entry *AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

// Entries returns all stored audit entries.
func (m *MemoryAuditLogger) Entries() []*AuditEntry {
	return m.entries
}

// Clear removes all stored entries.
func (m *MemoryAuditLogger) Clear() {
	m.entries = make([]*AuditEntry, 0)
}
