package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/dgraph-io/ristretto"
)

// RistrettoConfirmations is a high-performance implementation of Confirmations
// using Ristretto cache. Recommended for production single-instance deployments.
// For distributed deployments, use Redis or similar.
type RistrettoConfirmations struct {
	cache         *ristretto.Cache
	idempotency   *ristretto.Cache
	defaultTTL    time.Duration
	mu            sync.RWMutex
	actionsByUser map[string]map[string]struct{} // userID -> set of actionIDs
}

// RistrettoConfig configures the Ristretto confirmations store.
type RistrettoConfig struct {
	// NumCounters is the number of keys to track frequency of (10x expected items).
	NumCounters int64
	// MaxCost is the maximum size of the cache in bytes.
	MaxCost int64
	// BufferItems is the number of keys per Get buffer.
	BufferItems int64
	// DefaultTTL is the default expiration time for pending actions.
	DefaultTTL time.Duration
}

// DefaultRistrettoConfig returns sensible defaults for a confirmation store.
func DefaultRistrettoConfig() *RistrettoConfig {
	return &RistrettoConfig{
		NumCounters: 1e5,              // 100K counters
		MaxCost:     1 << 27,          // 128MB
		BufferItems: 64,               // 64 keys per buffer
		DefaultTTL:  15 * time.Minute, // 15 minute expiration
	}
}

// NewRistrettoConfirmations creates a high-performance confirmation store.
func NewRistrettoConfirmations(cfg *RistrettoConfig) (*RistrettoConfirmations, error) {
	if cfg == nil {
		cfg = DefaultRistrettoConfig()
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: cfg.BufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create action cache: %w", err)
	}

	idempotency, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: cfg.BufferItems,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create idempotency cache: %w", err)
	}

	return &RistrettoConfirmations{
		cache:         cache,
		idempotency:   idempotency,
		defaultTTL:    cfg.DefaultTTL,
		actionsByUser: make(map[string]map[string]struct{}),
	}, nil
}

func (r *RistrettoConfirmations) Store(ctx context.Context, action *core.PendingAction) error {
	ttl := r.ttlFor(action)

	// Store action by ID
	key := r.actionKey(action.UserID, action.ID)
	r.cache.SetWithTTL(key, action, 1, ttl)

	// Store idempotency mapping if present
	if action.IdempotencyKey != "" {
		idempKey := r.idempotencyKey(action.UserID, action.IdempotencyKey)
		r.idempotency.SetWithTTL(idempKey, action.ID, 1, ttl)
	}

	// Track action for user (for cleanup)
	r.mu.Lock()
	if r.actionsByUser[action.UserID] == nil {
		r.actionsByUser[action.UserID] = make(map[string]struct{})
	}
	r.actionsByUser[action.UserID][action.ID] = struct{}{}
	r.mu.Unlock()

	// Wait for value to be set
	r.cache.Wait()
	r.idempotency.Wait()

	return nil
}

func (r *RistrettoConfirmations) Get(ctx context.Context, userID, actionID string) (*core.PendingAction, error) {
	key := r.actionKey(userID, actionID)
	val, found := r.cache.Get(key)
	if !found {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	action := val.(*core.PendingAction)
	if action.ExpiresAt < time.Now().Unix() {
		r.delete(action)
		return nil, fmt.Errorf("action expired: %s", actionID)
	}

	return action, nil
}

func (r *RistrettoConfirmations) GetByIdempotency(ctx context.Context, userID, key string) (*core.PendingAction, error) {
	idempKey := r.idempotencyKey(userID, key)
	val, found := r.idempotency.Get(idempKey)
	if !found {
		return nil, nil
	}

	actionID := val.(string)
	action, err := r.Get(ctx, userID, actionID)
	if err != nil {
		// Action expired or not found, clean up idempotency mapping
		r.idempotency.Del(idempKey)
		return nil, nil
	}

	return action, nil
}

func (r *RistrettoConfirmations) Confirm(ctx context.Context, userID, actionID string) (*core.PendingAction, error) {
	action, err := r.Get(ctx, userID, actionID)
	if err != nil {
		return nil, err
	}

	r.delete(action)
	return action, nil
}

func (r *RistrettoConfirmations) Cancel(ctx context.Context, userID, actionID string) error {
	action, err := r.Get(ctx, userID, actionID)
	if err != nil {
		return err
	}

	r.delete(action)
	return nil
}

func (r *RistrettoConfirmations) Cleanup(ctx context.Context) (int, error) {
	// Ristretto handles TTL-based eviction automatically.
	// This method cleans up expired entries from our tracking map.
	r.mu.Lock()
	defer r.mu.Unlock()

	count := 0
	now := time.Now().Unix()

	for userID, actions := range r.actionsByUser {
		for actionID := range actions {
			key := r.actionKey(userID, actionID)
			val, found := r.cache.Get(key)
			if !found {
				delete(actions, actionID)
				count++
				continue
			}

			action := val.(*core.PendingAction)
			if action.ExpiresAt < now {
				r.cache.Del(key)
				if action.IdempotencyKey != "" {
					r.idempotency.Del(r.idempotencyKey(userID, action.IdempotencyKey))
				}
				delete(actions, actionID)
				count++
			}
		}

		if len(actions) == 0 {
			delete(r.actionsByUser, userID)
		}
	}

	return count, nil
}

// Close releases resources used by the cache.
func (r *RistrettoConfirmations) Close() {
	r.cache.Close()
	r.idempotency.Close()
}

func (r *RistrettoConfirmations) delete(action *core.PendingAction) {
	key := r.actionKey(action.UserID, action.ID)
	r.cache.Del(key)

	if action.IdempotencyKey != "" {
		r.idempotency.Del(r.idempotencyKey(action.UserID, action.IdempotencyKey))
	}

	r.mu.Lock()
	if actions, ok := r.actionsByUser[action.UserID]; ok {
		delete(actions, action.ID)
		if len(actions) == 0 {
			delete(r.actionsByUser, action.UserID)
		}
	}
	r.mu.Unlock()
}

func (r *RistrettoConfirmations) actionKey(userID, actionID string) string {
	return userID + ":" + actionID
}

func (r *RistrettoConfirmations) idempotencyKey(userID, key string) string {
	return userID + ":idemp:" + key
}

func (r *RistrettoConfirmations) ttlFor(action *core.PendingAction) time.Duration {
	if action.ExpiresAt > 0 {
		ttl := time.Until(time.Unix(action.ExpiresAt, 0))
		if ttl > 0 {
			return ttl
		}
	}
	return r.defaultTTL
}

// Verify RistrettoConfirmations implements Confirmations.
var _ Confirmations = (*RistrettoConfirmations)(nil)
