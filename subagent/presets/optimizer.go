package presets

import (
	"github.com/becomeliminal/nim-go-sdk/engine"
	"github.com/becomeliminal/nim-go-sdk/subagent"
)

// OptimizerSystemPrompt is the system prompt for the yield optimizer sub-agent.
const OptimizerSystemPrompt = `You are a yield optimization specialist.

Your role is to analyze savings options and provide yield optimization advice.
Focus on:
- Current savings positions and rates
- Available vault options and APYs
- Yield optimization opportunities
- Risk-adjusted return analysis

Guidelines:
- Present clear comparisons of options
- Consider current positions before suggesting changes
- Highlight both potential gains and risks
- Never execute deposits or withdrawals - only advise

Available tools: get_savings_balance, get_vault_rates`

// NewOptimizer creates a yield optimizer sub-agent.
// This agent can analyze savings positions and vault rates to suggest optimizations.
func NewOptimizer(eng *engine.Engine) *subagent.SubAgent {
	return subagent.NewSubAgent(eng, subagent.SubAgentConfig{
		Name:         "optimizer",
		SystemPrompt: OptimizerSystemPrompt,
		AvailableTools: []string{
			"get_savings_balance",
			"get_vault_rates",
		},
		MaxTurns:  5,
		MaxTokens: 1024,
	})
}

// NewOptimizerDelegationTool creates a delegation tool for the optimizer.
func NewOptimizerDelegationTool(eng *engine.Engine) *subagent.DelegationTool {
	return subagent.NewDelegationTool(subagent.DelegationConfig{
		SubAgent:    NewOptimizer(eng),
		ToolName:    "optimize_yield",
		Description: "Delegate yield optimization analysis to the optimizer specialist. Use this for savings advice and APY comparisons.",
		TaskFormatter: func(query string) string {
			return "Analyze and provide yield optimization advice for: " + query
		},
		QueryDescription: "The savings optimization question (e.g., 'How can I get better yields on my USDC?')",
	})
}
