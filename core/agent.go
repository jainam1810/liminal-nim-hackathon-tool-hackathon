package core

import (
	"context"
)

// Agent is the interface that all agents must implement.
// An agent processes user input and produces output using tools.
type Agent interface {
	// Run executes the agent with the given input and returns output.
	Run(ctx context.Context, input *Input) (*Output, error)

	// Capabilities returns the agent's capabilities and configuration.
	Capabilities() *Capabilities

	// Name returns the agent's unique identifier.
	Name() string
}

// Capabilities describes an agent's configuration and abilities.
type Capabilities struct {
	// CanRequestConfirmation indicates whether this agent can pause
	// execution to request user confirmation for write operations.
	CanRequestConfirmation bool

	// AvailableTools lists the tool names this agent can use.
	AvailableTools []string

	// Model is the Claude model to use (e.g., "claude-sonnet-4-20250514").
	Model string

	// MaxTokens is the maximum response tokens per turn.
	MaxTokens int64

	// MaxTurns is the maximum number of agentic turns.
	MaxTurns int

	// SystemPrompt is the system prompt for the agent.
	SystemPrompt string
}

// Input represents the input to an agent run.
type Input struct {
	// UserMessage is the user's message to process.
	UserMessage string

	// Context contains user identity, preferences, and execution limits.
	Context *Context

	// History contains previous messages in the conversation.
	History []Message

	// StreamCallback is an optional callback for streaming responses.
	StreamCallback func(chunk string, done bool)
}

// Output represents the output from an agent run.
type Output struct {
	// Type indicates the kind of output.
	Type OutputType

	// Text is the agent's text response.
	Text string

	// PendingAction is set when Type is OutputConfirmationNeeded.
	PendingAction *PendingAction

	// ToolsUsed records all tools invoked during this run.
	ToolsUsed []ToolExecution

	// ResponseBlocks contains the full response for persistence.
	ResponseBlocks []ContentBlock

	// TokensUsed tracks Claude API token consumption for this run.
	TokensUsed TokenUsage

	// Error is set when Type is OutputError.
	Error error
}

// OutputType indicates the kind of output from an agent run.
type OutputType int

const (
	// OutputComplete indicates the agent finished successfully.
	OutputComplete OutputType = iota

	// OutputConfirmationNeeded indicates a write operation needs user confirmation.
	OutputConfirmationNeeded

	// OutputError indicates an error occurred.
	OutputError
)

// DefaultCapabilities returns sensible default capabilities.
func DefaultCapabilities() *Capabilities {
	return &Capabilities{
		CanRequestConfirmation: true,
		Model:                  "claude-sonnet-4-20250514",
		MaxTokens:              4096,
		MaxTurns:               20,
	}
}

// SubAgentCapabilities returns capabilities suitable for sub-agents.
func SubAgentCapabilities() *Capabilities {
	return &Capabilities{
		CanRequestConfirmation: false, // Sub-agents cannot request confirmation
		Model:                  "claude-sonnet-4-20250514",
		MaxTokens:              2048,
		MaxTurns:               10,
	}
}
