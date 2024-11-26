package engine

import (
	"github.com/jonesrussell/goprowl/search/storage"
)

// DefaultSearchEngine implements the SearchEngine interface
type DefaultSearchEngine struct {
	storage storage.StorageAdapter
}

// New creates a new SearchEngine instance
func New(storage storage.StorageAdapter) SearchEngine {
	return &DefaultSearchEngine{
		storage: storage,
	}
}

// Index adds or updates a document in the search engine
func (e *DefaultSearchEngine) Index(doc Document) error {
	return e.storage.Store(doc.ID(), doc.Content())
}

// BatchIndex adds or updates multiple documents in the search engine
func (e *DefaultSearchEngine) BatchIndex(docs []Document) error {
	// TODO: Implement batch storage operation
	// For now, we'll do it one by one
	for _, doc := range docs {
		if err := e.Index(doc); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes a document from the search engine
func (e *DefaultSearchEngine) Delete(id string) error {
	return e.storage.Delete(id)
}

// Suggest returns auto-completion suggestions for the given prefix
func (e *DefaultSearchEngine) Suggest(prefix string) []string {
	// TODO: Implement proper prefix-based suggestion logic
	// For now, return empty slice
	return []string{}
}

// Search performs a search query
func (e *DefaultSearchEngine) Search(query Query) (*SearchResult, error) {
	// Convert query to storage format
	storageQuery := map[string]interface{}{
		"terms":   query.Terms(),
		"filters": query.Filters(),
	}

	// Perform search
	docs, err := e.storage.Search(storageQuery)
	if err != nil {
		return nil, err
	}

	// Convert results
	hits := make([]Document, 0, len(docs))
	for _, doc := range docs {
		hits = append(hits, &BaseDocument{
			id:       doc["id"].(string),
			docType:  doc["type"].(string),
			content:  doc,
			metadata: make(map[string]interface{}),
		})
	}

	return &SearchResult{
		Total:    int64(len(hits)),
		Hits:     hits,
		Facets:   make(map[string][]Facet),
		Metadata: make(map[string]interface{}),
	}, nil
}

// Stats returns current search engine statistics
func (e *DefaultSearchEngine) Stats() *SearchStats {
	return &SearchStats{
		DocumentCount: 0, // TODO: Implement
		IndexSize:     0, // TODO: Implement
		QueryCount:    0, // TODO: Implement
	}
}

// Reindex rebuilds the search index
func (e *DefaultSearchEngine) Reindex() error {
	// TODO: Implement reindexing logic
	return nil
}
