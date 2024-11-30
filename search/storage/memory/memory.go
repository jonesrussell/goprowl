package memory

import (
	"context"
	"sync"

	"github.com/jonesrussell/goprowl/search/core/types"
)

// MemoryStorage implements StorageAdapter interface with in-memory storage
type MemoryStorage struct {
	docs map[string]types.Document
	mu   sync.RWMutex
}

// New creates a new memory storage instance
func New() *MemoryStorage {
	return &MemoryStorage{
		docs: make(map[string]types.Document),
	}
}

// Store stores a document in memory
func (m *MemoryStorage) Store(ctx context.Context, doc types.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.docs[doc.GetURL()] = doc
		return nil
	}
}

// Get retrieves a document by ID
func (m *MemoryStorage) Get(ctx context.Context, id string) (types.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		doc, exists := m.docs[id]
		if !exists {
			return nil, types.ErrDocumentNotFound
		}
		return doc, nil
	}
}

// GetAll retrieves all documents
func (m *MemoryStorage) GetAll(ctx context.Context) ([]types.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		docs := make([]types.Document, 0, len(m.docs))
		for _, doc := range m.docs {
			docs = append(docs, doc)
		}
		return docs, nil
	}
}

// Delete removes a document by ID
func (m *MemoryStorage) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if _, exists := m.docs[id]; !exists {
			return types.ErrDocumentNotFound
		}
		delete(m.docs, id)
		return nil
	}
}

// Clear removes all documents
func (m *MemoryStorage) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.docs = make(map[string]types.Document)
		return nil
	}
}

// BatchStore stores multiple documents in a batch operation
func (m *MemoryStorage) BatchStore(ctx context.Context, docs []types.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for _, doc := range docs {
			m.docs[doc.GetURL()] = doc
		}
		return nil
	}
}

// Close implements StorageAdapter interface
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.docs = nil
	return nil
}

// List returns all documents (alias for GetAll)
func (m *MemoryStorage) List(ctx context.Context) ([]types.Document, error) {
	return m.GetAll(ctx)
}
