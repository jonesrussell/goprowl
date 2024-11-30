package query

import (
	"context"
	"fmt"
	"strings"
)

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
	q := NewQuery()
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

	printDebugOutput(q)
	return q, nil
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

func printDebugOutput(q *Query) {
	fmt.Printf("Parsed Query:\n")
	fmt.Printf("HasAndOperator: %v\n", q.HasAndOperator)
	for i, term := range q.Terms {
		fmt.Printf("Term %d: Text='%s', Required=%v, Type=%v\n",
			i, term.Text, term.Required, term.Type)
	}
}
