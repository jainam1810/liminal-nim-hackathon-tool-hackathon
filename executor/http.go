// Package executor provides ToolExecutor implementations.
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
)

// HTTPExecutor implements ToolExecutor by calling the agent_gateway over HTTP.
// This is the public implementation used by external developers.
type HTTPExecutor struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// HTTPExecutorConfig configures the HTTP executor.
type HTTPExecutorConfig struct {
	// BaseURL is the agent_gateway URL (e.g., "https://api.liminal.cash").
	BaseURL string

	// APIKey is the Liminal API key for authentication.
	APIKey string

	// Timeout is the HTTP request timeout.
	Timeout time.Duration
}

// NewHTTPExecutor creates a new HTTP-based tool executor.
func NewHTTPExecutor(cfg HTTPExecutorConfig) *HTTPExecutor {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &HTTPExecutor{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Execute runs a read-only tool via HTTP.
func (e *HTTPExecutor) Execute(ctx context.Context, req *core.ExecuteRequest) (*core.ExecuteResponse, error) {
	endpoint := e.endpointForTool(req.Tool)
	return e.doRequest(ctx, "POST", endpoint, req)
}

// ExecuteWrite runs a write tool that may require confirmation.
func (e *HTTPExecutor) ExecuteWrite(ctx context.Context, req *core.ExecuteRequest) (*core.ExecuteResponse, error) {
	endpoint := e.endpointForTool(req.Tool)
	return e.doRequest(ctx, "POST", endpoint, req)
}

// Confirm executes a previously confirmed write operation.
func (e *HTTPExecutor) Confirm(ctx context.Context, userID, confirmationID string) (*core.ExecuteResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/agent/confirmations/%s/confirm", confirmationID)
	return e.doRequest(ctx, "POST", endpoint, nil)
}

// Cancel cancels a pending confirmation.
func (e *HTTPExecutor) Cancel(ctx context.Context, userID, confirmationID string) error {
	endpoint := fmt.Sprintf("/api/v1/agent/confirmations/%s/cancel", confirmationID)
	_, err := e.doRequest(ctx, "POST", endpoint, nil)
	return err
}

// endpointForTool maps tool names to HTTP endpoints.
func (e *HTTPExecutor) endpointForTool(tool string) string {
	// Map tool names to agent_gateway endpoints
	endpoints := map[string]string{
		"get_balance":         "/api/v1/agent/wallet/balance",
		"get_savings_balance": "/api/v1/agent/savings/balance",
		"get_vault_rates":     "/api/v1/agent/savings/vaults",
		"get_transactions":    "/api/v1/agent/transactions",
		"get_profile":         "/api/v1/agent/profile",
		"search_users":        "/api/v1/agent/users/search",
		"send_money":          "/api/v1/agent/payments/send",
		"deposit_savings":     "/api/v1/agent/savings/deposit",
		"withdraw_savings":    "/api/v1/agent/savings/withdraw",
	}

	if endpoint, ok := endpoints[tool]; ok {
		return endpoint
	}
	// Default: use tool name as endpoint
	return fmt.Sprintf("/api/v1/agent/tools/%s", tool)
}

// doRequest performs an HTTP request to the agent_gateway.
func (e *HTTPExecutor) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*core.ExecuteResponse, error) {
	url := e.baseURL + endpoint

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", e.apiKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &core.ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)),
		}, nil
	}

	var result core.ExecuteResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// If response isn't our expected format, wrap the raw response
		result = core.ExecuteResponse{
			Success: true,
			Data:    respBody,
		}
	}

	return &result, nil
}
