package query

import (
	"fmt"
	"strings"

	"github.com/jonesrussell/goprowl/search/core/types"
)

// QueryValidator handles validation of search queries
type QueryValidator struct {
	maxTerms      int
	maxTermLength int
}

// NewQueryValidator creates a new QueryValidator instance
func NewQueryValidator() *QueryValidator {
	return &QueryValidator{
		maxTerms:      10,  // Maximum number of terms in a query
		maxTermLength: 100, // Maximum length of a single term
	}
}

// ValidateQuery validates a query and its terms
func (v *QueryValidator) ValidateQuery(q *types.Query) error {
	if q == nil {
		return fmt.Errorf("query cannot be nil")
	}

	if len(q.Terms) == 0 {
		return fmt.Errorf("query must contain at least one term")
	}

	if len(q.Terms) > v.maxTerms {
		return fmt.Errorf("query exceeds maximum of %d terms", v.maxTerms)
	}

	for _, term := range q.Terms {
		if err := v.validateTerm(term); err != nil {
			return fmt.Errorf("invalid term '%s': %w", term.Text, err)
		}
	}

	return nil
}

// validateTerm validates a single query term
func (v *QueryValidator) validateTerm(term *types.QueryTerm) error {
	if term == nil {
		return fmt.Errorf("term cannot be nil")
	}

	if strings.TrimSpace(term.Text) == "" {
		return fmt.Errorf("term text cannot be empty")
	}

	if len(term.Text) > v.maxTermLength {
		return fmt.Errorf("term exceeds maximum length of %d characters", v.maxTermLength)
	}

	// Validate field if specified
	if term.Field != "" {
		if err := v.validateField(term.Field); err != nil {
			return fmt.Errorf("invalid field: %w", err)
		}
	}

	return nil
}

// validateField validates a field name
func (v *QueryValidator) validateField(field string) error {
	// Add field validation logic here
	// For example, check against a list of valid fields
	validFields := map[string]bool{
		"title":   true,
		"content": true,
		"type":    true,
		"url":     true,
	}

	if !validFields[field] {
		return fmt.Errorf("unknown field '%s'", field)
	}

	return nil
}
