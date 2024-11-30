package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/jonesrussell/goprowl/search/engine/query"
)

// Document defines the interface for searchable documents
type Document interface {
	ID() string
	Type() string
	Content() map[string]interface{}
	Metadata() map[string]interface{}
	Permission() *Permission
}

// SearchEngine defines the interface for search operations
type SearchEngine interface {
	Search(ctx context.Context, q *query.Query) (*SearchResults, error)
	Index(ctx context.Context, doc Document) error
	BatchIndex(ctx context.Context, docs []Document) error
	Delete(id string) error
	List(ctx context.Context) ([]Document, error)
	Clear() error
	Stats() *SearchStats
}

// SearchStats represents search engine statistics
type SearchStats struct {
	LastIndexed   time.Time
	DocumentCount int64
}

// SearchOptions represents search configuration options
type SearchOptions struct {
	QueryString string
	Filters     map[string]interface{}
	Page        int
	PageSize    int
}

// SearchResults represents search results
type SearchResults struct {
	Hits     []SearchResult
	Metadata map[string]interface{}
}

// SearchResult represents a single search result
type SearchResult struct {
	Content map[string]interface{}
	Score   float64
}

// SearchError represents a search operation error
type SearchError struct {
	Op  string
	Err error
}

func (e *SearchError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Permission represents access control for documents
type Permission struct {
	Read  []string
	Write []string
}
