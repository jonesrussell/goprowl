package engine

import (
	"strings"

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

// Search performs the search operation using the engine's search capabilities
func (e *DefaultSearchEngine) Search(query Query) (*SearchResult, error) {
	// Get all documents from storage
	ids, err := e.storage.List()
	if err != nil {
		return nil, err
	}

	results := &SearchResult{
		Hits: make([]Document, 0),
	}

	// Process each document
	for _, id := range ids {
		doc, err := e.storage.Get(id)
		if err != nil {
			continue
		}

		// Apply search criteria and ranking here
		// This is where the actual search logic should live
		if matchesQuery(doc, query) {
			// Convert to Document interface
			results.Hits = append(results.Hits, convertToDocument(id, doc))
		}
	}

	results.Total = int64(len(results.Hits))
	return results, nil
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

// matchesQuery checks if a document matches the search criteria
func matchesQuery(doc map[string]interface{}, query Query) bool {
	// Get search terms from the query
	terms := query.Terms()
	if len(terms) == 0 {
		return true // No terms means match all
	}

	// Check if any content field contains any of the search terms
	content, ok := doc["content"].(string)
	if !ok {
		return false
	}

	contentLower := strings.ToLower(content)
	for _, term := range terms {
		if strings.Contains(contentLower, strings.ToLower(term)) {
			return true
		}
	}

	return false
}

// convertToDocument creates a Document interface from raw storage data
func convertToDocument(id string, data map[string]interface{}) Document {
	doc := NewDocument(id, "page") // Assuming "page" as default type

	// Copy content
	if content, ok := data["content"]; ok {
		doc.SetContent("content", content)
	}

	// Copy metadata if it exists
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		for key, value := range metadata {
			doc.SetMetadata(key, value)
		}
	}

	return doc
}
