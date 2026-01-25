package tools

import (
	"context"
	"encoding/json"

	"github.com/becomeliminal/nim-go-sdk/core"
)

// ThinkToolName is the name of the think tool.
const ThinkToolName = "think"

// ThinkTool allows the agent to think through problems step by step.
// This tool has no side effects and simply acknowledges the thought.
// It's useful for complex reasoning and planning.
type ThinkTool struct{}

// NewThinkTool creates a new think tool.
func NewThinkTool() *ThinkTool {
	return &ThinkTool{}
}

// Name returns the tool's name.
func (t *ThinkTool) Name() string {
	return ThinkToolName
}

// Description returns the tool's description.
func (t *ThinkTool) Description() string {
	return `Use this tool to think through complex problems step by step.
The thought content is for your internal reasoning and will not be shown to the user.
Use this when you need to:
- Plan a sequence of actions
- Analyze information before responding
- Work through calculations or logic
- Consider multiple approaches`
}

// Schema returns the tool's input schema.
func (t *ThinkTool) Schema() map[string]interface{} {
	return ObjectSchema(map[string]interface{}{
		"thought": StringProperty("Your step-by-step reasoning or analysis"),
	}, "thought")
}

// RequiresConfirmation returns false - thinking never requires confirmation.
func (t *ThinkTool) RequiresConfirmation() bool {
	return false
}

// Execute acknowledges the thought without any side effects.
func (t *ThinkTool) Execute(ctx context.Context, params *core.ToolParams) (*core.ToolResult, error) {
	// Parse the thought (we don't actually use it, but validate it)
	var input struct {
		Thought string `json:"thought"`
	}
	if err := json.Unmarshal(params.Input, &input); err != nil {
		return &core.ToolResult{
			Success: false,
			Error:   "invalid input: thought is required",
		}, nil
	}

	if input.Thought == "" {
		return &core.ToolResult{
			Success: false,
			Error:   "thought cannot be empty",
		}, nil
	}

	// Simply acknowledge the thought
	return &core.ToolResult{
		Success: true,
		Data:    map[string]string{"status": "thought recorded"},
	}, nil
}

// GetSummary returns a summary (not used for think tool).
func (t *ThinkTool) GetSummary(input json.RawMessage) string {
	return "Internal reasoning"
}
