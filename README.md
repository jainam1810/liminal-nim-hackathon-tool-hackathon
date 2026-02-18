# Nim Go SDK

A Go SDK for building AI-powered financial assistants using Claude.

---

# Checkout our tool for Detecting Suspicious Transaction for Liminal - 
https://drive.google.com/file/d/1KvJ2GQn74eZPzK1ZJr6XbAjoyAwrSWVS/view?usp=sharing

## 💜 For Hackathon Participants

**Building for the Liminal Vibe Banking Hackathon?** Start here:

👉 **[Hackathon Starter Guide](./examples/hackathon-starter/)** — Get up and running in 5 minutes with a complete AI banking agent.

**What you get:**
- ✅ Complete backend with all 9 Liminal banking tools
- ✅ Beautiful React chat interface with nim-chat widget
- ✅ Example custom tool (spending analyzer) to learn from
- ✅ Step-by-step guide for getting API keys via TestFlight
- ✅ Ready-to-customize system prompt and AI personality

**Quick start:**
```bash
cd examples/hackathon-starter
cp .env.example .env
# Add your API keys to .env
go run main.go
```

Then in another terminal:
```bash
cd frontend && npm install && npm run dev
```

See the [Hackathon Starter README](./examples/hackathon-starter/README.md) for complete instructions, project ideas, and tips!

---

## Features

- **Ready-to-run WebSocket server** - Handles all WebSocket/streaming complexity
- **Custom tool support** - Extend the agent with your own tools
- **Liminal integration** - Connect to Liminal's financial APIs
- **Confirmation flow** - Built-in support for write operation approvals
- **Streaming responses** - Real-time text streaming from Claude

## Quick Start

```go
package main

import (
    "log"

    "github.com/becomeliminal/nim-go-sdk/server"
    "github.com/becomeliminal/nim-go-sdk/tools"
)

func main() {
    srv, err := server.New(server.Config{
        AnthropicKey: "sk-ant-...",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Add a custom tool
    srv.AddTool(tools.New("get_weather").
        Description("Get weather for a location").
        Schema(tools.ObjectSchema(map[string]interface{}{
            "location": tools.StringProperty("City name"),
        }, "location")).
        HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
            // Your logic here
            return map[string]interface{}{"temp": "72°F"}, nil
        }).
        Build())

    srv.Run(":8080")
}
```

## Installation

```bash
go get github.com/becomeliminal/nim-go-sdk
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    nim-go-sdk                                   │
│                                                                 │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌──────────────────┐   │
│  │  core/  │  │ engine/ │  │ server/ │  │     tools/       │   │
│  │ Types   │  │ Engine  │  │WebSocket│  │ Tool builders    │   │
│  │ Tool IF │  │ Registry│  │ Handler │  │ Liminal defs     │   │
│  │Executor │  │ Session │  │Streaming│  │ Schema helpers   │   │
│  └─────────┘  └─────────┘  └─────────┘  └──────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Packages

### `core/`

Core types and interfaces:

- `Tool` - Interface for tools
- `ToolExecutor` - Interface for executing Liminal tools
- `Message`, `ContentBlock` - Message types
- `Context`, `ExecutionLimits` - Execution context

### `engine/`

Agent execution engine:

- `Engine` - Runs the agent loop with Claude
- `ToolRegistry` - Manages available tools
- `Session` - Conversation state

### `server/`

WebSocket server:

- `Server` - Ready-to-run WebSocket server
- `Config` - Server configuration
- Protocol types for client/server messages

### `executor/`

ToolExecutor implementations:

- `HTTPExecutor` - Calls Liminal API over HTTP

### `tools/`

Tool building utilities:

- `Builder` - Fluent tool builder
- Schema helpers for JSON Schema
- `LiminalTools()` - Pre-defined Liminal tool definitions

## WebSocket Protocol

### Client Messages

```json
{"type": "new_conversation"}
{"type": "resume_conversation", "conversationId": "..."}
{"type": "message", "content": "What's my balance?"}
{"type": "confirm", "actionId": "..."}
{"type": "cancel", "actionId": "..."}
```

### Server Messages

```json
{"type": "conversation_started", "conversationId": "..."}
{"type": "text_chunk", "content": "Let me check..."}
{"type": "text", "content": "Your balance is $100"}
{"type": "confirm_request", "actionId": "...", "tool": "send_money", "summary": "Send $50 to @alice"}
{"type": "complete", "tokenUsage": {...}}
{"type": "error", "content": "..."}
```

## Creating Custom Tools

### Using Builder

```go
tool := tools.New("my_tool").
    Description("Description for Claude").
    Schema(tools.ObjectSchema(map[string]interface{}{
        "param1": tools.StringProperty("Description"),
        "param2": tools.NumberProperty("Description"),
    }, "param1")).
    HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
        var params struct {
            Param1 string  `json:"param1"`
            Param2 float64 `json:"param2"`
        }
        json.Unmarshal(input, &params)

        return map[string]interface{}{"result": "..."}, nil
    }).
    Build()
```

### Write Operations (Requiring Confirmation)

```go
tool := tools.New("dangerous_action").
    Description("Does something that needs approval").
    RequiresConfirmation().
    SummaryTemplate("Perform action on {{.target}}").
    // ...
    Build()
```

## Using Liminal Tools

To use Liminal's financial tools:

```go
import (
    "github.com/becomeliminal/nim-go-sdk/executor"
    "github.com/becomeliminal/nim-go-sdk/tools"
)

// Create executor
exec := executor.NewHTTPExecutor(executor.HTTPExecutorConfig{
    BaseURL: "https://api.liminal.cash",
    APIKey:  "nim_...",
})

// Add Liminal tools to server
srv.AddTools(tools.LiminalTools(exec)...)
```

Available Liminal tools:
- `get_balance` - Wallet balance
- `get_savings_balance` - Savings positions
- `get_vault_rates` - Savings APY rates
- `get_transactions` - Transaction history
- `get_profile` - User profile
- `search_users` - Find users
- `send_money` - Send payments (confirmation required)
- `deposit_savings` - Deposit to savings (confirmation required)
- `withdraw_savings` - Withdraw from savings (confirmation required)

## Examples

See the `examples/` directory:

- **`hackathon-starter/`** - Complete starter for the Liminal Vibe Banking Hackathon (recommended!)
- `basic/` - Simple server with one custom tool
- `custom-tools/` - Multiple custom tools (task manager)
- `full-agent/` - Full agent with Liminal integration

## Environment Variables

- `ANTHROPIC_API_KEY` - Required. Your Anthropic API key.
- `LIMINAL_BASE_URL` - Optional. Liminal API URL (default: https://api.liminal.cash)

Note: Liminal authentication is automatic via JWT tokens from the login flow. No API key needed.

## License

MIT
