package subagent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// DelegationTool wraps a sub-agent as a tool that can be called by the parent agent.
type DelegationTool struct {
	subagent      *SubAgent
	taskFormatter func(query string) string
	definition    core.ToolDefinition
}

// DelegationConfig configures a delegation tool.
type DelegationConfig struct {
	// SubAgent is the sub-agent to delegate to.
	SubAgent *SubAgent

	// ToolName overrides the tool name. Defaults to "delegate_to_<agent_name>".
	ToolName string

	// Description overrides the tool description.
	Description string

	// TaskFormatter formats the query into a task for the sub-agent.
	// If nil, the query is passed directly.
	TaskFormatter func(query string) string

	// QueryDescription describes what the query parameter should contain.
	QueryDescription string
}

// NewDelegationTool creates a tool that delegates to a sub-agent.
func NewDelegationTool(cfg DelegationConfig) *DelegationTool {
	toolName := cfg.ToolName
	if toolName == "" {
		toolName = fmt.Sprintf("delegate_to_%s", cfg.SubAgent.Name())
	}

	description := cfg.Description
	if description == "" {
		description = fmt.Sprintf("Delegate a task to the %s specialist agent.", cfg.SubAgent.Name())
	}

	queryDesc := cfg.QueryDescription
	if queryDesc == "" {
		queryDesc = "The task or question to delegate to the specialist agent."
	}

	return &DelegationTool{
		subagent:      cfg.SubAgent,
		taskFormatter: cfg.TaskFormatter,
		definition: core.ToolDefinition{
			ToolName:        toolName,
			ToolDescription: description,
			InputSchema: tools.ObjectSchema(map[string]interface{}{
				"query": tools.StringProperty(queryDesc),
			}, "query"),
		},
	}
}

// Name returns the tool's name.
func (d *DelegationTool) Name() string {
	return d.definition.ToolName
}

// Description returns the tool's description.
func (d *DelegationTool) Description() string {
	return d.definition.ToolDescription
}

// Schema returns the tool's input schema.
func (d *DelegationTool) Schema() map[string]interface{} {
	return d.definition.InputSchema
}

// RequiresConfirmation returns false - delegation doesn't require confirmation.
func (d *DelegationTool) RequiresConfirmation() bool {
	return false
}

// Execute runs the sub-agent with the given query.
func (d *DelegationTool) Execute(ctx context.Context, params *core.ToolParams) (*core.ToolResult, error) {
	// Parse input
	var input struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(params.Input, &input); err != nil {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("invalid input: %v", err),
		}, nil
	}

	if input.Query == "" {
		return &core.ToolResult{
			Success: false,
			Error:   "query is required",
		}, nil
	}

	// Format task
	task := input.Query
	if d.taskFormatter != nil {
		task = d.taskFormatter(input.Query)
	}

	// Create a minimal context for the sub-agent
	// The sub-agent context should be created from the parent context
	// but we only have userID here, so we create a basic one
	subCtx := &core.Context{
		UserID:    params.UserID,
		RequestID: params.RequestID,
		Limits:    core.SubAgentLimits(),
	}

	// Run sub-agent
	output, err := d.subagent.Run(ctx, &core.Input{
		UserMessage: task,
		Context:     subCtx,
	})
	if err != nil {
		return &core.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("sub-agent error: %v", err),
		}, nil
	}

	// Convert output to result
	result := ToResult(d.subagent.Name(), output)

	if !result.Success {
		return &core.ToolResult{
			Success: false,
			Error:   result.Error,
		}, nil
	}

	return &core.ToolResult{
		Success: true,
		Data:    result.Response,
		Metadata: map[string]interface{}{
			"agent":       result.AgentName,
			"tools_used":  len(result.ToolsUsed),
			"tokens_used": result.TokensUsed.TotalTokens(),
		},
	}, nil
}

// GetSummary returns a summary of the delegation.
func (d *DelegationTool) GetSummary(input json.RawMessage) string {
	return fmt.Sprintf("Delegate to %s specialist", d.subagent.Name())
}

// DelegationToolFromAgent creates a delegation tool with sensible defaults.
func DelegationToolFromAgent(agent *SubAgent) core.Tool {
	return NewDelegationTool(DelegationConfig{
		SubAgent: agent,
	})
}
