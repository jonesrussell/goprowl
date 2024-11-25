package query

import (
	"strings"
)

// QueryProcessor handles query parsing and optimization
type QueryProcessor struct {
	minTermLength int
	stopWords     map[string]struct{}
}

// New creates a new QueryProcessor
func New() *QueryProcessor {
	return &QueryProcessor{
		minTermLength: 2,
		stopWords:     defaultStopWords(),
	}
}

// ProcessQuery parses and optimizes a search query
func (qp *QueryProcessor) ProcessQuery(rawQuery string) *Query {
	terms := qp.tokenize(rawQuery)
	terms = qp.removeStopWords(terms)
	terms = qp.filterShortTerms(terms)

	return &Query{
		Terms:   terms,
		Filters: make(map[string]interface{}),
		Page:    &Page{Number: 1, Size: 10},
		Sort:    []SortField{{Field: "_score", Ascending: false}},
	}
}

// Query represents a processed search query
type Query struct {
	Terms   []string
	Filters map[string]interface{}
	Page    *Page
	Sort    []SortField
}

// Page represents pagination information
type Page struct {
	Number int
	Size   int
}

// SortField represents a field to sort by and its direction
type SortField struct {
	Field     string
	Ascending bool
}

func (qp *QueryProcessor) tokenize(text string) []string {
	return strings.Fields(strings.ToLower(text))
}

func (qp *QueryProcessor) removeStopWords(terms []string) []string {
	filtered := make([]string, 0, len(terms))
	for _, term := range terms {
		if _, isStopWord := qp.stopWords[term]; !isStopWord {
			filtered = append(filtered, term)
		}
	}
	return filtered
}

func (qp *QueryProcessor) filterShortTerms(terms []string) []string {
	filtered := make([]string, 0, len(terms))
	for _, term := range terms {
		if len(term) >= qp.minTermLength {
			filtered = append(filtered, term)
		}
	}
	return filtered
}

func defaultStopWords() map[string]struct{} {
	words := []string{"a", "an", "and", "are", "as", "at", "be", "by", "for", "from",
		"has", "he", "in", "is", "it", "its", "of", "on", "that", "the",
		"to", "was", "were", "will", "with"}
	stopWords := make(map[string]struct{})
	for _, word := range words {
		stopWords[word] = struct{}{}
	}
	return stopWords
}
