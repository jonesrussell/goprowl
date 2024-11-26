package engine

import (
	"context"
	"time"
)

// Permission represents access control for documents
type Permission struct {
	Read  []string
	Write []string
}

// Page represents pagination information
type Page struct {
	Number int
	Size   int
}

// SortField represents sorting configuration
type SortField struct {
	Field     string
	Ascending bool
}

// SearchStats contains search engine statistics
type SearchStats struct {
	DocumentCount  int64
	IndexSize      int64
	LastIndexed    time.Time
	QueryCount     int64
	AverageLatency time.Duration
}

// Facet represents a category or filter value with its count
type Facet struct {
	Value string
	Count int64
}

// SearchResult represents search results
type SearchResult struct {
	Hits     []Document
	Facets   map[string][]Facet
	Metadata map[string]interface{}
}

// Document interface as defined in the spec
type Document interface {
	ID() string
	Type() string
	Content() map[string]interface{}
	Metadata() map[string]interface{}
	Permission() *Permission
}

// Query interface as defined in the spec
type Query interface {
	Terms() []string
	Filters() map[string]interface{}
	Pagination() *Page
	Sort() []SortField
}

// SearchEngine interface as defined in the spec
type SearchEngine interface {
	// Indexing operations
	Index(doc Document) error
	BatchIndex(docs []Document) error
	Delete(id string) error

	// Searching operations
	Search(query Query) (*SearchResult, error)
	SearchWithOptions(ctx context.Context, opts SearchOptions) ([]SearchResult, error)
	GetTotalResults(ctx context.Context, query string) (int, error)
	Suggest(prefix string) []string

	// Management operations
	Reindex() error
	Stats() *SearchStats
}

// SearchOptions represents options for searching
type SearchOptions struct {
	Query     string
	FromDate  *time.Time
	ToDate    *time.Time
	TitleOnly bool
	Limit     int
	Offset    int
}
