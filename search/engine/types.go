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

// SearchResult represents a single search result
type SearchResult struct {
	Content  map[string]interface{} // Contains URL, Title, Snippet, etc.
	Score    float64
	Metadata map[string]interface{}
}

// SearchResults represents a collection of search results with metadata
type SearchResults struct {
	Hits     []SearchResult
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
	Search(query Query) (*SearchResults, error)
	SearchWithOptions(ctx context.Context, opts SearchOptions) ([]SearchResult, error)
	GetTotalResults(ctx context.Context, query string) (int, error)
	Suggest(prefix string) []string

	// Management operations
	Reindex() error
	Stats() *SearchStats

	// List operation
	List() ([]Document, error)

	// Cleanup operation
	Clear() error
}

// Searcher defines the interface for search operations
type Searcher interface {
	// Search performs a search using the given query
	Search(ctx context.Context, query string) ([]SearchResult, error)

	// List returns all indexed documents
	List(ctx context.Context) ([]SearchResult, error)
}
