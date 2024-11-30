package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/goprowl/search/engine/query"
)

// SearchEngine interface defines the contract for search engine implementations
type SearchEngine[T Document] interface {
	// Indexing operations
	Index(doc Document) error
	BatchIndex(docs []Document) error
	Delete(id string) error

	// Searching operations
	Search(ctx context.Context, q *query.Query) (*SearchResults[T], error)
	SearchWithOptions(ctx context.Context, opts SearchOptions) ([]SearchResult, error)
	GetTotalResults(ctx context.Context, queryStr string) (int, error)
	Suggest(prefix string) []string

	// Management operations
	Reindex() error
	Stats() *SearchStats
	List() ([]Document, error)
	Clear() error
}

// SearchResult represents a single search result
type SearchResult struct {
	DocID   string
	Score   float64
	Content map[string]interface{}
}

// SearchResults represents a collection of search results
type SearchResults[T Document] struct {
	Hits     []T
	Total    int64
	Took     time.Duration
	Metadata map[string]interface{}
}

// SearchStats holds search engine statistics
type SearchStats struct {
	DocumentCount int64
	LastIndexed   time.Time
	IndexSize     int64
}

// SearchOptions represents options for search operations
type SearchOptions struct {
	QueryString string
	Page        int
	PageSize    int
	Filters     map[string]interface{}
	SortField   string
	SortDesc    bool
}

// SortField represents a field to sort by
type SortField struct {
	Field      string
	Descending bool
}

// Document interface defines the contract for searchable documents
type Document interface {
	ID() string
	Type() string
	Content() map[string]interface{}
	Metadata() map[string]interface{}
	Permission() *Permission
}

// Permission represents document access permissions
type Permission struct {
	Read  []string
	Write []string
}

// Add error types using the new errors.Join pattern
type SearchError struct {
	Op   string
	Kind ErrorKind
	Err  error
}

type ErrorKind int

const (
	ErrorKindNotFound ErrorKind = iota
	ErrorKindInvalidInput
	ErrorKindInternal
)

func (e *SearchError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *SearchError) Unwrap() error {
	return e.Err
}

// Add comprehensive documentation with examples
// Document represents a searchable item in the search engine.
//
// Example:
//
//	doc := &Document{
//	    ID:      "123",
//	    Content: "Sample content",
//	    Type:    "article",
//	}
