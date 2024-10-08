package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/conneroisu/semanticrouter-go"
)

// Store is a simple key-value store for embeddings.
type Store struct {
	mu    sync.RWMutex
	store map[string][]float64
}

// NewStore creates a new Store from a redis client.
func NewStore() *Store {
	return &Store{store: make(map[string][]float64)}
}

// Get gets a value from the
func (s *Store) Get(
	_ context.Context,
	utterance string,
) (embedding []float64, err error) {
	embedding, ok := s.store[utterance]
	if !ok {
		return nil, fmt.Errorf("key does not exist: %w", err)
	}
	return embedding, nil
}

// Set sets a value in the in-memory store.
//
// It is concurrency safe.
func (s *Store) Set(
	_ context.Context,
	utterance semanticrouter.Utterance,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[utterance.Utterance] = utterance.Embed
	return nil
}

// Close closes the store.
//
// It is concurrency safe.
func (s *Store) Close() error {
	if s.store == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = nil
	return nil
}
