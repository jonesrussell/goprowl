package engine

import (
	"sort"
	"strings"
	"sync"
)

// InvertedIndex represents the core index structure
type InvertedIndex struct {
	mu sync.RWMutex
	// map[term]map[documentID]termFrequency
	index map[string]map[string]int
	// map[documentID]documentLength
	docLengths map[string]int
	// total number of documents
	documentCount int64
}

// New creates a new inverted index
func New() *InvertedIndex {
	return &InvertedIndex{
		index:         make(map[string]map[string]int),
		docLengths:    make(map[string]int),
		documentCount: 0,
	}
}

// IndexDocument adds or updates a document in the index
func (idx *InvertedIndex) IndexDocument(docID string, content string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Tokenize the content (simple implementation)
	terms := tokenize(content)

	// Calculate term frequencies for this document
	termFreqs := make(map[string]int)
	docLength := 0
	for _, term := range terms {
		termFreqs[term]++
		docLength++
	}

	// Update document length
	idx.docLengths[docID] = docLength

	// Update inverted index
	for term, freq := range termFreqs {
		if idx.index[term] == nil {
			idx.index[term] = make(map[string]int)
		}
		idx.index[term][docID] = freq
	}

	idx.documentCount++
}

// Search performs a search query and returns ranked results
func (idx *InvertedIndex) Search(query string) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	queryTerms := tokenize(query)
	scores := make(map[string]float64)

	// Calculate TF-IDF scores for each matching document
	for _, term := range queryTerms {
		if postings, exists := idx.index[term]; exists {
			idf := calculateIDF(idx.documentCount, int64(len(postings)))

			for docID, tf := range postings {
				docLength := idx.docLengths[docID]
				normalizedTF := float64(tf) / float64(docLength)
				scores[docID] += normalizedTF * idf
			}
		}
	}

	// Convert scores to sorted results
	results := rankResults(scores)
	return results
}

// SearchResult represents a ranked search result
type SearchResult struct {
	DocID string
	Score float64
}

// Helper functions

func tokenize(text string) []string {
	// Simple tokenization - split on whitespace and convert to lowercase
	return strings.Fields(strings.ToLower(text))
}

func calculateIDF(totalDocs, docsWithTerm int64) float64 {
	return float64(1.0 + totalDocs/docsWithTerm)
}

func rankResults(scores map[string]float64) []SearchResult {
	results := make([]SearchResult, 0, len(scores))
	for docID, score := range scores {
		results = append(results, SearchResult{
			DocID: docID,
			Score: score,
		})
	}

	// Sort results by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}
