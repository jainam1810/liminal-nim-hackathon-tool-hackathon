package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/google/uuid"
)

// Engine is the agent runner that executes tools and manages Claude API interactions.
type Engine struct {
	client     *anthropic.Client
	registry   *ToolRegistry
	guardrails Guardrails  // Optional: rate limiting and circuit breaker
	audit      AuditLogger // Optional: audit logging
}

// Option configures the engine.
type Option func(*Engine)

// WithGuardrails sets the guardrails implementation for rate limiting.
func WithGuardrails(g Guardrails) Option {
	return func(e *Engine) {
		e.guardrails = g
	}
}

// WithAudit sets the audit logger implementation.
func WithAudit(a AuditLogger) Option {
	return func(e *Engine) {
		e.audit = a
	}
}

// NewEngine creates a new engine with the given Anthropic client and registry.
func NewEngine(client *anthropic.Client, registry *ToolRegistry, opts ...Option) *Engine {
	e := &Engine{
		client:   client,
		registry: registry,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Registry returns the engine's tool registry.
func (e *Engine) Registry() *ToolRegistry {
	return e.registry
}

// Input represents the input to an agent run.
type Input struct {
	// UserMessage is the user's message to process.
	UserMessage string

	// Context contains user identity, preferences, and execution limits.
	Context *core.Context

	// History contains previous messages in the conversation.
	History []core.Message

	// SystemPrompt is the system prompt to use.
	SystemPrompt string

	// Model is the Claude model to use.
	Model string

	// MaxTokens is the maximum response tokens.
	MaxTokens int64

	// AgentName identifies the agent for audit logging.
	// Defaults to "default" if not specified.
	AgentName string

	// AvailableTools filters which tools from the registry are available.
	// If empty, all registered tools are available.
	AvailableTools []string

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
	PendingAction *core.PendingAction

	// ToolsUsed records all tools invoked during this run.
	ToolsUsed []core.ToolExecution

	// ResponseBlocks contains the full response for persistence.
	ResponseBlocks []core.ContentBlock

	// TokensUsed tracks Claude API token consumption for this run.
	TokensUsed core.TokenUsage

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

// Run executes the agent loop until completion or confirmation is needed.
func (e *Engine) Run(ctx context.Context, input *Input) (*Output, error) {
	// Check guardrails if configured
	if e.guardrails != nil && input.Context != nil {
		result, err := e.guardrails.Check(ctx, input.Context.UserID)
		if err != nil {
			return &Output{
				Type:  OutputError,
				Error: fmt.Errorf("guardrails check failed: %w", err),
			}, nil
		}
		if !result.Allowed {
			return &Output{
				Type:  OutputError,
				Error: fmt.Errorf("request blocked by guardrails: %s", result.Warning),
			}, nil
		}
	}

	// Apply defaults
	model := input.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	maxTokens := input.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}
	systemPrompt := input.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = DefaultSystemPrompt
	}

	// Get limits from context
	maxTurns := 20
	canConfirm := true
	if input.Context != nil && input.Context.Limits != nil {
		maxTurns = input.Context.Limits.MaxTurns
		canConfirm = input.Context.Limits.CanConfirm
		if input.Context.Limits.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, input.Context.Limits.Timeout)
			defer cancel()
		}
	}

	// Create session
	userID := ""
	conversationID := ""
	if input.Context != nil {
		userID = input.Context.UserID
		conversationID = input.Context.ConversationID
	}
	session := NewSession(userID, conversationID)

	// Track cumulative token usage
	var totalTokens core.TokenUsage

	// Restore history
	session.RestoreHistory(input.History)

	// Add user message
	if input.UserMessage != "" {
		session.AddUserMessage(input.UserMessage)
	}

	// Get tools (filtered if AvailableTools is specified)
	var apiTools []anthropic.ToolUnionParam
	if len(input.AvailableTools) > 0 {
		apiTools = e.registry.ToAPIToolsFiltered(FilterByNames(input.AvailableTools...))
	} else {
		apiTools = e.registry.ToAPITools()
	}

	// Get agent name for audit logging
	agentName := input.AgentName
	if agentName == "" {
		agentName = "default"
	}

	// Get parent ID for audit chain
	var auditParentID *string
	if input.Context != nil && input.Context.AuditParentID != nil {
		auditParentID = input.Context.AuditParentID
	}

	for {
		// Check context cancellation
		if ctx.Err() != nil {
			return &Output{
				Type:       OutputError,
				Error:      fmt.Errorf("timed out: %w", ctx.Err()),
				TokensUsed: totalTokens,
			}, nil
		}

		// Check turn limit
		if session.TurnCount >= maxTurns {
			return &Output{
				Type:       OutputError,
				Error:      fmt.Errorf("exceeded maximum turns (%d)", maxTurns),
				TokensUsed: totalTokens,
			}, nil
		}

		session.IncrementTurnCount()

		// Build the message request
		params := anthropic.MessageNewParams{
			Model:     anthropic.Model(model),
			MaxTokens: maxTokens,
			Messages:  session.Messages(),
			System: []anthropic.TextBlockParam{
				{Text: systemPrompt},
			},
		}

		if len(apiTools) > 0 {
			params.Tools = apiTools
		}

		// Call Claude API
		var resp *anthropic.Message
		var err error

		if input.StreamCallback != nil {
			resp, err = e.createMessageStreaming(ctx, params, input.StreamCallback)
		} else {
			resp, err = e.client.Messages.New(ctx, params)
		}

		if err != nil {
			return &Output{
				Type:       OutputError,
				Error:      fmt.Errorf("claude API error: %w", err),
				TokensUsed: totalTokens,
			}, err
		}

		// Accumulate token usage
		totalTokens.InputTokens += int(resp.Usage.InputTokens)
		totalTokens.OutputTokens += int(resp.Usage.OutputTokens)

		// Process response blocks
		var toolResults []anthropic.ContentBlockParamUnion
		var textResponse string
		var toolsUsed []core.ToolExecution
		var confirmationNeeded *core.PendingAction

		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				textResponse += block.Text

			case "tool_use":
				toolName := block.Name
				toolInput := block.Input

				tool, ok := e.registry.Get(toolName)
				if !ok {
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						block.ID,
						fmt.Sprintf("unknown tool: %s", toolName),
						true,
					))
					continue
				}

				// Check if write operation requiring confirmation
				if tool.RequiresConfirmation() {
					if !canConfirm {
						toolResults = append(toolResults, anthropic.NewToolResultBlock(
							block.ID,
							"error: this operation requires user confirmation",
							true,
						))
						continue
					}

					inputBytes, _ := json.Marshal(toolInput)
					confirmationNeeded = &core.PendingAction{
						ID:             uuid.New().String(),
						IdempotencyKey: GenerateIdempotencyKey(session.UserID, toolName, inputBytes),
						SessionID:      session.ID,
						UserID:         session.UserID,
						Tool:           toolName,
						Input:          inputBytes,
						Summary:        tool.GetSummary(inputBytes),
						BlockID:        block.ID,
						CreatedAt:      time.Now().Unix(),
						ExpiresAt:      time.Now().Add(10 * time.Minute).Unix(),
					}
					break
				}

				// Execute read-only tool
				startTime := time.Now()
				inputBytes, _ := json.Marshal(toolInput)

				result, err := tool.Execute(ctx, &core.ToolParams{
					UserID:    session.UserID,
					Input:     inputBytes,
					RequestID: session.ID,
				})

				durationMs := time.Since(startTime).Milliseconds()
				execution := core.ToolExecution{
					Tool:       toolName,
					Input:      toolInput,
					DurationMs: durationMs,
				}

				// Log audit entry if configured
				if e.audit != nil {
					var outputBytes json.RawMessage
					var errStr *string
					if result != nil {
						outputBytes, _ = json.Marshal(result.Data)
						if result.Error != "" {
							errStr = &result.Error
						}
					}
					if err != nil {
						errMsg := err.Error()
						errStr = &errMsg
					}
					e.audit.Log(ctx, &AuditEntry{
						ID:         uuid.New().String(),
						UserID:     session.UserID,
						SessionID:  session.ID,
						RequestID:  session.ID,
						ParentID:   auditParentID,
						AgentName:  agentName,
						ToolName:   toolName,
						ToolInput:  inputBytes,
						ToolOutput: outputBytes,
						Error:      errStr,
						DurationMs: durationMs,
						IsWriteOp:  tool.RequiresConfirmation(),
						Timestamp:  startTime.Unix(),
					})
				}

				if err != nil {
					execution.Error = err.Error()
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						block.ID,
						err.Error(),
						true,
					))
				} else if result != nil && !result.Success {
					execution.Error = result.Error
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						block.ID,
						result.Error,
						true,
					))
				} else {
					if result != nil {
						execution.Result = result.Data
					}
					resultBytes, _ := json.Marshal(result.Data)
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						block.ID,
						string(resultBytes),
						false,
					))
				}

				toolsUsed = append(toolsUsed, execution)
			}

			if confirmationNeeded != nil {
				break
			}
		}

		// Build response blocks for persistence
		responseBlocks := responseToBlocks(resp)

		// If confirmation needed, return for user approval
		if confirmationNeeded != nil {
			session.AddAssistantResponse(resp)

			return &Output{
				Type:           OutputConfirmationNeeded,
				Text:           textResponse,
				PendingAction:  confirmationNeeded,
				ToolsUsed:      toolsUsed,
				ResponseBlocks: responseBlocks,
				TokensUsed:     totalTokens,
			}, nil
		}

		// If no tool calls, we're done
		if len(toolResults) == 0 {
			session.AddAssistantMessage(textResponse)

			if input.StreamCallback != nil {
				input.StreamCallback("", true)
			}

			// Record success with guardrails
			if e.guardrails != nil && input.Context != nil {
				e.guardrails.RecordSuccess(ctx, input.Context.UserID)
			}

			return &Output{
				Type:       OutputComplete,
				Text:       textResponse,
				ToolsUsed:  toolsUsed,
				TokensUsed: totalTokens,
			}, nil
		}

		// Continue loop with tool results
		session.AddAssistantResponse(resp)
		session.AddToolResults(toolResults)
	}
}

// ExecuteTool executes a confirmed write operation.
func (e *Engine) ExecuteTool(ctx context.Context, userID, toolName string, input json.RawMessage, confirmationID string) (*core.ToolResult, error) {
	tool, ok := e.registry.Get(toolName)
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	return tool.Execute(ctx, &core.ToolParams{
		UserID:         userID,
		Input:          input,
		ConfirmationID: confirmationID,
		RequestID:      confirmationID,
	})
}

// createMessageStreaming handles streaming API calls.
func (e *Engine) createMessageStreaming(ctx context.Context, params anthropic.MessageNewParams, callback func(string, bool)) (*anthropic.Message, error) {
	stream := e.client.Messages.NewStreaming(ctx, params)
	defer stream.Close()

	// Accumulate the message from events
	message := anthropic.Message{}

	for stream.Next() {
		event := stream.Current()

		// Accumulate into the message
		if err := message.Accumulate(event); err != nil {
			// Log but continue - accumulation errors are non-fatal
		}

		// Handle different event types
		switch evt := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch delta := evt.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				callback(delta.Text, false)
			}
		case anthropic.MessageStopEvent:
			// Stream complete
		}
	}

	if err := stream.Err(); err != nil {
		return nil, err
	}

	return &message, nil
}

// responseToBlocks converts a Claude response to core.ContentBlock slice.
func responseToBlocks(resp *anthropic.Message) []core.ContentBlock {
	blocks := make([]core.ContentBlock, 0, len(resp.Content))
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			blocks = append(blocks, core.NewTextBlock(block.Text))
		case "tool_use":
			inputBytes, _ := json.Marshal(block.Input)
			blocks = append(blocks, core.NewToolUseBlock(block.ID, block.Name, inputBytes))
		}
	}
	return blocks
}

// RunAgent executes an Agent using the engine.
// This method uses the agent's Capabilities to configure the execution.
func (e *Engine) RunAgent(ctx context.Context, agent core.Agent, input *core.Input) (*core.Output, error) {
	caps := agent.Capabilities()

	// Build engine input from core input and agent capabilities
	engineInput := &Input{
		UserMessage:    input.UserMessage,
		Context:        input.Context,
		History:        input.History,
		SystemPrompt:   caps.SystemPrompt,
		Model:          caps.Model,
		MaxTokens:      caps.MaxTokens,
		AgentName:      agent.Name(),
		AvailableTools: caps.AvailableTools,
	}

	// Override context limits with agent capabilities if not already set
	if engineInput.Context != nil && engineInput.Context.Limits == nil {
		engineInput.Context.Limits = &core.ExecutionLimits{
			MaxTurns:   caps.MaxTurns,
			MaxTokens:  caps.MaxTokens,
			CanConfirm: caps.CanRequestConfirmation,
		}
	}

	// Set stream callback if provided
	if input.StreamCallback != nil {
		engineInput.StreamCallback = input.StreamCallback
	}

	// Run the engine
	output, err := e.Run(ctx, engineInput)
	if err != nil {
		return nil, err
	}

	// Convert to core output
	return &core.Output{
		Type:           core.OutputType(output.Type),
		Text:           output.Text,
		PendingAction:  output.PendingAction,
		ToolsUsed:      output.ToolsUsed,
		ResponseBlocks: output.ResponseBlocks,
		TokensUsed:     output.TokensUsed,
		Error:          output.Error,
	}, nil
}

// DefaultSystemPrompt is the default system prompt for the agent.
const DefaultSystemPrompt = `You are a helpful financial assistant.

GUIDELINES:
- Be conversational and helpful
- Ask clarifying questions when needed
- Use tools when you have enough information
- All money movements require user confirmation

AVAILABLE ACTIONS:
- Check balances and transactions
- Send money to other users
- Manage savings deposits and withdrawals
- Look up user profiles`
