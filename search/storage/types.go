package storage

import (
	"context"
	"fmt"
	"time"
)

// Document represents a stored document in the storage layer
type Document struct {
	URL       string                 // Unique URL of the document
	Title     string                 // Document title
	Content   string                 // Main content of the document
	Type      string                 // Document type (e.g., "webpage")
	CreatedAt time.Time              // Timestamp when document was created
	Metadata  map[string]interface{} // Additional metadata
}

// StorageAdapter defines the interface for storage implementations
type StorageAdapter interface {
	// Store saves a document to storage
	// Returns an error if the operation fails
	Store(ctx context.Context, doc *Document, opts ...StoreOption) error

	// Get retrieves a document by ID
	// Returns ErrDocumentNotFound if the document doesn't exist
	Get(ctx context.Context, id string, opts ...GetOption) (*Document, error)

	// Delete removes a document from storage
	// Returns an error if the document doesn't exist or operation fails
	Delete(ctx context.Context, id string) error

	// List returns all documents in storage
	// Returns an empty slice if no documents exist
	List(ctx context.Context) ([]*Document, error)

	// BatchStore stores multiple documents to storage
	// Returns an error if any document fails to store
	BatchStore(ctx context.Context, docs []*Document) error

	// GetAll retrieves all documents from storage
	// Returns an empty slice if no documents exist
	GetAll(ctx context.Context) ([]*Document, error)

	// Search searches for documents based on a query string
	// Returns matching documents or an empty slice if no matches
	Search(ctx context.Context, query string) ([]*Document, error)

	// Clear removes all documents from storage
	// Returns an error if the operation fails
	Clear(ctx context.Context) error
}

// ErrDocumentNotFound is returned when a document cannot be found in storage
var ErrDocumentNotFound = fmt.Errorf("document not found")

// Add functional options
type StoreOption func(*storeOptions)
type storeOptions struct {
	expiry time.Duration
	async  bool
}
