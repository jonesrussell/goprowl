package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/jonesrussell/goprowl/search/engine/query"
	"github.com/jonesrussell/goprowl/search/storage"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type BasicSearchEngine struct {
	storage storage.StorageAdapter
	stats   *SearchStats
	index   bleve.Index
	metrics *searchMetrics
}

// FacetResult represents a single facet value and its count
type FacetResult struct {
	Value string
	Count int64
}

func (e *BasicSearchEngine) Search(ctx context.Context, q *query.Query) (*SearchResults, error) {
	select {
	case <-ctx.Done():
		return nil, &SearchError{Op: "Search", Err: ctx.Err()}
	default:
		docs, err := e.storage.GetAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get documents: %w", err)
		}

		scored := make([]struct {
			doc   Document
			score float64
		}, 0)

		for _, doc := range docs {
			score := e.calculateRelevancy(doc, q.Terms)
			if !e.matchesFilters(doc, q.Filters) {
				continue
			}
			if score > 0 {
				scored = append(scored, struct {
					doc   Document
					score float64
				}{
					doc:   NewBasicDocument(doc),
					score: score,
				})
			}
		}

		// Sort by score
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})

		// Apply pagination
		start := (q.Page - 1) * q.PageSize
		end := start + q.PageSize
		if end > len(scored) {
			end = len(scored)
		}

		// Convert scored documents to SearchResults
		hits := make([]SearchResult, 0)
		if start < end {
			for _, s := range scored[start:end] {
				hits = append(hits, SearchResult{
					Content: s.doc.Content(),
					Score:   s.score,
				})
			}
		}

		// Create facets
		facets := make(map[string][]FacetResult)
		typeCounts := make(map[string]int64)
		for _, doc := range docs {
			if basicDoc, ok := doc.(storage.Document); ok {
				typeCounts[basicDoc.GetType()]++
			}
		}

		typeFacets := make([]FacetResult, 0)
		for typ, count := range typeCounts {
			typeFacets = append(typeFacets, FacetResult{
				Value: typ,
				Count: count,
			})
		}
		facets["type"] = typeFacets

		return &SearchResults{
			Hits: hits,
			Metadata: map[string]interface{}{
				"total":      int64(len(scored)),
				"query_time": time.Now(),
				"facets":     facets,
			},
		}, nil
	}
}

func (e *BasicSearchEngine) List(ctx context.Context) ([]Document, error) {
	docs, err := e.storage.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	results := make([]Document, 0, len(docs))
	for _, doc := range docs {
		results = append(results, NewBasicDocument(doc))
	}

	return results, nil
}

func (e *BasicSearchEngine) BatchIndex(ctx context.Context, docs []Document) error {
	g, ctx := errgroup.WithContext(ctx)
	sem := semaphore.NewWeighted(10)

	for _, doc := range docs {
		doc := doc
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)
			return e.Index(ctx, doc)
		})
	}

	return g.Wait()
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

func (e *BasicSearchEngine) calculateRelevancy(doc storage.Document, terms []*query.QueryTerm) float64 {
	score := 0.0

	if len(terms) == 0 {
		return score
	}

	// For AND queries, check if all terms match
	isAndQuery := false
	for _, term := range terms {
		if term.Required {
			isAndQuery = true
			break
		}
	}

	if isAndQuery {
		for _, term := range terms {
			if !strings.Contains(strings.ToLower(doc.Title), strings.ToLower(term.Text)) &&
				!strings.Contains(strings.ToLower(doc.Content), strings.ToLower(term.Text)) {
				return 0.0 // If any required term doesn't match, return 0
			}
		}
		// If we get here, all required terms matched
		score = 1.0
		return score
	}

	// Handle different term types
	for _, term := range terms {
		switch term.Type {
		case query.TypePhrase:
			// For phrases, check exact matches
			if strings.Contains(strings.ToLower(doc.Title), strings.ToLower(term.Text)) {
				score += 3.0
			}
			if strings.Contains(strings.ToLower(doc.Content), strings.ToLower(term.Text)) {
				score += 2.0
			}

		default:
			// For regular terms
			if strings.Contains(strings.ToLower(doc.Title), strings.ToLower(term.Text)) {
				score += 2.0
			}
			if strings.Contains(strings.ToLower(doc.Content), strings.ToLower(term.Text)) {
				score += 1.0
			}
		}
	}

	return score
}

func (e *BasicSearchEngine) matchesFilters(doc storage.Document, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "type":
			if basicDoc, ok := doc.(storage.Document); ok {
				if basicDoc.GetType() != value.(string) {
					return false
				}
			}
		}
	}
	return true
}

func New(storage storage.StorageAdapter) (SearchEngine, error) {
	// Create new bleve index in memory
	indexMapping := mapping.NewIndexMapping()
	index, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create search index: %w", err)
	}

	return &BasicSearchEngine{
		storage: storage,
		stats: &SearchStats{
			LastIndexed: time.Now(),
		},
		index: index,
	}, nil
}

// SearchWithOptions implements the SearchEngine interface
func (e *BasicSearchEngine) SearchWithOptions(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	q := &query.Query{
		Terms:    make([]*query.QueryTerm, 0),
		Filters:  opts.Filters,
		Page:     opts.Page,
		PageSize: opts.PageSize,
	}

	// Parse the query string
	processor := query.NewQueryProcessor()
	parsedQuery, err := processor.ParseQuery(ctx, opts.QueryString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}
	q.Terms = parsedQuery.Terms

	// Use existing Search method
	results, err := e.Search(ctx, q)
	if err != nil {
		return nil, err
	}

	return results.Hits, nil
}

// GetTotalResults implements the SearchEngine interface
func (e *BasicSearchEngine) GetTotalResults(ctx context.Context, queryString string) (int, error) {
	processor := query.NewQueryProcessor()
	query, err := processor.ParseQuery(ctx, queryString)
	if err != nil {
		return 0, fmt.Errorf("failed to parse query: %w", err)
	}

	// Use existing Search method
	result, err := e.Search(ctx, query)
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

// Clear implements the SearchEngine interface by removing all documents
func (e *BasicSearchEngine) Clear() error {
	// Clear the bleve index
	if err := e.index.Close(); err != nil {
		return fmt.Errorf("failed to close index: %w", err)
	}

	// Create new empty index
	indexMapping := mapping.NewIndexMapping()
	newIndex, err := bleve.NewMemOnly(indexMapping)
	if err != nil {
		return fmt.Errorf("failed to create new index: %w", err)
	}
	e.index = newIndex

	// Clear the storage
	if err := e.storage.Clear(context.Background()); err != nil {
		return fmt.Errorf("failed to clear storage: %w", err)
	}

	// Reset stats
	e.stats = &SearchStats{
		LastIndexed:   time.Now(),
		DocumentCount: 0,
	}

	return nil
}

// Add structured metrics
type searchMetrics struct {
	searchDuration   *prometheus.HistogramVec
	searchErrors     *prometheus.CounterVec
	documentsIndexed *prometheus.Counter
}

func (e *BasicSearchEngine) recordMetrics(start time.Time, q *query.Query, err error) {
	duration := time.Since(start)
	queryType := "default" // or get from query options
	e.metrics.searchDuration.WithLabelValues(queryType).Observe(duration.Seconds())
	if err != nil {
		e.metrics.searchErrors.WithLabelValues(err.Error()).Inc()
	}
}

func (e *BasicSearchEngine) Index(ctx context.Context, doc Document) error {
	if doc == nil {
		return fmt.Errorf("cannot index nil document")
	}

	content := doc.Content()
	if content == nil {
		return fmt.Errorf("document content is nil")
	}

	// Add document to storage
	if err := e.storage.Store(ctx, doc); err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}

	// Update stats
	e.stats.DocumentCount++
	e.stats.LastIndexed = time.Now()

	// Record metric
	e.metrics.documentsIndexed.Inc()

	return nil
}
