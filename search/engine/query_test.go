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

	// Debug: Print initial documents
	t.Log("Initial documents:")
	for _, doc := range docs {
		t.Logf("URL: %s, Title: %s, Content: %s", doc.URL, doc.Title, doc.Content)
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

			// Debug: Print search parameters
			t.Logf("Searching with query: %s", tc.queryStr)
			t.Logf("Expected URLs: %v", tc.expectedURLs)

			results, err := engine.SearchWithOptions(ctx, opts)
			assert.NoError(t, err)
			assert.NotNil(t, results)

			// Debug: Print raw results
			t.Log("Search results:")
			for _, result := range results {
				t.Logf("URL: %s", result.Content["url"])
				t.Logf("Title: %s", result.Content["title"])
				t.Logf("Content: %s", result.Content["content"])
				t.Logf("Score: %f", result.Score)
				t.Log("---")
			}

			// Verify results
			assert.Equal(t, len(tc.expectedURLs), len(results),
				"Expected %d results, got %d", len(tc.expectedURLs), len(results))

			foundURLs := make([]string, len(results))
			for i, result := range results {
				foundURLs[i] = result.Content["url"].(string)
			}

			// Debug: Print URL comparison
			t.Logf("Expected URLs: %v", tc.expectedURLs)
			t.Logf("Found URLs: %v", foundURLs)

			assert.ElementsMatch(t, tc.expectedURLs, foundURLs,
				"Results don't match expected URLs")
		})
	}
}
