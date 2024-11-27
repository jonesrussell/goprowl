package storage

import (
	"context"
	"fmt"
	"time"
)

// Document represents a stored document in the storage layer
type Document struct {
	URL       string
	Title     string
	Content   string
	Type      string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}

// StorageAdapter defines the interface for storage implementations
type StorageAdapter interface {
	// Store saves a document to storage
	Store(ctx context.Context, doc *Document) error

	// Get retrieves a document by ID
	Get(ctx context.Context, id string) (*Document, error)

	// Delete removes a document from storage
	Delete(ctx context.Context, id string) error

	// List returns all document IDs
	List(ctx context.Context) ([]*Document, error)

	// BatchStore stores multiple documents to storage
	BatchStore(ctx context.Context, docs []*Document) error

	// GetAll retrieves all documents from storage
	GetAll(ctx context.Context) ([]*Document, error)

	// Search searches for documents based on a query
	Search(ctx context.Context, query string) ([]*Document, error)

	// Clear clears the storage
	Clear(ctx context.Context) error
}

var ErrDocumentNotFound = fmt.Errorf("document not found")
