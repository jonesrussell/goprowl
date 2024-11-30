package types

import (
	"context"
	"errors"
	"time"
)

// StorageAdapter defines the interface for storage implementations
type StorageAdapter interface {
	// Store stores a single document
	Store(ctx context.Context, doc Document) error

	// BatchStore stores multiple documents
	BatchStore(ctx context.Context, docs []Document) error

	// Get retrieves a document by ID
	Get(ctx context.Context, id string) (Document, error)

	// GetAll retrieves all documents
	GetAll(ctx context.Context) ([]Document, error)

	// Delete removes a document by ID
	Delete(ctx context.Context, id string) error

	// Clear removes all documents
	Clear(ctx context.Context) error

	// Close closes the storage
	Close() error
}

// Document represents a stored document
type Document interface {
	GetURL() string
	GetTitle() string
	GetContent() string
	GetType() string
	GetCreatedAt() time.Time
	GetMetadata() map[string]interface{}
}

// DocumentOption represents options for document operations
type DocumentOption func(*DocumentOptions)

// DocumentOptions contains configuration for document operations
type DocumentOptions struct {
	ExpiresAt time.Time
	Priority  int
	Tags      []string
}

// Common errors
var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrInvalidDocument  = errors.New("invalid document")
	ErrStorageClosed    = errors.New("storage is closed")
)
