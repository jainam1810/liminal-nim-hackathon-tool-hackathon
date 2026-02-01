// Hackathon Starter: Complete AI Financial Agent
// Build intelligent financial tools with nim-go-sdk + Liminal banking APIs
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/examples/hackathon-starter/blockchain" // Imported
	"github.com/becomeliminal/nim-go-sdk/executor"
	"github.com/becomeliminal/nim-go-sdk/server"
	"github.com/becomeliminal/nim-go-sdk/tools"
	"github.com/joho/godotenv"
)

var GlobalChain *blockchain.Blockchain

func main() {
	// ============================================================================
	// CONFIGURATION
	// ============================================================================
	// Load .env file if it exists (optional - will use system env vars if not found)
	// Load .env file if it exists (optional - will use system env vars if not found)
	_ = godotenv.Load()

	// Initialize Blockchain
	if chain, err := blockchain.LoadFromFile("blockchain.json"); err == nil {
		GlobalChain = chain
		log.Printf("🔗 Blockchain loaded from file (Height: %d)", len(GlobalChain.Chain))
	} else {
		GlobalChain = blockchain.NewBlockchain()
		log.Println("🔗 Blockchain initialized with Genesis block (New)")
	}

	// Load configuration from environment variables
	// Create a .env file or export these in your shell

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Fatal("❌ ANTHROPIC_API_KEY environment variable is required")
	}

	liminalBaseURL := os.Getenv("LIMINAL_BASE_URL")
	if liminalBaseURL == "" {
		liminalBaseURL = "https://api.liminal.cash"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// ============================================================================
	// LIMINAL EXECUTOR SETUP
	// ============================================================================
	// The HTTPExecutor handles all API calls to Liminal banking services.
	// Authentication is handled automatically via JWT tokens passed from the
	// frontend login flow (email/OTP). No API key needed!

	liminalExecutor := executor.NewHTTPExecutor(executor.HTTPExecutorConfig{
		BaseURL: liminalBaseURL,
	})
	log.Println("✅ Liminal API configured")

	// ============================================================================
	// SERVER SETUP
	// ============================================================================
	// Create the nim-go-sdk server with Claude AI
	// The server handles WebSocket connections and manages conversations
	// Authentication is automatic: JWT tokens from the login flow are extracted
	// from WebSocket connections and forwarded to Liminal API calls

	srv, err := server.New(server.Config{
		AnthropicKey:    anthropicKey,
		SystemPrompt:    hackathonSystemPrompt,
		Model:           "claude-sonnet-4-20250514",
		MaxTokens:       4096,
		LiminalExecutor: liminalExecutor, // SDK automatically handles JWT extraction and forwarding
	})
	if err != nil {
		log.Fatal(err)
	}

	// ============================================================================
	// ADD LIMINAL BANKING TOOLS
	// ============================================================================
	// These are the 9 core Liminal tools that give your AI access to real banking:
	//
	// READ OPERATIONS (no confirmation needed):
	//   1. get_balance - Check wallet balance
	//   2. get_savings_balance - Check savings positions and APY
	//   3. get_vault_rates - Get current savings rates
	//   4. get_transactions - View transaction history
	//   5. get_profile - Get user profile info
	//   6. search_users - Find users by display tag
	//
	// WRITE OPERATIONS (require user confirmation):
	//   7. send_money - Send money to another user
	//   8. deposit_savings - Deposit funds into savings
	//   9. withdraw_savings - Withdraw funds from savings

	srv.AddTools(tools.LiminalTools(liminalExecutor)...)
	log.Println("✅ Added 9 Liminal banking tools")

	// ============================================================================
	// ADD CUSTOM TOOLS
	// ============================================================================
	// This is where you'll add your hackathon project's custom tools!
	// Below is an example spending analyzer tool to get you started.

	srv.AddTool(createSpendingAnalyzerTool(liminalExecutor))
	log.Println("✅ Added custom spending analyzer tool")

	// Register anomaly detector tool
	srv.AddTool(createAnomalyDetectorTool())
	log.Println("✅ Added anomaly detector tool")

	// Register Blockchain View tool
	srv.AddTool(createViewBlockchainTool())
	log.Println("✅ Added blockchain view tool")

	// Register AI Wallet Advice Planner
	// srv.AddTool(createWalletAdviceTool(liminalExecutor))
	// log.Println("✅ Added wallet advice planner tool")

	// TODO: Add more custom tools here!
	// Examples:
	//   - Savings goal tracker
	//   - Spending category analyzer
	//   - Bill payment predictor
	//   - Cash flow forecaster

	// ============================================================================
	// START SERVER
	// ============================================================================

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🚀 Hackathon Starter Server Running")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("📡 WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("Ready for connections! Start your frontend with: cd frontend && npm run dev")
	log.Println()

	if err := srv.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

// ============================================================================
// SYSTEM PROMPT
// ============================================================================
// This prompt defines your AI agent's personality and behavior
// Customize this to match your hackathon project's focus!

const hackathonSystemPrompt = `You are Nim, a friendly AI financial assistant built for the Liminal Vibe Banking Hackathon.

WHAT YOU DO:
You help users manage their money using Liminal's stablecoin banking platform. You can check balances, review transactions, send money, and manage savings - all through natural conversation.

CONVERSATIONAL STYLE:
- Be warm, friendly, and conversational - not robotic
- Use casual language when appropriate, but stay professional about money
- Ask clarifying questions when something is unclear
- Remember context from earlier in the conversation
- Explain things simply without being condescending

WHEN TO USE TOOLS:
- Use tools immediately for simple queries ("what's my balance?")
- For actions, gather all required info first ("send $50 to @alice")
- Always confirm before executing money movements
- Don't use tools for general questions about how things work

MONEY MOVEMENT RULES (IMPORTANT):
- ALL money movements require explicit user confirmation
- Show a clear summary before confirming:
  * send_money: "Send $50 USD to @alice"
  * deposit_savings: "Deposit $100 USD into savings"
  * withdraw_savings: "Withdraw $50 USD from savings"
- Never assume amounts or recipients
- Always use the exact currency the user specified

AVAILABLE BANKING TOOLS:
- Check wallet balance (get_balance)
- Check savings balance and APY (get_savings_balance)
- View savings rates (get_vault_rates)
- View transaction history (get_transactions)
- Get profile info (get_profile)
- Search for users (search_users)
- Send money (send_money) - requires confirmation
- Deposit to savings (deposit_savings) - requires confirmation
- Withdraw from savings (withdraw_savings) - requires confirmation

CUSTOM ANALYTICAL TOOLS:
- Analyze spending patterns (analyze_spending)
- Detect suspicious transactions (detect_lil_anomalies) - Uses LIVE CoinGecko market data
- Get personalized financial advice (wallet_advice_planner) - Creates a risk-adjusted savings/investment plan



TIPS FOR GREAT INTERACTIONS:
- Proactively suggest relevant actions ("Want me to move some to savings?")
- Explain the "why" behind suggestions
- Celebrate financial wins ("Nice! Your savings earned $5 this month!")
- Be encouraging about savings goals
- Make finance feel less intimidating

Remember: You're here to make banking delightful and help users build better financial habits!`

// ============================================================================
// CUSTOM TOOL: SPENDING ANALYZER
// ============================================================================
// This is an example custom tool that demonstrates how to:
// 1. Define tool parameters with JSON schema
// 2. Call other Liminal tools from within your tool
// 3. Process and analyze the data
// 4. Return useful insights
//
// Use this as a template for your own hackathon tools!

func createSpendingAnalyzerTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("analyze_spending").
		Description("Analyze the user's spending patterns over a specified time period. Returns insights about spending velocity, categories, and trends.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"days": tools.IntegerProperty("Number of days to analyze (default: 30)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
			var params struct {
				Days int `json:"days"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Default to 30 days if not specified
			if params.Days == 0 {
				params.Days = 30
			}

			// STEP 1: Fetch transaction history
			// We'll call the Liminal get_transactions tool through the executor
			txRequest := map[string]interface{}{
				"limit": 100, // Get up to 100 transactions
			}
			txRequestJSON, _ := json.Marshal(txRequest)

			txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_transactions",
				Input:     txRequestJSON,
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch transactions: %v", err),
				}, nil
			}

			if !txResponse.Success {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("transaction fetch failed: %s", txResponse.Error),
				}, nil
			}

			// STEP 2: Parse transaction data
			// In a real implementation, you'd parse the actual response structure
			// For now, we'll create a structured analysis

			var transactions []map[string]interface{}
			var txData map[string]interface{}
			if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
				if txArray, ok := txData["transactions"].([]interface{}); ok {
					for _, tx := range txArray {
						if txMap, ok := tx.(map[string]interface{}); ok {
							transactions = append(transactions, txMap)
						}
					}
				}
			}

			// STEP 3: Analyze the data
			analysis := analyzeTransactions(transactions, params.Days)

			// STEP 4: Return insights
			result := map[string]interface{}{
				"period_days":        params.Days,
				"total_transactions": len(transactions),
				"analysis":           analysis,
				"generated_at":       time.Now().Format(time.RFC3339),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// analyzeTransactions processes transaction data and returns insights
func analyzeTransactions(transactions []map[string]interface{}, days int) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{
			"summary": "No transactions found in the specified period",
		}
	}

	// Calculate basic metrics
	var totalSpent, totalReceived float64
	var spendCount, receiveCount int

	// This is a simplified example - you'd do real analysis here:
	// - Group by category/merchant
	// - Calculate daily/weekly averages
	// - Identify spending spikes
	// - Compare to previous periods
	// - Detect recurring payments

	for _, tx := range transactions {
		// Example analysis logic
		txType, _ := tx["type"].(string)
		amount, _ := tx["amount"].(float64)

		switch txType {
		case "send":
			totalSpent += amount
			spendCount++
		case "receive":
			totalReceived += amount
			receiveCount++
		}
	}

	avgDailySpend := totalSpent / float64(days)

	return map[string]interface{}{
		"total_spent":     fmt.Sprintf("%.2f", totalSpent),
		"total_received":  fmt.Sprintf("%.2f", totalReceived),
		"spend_count":     spendCount,
		"receive_count":   receiveCount,
		"avg_daily_spend": fmt.Sprintf("%.2f", avgDailySpend),
		"velocity":        calculateVelocity(spendCount, days),
		"insights": []string{
			fmt.Sprintf("You made %d spending transactions over %d days", spendCount, days),
			fmt.Sprintf("Average daily spend: $%.2f", avgDailySpend),
			"Consider setting up savings goals to build financial cushion",
		},
	}
}

// calculateVelocity determines spending frequency
func calculateVelocity(transactionCount, days int) string {
	txPerWeek := float64(transactionCount) / float64(days) * 7

	switch {
	case txPerWeek < 2:
		return "low"
	case txPerWeek < 7:
		return "moderate"
	default:
		return "high"
	}
}

// ============================================================================
// HACKATHON IDEAS
// ============================================================================
// Here are some ideas for custom tools you could build:
//
// 1. SAVINGS GOAL TRACKER
//   - Track progress toward savings goals
//   - Calculate how long until goal is reached
//   - Suggest optimal deposit amounts
//
// 2. BUDGET ANALYZER
//   - Set spending limits by category
//   - Alert when approaching limits
//   - Compare actual vs. planned spending
//
// 3. RECURRING PAYMENT DETECTOR
//   - Identify subscription payments
//   - Warn about upcoming bills
//   - Suggest savings opportunities
//
// 4. CASH FLOW FORECASTER
//   - Predict future balance based on patterns
//   - Identify potential low balance periods
//   - Suggest when to save vs. spend
//
// 5. SMART SAVINGS ADVISOR
//   - Analyze spare cash available
//   - Recommend savings deposits
//   - Calculate interest projections
//
// 6. SPENDING INSIGHTS
//   - Categorize spending automatically
//   - Compare to typical user patterns
//   - Highlight unusual activity
//
// 7. FINANCIAL HEALTH SCORE
//   - Calculate overall financial wellness
//   - Track improvements over time
//   - Provide actionable recommendations
//
// 8. PEER COMPARISON (anonymous)
//   - Compare savings rate to anonymized peers
//   - Show percentile rankings
//   - Motivate better habits
//
// 9. TAX ESTIMATION
//   - Track potential tax obligations
//   - Suggest amounts to set aside
//   - Generate tax reports
//
// 10. EMERGENCY FUND BUILDER
//   - Calculate needed emergency fund size
//   - Track progress toward goal
//   - Suggest automated savings plan
//
// ============================================================================
// createAnomalyDetectorTool builds the "detect_lil_anomalies" tool.
func createAnomalyDetectorTool() core.Tool {
	return tools.New("detect_lil_anomalies").
		Description("Simulate LIL token transactions, convert them to fiat using **LIVE** crypto market rates (pegged to Cardano/ADA via CoinGecko), and detect suspicious or anomalous payments.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"fiat_currency":           tools.StringProperty("Fiat currency code to convert LIL into (e.g. 'USD', 'GBP'). Default: 'USD'."),
			"dummy_transaction_count": tools.IntegerProperty("How many dummy transactions to simulate (default: 30)."),
			"large_amount_multiplier": tools.NumberProperty("How many times above typical spend counts as 'large' (default: 5.0)."),
			"burst_window_minutes":    tools.IntegerProperty("Time window in minutes for burst detection (default: 60)."),
			"burst_min_count":         tools.IntegerProperty("Minimum small transactions in the window to flag a burst (default: 8)."),
			"small_amount_threshold":  tools.NumberProperty("Max LIL amount to be treated as 'small' in burst rule (default: 0.5)."),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// 1) Parse input
			var params struct {
				FiatCurrency          string  `json:"fiat_currency"`
				DummyCount            int     `json:"dummy_transaction_count"`
				LargeAmountMultiplier float64 `json:"large_amount_multiplier"`
				BurstWindowMinutes    int     `json:"burst_window_minutes"`
				BurstMinCount         int     `json:"burst_min_count"`
				SmallAmountThreshold  float64 `json:"small_amount_threshold"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// 2) Apply defaults
			if params.FiatCurrency == "" {
				params.FiatCurrency = "USD"
			}
			if params.DummyCount <= 0 {
				params.DummyCount = 30
			}
			if params.LargeAmountMultiplier <= 0 {
				params.LargeAmountMultiplier = 5.0
			}
			if params.BurstWindowMinutes <= 0 {
				params.BurstWindowMinutes = 60
			}
			if params.BurstMinCount <= 0 {
				params.BurstMinCount = 8
			}
			if params.SmallAmountThreshold <= 0 {
				params.SmallAmountThreshold = 0.5
			}

			// 3) Fetch (fake) LIL -> fiat exchange rate.
			// For hackathon: simple stub, later swap for real HTTP call.
			rate, err := fetchLILRateStub(params.FiatCurrency)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("failed to fetch LIL rate: %v", err),
				}, nil
			}

			// 4) Generate dummy LIL transactions for this user.
			dummyTxs := generateDummyLILTransactions(params.DummyCount)

			// 5) Convert to fiat and build enriched objects.
			enriched := enrichWithFiat(dummyTxs, rate, params.FiatCurrency)

			// 6) Run anomaly detection.
			// 6) Run anomaly detection.
			anomalies, summary := detectAnomaliesOnDummy(enriched, params)

			// 6a) Record in Blockchain
			// We serialize the entire enriched transaction set to a JSON string for the block data
			txDataBytes, _ := json.Marshal(map[string]interface{}{
				"transactions": enriched,
				"anomalies":    anomalies,
				"summary":      summary,
			})
			newBlock := GlobalChain.AddBlock(string(txDataBytes))
			log.Printf("🔗 Blockchain: Added Block #%d with %d transactions (Hash: %s)", newBlock.Index, len(enriched), newBlock.Hash)

			// Save to file
			if err := GlobalChain.SaveToFile("blockchain.json"); err != nil {
				log.Printf("⚠️ Failed to save blockchain: %v", err)
			} else {
				log.Println("💾 Blockchain saved to blockchain.json")
			}

			// 6b) Send Email Notification
			// Hardcoded recipients as per hackathon requirement
			recipients := []string{"jainamvaria1010@gmail.com", "pks850pks8311@gmail.com"}
			emailStatus := sendAnomalyReportEmail(recipients, anomalies, enriched) // Sync call to get status

			// 7) Build result payload.
			result := map[string]interface{}{
				"fiat_currency":      params.FiatCurrency,
				"exchange_rate":      rate,
				"total_transactions": len(enriched),
				"transactions":       enriched,
				"anomalies":          anomalies,
				"summary":            summary,
				"email_status":       emailStatus,
				"block_index":        newBlock.Index,
				"block_hash":         newBlock.Hash,
				"generated_at":       time.Now().Format(time.RFC3339),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// ============================================================================
// BLOCKCHAIN VIEW TOOL
// ============================================================================

func createViewBlockchainTool() core.Tool {
	return tools.New("view_blockchain").
		Description("View the current state of the transaction blockchain, including all blocks and their validity.").
		Schema(tools.ObjectSchema(map[string]interface{}{})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			valid := GlobalChain.IsChainValid()

			// We return a simplified view to avoid blowing up token limits if chain is huge
			// For a hackathon demo, returning everything is fine if small.
			// Let's return the last 5 blocks and the chain metadata.

			chainLen := len(GlobalChain.Chain)
			startIndex := 0
			if chainLen > 5 {
				startIndex = chainLen - 5
			}

			recentBlocks := GlobalChain.Chain[startIndex:]

			result := map[string]interface{}{
				"chain_valid":   valid,
				"total_blocks":  chainLen,
				"difficulty":    GlobalChain.Difficulty,
				"recent_blocks": recentBlocks,
				"last_updated":  time.Now().Format(time.RFC3339),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// LILRateResponse is a simple struct if you later swap to a real HTTP API.
type LILRateResponse struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
}

// RateCache holds the cached exchange rate
var rateCache = struct {
	sync.RWMutex
	Prices      map[string]float64
	LastUpdated time.Time
}{
	Prices: make(map[string]float64),
}

// fetchLILRateStub returns a live rate from CoinGecko (using Cardano as proxy).
func fetchLILRateStub(fiat string) (float64, error) {
	fiat = strings.ToLower(fiat)
	if fiat == "" {
		fiat = "usd"
	}

	// 1. Check cache (TTL 5 minutes)
	rateCache.RLock()
	if time.Since(rateCache.LastUpdated) < 5*time.Minute {
		if p, ok := rateCache.Prices[fiat]; ok {
			rateCache.RUnlock()
			return p, nil
		}
	}
	rateCache.RUnlock()

	// 2. Fetch from CoinGecko (Cardano as LIL proxy)
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=cardano&vs_currencies=%s", fiat)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("⚠️ Failed to fetch live rate: %v. Using fallback.", err)
		return 0.42, nil // Fallback
	}
	defer resp.Body.Close()

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("⚠️ Failed to decode live rate: %v. Using fallback.", err)
		return 0.42, nil
	}

	price := result["cardano"][fiat]
	if price == 0 {
		return 0.42, nil
	}

	// 3. Update cache
	rateCache.Lock()
	if rateCache.Prices == nil {
		rateCache.Prices = make(map[string]float64)
	}
	rateCache.Prices[fiat] = price
	rateCache.LastUpdated = time.Now()
	rateCache.Unlock()

	return price, nil
}

// DummyTx represents a basic LIL transaction (before fiat enrichment).
type DummyTx struct {
	ID        string
	Timestamp time.Time
	AmountLIL float64
	Type      string
	Merchant  string
	Category  string
}

// EnrichedTx includes fiat conversion.
type EnrichedTx struct {
	ID         string  `json:"id"`
	Timestamp  string  `json:"timestamp"`
	AmountLIL  float64 `json:"amount_lil"`
	AmountFiat float64 `json:"amount_fiat"`
	Type       string  `json:"type"`
	Merchant   string  `json:"merchant"`
	Category   string  `json:"category"`
}

// generateDummyLILTransactions simulates a mixture of normal and suspicious LIL spends.
func generateDummyLILTransactions(n int) []DummyTx {
	if n < 10 {
		n = 10
	}

	now := time.Now()
	txs := make([]DummyTx, 0, n)

	merchants := []string{"Coffee Shop", "Grocery Store", "Online Retailer", "Streaming Service", "Food Delivery"}
	categories := []string{"Food", "Groceries", "Shopping", "Subscriptions", "Entertainment"}

	// Reserve indices for special anomalies
	largeIdx1, largeIdx2 := n-2, n-1 // Last two for large amounts
	burstCount := 8
	if burstCount > n-2 {
		burstCount = n - 2 // Leave room for large transactions
	}

	// 1) Generate normal transactions (indices 0 to n-3)
	for i := 0; i < n-2; i++ {
		// spread over last 24h
		offsetMinutes := -rand.Intn(24 * 60)
		ts := now.Add(time.Duration(offsetMinutes) * time.Minute)

		amount := 0.1 + rand.Float64()*2.0 // 0.1 to ~2.1 LIL typical
		typ := "send"

		m := merchants[rand.Intn(len(merchants))]
		c := categories[rand.Intn(len(categories))]

		txs = append(txs, DummyTx{
			ID:        fmt.Sprintf("tx_%d", i+1),
			Timestamp: ts,
			AmountLIL: amount,
			Type:      typ,
			Merchant:  m,
			Category:  c,
		})
	}

	// 2) Add two very large spends (these won't be overwritten)
	txs = append(txs, DummyTx{
		ID:        fmt.Sprintf("tx_%d", largeIdx1+1),
		Timestamp: now.Add(-time.Duration(rand.Intn(12*60)) * time.Minute),
		AmountLIL: 15.0, // clearly large compared to ~2 LIL typical
		Type:      "send",
		Merchant:  "Luxury Store",
		Category:  "Shopping",
	})
	txs = append(txs, DummyTx{
		ID:        fmt.Sprintf("tx_%d", largeIdx2+1),
		Timestamp: now.Add(-time.Duration(rand.Intn(12*60)) * time.Minute),
		AmountLIL: 20.0,
		Type:      "send",
		Merchant:  "High-End Electronics",
		Category:  "Shopping",
	})

	// 3) Inject a burst of small payments within 1 hour to simulate card testing
	//    Modify only first burstCount normal transactions (indices 0 to burstCount-1)
	base := now.Add(-2 * time.Hour)
	for i := 0; i < burstCount && i < len(txs)-2; i++ {
		txs[i].Timestamp = base.Add(time.Duration(i*5) * time.Minute) // every 5 minutes
		txs[i].AmountLIL = 0.2                                        // all small
		txs[i].Merchant = "Test Merchant"
		txs[i].Category = "Unknown"
	}

	return txs
}

// enrichWithFiat converts dummy txs to EnrichedTx using the given rate.
func enrichWithFiat(dummies []DummyTx, rate float64, _ string) []EnrichedTx {
	out := make([]EnrichedTx, 0, len(dummies))
	for _, d := range dummies {
		out = append(out, EnrichedTx{
			ID:         d.ID,
			Timestamp:  d.Timestamp.Format(time.RFC3339),
			AmountLIL:  d.AmountLIL,
			AmountFiat: d.AmountLIL * rate,
			Type:       d.Type,
			Merchant:   d.Merchant,
			Category:   d.Category,
		})
	}
	return out
}

// Anomaly describes why a given tx was flagged.
type Anomaly struct {
	TransactionID string   `json:"transaction_id"`
	ReasonCodes   []string `json:"reason_codes"`
	Explanations  []string `json:"explanations"`
	RiskScore     float64  `json:"risk_score"`
}

// detectAnomaliesOnDummy runs rules on the enriched txs.
func detectAnomaliesOnDummy(txs []EnrichedTx, params struct {
	FiatCurrency          string  `json:"fiat_currency"`
	DummyCount            int     `json:"dummy_transaction_count"`
	LargeAmountMultiplier float64 `json:"large_amount_multiplier"`
	BurstWindowMinutes    int     `json:"burst_window_minutes"`
	BurstMinCount         int     `json:"burst_min_count"`
	SmallAmountThreshold  float64 `json:"small_amount_threshold"`
}) ([]Anomaly, map[string]interface{}) {
	if len(txs) == 0 {
		return nil, map[string]interface{}{
			"total_flagged":   0,
			"high_risk_count": 0,
			"rules_used":      []string{},
			"note":            "No transactions generated.",
		}
	}

	// 1) Large-amount anomalies
	largeAnoms := detectLargeAmountAnomalies(txs, params.LargeAmountMultiplier)

	// 2) Burst-of-small anomalies
	burstAnoms := detectBurstAnomalies(txs, params.BurstWindowMinutes, params.SmallAmountThreshold, params.BurstMinCount)

	// 3) Merge by transaction id
	merged := mergeAnomalies(largeAnoms, burstAnoms)

	// 4) Compute summary
	total := len(merged)
	highRisk := 0
	for _, a := range merged {
		if a.RiskScore >= 0.8 {
			highRisk++
		}
	}

	summary := map[string]interface{}{
		"total_flagged":   total,
		"high_risk_count": highRisk,
		"rules_used": []string{
			"large_amount_for_user",
			"burst_of_small_transactions",
		},
	}

	// Convert map to slice for deterministic ordering
	out := make([]Anomaly, 0, len(merged))
	for _, a := range merged {
		out = append(out, a)
	}
	return out, summary
}

// detectLargeAmountAnomalies flags txs much larger than typical.
func detectLargeAmountAnomalies(txs []EnrichedTx, multiplier float64) map[string]Anomaly {
	amounts := make([]float64, 0, len(txs))
	for _, tx := range txs {
		amounts = append(amounts, tx.AmountLIL)
	}
	sort.Float64s(amounts)
	median := amounts[len(amounts)/2]

	if median <= 0 {
		return map[string]Anomaly{}
	}

	threshold := multiplier * median
	result := make(map[string]Anomaly)

	for _, tx := range txs {
		if tx.AmountLIL > threshold {
			ratio := tx.AmountLIL / median
			reason := "large_amount_for_user"
			expl := fmt.Sprintf("This payment is %.1fx larger than your typical LIL spend in this sample.", ratio)

			// Dynamic risk score based on how extreme the ratio is
			// Base 0.5, increases with ratio up to 0.95
			riskScore := 0.5 + (ratio-multiplier)/(ratio+10)*0.45
			if riskScore > 0.95 {
				riskScore = 0.95
			}
			if riskScore < 0.5 {
				riskScore = 0.5
			}

			result[tx.ID] = Anomaly{
				TransactionID: tx.ID,
				ReasonCodes:   []string{reason},
				Explanations:  []string{expl},
				RiskScore:     riskScore,
			}
		}
	}

	return result
}

// detectBurstAnomalies flags bursts of small txs in a short window.
func detectBurstAnomalies(txs []EnrichedTx, windowMinutes int, smallThreshold float64, minCount int) map[string]Anomaly {
	if len(txs) == 0 {
		return map[string]Anomaly{}
	}

	// Sort by time
	sorted := make([]EnrichedTx, len(txs))
	copy(sorted, txs)
	sort.Slice(sorted, func(i, j int) bool {
		ti, erri := time.Parse(time.RFC3339, sorted[i].Timestamp)
		tj, errj := time.Parse(time.RFC3339, sorted[j].Timestamp)
		if erri != nil || errj != nil {
			return false // Keep original order if parsing fails
		}
		return ti.Before(tj)
	})

	window := time.Duration(windowMinutes) * time.Minute
	flagged := make(map[string]Anomaly)

	for i := 0; i < len(sorted); i++ {
		startTime, err := time.Parse(time.RFC3339, sorted[i].Timestamp)
		if err != nil {
			continue // Skip transactions with unparseable timestamps
		}
		var smallInWindow []EnrichedTx

		for j := i; j < len(sorted); j++ {
			tj, err := time.Parse(time.RFC3339, sorted[j].Timestamp)
			if err != nil {
				continue
			}
			if tj.Sub(startTime) > window {
				break
			}
			if sorted[j].AmountLIL <= smallThreshold {
				smallInWindow = append(smallInWindow, sorted[j])
			}
		}

		if len(smallInWindow) >= minCount {
			// Dynamic risk score based on how many transactions in the burst
			// More transactions = higher risk
			riskScore := 0.5 + float64(len(smallInWindow)-minCount)/float64(minCount)*0.3
			if riskScore > 0.9 {
				riskScore = 0.9
			}

			for _, tx := range smallInWindow {
				if _, ok := flagged[tx.ID]; ok {
					// Already flagged, skip
					continue
				}
				expl := fmt.Sprintf("%d small payments within %d minutes may indicate testing or automated retries.", len(smallInWindow), windowMinutes)
				flagged[tx.ID] = Anomaly{
					TransactionID: tx.ID,
					ReasonCodes:   []string{"burst_of_small_transactions"},
					Explanations:  []string{expl},
					RiskScore:     riskScore,
				}
			}
			// You can break here if you only want the first burst, but we'll continue scanning.
		}
	}

	return flagged
}

// mergeAnomalies combines anomalies from different rules and adjusts risk.
func mergeAnomalies(a, b map[string]Anomaly) map[string]Anomaly {
	out := make(map[string]Anomaly)

	for id, an := range a {
		out[id] = an
	}
	for id, anB := range b {
		if existing, ok := out[id]; ok {
			// Merge reason codes & explanations
			reasons := append(existing.ReasonCodes, anB.ReasonCodes...)
			expls := append(existing.Explanations, anB.Explanations...)
			// Higher risk if multiple rules hit
			score := existing.RiskScore
			if anB.RiskScore > score {
				score = anB.RiskScore
			}
			if score < 0.8 {
				score = 0.8
			}
			out[id] = Anomaly{
				TransactionID: id,
				ReasonCodes:   reasons,
				Explanations:  expls,
				RiskScore:     score,
			}
		} else {
			out[id] = anB
		}
	}
	return out
}
