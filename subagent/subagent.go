// Package subagent provides a framework for creating specialized sub-agents.
// Sub-agents are agents with restricted capabilities that can be delegated
// specific tasks by a parent agent.
package subagent

import (
	"context"
	"fmt"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/engine"
)

// SubAgent is a specialized agent with restricted capabilities.
// It implements the core.Agent interface and can be run by the engine.
type SubAgent struct {
	name           string
	systemPrompt   string
	availableTools []string
	model          string
	maxTokens      int64
	maxTurns       int
	engine         *engine.Engine
}

// SubAgentConfig configures a sub-agent.
type SubAgentConfig struct {
	// Name is the unique identifier for this sub-agent.
	Name string

	// SystemPrompt is the specialized system prompt for this sub-agent.
	SystemPrompt string

	// AvailableTools lists the tool names this sub-agent can use.
	AvailableTools []string

	// Model is the Claude model to use. Defaults to claude-sonnet-4-20250514.
	Model string

	// MaxTokens is the maximum response tokens per turn. Defaults to 2048.
	MaxTokens int64

	// MaxTurns is the maximum number of agentic turns. Defaults to 10.
	MaxTurns int
}

// NewSubAgent creates a new sub-agent with the given configuration.
func NewSubAgent(eng *engine.Engine, cfg SubAgentConfig) *SubAgent {
	// Apply defaults
	model := cfg.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	maxTokens := cfg.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2048
	}
	maxTurns := cfg.MaxTurns
	if maxTurns == 0 {
		maxTurns = 10
	}

	return &SubAgent{
		name:           cfg.Name,
		systemPrompt:   cfg.SystemPrompt,
		availableTools: cfg.AvailableTools,
		model:          model,
		maxTokens:      maxTokens,
		maxTurns:       maxTurns,
		engine:         eng,
	}
}

// Name returns the sub-agent's unique identifier.
func (s *SubAgent) Name() string {
	return s.name
}

// Capabilities returns the sub-agent's configuration.
func (s *SubAgent) Capabilities() *core.Capabilities {
	return &core.Capabilities{
		CanRequestConfirmation: false, // Sub-agents cannot request confirmation
		AvailableTools:         s.availableTools,
		Model:                  s.model,
		MaxTokens:              s.maxTokens,
		MaxTurns:               s.maxTurns,
		SystemPrompt:           s.systemPrompt,
	}
}

// Run executes the sub-agent with the given input.
func (s *SubAgent) Run(ctx context.Context, input *core.Input) (*core.Output, error) {
	// Ensure sub-agent context has restricted limits
	if input.Context != nil && input.Context.Limits == nil {
		input.Context.Limits = core.SubAgentLimits()
	}

	// Override limits to prevent confirmation requests
	if input.Context != nil && input.Context.Limits != nil {
		input.Context.Limits.CanConfirm = false
	}

	// Run via engine with filtered tools
	return s.engine.RunAgent(ctx, s, input)
}

// RunWithTask executes the sub-agent with a formatted task message.
func (s *SubAgent) RunWithTask(ctx context.Context, parentCtx *core.Context, task string) (*core.Output, error) {
	// Create sub-agent context from parent
	subCtx := parentCtx.ForSubAgent(fmt.Sprintf("%s-%s", parentCtx.RequestID, s.name))

	input := &core.Input{
		UserMessage: task,
		Context:     subCtx,
		History:     []core.Message{},
	}

	return s.Run(ctx, input)
}

// SubAgentResult represents the result of a sub-agent execution.
type SubAgentResult struct {
	// AgentName is the name of the sub-agent that was executed.
	AgentName string `json:"agent_name"`

	// Success indicates whether the sub-agent completed successfully.
	Success bool `json:"success"`

	// Response is the sub-agent's text response.
	Response string `json:"response"`

	// Error is set if the sub-agent failed.
	Error string `json:"error,omitempty"`

	// ToolsUsed lists the tools invoked during execution.
	ToolsUsed []core.ToolExecution `json:"tools_used,omitempty"`

	// TokensUsed tracks token consumption.
	TokensUsed core.TokenUsage `json:"tokens_used"`
}

// ToResult converts an Output to a SubAgentResult.
func ToResult(agentName string, output *core.Output) *SubAgentResult {
	result := &SubAgentResult{
		AgentName:  agentName,
		TokensUsed: output.TokensUsed,
		ToolsUsed:  output.ToolsUsed,
	}

	switch output.Type {
	case core.OutputComplete:
		result.Success = true
		result.Response = output.Text
	case core.OutputError:
		result.Success = false
		if output.Error != nil {
			result.Error = output.Error.Error()
		}
	case core.OutputConfirmationNeeded:
		// Sub-agents should never reach this state
		result.Success = false
		result.Error = "sub-agent attempted to request confirmation"
	}

	return result
}
