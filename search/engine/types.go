package engine

import "time"

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
	Total    int64
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
	// Indexing
	Index(doc Document) error
	BatchIndex(docs []Document) error
	Delete(id string) error

	// Searching
	Search(query Query) (*SearchResult, error)
	Suggest(prefix string) []string

	// Management
	Reindex() error
	Stats() *SearchStats
}
