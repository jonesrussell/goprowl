package engine

import (
	"context"
	"testing"

	"github.com/jonesrussell/goprowl/search/storage"
	memorystorage "github.com/jonesrussell/goprowl/search/storage/memory"
	"github.com/stretchr/testify/assert"
)

func TestSearchEngineIntegration(t *testing.T) {
	// Create test documents
	docs := []*storage.Document{
		{
			URL:     "1",
			Title:   "Go Programming Guide",
			Content: "Learn how to program in Go programming. Go is fast and efficient.",
			Type:    "article",
		},
		{
			URL:     "2",
			Title:   "Python vs Go",
			Content: "Comparing Python and Go programming languages.",
			Type:    "article",
		},
		{
			URL:     "3",
			Title:   "Web Development",
			Content: "Building web applications using Go frameworks.",
			Type:    "article",
		},
	}

	// Create a test engine with memory storage
	memStore := memorystorage.New()
	for _, doc := range docs {
		err := memStore.Store(context.Background(), doc)
		assert.NoError(t, err)
	}

	engine, err := New(memStore)
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		queryStr     string
		expectedURLs []string
	}{
		{
			name:         "Simple term search",
			queryStr:     "go",
			expectedURLs: []string{"1", "2", "3"},
		},
		{
			name:         "Phrase search",
			queryStr:     "\"go programming\"",
			expectedURLs: []string{"1", "2"},
		},
		{
			name:         "Boolean AND search",
			queryStr:     "go AND web",
			expectedURLs: []string{"3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			opts := SearchOptions{
				QueryString: tc.queryStr,
				Page:        1,
				PageSize:    10,
			}

			results, err := engine.SearchWithOptions(ctx, opts)
			assert.NoError(t, err)
			assert.NotNil(t, results)

			// Verify results
			assert.Equal(t, len(tc.expectedURLs), len(results))
			foundURLs := make([]string, len(results))
			for i, result := range results {
				foundURLs[i] = result.Content["url"].(string)
			}
			assert.ElementsMatch(t, tc.expectedURLs, foundURLs)
		})
	}
}
