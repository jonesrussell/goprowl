package engine

import (
	"strings"
)

type QueryType int

const (
	TypeSimple QueryType = iota
	TypePhrase
	TypeFuzzy
	TypeBoolean
)

// BasicQuery implements the Query interface
type BasicQuery struct {
	queryTerms     []*QueryTerm
	filters        map[string]interface{}
	pagination     *Pagination
	SortField      string
	SortDescending bool
}

// QueryProcessor handles advanced query parsing
type QueryProcessor struct{}

func NewQueryProcessor() *QueryProcessor {
	return &QueryProcessor{}
}

// ParseQuery parses a query string into structured terms
func (p *QueryProcessor) ParseQuery(queryStr string) (*BasicQuery, error) {
	terms := strings.Fields(queryStr)
	var queryTerms []*QueryTerm

	for i := 0; i < len(terms); i++ {
		term := terms[i]

		// Handle boolean operators
		switch strings.ToUpper(term) {
		case "AND":
			if i+1 < len(terms) {
				i++
				queryTerms = append(queryTerms, &QueryTerm{
					Text:     terms[i],
					Type:     TypeSimple,
					Required: true,
				})
			}
			continue
		case "NOT":
			if i+1 < len(terms) {
				i++
				queryTerms = append(queryTerms, &QueryTerm{
					Text:     terms[i],
					Type:     TypeSimple,
					Excluded: true,
				})
			}
			continue
		}

		// Handle phrase matching
		if strings.HasPrefix(term, "\"") {
			phrase := []string{strings.TrimPrefix(term, "\"")}
			for i++; i < len(terms); i++ {
				phrase = append(phrase, terms[i])
				if strings.HasSuffix(terms[i], "\"") {
					phrase[len(phrase)-1] = strings.TrimSuffix(phrase[len(phrase)-1], "\"")
					break
				}
			}
			queryTerms = append(queryTerms, &QueryTerm{
				Text: strings.Join(phrase, " "),
				Type: TypePhrase,
			})
			continue
		}

		// Handle fuzzy matching
		if strings.Contains(term, "~") {
			parts := strings.Split(term, "~")
			fuzziness := 1 // Default fuzziness
			if len(parts) > 1 && parts[1] != "" {
				fuzziness = int(parts[1][0] - '0')
			}
			queryTerms = append(queryTerms, &QueryTerm{
				Text:      parts[0],
				Type:      TypeFuzzy,
				Fuzziness: fuzziness,
			})
			continue
		}

		// Handle field-specific search
		if strings.Contains(term, ":") {
			parts := strings.Split(term, ":")
			queryTerms = append(queryTerms, &QueryTerm{
				Field: parts[0],
				Text:  parts[1],
				Type:  TypeSimple,
			})
			continue
		}

		// Simple term
		queryTerms = append(queryTerms, &QueryTerm{
			Text: term,
			Type: TypeSimple,
		})
	}

	return &BasicQuery{
		queryTerms: queryTerms,
		filters:    make(map[string]interface{}),
		pagination: &Pagination{
			Page: 1,
			Size: 10,
		},
	}, nil
}

// BasicQuery implementation methods
func (q *BasicQuery) Terms() []*QueryTerm {
	return q.queryTerms
}

func (q *BasicQuery) Filters() map[string]interface{} {
	return q.filters
}

func (q *BasicQuery) Pagination() *Pagination {
	return q.pagination
}

func (q *BasicQuery) SetPagination(page, pageSize int) {
	q.pagination = &Pagination{
		Page: page,
		Size: pageSize,
	}
}

func (q *BasicQuery) SetSorting(field string, descending bool) {
	q.SortField = field
	q.SortDescending = descending
}
