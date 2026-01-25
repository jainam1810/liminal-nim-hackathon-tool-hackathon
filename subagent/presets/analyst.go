// Package presets provides pre-configured sub-agents for common tasks.
package presets

import (
	"github.com/becomeliminal/nim-go-sdk/engine"
	"github.com/becomeliminal/nim-go-sdk/subagent"
)

// AnalystSystemPrompt is the system prompt for the spending analyst sub-agent.
const AnalystSystemPrompt = `You are a financial analyst specialist.

Your role is to analyze the user's financial data and provide insights.
Focus on:
- Spending patterns and trends
- Budget analysis
- Transaction categorization
- Financial health indicators

Guidelines:
- Be concise and data-driven
- Highlight key findings first
- Provide actionable insights when relevant
- Never make transfers or modify data - you are read-only

Available tools: get_transactions, get_savings_balance`

// NewAnalyst creates a spending analyst sub-agent.
// This agent can analyze transactions and savings to provide financial insights.
func NewAnalyst(eng *engine.Engine) *subagent.SubAgent {
	return subagent.NewSubAgent(eng, subagent.SubAgentConfig{
		Name:         "analyst",
		SystemPrompt: AnalystSystemPrompt,
		AvailableTools: []string{
			"get_transactions",
			"get_savings_balance",
		},
		MaxTurns:  5,
		MaxTokens: 1024,
	})
}

// NewAnalystDelegationTool creates a delegation tool for the analyst.
func NewAnalystDelegationTool(eng *engine.Engine) *subagent.DelegationTool {
	return subagent.NewDelegationTool(subagent.DelegationConfig{
		SubAgent:    NewAnalyst(eng),
		ToolName:    "analyze_spending",
		Description: "Delegate financial analysis to the analyst specialist. Use this for spending analysis, transaction summaries, and budget insights.",
		TaskFormatter: func(query string) string {
			return "Analyze the following and provide insights: " + query
		},
		QueryDescription: "The financial analysis question or request (e.g., 'What did I spend on food this month?')",
	})
}
