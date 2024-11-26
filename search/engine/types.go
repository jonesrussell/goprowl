package engine

import (
	"context"
	"time"
)

// Query interface defines the contract for all query types
type Query interface {
	Terms() []*QueryTerm
	Filters() map[string]interface{}
	Pagination() *Pagination
}

// QueryTerm represents a structured query term
type QueryTerm struct {
	Text      string
	Type      QueryType
	Field     string
	Fuzziness int
	Required  bool
	Excluded  bool
}

// Pagination holds pagination information
type Pagination struct {
	Page int
	Size int
}

// SearchOptions represents options for search operations
type SearchOptions struct {
	Query     string
	Filters   map[string]interface{}
	Page      int
	PageSize  int
	SortBy    string
	SortOrder string
}

// SearchResult represents a search result
type SearchResult struct {
	Hits     []Document
	Facets   map[string][]Facet
	Metadata map[string]interface{}
}

// Facet represents a facet in search results
type Facet struct {
	Value string
	Count int64
}

// SearchStats holds search engine statistics
type SearchStats struct {
	DocumentCount int64
	LastIndexed   time.Time
	IndexSize     int64
}

// Permission represents document access permissions
type Permission struct {
	Read  []string
	Write []string
}

// Document interface defines the contract for searchable documents
type Document interface {
	ID() string
	Type() string
	Content() map[string]interface{}
	Metadata() map[string]interface{}
	Permission() *Permission
}

// SearchEngine interface defines the contract for search engine implementations
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

	// List operation
	List() ([]Document, error)
}
