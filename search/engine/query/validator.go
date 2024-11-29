package query

import (
	"fmt"
)

type QueryValidator struct {
	maxTerms    int
	maxLength   int
	validFields []string
}

func NewQueryValidator() *QueryValidator {
	return &QueryValidator{
		maxTerms:    20,                                  // Configurable
		maxLength:   200,                                 // Configurable
		validFields: []string{"title", "content", "url"}, // Add your fields
	}
}

func (v *QueryValidator) Validate(query *Query) error {
	if len(query.Terms) > v.maxTerms {
		return fmt.Errorf("query exceeds maximum of %d terms", v.maxTerms)
	}

	for _, term := range query.Terms {
		if err := v.validateTerm(term); err != nil {
			return fmt.Errorf("invalid term %q: %w", term.Text, err)
		}
	}

	return nil
}

func (v *QueryValidator) validateTerm(term *QueryTerm) error {
	if len(term.Text) > v.maxLength {
		return fmt.Errorf("term exceeds maximum length of %d characters", v.maxLength)
	}

	if term.Field != "" && !v.isValidField(term.Field) {
		return fmt.Errorf("invalid field: %s", term.Field)
	}

	return nil
}

func (v *QueryValidator) isValidField(field string) bool {
	for _, validField := range v.validFields {
		if field == validField {
			return true
		}
	}
	return false
}
