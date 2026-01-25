package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
)

// MemoryConfirmations is an in-memory implementation of Confirmations.
// Suitable for development and testing. Not suitable for production
// as data is lost on restart and doesn't work across multiple instances.
type MemoryConfirmations struct {
	mu            sync.RWMutex
	actions       map[string]*core.PendingAction // actionID -> action
	byIdempotency map[string]string              // idempotencyKey -> actionID
}

// NewMemoryConfirmations creates an in-memory confirmation store.
func NewMemoryConfirmations() *MemoryConfirmations {
	return &MemoryConfirmations{
		actions:       make(map[string]*core.PendingAction),
		byIdempotency: make(map[string]string),
	}
}

func (m *MemoryConfirmations) Store(ctx context.Context, action *core.PendingAction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.actions[action.ID] = action
	if action.IdempotencyKey != "" {
		m.byIdempotency[action.IdempotencyKey] = action.ID
	}
	return nil
}

func (m *MemoryConfirmations) Get(ctx context.Context, userID, actionID string) (*core.PendingAction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	action, ok := m.actions[actionID]
	if !ok {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}
	if action.UserID != userID {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}
	if action.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("action expired: %s", actionID)
	}
	return action, nil
}

func (m *MemoryConfirmations) GetByIdempotency(ctx context.Context, userID, key string) (*core.PendingAction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	actionID, ok := m.byIdempotency[key]
	if !ok {
		return nil, nil
	}

	action, ok := m.actions[actionID]
	if !ok {
		return nil, nil
	}
	if action.UserID != userID {
		return nil, nil
	}
	if action.ExpiresAt < time.Now().Unix() {
		return nil, nil
	}
	return action, nil
}

func (m *MemoryConfirmations) Confirm(ctx context.Context, userID, actionID string) (*core.PendingAction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	action, ok := m.actions[actionID]
	if !ok {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}
	if action.UserID != userID {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}
	if action.ExpiresAt < time.Now().Unix() {
		m.deleteUnlocked(action)
		return nil, fmt.Errorf("action expired: %s", actionID)
	}

	m.deleteUnlocked(action)
	return action, nil
}

func (m *MemoryConfirmations) Cancel(ctx context.Context, userID, actionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	action, ok := m.actions[actionID]
	if !ok {
		return fmt.Errorf("action not found: %s", actionID)
	}
	if action.UserID != userID {
		return fmt.Errorf("action not found: %s", actionID)
	}

	m.deleteUnlocked(action)
	return nil
}

func (m *MemoryConfirmations) Cleanup(ctx context.Context) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().Unix()
	count := 0
	for _, action := range m.actions {
		if action.ExpiresAt < now {
			m.deleteUnlocked(action)
			count++
		}
	}
	return count, nil
}

func (m *MemoryConfirmations) deleteUnlocked(action *core.PendingAction) {
	delete(m.actions, action.ID)
	if action.IdempotencyKey != "" {
		delete(m.byIdempotency, action.IdempotencyKey)
	}
}

// Verify MemoryConfirmations implements Confirmations.
var _ Confirmations = (*MemoryConfirmations)(nil)
