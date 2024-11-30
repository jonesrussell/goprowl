package bleve

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/blevesearch/bleve/v2"
	blevedoc "github.com/blevesearch/bleve/v2/document"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/jonesrussell/goprowl/search/core/types"
	"github.com/jonesrussell/goprowl/search/storage/document"
)

// BleveDocument represents the document structure for Bleve indexing
type BleveDocument struct {
	URL       string
	Title     string
	Content   string
	Type      string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}

type BleveStorage struct {
	index bleve.Index
	path  string
	mu    sync.RWMutex
}

func New(path string) (*BleveStorage, error) {
	// Open or create index
	index, err := bleve.Open(path)
	if err == bleve.ErrorIndexPathDoesNotExist {
		mapping := createMapping()
		index, err = bleve.New(path, mapping)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create/open bleve index: %w", err)
	}
	return &BleveStorage{index: index, path: path}, nil
}

func createMapping() mapping.IndexMapping {
	indexMapping := bleve.NewIndexMapping()

	docMapping := bleve.NewDocumentMapping()

	// Add field mappings
	textFieldMapping := bleve.NewTextFieldMapping()
	dateFieldMapping := bleve.NewDateTimeFieldMapping()

	docMapping.AddFieldMappingsAt("url", textFieldMapping)
	docMapping.AddFieldMappingsAt("title", textFieldMapping)
	docMapping.AddFieldMappingsAt("content", textFieldMapping)
	docMapping.AddFieldMappingsAt("type", textFieldMapping)
	docMapping.AddFieldMappingsAt("created_at", dateFieldMapping)

	indexMapping.AddDocumentMapping("_default", docMapping)

	return indexMapping
}

// Store implements the StorageAdapter interface
func (s *BleveStorage) Store(ctx context.Context, doc types.Document) error {
	done := make(chan error, 1)

	go func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		fields := map[string]interface{}{
			"url":        doc.GetURL(),
			"title":      doc.GetTitle(),
			"content":    doc.GetContent(),
			"type":       doc.GetType(),
			"created_at": doc.GetCreatedAt().Format(time.RFC3339),
		}

		done <- s.index.Index(doc.GetURL(), fields)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("storage operation cancelled: %w", ctx.Err())
	case err := <-done:
		return err
	}
}

// Get retrieves a document by ID
func (s *BleveStorage) Get(ctx context.Context, id string) (types.Document, error) {
	doc, err := s.index.Document(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if doc == nil {
		return nil, types.ErrDocumentNotFound
	}

	// Convert document fields to map
	docData := make(map[string]interface{})
	if doc, ok := doc.(*blevedoc.Document); ok {
		for _, field := range doc.Fields {
			switch field := field.(type) {
			case *blevedoc.TextField:
				docData[field.Name()] = field.Value()
			case *blevedoc.DateTimeField:
				docData[field.Name()] = field.Value()
			case *blevedoc.NumericField:
				num, err := field.Number()
				if err != nil {
					return nil, fmt.Errorf("failed to get numeric field value: %w", err)
				}
				docData[field.Name()] = num
			}
		}
	}

	// Create a new document
	createdAt, _ := time.Parse(time.RFC3339, docData["created_at"].(string))
	return document.NewDocument(
		docData["url"].(string),
		docData["title"].(string),
		docData["content"].(string),
		docData["type"].(string),
		createdAt,
		docData,
	), nil
}

// GetAll retrieves all documents
func (s *BleveStorage) GetAll(ctx context.Context) ([]types.Document, error) {
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 10000 // Consider implementing pagination

	searchResult, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get all documents: %w", err)
	}

	docs := make([]types.Document, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		doc, err := s.Get(ctx, hit.ID)
		if err != nil {
			continue
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

const maxBatchSize = 1000

func (s *BleveStorage) BatchStore(ctx context.Context, docs []types.Document) error {
	for i := 0; i < len(docs); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(docs) {
			end = len(docs)
		}

		batch := s.index.NewBatch()
		for _, doc := range docs[i:end] {
			fields := map[string]interface{}{
				"url":        doc.GetURL(),
				"title":      doc.GetTitle(),
				"content":    doc.GetContent(),
				"type":       doc.GetType(),
				"created_at": doc.GetCreatedAt().Format(time.RFC3339),
			}

			// Add metadata
			for key, value := range doc.GetMetadata() {
				if !isReservedField(key) {
					fields[key] = value
				}
			}

			if err := batch.Index(doc.GetURL(), fields); err != nil {
				return fmt.Errorf("failed to add document to batch: %w", err)
			}
		}

		if err := s.index.Batch(batch); err != nil {
			return fmt.Errorf("batch store failed at index %d: %w", i, err)
		}
	}
	return nil
}

// Delete removes a document by ID
func (s *BleveStorage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.index.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// Clear removes all documents
func (s *BleveStorage) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.index.Close(); err != nil {
		return fmt.Errorf("failed to close index: %w", err)
	}

	if err := os.RemoveAll(s.path); err != nil {
		return fmt.Errorf("failed to remove index directory: %w", err)
	}

	mapping := createMapping()
	index, err := bleve.New(s.path, mapping)
	if err != nil {
		return fmt.Errorf("failed to create new index: %w", err)
	}
	s.index = index

	return nil
}

// Close implements the StorageAdapter interface
func (s *BleveStorage) Close() error {
	return s.index.Close()
}

// Helper function to check if a field name is reserved
func isReservedField(field string) bool {
	reserved := map[string]bool{
		"url":        true,
		"title":      true,
		"content":    true,
		"type":       true,
		"created_at": true,
	}
	return reserved[field]
}
