package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// IdempotencyBucketDuration is the time window for idempotency key generation.
// Actions with the same user, tool, and input within this window will have
// the same idempotency key.
const IdempotencyBucketDuration = 10 * time.Minute

// GenerateIdempotencyKey creates a unique key for deduplicating confirmations.
// Keys are deterministic based on userID, tool name, canonicalized input, and
// a 10-minute time bucket. This prevents duplicate confirmations for the same
// action within a short time window.
func GenerateIdempotencyKey(userID, tool string, input json.RawMessage) string {
	// Time bucket (10-minute windows)
	bucket := time.Now().Unix() / int64(IdempotencyBucketDuration.Seconds())

	// Canonicalize JSON by parsing and re-marshaling
	var parsed interface{}
	if err := json.Unmarshal(input, &parsed); err != nil {
		// If parsing fails, use raw input
		parsed = string(input)
	}
	canonical, _ := json.Marshal(parsed)

	// SHA256 hash of combined data
	data := fmt.Sprintf("%s:%s:%s:%d", userID, tool, string(canonical), bucket)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateIdempotencyKeyWithTime creates an idempotency key using a specific timestamp.
// Useful for testing and replay scenarios.
func GenerateIdempotencyKeyWithTime(userID, tool string, input json.RawMessage, t time.Time) string {
	bucket := t.Unix() / int64(IdempotencyBucketDuration.Seconds())

	var parsed interface{}
	if err := json.Unmarshal(input, &parsed); err != nil {
		parsed = string(input)
	}
	canonical, _ := json.Marshal(parsed)

	data := fmt.Sprintf("%s:%s:%s:%d", userID, tool, string(canonical), bucket)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
