package engine

import (
	"context"
)

// Guardrails provides rate limiting and circuit breaker functionality.
// This is an interface - implementations (e.g., Redis-backed) are provided
// by the consuming application.
type Guardrails interface {
	// Check verifies whether the user is allowed to proceed.
	// Returns a result indicating if the request is allowed and any warnings.
	Check(ctx context.Context, userID string) (*GuardrailResult, error)

	// RecordSuccess records a successful operation for the user.
	// This is used to track usage and reset circuit breakers.
	RecordSuccess(ctx context.Context, userID string)

	// RecordFailure records a failed operation for the user.
	// Repeated failures may trigger circuit breaker protection.
	RecordFailure(ctx context.Context, userID string)
}

// GuardrailResult contains the result of a guardrail check.
type GuardrailResult struct {
	// Allowed indicates whether the request should proceed.
	Allowed bool

	// Warning contains an optional warning message to show the user.
	// May be set even when Allowed is true (e.g., "approaching rate limit").
	Warning string

	// CircuitState indicates the current circuit breaker state.
	// Values: "closed" (normal), "open" (blocked), "half-open" (testing).
	CircuitState string

	// RemainingRequests is the number of requests remaining in the current window.
	RemainingRequests int

	// RetryAfter is set when Allowed is false, indicating when to retry.
	RetryAfter int64 // Unix timestamp
}

// NoOpGuardrails is a guardrails implementation that allows everything.
// Useful for development and testing.
type NoOpGuardrails struct{}

// Check always returns allowed.
func (n *NoOpGuardrails) Check(ctx context.Context, userID string) (*GuardrailResult, error) {
	return &GuardrailResult{
		Allowed:           true,
		CircuitState:      "closed",
		RemainingRequests: -1, // Unlimited
	}, nil
}

// RecordSuccess is a no-op.
func (n *NoOpGuardrails) RecordSuccess(ctx context.Context, userID string) {}

// RecordFailure is a no-op.
func (n *NoOpGuardrails) RecordFailure(ctx context.Context, userID string) {}
