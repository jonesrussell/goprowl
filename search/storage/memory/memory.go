package memory

import (
	"context"
	"sync"
	"time"

	"github.com/jonesrussell/goprowl/search/storage"
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
func (m *MemoryStorage) Store(ctx context.Context, doc *storage.Document) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	content := map[string]interface{}{
		"url":        doc.URL,
		"title":      doc.Title,
		"content":    doc.Content,
		"type":       doc.Type,
		"created_at": doc.CreatedAt,
		"metadata":   doc.Metadata,
	}

	m.store[doc.URL] = content
	return nil
}

// BatchStore stores multiple documents to storage
func (m *MemoryStorage) BatchStore(ctx context.Context, docs []*storage.Document) error {
	for _, doc := range docs {
		if err := m.Store(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// GetAll retrieves all documents from storage
func (m *MemoryStorage) GetAll(ctx context.Context) ([]*storage.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	docs := make([]*storage.Document, 0, len(m.store))
	for _, content := range m.store {
		doc := &storage.Document{
			URL:       content["url"].(string),
			Title:     content["title"].(string),
			Content:   content["content"].(string),
			Type:      content["type"].(string),
			CreatedAt: content["created_at"].(time.Time),
			Metadata:  content["metadata"].(map[string]interface{}),
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// Get retrieves a document from memory
func (m *MemoryStorage) Get(ctx context.Context, id string) (*storage.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if content, exists := m.store[id]; exists {
		return &storage.Document{
			URL:       content["url"].(string),
			Title:     content["title"].(string),
			Content:   content["content"].(string),
			Type:      content["type"].(string),
			CreatedAt: content["created_at"].(time.Time),
			Metadata:  content["metadata"].(map[string]interface{}),
		}, nil
	}
	return nil, storage.ErrDocumentNotFound
}

// Delete removes a document from memory
func (m *MemoryStorage) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.store, id)
	return nil
}

// List returns all document IDs
func (m *MemoryStorage) List(ctx context.Context) ([]*storage.Document, error) {
	return m.GetAll(ctx)
}

// Search performs a basic search operation
func (m *MemoryStorage) Search(ctx context.Context, query string) ([]*storage.Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]*storage.Document, 0)
	for _, content := range m.store {
		doc := &storage.Document{
			URL:       content["url"].(string),
			Title:     content["title"].(string),
			Content:   content["content"].(string),
			Type:      content["type"].(string),
			CreatedAt: content["created_at"].(time.Time),
			Metadata:  content["metadata"].(map[string]interface{}),
		}
		results = append(results, doc)
	}
	return results, nil
}

// Add this method to MemoryStorage
func (m *MemoryStorage) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = make(map[string]map[string]interface{})
	return nil
}
