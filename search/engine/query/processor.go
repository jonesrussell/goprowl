package query

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

type QueryTerm struct {
	Text      string
	Field     string
	Type      QueryType
	Fuzziness int     // For fuzzy matching
	Required  bool    // For boolean AND
	Excluded  bool    // For boolean NOT
	Boost     float64 // Term importance
}

type QueryProcessor struct {
	// Remove unused fields and add any necessary state here
}

func NewQueryProcessor() *QueryProcessor {
	return &QueryProcessor{}
}

// ParseQuery is a more descriptive name than just Parse
func (p *QueryProcessor) ParseQuery(queryString string) ([]*QueryTerm, error) {
	terms := strings.Fields(queryString)
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
					Required: true,
					Type:     TypeSimple,
				})
			}
			continue
		case "OR":
			continue
		case "NOT":
			if i+1 < len(terms) {
				i++
				queryTerms = append(queryTerms, &QueryTerm{
					Text:     terms[i],
					Excluded: true,
					Type:     TypeSimple,
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

	return queryTerms, nil
}
