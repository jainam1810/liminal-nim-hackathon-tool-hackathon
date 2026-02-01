package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/becomeliminal/nim-go-sdk/core"
)

// ============================================================================
// CONFIGURATION & CONSTANTS
// ============================================================================

const (
	// Default constraints
	DefaultMinBufferMonths    = 3.0
	DefaultMaxHighRiskPctLow  = 0.10 // Max 10% high risk for Low profile
	DefaultMaxHighRiskPctMed  = 0.40 // Max 40% high risk for Medium profile
	DefaultMaxHighRiskPctHigh = 0.70 // Max 70% high risk for High profile

	// Analysis constants
	HistoryWindowMonths = 6
)

type GoalType string

const (
	GoalRetirement       GoalType = "retirement"
	GoalEmergencyFund    GoalType = "emergency_fund"
	GoalLargePurchase    GoalType = "large_purchase"
	GoalEducation        GoalType = "education"
	GoalGeneralGrowth    GoalType = "general_growth"
	GoalShortTermSavings GoalType = "short_term_savings"
)

type RiskPreference string

const (
	RiskLow    RiskPreference = "low"
	RiskMedium RiskPreference = "medium"
	RiskHigh   RiskPreference = "high"
)

// ============================================================================
// INPUT / OUTPUT STRUCTURES
// ============================================================================

type WalletAdviceInput struct {
	UserID              string          `json:"user_id"`
	NaturalLanguageGoal string          `json:"natural_language_goal"`
	ExplicitRisk        *RiskPreference `json:"explicit_risk,omitempty"`
	TimeHorizonMonths   *int            `json:"time_horizon_months,omitempty"`
	TargetAmount        *float64        `json:"target_amount,omitempty"`
	Profile             *UserProfile    `json:"profile,omitempty"`
	Constraints         *Constraints    `json:"constraints,omitempty"`
}

type UserProfile struct {
	Age        *int   `json:"age,omitempty"`
	Country    string `json:"country,omitempty"`
	Employment string `json:"employment,omitempty"`  // student, employed, etc.
	IncomeBand string `json:"income_band,omitempty"` // low/medium/high
}

type Constraints struct {
	MinEmergencyBufferMonths float64  `json:"min_emergency_buffer_months"`
	MaxHighRiskAllocationPct float64  `json:"max_high_risk_allocation_pct"`
	PreferredCurrencies      []string `json:"preferred_currencies,omitempty"`
}

type WalletAdviceOutput struct {
	Status             Status              `json:"status"`
	Goal               NormalisedGoal      `json:"goal"`
	RiskProfile        ComputedRiskProfile `json:"risk_profile"`
	AccountSnapshot    AccountSnapshot     `json:"account_snapshot"`
	CashFlowAnalysis   CashFlowAnalysis    `json:"cash_flow_analysis"`
	ProductCatalogue   []ProductOption     `json:"product_catalogue"`
	RecommendedPlan    RecommendedPlan     `json:"recommended_plan"`
	MonitoringSignals  []MonitoringSignal  `json:"monitoring_signals"`
	ParametersUsed     ParametersUsed      `json:"parameters_used"`
	NarrativeSummaries NarrativeSummaries  `json:"narrative_summaries"`
}

type Status struct {
	Code    string `json:"code"` // ok, insufficient_data, error
	Message string `json:"message"`
}

type NormalisedGoal struct {
	Type              GoalType `json:"type"`
	TimeHorizonMonths int      `json:"time_horizon_months"`
	TargetAmount      float64  `json:"target_amount,omitempty"`
	Confidence        float64  `json:"confidence"`
	Notes             string   `json:"notes,omitempty"`
}

type ComputedRiskProfile struct {
	Score         int            `json:"score"` // 0-100
	Level         RiskPreference `json:"level"` // low, medium, high
	Justification string         `json:"justification"`
}

type AccountSnapshot struct {
	TotalBalanceUSD  float64        `json:"total_balance_usd"`
	Balances         []BalanceEntry `json:"balances"`
	SavingsPositions []SavingsEntry `json:"savings_positions"`
	LiquidCapital    float64        `json:"liquid_capital"`
	LockedCapital    float64        `json:"locked_capital"`
}

type BalanceEntry struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	USDValue float64 `json:"usd_value"` // Estimated
}

type SavingsEntry struct {
	VaultName string  `json:"vault_name"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	APY       float64 `json:"apy"`
}

type CashFlowAnalysis struct {
	MonthlyIncome    float64 `json:"monthly_income"`
	MonthlyExpenses  float64 `json:"monthly_expenses"`
	FreeCashFlow     float64 `json:"free_cash_flow"`
	VolatilityScore  float64 `json:"volatility_score"` // 0-1 (stable to volatile)
	BufferMonths     float64 `json:"buffer_months"`
	DataWindowMonths int     `json:"data_window_months"`
}

type ProductOption struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	RiskBand    string  `json:"risk_band"` // Low, Medium, High
	ExpectedAPY float64 `json:"expected_apy"`
	Currency    string  `json:"currency"`
	MinDeposit  float64 `json:"min_deposit"`
	LockupDays  int     `json:"lockup_days"`
	Suitability string  `json:"suitability"` // Match score description
}

type RecommendedPlan struct {
	TargetBufferAmount float64            `json:"target_buffer_amount"`
	InvestableCapital  float64            `json:"investable_capital"` // Amount safe to invest
	Allocations        []Allocation       `json:"allocations"`
	Projections        []Projection       `json:"projections"`
	RiskBreakdown      map[string]float64 `json:"risk_breakdown"` // % in low/med/high
}

type Allocation struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Action      string  `json:"action"` // "Move X from Wallet to Vault Y"
}

type Projection struct {
	Months        int     `json:"months"`
	ExpectedValue float64 `json:"expected_value"`
	Scenario      string  `json:"scenario"` // Conservative, Base, Aggressive
}

type MonitoringSignal struct {
	SignalID  string `json:"signal_id"`
	Message   string `json:"message"`
	Threshold string `json:"threshold"`
}

type ParametersUsed struct {
	UserGoal  string  `json:"user_goal"`
	RiskLevel string  `json:"risk_level"`
	MinBuffer float64 `json:"min_buffer_months"`
}

type NarrativeSummaries struct {
	SituationAnalysis string `json:"situation_analysis"`
	PlanStrategy      string `json:"plan_strategy"`
	RiskExplanation   string `json:"risk_explanation"`
}

// ============================================================================
// DATA CLIENT
// ============================================================================

type DataClient struct {
	Executor core.ToolExecutor
	UserID   string
	Ctx      context.Context
}

func (dc *DataClient) GetBalances() ([]map[string]interface{}, error) {
	resp, err := dc.Executor.Execute(dc.Ctx, &core.ExecuteRequest{
		UserID: dc.UserID,
		Tool:   "get_balance",
		Input:  []byte("{}"),
	})
	if err != nil {
		return nil, err
	}
	// Note: get_balance usually returns a single balance object or map.
	// For this tool we assume it returns a structure we can parse into currencies.
	// Hackathon starter get_balance matches what we need.
	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}

	// Normalize to list
	var balances []map[string]interface{}
	// Check if data itself is the balance map or wrapper
	if bal, ok := data["balance"].(float64); ok {
		// Simple single currency case (likely USD from starter)
		balances = append(balances, map[string]interface{}{
			"currency": "USD",
			"amount":   bal,
		})
	}
	return balances, nil
}

func (dc *DataClient) GetTransactions(limit int) ([]map[string]interface{}, error) {
	req := map[string]interface{}{"limit": limit}
	input, _ := json.Marshal(req)
	resp, err := dc.Executor.Execute(dc.Ctx, &core.ExecuteRequest{
		UserID: dc.UserID,
		Tool:   "get_transactions",
		Input:  input,
	})
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, err
	}

	if txs, ok := data["transactions"].([]interface{}); ok {
		var result []map[string]interface{}
		for _, t := range txs {
			if tm, ok := t.(map[string]interface{}); ok {
				result = append(result, tm)
			}
		}
		return result, nil
	}
	return nil, nil
}

func (dc *DataClient) GetVaultRates() ([]map[string]interface{}, error) {
	resp, err := dc.Executor.Execute(dc.Ctx, &core.ExecuteRequest{
		UserID: dc.UserID,
		Tool:   "get_vault_rates",
		Input:  []byte("{}"),
	})
	if err != nil {
		return nil, err
	}

	// Parse list of vaults
	var data map[string]interface{} // assume wrapped result
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		// Could be direct array
		var arr []map[string]interface{}
		if err2 := json.Unmarshal(resp.Data, &arr); err2 == nil {
			return arr, nil
		}
		return nil, err
	}

	if vaults, ok := data["rates"].([]interface{}); ok {
		var result []map[string]interface{}
		for _, v := range vaults {
			if vm, ok := v.(map[string]interface{}); ok {
				result = append(result, vm)
			}
		}
		return result, nil
	}
	return nil, nil
}

// ============================================================================
// LOGIC IMPLEMENTATION
// ============================================================================

// HandleWalletAdvice is the main entry point logic
func HandleWalletAdvice(ctx context.Context, input WalletAdviceInput, dc *DataClient) (WalletAdviceOutput, error) {
	out := WalletAdviceOutput{
		Status: Status{Code: "ok", Message: "Plan generated successfully"},
	}

	// 1. Fetch Data
	balances, err := dc.GetBalances()
	if err != nil {
		out.Status = Status{Code: "error", Message: fmt.Sprintf("Failed to fetch balances: %v", err)}
		return out, nil
	}

	txs, err := dc.GetTransactions(100)
	if err != nil {
		log.Printf("Warning: failed to fetch txs: %v", err)
	}

	vaults, err := dc.GetVaultRates()
	if err != nil {
		log.Printf("Warning: failed to fetch vaults: %v", err)
	}

	// 2. Build Snapshot
	out.AccountSnapshot = buildSnapshot(balances)

	// 3. Analyze Cash Flow (very heuristic for hackathon)
	out.CashFlowAnalysis = analyzeCashFlow(txs, out.AccountSnapshot.TotalBalanceUSD)

	// 4. Goal Assessment
	out.Goal = parseGoal(input.NaturalLanguageGoal, input.TimeHorizonMonths, input.TargetAmount)

	// 5. Risk Profiling
	out.RiskProfile = calculateRiskProfile(input.ExplicitRisk, out.CashFlowAnalysis, input.Profile)

	// 6. Product Options
	out.ProductCatalogue = selectProductOptions(vaults, out.RiskProfile.Level)

	// 7. Plan Generation
	out.RecommendedPlan, out.NarrativeSummaries = generatePlan(out, input.Constraints)

	// 8. Other Metadata
	out.UpdatedParametersUsed(input)

	return out, nil
}

func buildSnapshot(balances []map[string]interface{}) AccountSnapshot {
	snap := AccountSnapshot{
		Balances: []BalanceEntry{},
	}

	total := 0.0
	for _, b := range balances {
		curr := "USD"
		if c, ok := b["currency"].(string); ok {
			curr = c
		}
		amt, _ := b["amount"].(float64)

		// Simple USD estimation (assuming 1:1 for now or hardcoded rates)
		usdVal := amt
		// Simple stub converter
		if curr == "LIL" {
			usdVal = amt * 0.50
		} // Matches our live rate proxy roughly

		snap.Balances = append(snap.Balances, BalanceEntry{
			Currency: curr,
			Amount:   amt,
			USDValue: usdVal,
		})
		total += usdVal
	}
	snap.TotalBalanceUSD = total
	snap.LiquidCapital = total // start assumption is all liquid
	return snap
}

func analyzeCashFlow(txs []map[string]interface{}, currentBalance float64) CashFlowAnalysis {
	if len(txs) == 0 {
		return CashFlowAnalysis{
			MonthlyIncome:   2000, // Fallback/Assumption
			MonthlyExpenses: 1500,
			FreeCashFlow:    500,
			BufferMonths:    currentBalance / 1500,
			VolatilityScore: 0.5,
		}
	}

	var income, expense float64
	// Simple heuristic: 'receive' = income, 'send' = expense
	// Window: assume txs are recent.
	// For better accuracy we'd check timestamps.

	count := 0
	for _, tx := range txs {
		typ, _ := tx["type"].(string)
		amt, _ := tx["amount"].(float64)

		if typ == "receive" {
			income += amt
		} else {
			expense += amt
		}
		count++
	}

	// Basic average assuming the batch represents ~1 month of activity for the demo
	// In real app we'd divide by actual date range
	months := 1.0

	monthlyInc := income / months
	monthlyExp := expense / months

	// Fallbacks if data is too sparse
	if monthlyExp == 0 {
		monthlyExp = 500
	}

	return CashFlowAnalysis{
		MonthlyIncome:    monthlyInc,
		MonthlyExpenses:  monthlyExp,
		FreeCashFlow:     monthlyInc - monthlyExp,
		BufferMonths:     currentBalance / monthlyExp,
		DataWindowMonths: int(months),
		VolatilityScore:  0.2, // Stub
	}
}

func parseGoal(desc string, horizon *int, amount *float64) NormalisedGoal {
	desc = strings.ToLower(desc)
	goal := NormalisedGoal{
		Type:              GoalGeneralGrowth,
		Confidence:        0.8,
		TimeHorizonMonths: 12, // Default
	}

	if amount != nil {
		goal.TargetAmount = *amount
	}
	if horizon != nil {
		goal.TimeHorizonMonths = *horizon
	}

	// Heuristics
	if strings.Contains(desc, "retire") {
		goal.Type = GoalRetirement
		if horizon == nil {
			goal.TimeHorizonMonths = 120
		}
	} else if strings.Contains(desc, "emergency") || strings.Contains(desc, "safety") {
		goal.Type = GoalEmergencyFund
		if horizon == nil {
			goal.TimeHorizonMonths = 6
		}
	} else if strings.Contains(desc, "house") || strings.Contains(desc, "car") || strings.Contains(desc, "buy") {
		goal.Type = GoalLargePurchase
		if horizon == nil {
			goal.TimeHorizonMonths = 36
		}
	} else if strings.Contains(desc, "trip") || strings.Contains(desc, "vacation") {
		goal.Type = GoalShortTermSavings
		if horizon == nil {
			goal.TimeHorizonMonths = 9
		}
	}

	goal.Notes = "Goal inferred from natural language input."
	return goal
}

func calculateRiskProfile(explicit *RiskPreference, cashFlow CashFlowAnalysis, profile *UserProfile) ComputedRiskProfile {
	// 1. Base Score
	score := 50 // Medium

	// 2. Adjust based on buffer
	if cashFlow.BufferMonths < 1 {
		score -= 20
	} else if cashFlow.BufferMonths > 6 {
		score += 10
	}

	// 3. Adjust based on age
	if profile != nil && profile.Age != nil {
		age := *profile.Age
		if age < 30 {
			score += 10
		} else if age > 60 {
			score -= 10
		}
	}

	// 4. Override with explicit
	if explicit != nil {
		switch *explicit {
		case RiskLow:
			return ComputedRiskProfile{Score: 20, Level: RiskLow, Justification: "User explicitly requested low risk."}
		case RiskHigh:
			return ComputedRiskProfile{Score: 80, Level: RiskHigh, Justification: "User explicitly requested high risk."}
		default:
			score = 50
		}
	}

	// Map final score
	level := RiskMedium
	just := "Balanced profile based on healthy cash buffers and income."

	if score < 35 {
		level = RiskLow
		just = "Conservative profile recommended due to limited cash reserves."
	} else if score > 70 {
		level = RiskHigh
		just = "Aggressive profile suitable due to strong cash reserves and timeline."
	}

	return ComputedRiskProfile{
		Score:         score,
		Level:         level,
		Justification: just,
	}
}

func selectProductOptions(vaults []map[string]interface{}, riskLevel RiskPreference) []ProductOption {
	var options []ProductOption

	// Process real vaults if available
	for i, v := range vaults {
		name, _ := v["name"].(string)
		if name == "" {
			name = fmt.Sprintf("Vault #%d", i+1)
		}

		apy, _ := v["apy"].(float64)
		curr, _ := v["currency"].(string)
		if curr == "" {
			curr = "USD"
		}

		// Heuristic risk banding based on APY
		band := "Low"
		if apy > 4.0 {
			band = "Medium"
		}
		if apy > 7.0 {
			band = "High"
		}

		options = append(options, ProductOption{
			ID:          fmt.Sprintf("real_vault_%d", i),
			Name:        name,
			RiskBand:    band,
			ExpectedAPY: apy,
			Currency:    curr,
			Suitability: fmt.Sprintf("Real %s vault with %.1f%% APY", curr, apy),
		})
	}

	// Add some built-in options if vaults are empty or for demo
	if len(options) == 0 {
		options = append(options, ProductOption{
			ID: "vault_flex_save", Name: "Flex Savings", RiskBand: "Low", ExpectedAPY: 3.5, Suitability: "High match for emergency funds",
		})
		options = append(options, ProductOption{
			ID: "vault_growth_plus", Name: "Growth Plus Vault", RiskBand: "Medium", ExpectedAPY: 5.2, LockupDays: 30, Suitability: "Good for general growth",
		})
		options = append(options, ProductOption{
			ID: "vault_defi_yield", Name: "DeFi Yield Protocol", RiskBand: "High", ExpectedAPY: 8.5, LockupDays: 90, Suitability: "Only for risk-tolerant portion",
		})
	}

	// Filter based on risk level
	var filtered []ProductOption
	for _, opt := range options {
		// Low profile -> Low risk products only
		if riskLevel == RiskLow && opt.RiskBand != "Low" {
			continue
		}
		// Medium -> Low & Medium
		if riskLevel == RiskMedium && opt.RiskBand == "High" {
			continue
		}
		filtered = append(filtered, opt)
	}
	return filtered
}

func generatePlan(context WalletAdviceOutput, constraints *Constraints) (RecommendedPlan, NarrativeSummaries) {
	plan := RecommendedPlan{
		RiskBreakdown: make(map[string]float64),
	}
	narrative := NarrativeSummaries{}

	// Constraints
	minBufferMonths := DefaultMinBufferMonths
	if constraints != nil && constraints.MinEmergencyBufferMonths > 0 {
		minBufferMonths = constraints.MinEmergencyBufferMonths
	}

	// 1. Calculate Emergency Buffer
	targetBuffer := context.CashFlowAnalysis.MonthlyExpenses * minBufferMonths
	plan.TargetBufferAmount = targetBuffer

	available := context.AccountSnapshot.LiquidCapital

	// 2. Allocate
	if available < targetBuffer {
		// Logic: Fill buffer first
		plan.InvestableCapital = 0
		plan.Allocations = append(plan.Allocations, Allocation{
			Action:      fmt.Sprintf("Keep all funds ($%.2f) in liquid basic savings", available),
			Amount:      available,
			ProductName: "Liquid Wallet / Basic Savings",
		})

		missing := targetBuffer - available
		narrative.SituationAnalysis = fmt.Sprintf("You currently have %.1f months of expenses covered. Recommended buffer is %.1f months.", context.CashFlowAnalysis.BufferMonths, minBufferMonths)
		narrative.PlanStrategy = fmt.Sprintf("Priority 1 is building your emergency fund. We need to save another $%.0f.", missing)
		narrative.RiskExplanation = "We are keeping everything low-risk and liquid until your safety net is established."

	} else {
		investable := available - targetBuffer
		plan.InvestableCapital = investable

		// Split investable based on risk profile
		// Simple Demo Logic: Put 100% of investable into best matching vault
		bestProduct := context.ProductCatalogue[len(context.ProductCatalogue)-1] // Pick highest yield available

		plan.Allocations = append(plan.Allocations, Allocation{
			Action:      fmt.Sprintf("Keep $%.2f as Emergency Buffer", targetBuffer),
			Amount:      targetBuffer,
			ProductName: "Liquid Wallet",
		})

		plan.Allocations = append(plan.Allocations, Allocation{
			Action:      fmt.Sprintf("Move $%.2f to %s for %.1f%% APY", investable, bestProduct.Name, bestProduct.ExpectedAPY),
			Amount:      investable,
			ProductName: bestProduct.Name,
			ProductID:   bestProduct.ID,
		})

		// Projection
		plan.Projections = []Projection{
			{Months: 12, Scenario: "Base", ExpectedValue: investable * (1 + bestProduct.ExpectedAPY/100)},
		}

		narrative.SituationAnalysis = "You have a fully funded emergency buffer, which is great!"
		narrative.PlanStrategy = fmt.Sprintf("We can perform a 'Risk-On' trade with your excess $%.0f capital to target higher yields.", investable)
		narrative.RiskExplanation = fmt.Sprintf("We selected %s as it matches your %s risk profile.", bestProduct.Name, context.RiskProfile.Level)
	}

	return plan, narrative
}

func (out *WalletAdviceOutput) UpdatedParametersUsed(input WalletAdviceInput) {
	out.ParametersUsed = ParametersUsed{
		UserGoal:  string(out.Goal.Type) + " (" + out.Goal.Notes + ")",
		RiskLevel: string(out.RiskProfile.Level),
		MinBuffer: 3.0, // Default
	}
	if input.Constraints != nil {
		out.ParametersUsed.MinBuffer = input.Constraints.MinEmergencyBufferMonths
	}
}

// ============================================================================
// TOOL CONSTRUCTION
// ============================================================================
