package presets

import (
	"github.com/becomeliminal/nim-go-sdk/engine"
	"github.com/becomeliminal/nim-go-sdk/subagent"
)

// ResearcherSystemPrompt is the system prompt for the researcher sub-agent.
const ResearcherSystemPrompt = `You are a user research specialist.

Your role is to help find and verify user information before transfers.
Focus on:
- Finding users by display tag or name
- Verifying user identity
- Providing profile information
- Confirming recipient details

Guidelines:
- Be thorough in verification
- Clearly present user information found
- Warn if multiple potential matches exist
- Never make transfers - only research

Available tools: search_users, get_profile`

// NewResearcher creates a user researcher sub-agent.
// This agent can search for users and retrieve profile information.
func NewResearcher(eng *engine.Engine) *subagent.SubAgent {
	return subagent.NewSubAgent(eng, subagent.SubAgentConfig{
		Name:         "researcher",
		SystemPrompt: ResearcherSystemPrompt,
		AvailableTools: []string{
			"search_users",
			"get_profile",
		},
		MaxTurns:  5,
		MaxTokens: 1024,
	})
}

// NewResearcherDelegationTool creates a delegation tool for the researcher.
func NewResearcherDelegationTool(eng *engine.Engine) *subagent.DelegationTool {
	return subagent.NewDelegationTool(subagent.DelegationConfig{
		SubAgent:    NewResearcher(eng),
		ToolName:    "research_recipient",
		Description: "Delegate recipient research to the researcher specialist. Use this to find and verify users before sending money.",
		TaskFormatter: func(query string) string {
			return "Find and verify this recipient: " + query
		},
		QueryDescription: "The user to search for (display tag like @alice, name, or description)",
	})
}
