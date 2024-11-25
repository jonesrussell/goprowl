package memory

import (
	"fmt"
	"sync"
)

// MemoryStorage implements StorageAdapter using in-memory map
type MemoryStorage struct {
	mu    sync.RWMutex
	store map[string]map[string]interface{}
}

// New creates a new MemoryStorage instance
func New() *MemoryStorage {
	return &MemoryStorage{
		store: make(map[string]map[string]interface{}),
	}
}

// Store saves a document to memory
func (m *MemoryStorage) Store(id string, data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.store[id] = data
	return nil
}

// Get retrieves a document from memory
func (m *MemoryStorage) Get(id string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	doc, exists := m.store[id]
	if !exists {
		return nil, fmt.Errorf("document with id %s not found", id)
	}
	return doc, nil
}

// Delete removes a document from memory
func (m *MemoryStorage) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.store[id]; !exists {
		return fmt.Errorf("document with id %s not found", id)
	}
	delete(m.store, id)
	return nil
}

// List returns all document IDs
func (m *MemoryStorage) List() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.store))
	for id := range m.store {
		ids = append(ids, id)
	}
	return ids, nil
}

// Search performs a basic search operation
func (m *MemoryStorage) Search(query map[string]interface{}) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]map[string]interface{}, 0)

	// Basic implementation - matches exact values in fields
	for _, doc := range m.store {
		matches := true
		for field, value := range query {
			if docValue, exists := doc[field]; !exists || docValue != value {
				matches = false
				break
			}
		}
		if matches {
			results = append(results, doc)
		}
	}

	return results, nil
}
