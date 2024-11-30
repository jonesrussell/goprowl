package memory

import (
	"context"
	"testing"
	"time"

	"github.com/jonesrussell/goprowl/search/core/types"
	"github.com/jonesrussell/goprowl/search/storage/document"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage(t *testing.T) {
	ctx := context.Background()
	store := New()

	// Test document
	doc := document.NewDocument(
		"test-url",
		"Test Document",
		"This is a test document",
		"article",
		time.Now(),
		map[string]interface{}{
			"author": "Test Author",
		},
	)

	// Test Store
	t.Run("Store", func(t *testing.T) {
		err := store.Store(ctx, doc)
		assert.NoError(t, err)

		// Verify storage
		retrieved, err := store.Get(ctx, doc.GetURL())
		assert.NoError(t, err)
		assert.Equal(t, doc.GetTitle(), retrieved.GetTitle())
	})

	// Test GetAll
	t.Run("GetAll", func(t *testing.T) {
		docs, err := store.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, doc.GetURL(), docs[0].GetURL())
	})

	// Test BatchStore
	t.Run("BatchStore", func(t *testing.T) {
		docs := []types.Document{
			document.NewDocument(
				"batch-1",
				"Batch Doc 1",
				"Batch content 1",
				"article",
				time.Now(),
				nil,
			),
			document.NewDocument(
				"batch-2",
				"Batch Doc 2",
				"Batch content 2",
				"article",
				time.Now(),
				nil,
			),
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
		assert.ErrorIs(t, err, types.ErrDocumentNotFound)
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
		assert.Equal(t, doc.GetURL(), docs[0].GetURL())
	})
}
