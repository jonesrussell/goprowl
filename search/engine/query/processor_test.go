package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryProcessor(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedTerms []struct {
			text     string
			required bool
			termType QueryType
		}
		hasAndOperator bool
	}{
		{
			name:  "Simple term",
			input: "go",
			expectedTerms: []struct {
				text     string
				required bool
				termType QueryType
			}{
				{text: "go", required: true, termType: TypeSimple},
			},
			hasAndOperator: false,
		},
		{
			name:  "Phrase search",
			input: "\"go programming\"",
			expectedTerms: []struct {
				text     string
				required bool
				termType QueryType
			}{
				{text: "go programming", required: true, termType: TypePhrase},
			},
			hasAndOperator: false,
		},
		{
			name:  "Boolean AND search",
			input: "go AND web",
			expectedTerms: []struct {
				text     string
				required bool
				termType QueryType
			}{
				{text: "go", required: true, termType: TypeSimple},
				{text: "web", required: true, termType: TypeSimple},
			},
			hasAndOperator: true,
		},
	}

	processor := NewQueryProcessor()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := processor.ParseQuery(tc.input)
			assert.NoError(t, err)
			assert.NotNil(t, query)

			// Verify AND operator
			assert.Equal(t, tc.hasAndOperator, query.HasAndOperator)

			// Verify terms
			assert.Equal(t, len(tc.expectedTerms), len(query.Terms))
			for i, expectedTerm := range tc.expectedTerms {
				assert.Equal(t, expectedTerm.text, query.Terms[i].Text)
				assert.Equal(t, expectedTerm.required, query.Terms[i].Required)
				assert.Equal(t, expectedTerm.termType, query.Terms[i].Type)
			}
		})
	}
}

func TestSplitKeepingQuotes(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple terms",
			input:    "go web",
			expected: []string{"go", "web"},
		},
		{
			name:     "Quoted phrase",
			input:    "\"go programming\" web",
			expected: []string{"\"go programming\"", "web"},
		},
		{
			name:     "AND operator",
			input:    "go AND web",
			expected: []string{"go", "AND", "web"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitKeepingQuotes(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
