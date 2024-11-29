package memory

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/goprowl/search/storage"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	store := New()

	// Test document
	doc := &storage.Document{
		URL:       "test-url",
		Title:     "Test Document",
		Content:   "This is a test document",
		Type:      "article",
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"author": "Test Author",
		},
	}

	// Test Store
	t.Run("Store", func(t *testing.T) {
		err := store.Store(ctx, doc)
		assert.NoError(t, err)

		// Verify storage
		retrieved, err := store.Get(ctx, doc.URL)
		assert.NoError(t, err)
		assert.Equal(t, doc.Title, retrieved.Title)
	})

	// Test GetAll
	t.Run("GetAll", func(t *testing.T) {
		docs, err := store.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, doc.URL, docs[0].URL)
	})

	// Test BatchStore
	t.Run("BatchStore", func(t *testing.T) {
		docs := []*storage.Document{
			{
				URL:     "batch-1",
				Title:   "Batch Doc 1",
				Content: "Batch content 1",
				Type:    "article",
			},
			{
				URL:     "batch-2",
				Title:   "Batch Doc 2",
				Content: "Batch content 2",
				Type:    "article",
			},
		}

		err := store.BatchStore(ctx, docs)
		assert.NoError(t, err)

		// Verify batch storage
		allDocs, err := store.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, allDocs, 3) // 1 original + 2 batch docs
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := store.Delete(ctx, "test-url")
		assert.NoError(t, err)

		// Verify deletion
		retrieved, err := store.Get(ctx, "test-url")
		assert.ErrorIs(t, err, storage.ErrDocumentNotFound)
		assert.Nil(t, retrieved)
	})

	// Test Clear
	t.Run("Clear", func(t *testing.T) {
		err := store.Clear(ctx)
		assert.NoError(t, err)

		// Verify cleared storage
		docs, err := store.GetAll(ctx)
		assert.NoError(t, err)
		assert.Empty(t, docs)
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		// Add a test document
		err := store.Store(ctx, doc)
		assert.NoError(t, err)

		// Get list of documents
		docs, err := store.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, doc.URL, docs[0].URL)
	})
}
