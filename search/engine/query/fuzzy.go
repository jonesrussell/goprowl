package query

import (
	"strconv"
	"strings"
)

type FuzzyMatcher struct {
	maxDistance int
}

func NewFuzzyMatcher(maxDistance int) *FuzzyMatcher {
	return &FuzzyMatcher{
		maxDistance: maxDistance,
	}
}

func (f *FuzzyMatcher) AddFuzzySupport(term *QueryTerm) {
	// Check for fuzzy syntax (e.g., "term~2")
	if strings.Contains(term.Text, "~") {
		parts := strings.Split(term.Text, "~")
		if len(parts) == 2 {
			term.Text = parts[0]
			if distance, err := strconv.Atoi(parts[1]); err == nil {
				term.FuzzyDistance = min(distance, f.maxDistance)
			}
		}
	}
}
