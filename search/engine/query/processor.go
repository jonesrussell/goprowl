package query

import (
	"context"
	"fmt"
	"strings"
)

// ErrorKind represents the type of error
type ErrorKind int

const (
	ErrorKindInvalidInput ErrorKind = iota
	ErrorKindParseFailed
	ErrorKindInternalError
)

// SearchError represents a query processing error
type SearchError struct {
	Op   string
	Kind ErrorKind
	Err  error
}

func (e *SearchError) Error() string {
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// QueryProcessor handles query parsing and processing
type QueryProcessor struct{}

// NewQueryProcessor creates a new QueryProcessor instance
func NewQueryProcessor() *QueryProcessor {
	return &QueryProcessor{}
}

// ParseQuery parses a query string into a structured Query
func (p *QueryProcessor) ParseQuery(ctx context.Context, queryStr string) (*Query, error) {
	sanitized := p.sanitizeQuery(queryStr)
	if err := p.validateQuery(sanitized); err != nil {
		return nil, &SearchError{
			Op:   "ParseQuery",
			Kind: ErrorKindInvalidInput,
			Err:  err,
		}
	}

	q := &Query{
		Terms:    make([]*QueryTerm, 0),
		Page:     1,
		PageSize: 10,
		Filters:  make(map[string]interface{}),
	}

	terms := splitKeepingQuotes(sanitized)
	hasAnd := containsAndOperator(terms)

	if len(terms) > 0 {
		if err := addFirstTerm(q, terms[0]); err != nil {
			return nil, err
		}
	}

	if err := addRemainingTerms(q, terms[1:], hasAnd); err != nil {
		return nil, err
	}

	return q, nil
}

// sanitizeQuery cleans and normalizes the query string
func (p *QueryProcessor) sanitizeQuery(query string) string {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Replace multiple spaces with single space
	query = strings.Join(strings.Fields(query), " ")

	// Remove special characters except quotes and basic operators
	allowed := func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '"',
			r == ' ',
			r == '+',
			r == '-':
			return r
		}
		return ' '
	}

	return strings.Map(allowed, query)
}

// validateQuery checks if the query is valid
func (p *QueryProcessor) validateQuery(query string) error {
	if len(query) == 0 {
		return fmt.Errorf("empty query")
	}

	// Check for unmatched quotes
	quoteCount := strings.Count(query, "\"")
	if quoteCount%2 != 0 {
		return fmt.Errorf("unmatched quotes in query")
	}

	// Check for invalid operators
	terms := strings.Fields(query)
	for i, term := range terms {
		if isOperator(term) && (i == 0 || i == len(terms)-1) {
			return fmt.Errorf("operator '%s' at invalid position", term)
		}
	}

	return nil
}

// isOperator checks if a term is an operator
func isOperator(term string) bool {
	operators := map[string]bool{
		"AND": true,
		"OR":  true,
		"NOT": true,
	}
	return operators[strings.ToUpper(term)]
}

func containsAndOperator(terms []string) bool {
	for _, t := range terms {
		if strings.EqualFold(strings.TrimSpace(t), "AND") {
			return true
		}
	}
	return false
}

func addFirstTerm(q *Query, firstTerm string) error {
	firstTerm = strings.TrimSpace(firstTerm)
	if firstTerm != "" && !strings.EqualFold(firstTerm, "AND") {
		queryTerm := &QueryTerm{
			Text:     strings.Trim(firstTerm, "\""),
			Type:     TypeSimple,
			Required: true,
		}

		if strings.HasPrefix(firstTerm, "\"") && strings.HasSuffix(firstTerm, "\"") {
			queryTerm.Type = TypePhrase
			queryTerm.Text = strings.Trim(queryTerm.Text, "\"")
		}
		q.AddTerm(queryTerm)
	}
	return nil
}

func addRemainingTerms(q *Query, terms []string, hasAnd bool) error {
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" || strings.EqualFold(term, "AND") {
			continue
		}

		queryTerm := &QueryTerm{
			Text:     strings.Trim(term, "\""),
			Type:     TypeSimple,
			Required: hasAnd,
		}

		if strings.HasPrefix(term, "\"") && strings.HasSuffix(term, "\"") {
			queryTerm.Type = TypePhrase
			queryTerm.Text = strings.Trim(queryTerm.Text, "\"")
			queryTerm.Required = true
		}

		q.AddTerm(queryTerm)
	}
	return nil
}

func splitKeepingQuotes(s string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for _, r := range s {
		switch r {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case ' ':
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
