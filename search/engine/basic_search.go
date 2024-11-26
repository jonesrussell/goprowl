package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jonesrussell/goprowl/search/storage"
)

type BasicSearchEngine struct {
	storage storage.StorageAdapter
	stats   *SearchStats
}

func (e *BasicSearchEngine) Search(query Query) (*SearchResult, error) {
	docs, err := e.storage.GetAll(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	// Convert storage documents to interface Documents and score them
	scored := make([]struct {
		doc   Document
		score float64
	}, 0)

	for _, doc := range docs {
		score := e.calculateRelevancy(doc, query.Terms())

		// Apply filters
		if !e.matchesFilters(doc, query.Filters()) {
			continue
		}

		if score > 0 {
			// Convert storage document to Document interface
			scored = append(scored, struct {
				doc   Document
				score float64
			}{
				doc:   NewBasicDocument(doc), // We'll create this helper
				score: score,
			})
		}
	}

	// Sort by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Apply pagination
	page := query.Pagination()
	start := (page.Number - 1) * page.Size
	end := start + page.Size
	if end > len(scored) {
		end = len(scored)
	}

	hits := make([]Document, 0)
	if start < end {
		for _, s := range scored[start:end] {
			hits = append(hits, s.doc)
		}
	}

	// Create facets (example with content type facet)
	facets := make(map[string][]Facet)
	typeCounts := make(map[string]int64)
	for _, doc := range docs {
		typeCounts[doc.Type]++
	}

	typeFacets := make([]Facet, 0)
	for typ, count := range typeCounts {
		typeFacets = append(typeFacets, Facet{
			Value: typ,
			Count: count,
		})
	}
	facets["type"] = typeFacets

	return &SearchResult{
		Hits:   hits,
		Facets: facets,
		Metadata: map[string]interface{}{
			"total":      int64(len(scored)),
			"query_time": time.Now(),
		},
	}, nil
}

// Helper to convert storage document to Document interface
type BasicDocument struct {
	id         string
	docType    string
	content    map[string]interface{}
	metadata   map[string]interface{}
	permission *Permission
}

func NewBasicDocument(doc *storage.Document) *BasicDocument {
	return &BasicDocument{
		id:      doc.URL, // Using URL as ID for now
		docType: "webpage",
		content: map[string]interface{}{
			"title":   doc.Title,
			"content": doc.Content,
			"url":     doc.URL,
		},
		metadata: map[string]interface{}{
			"created_at": doc.CreatedAt,
		},
		permission: &Permission{
			Read:  []string{"public"},
			Write: []string{"admin"},
		},
	}
}

// Implement Document interface
func (d *BasicDocument) ID() string                       { return d.id }
func (d *BasicDocument) Type() string                     { return d.docType }
func (d *BasicDocument) Content() map[string]interface{}  { return d.content }
func (d *BasicDocument) Metadata() map[string]interface{} { return d.metadata }
func (d *BasicDocument) Permission() *Permission          { return d.permission }

// Also implement other SearchEngine interface methods:
func (e *BasicSearchEngine) Index(doc Document) error {
	// Convert Document interface to storage.Document
	storageDoc := &storage.Document{
		URL:       doc.ID(), // Using ID as URL since that's our current pattern
		Title:     doc.Content()["title"].(string),
		Content:   doc.Content()["content"].(string),
		Type:      doc.Type(),
		CreatedAt: time.Now(),
	}

	// Store the document using the storage adapter
	err := e.storage.Store(context.Background(), storageDoc)
	if err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	// Update stats
	e.stats.DocumentCount++
	e.stats.LastIndexed = time.Now()

	return nil
}

func (e *BasicSearchEngine) BatchIndex(docs []Document) error {
	storageDocs := make([]*storage.Document, len(docs))
	for i, doc := range docs {
		storageDocs[i] = &storage.Document{
			URL:       doc.ID(),
			Title:     doc.Content()["title"].(string),
			Content:   doc.Content()["content"].(string),
			Type:      doc.Type(),
			CreatedAt: time.Now(),
		}
	}

	// Store all documents
	err := e.storage.BatchStore(context.Background(), storageDocs)
	if err != nil {
		return fmt.Errorf("failed to batch store documents: %w", err)
	}

	// Update stats
	e.stats.DocumentCount += int64(len(docs))
	e.stats.LastIndexed = time.Now()

	return nil
}

func (e *BasicSearchEngine) Delete(id string) error {
	// Implementation
	return nil
}

func (e *BasicSearchEngine) Suggest(prefix string) []string {
	// Implementation
	return nil
}

func (e *BasicSearchEngine) Reindex() error {
	// Implementation
	return nil
}

func (e *BasicSearchEngine) Stats() *SearchStats {
	return e.stats
}

func (e *BasicSearchEngine) calculateRelevancy(doc *storage.Document, terms []string) float64 {
	score := 0.0

	for _, term := range terms {
		// Title matching (higher weight)
		if strings.Contains(strings.ToLower(doc.Title), strings.ToLower(term)) {
			score += 2.0
		}

		// Content matching
		if strings.Contains(strings.ToLower(doc.Content), strings.ToLower(term)) {
			score += 1.0
		}
	}

	return score
}

func (e *BasicSearchEngine) matchesFilters(doc *storage.Document, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "type":
			if doc.Type != value.(string) {
				return false
			}
			// Add more filter cases as needed
		}
	}
	return true
}

func New(storage storage.StorageAdapter) SearchEngine {
	return &BasicSearchEngine{
		storage: storage,
		stats: &SearchStats{
			LastIndexed: time.Now(),
		},
	}
}

// SearchWithOptions implements the SearchEngine interface
func (e *BasicSearchEngine) SearchWithOptions(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	// Convert SearchOptions to Query
	query := NewBasicQuery(strings.Fields(opts.Query))

	// Use existing Search method
	result, err := e.Search(query)
	if err != nil {
		return nil, err
	}

	// Convert single result to slice for compatibility
	return []SearchResult{*result}, nil
}

// GetTotalResults implements the SearchEngine interface
func (e *BasicSearchEngine) GetTotalResults(ctx context.Context, queryString string) (int, error) {
	// Create a basic query from the query string
	query := NewBasicQuery(strings.Fields(queryString))

	// Use existing Search method
	result, err := e.Search(query)
	if err != nil {
		return 0, err
	}

	// Get total from metadata
	total, ok := result.Metadata["total"].(int64)
	if !ok {
		return 0, fmt.Errorf("invalid total count in search results")
	}

	return int(total), nil
}
