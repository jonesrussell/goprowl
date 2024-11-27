package bleve

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/document"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/jonesrussell/goprowl/search/storage"
)

type BleveStorage struct {
	index bleve.Index
	path  string
}

type BleveDocument struct {
	URL       string                 `json:"url"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
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

func (s *BleveStorage) Store(ctx context.Context, doc *storage.Document) error {
	bleveDoc := BleveDocument{
		URL:       doc.URL,
		Title:     doc.Title,
		Content:   doc.Content,
		Type:      doc.Type,
		Metadata:  doc.Metadata,
		CreatedAt: doc.CreatedAt,
	}

	return s.index.Index(doc.URL, bleveDoc)
}

func (s *BleveStorage) Get(ctx context.Context, id string) (*storage.Document, error) {
	doc, err := s.index.Document(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if doc == nil {
		return nil, storage.ErrDocumentNotFound
	}

	bleveDoc := &BleveDocument{}
	docData := make(map[string]interface{})

	// Convert document fields to map
	if doc, ok := doc.(*document.Document); ok {
		for _, field := range doc.Fields {
			switch field := field.(type) {
			case *document.TextField:
				docData[field.Name()] = field.Value()
			case *document.DateTimeField:
				docData[field.Name()] = field.Value()
			case *document.NumericField:
				num, err := field.Number()
				if err != nil {
					return nil, fmt.Errorf("failed to get numeric field value: %w", err)
				}
				docData[field.Name()] = num
			}
		}
	}

	// Marshal and unmarshal to convert the map to struct
	jsonData, err := json.Marshal(docData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document data: %w", err)
	}

	if err := json.Unmarshal(jsonData, bleveDoc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	return &storage.Document{
		URL:       bleveDoc.URL,
		Title:     bleveDoc.Title,
		Content:   bleveDoc.Content,
		Type:      bleveDoc.Type,
		Metadata:  bleveDoc.Metadata,
		CreatedAt: bleveDoc.CreatedAt,
	}, nil
}

func (s *BleveStorage) List(ctx context.Context) ([]*storage.Document, error) {
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Size = 1000 // Adjust as needed

	searchResult, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	docs := make([]*storage.Document, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		doc, err := s.Get(ctx, hit.ID)
		if err != nil {
			continue // Skip failed documents
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

func (s *BleveStorage) Search(ctx context.Context, query string) ([]*storage.Document, error) {
	q := bleve.NewQueryStringQuery(query)
	searchRequest := bleve.NewSearchRequest(q)

	searchResult, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	docs := make([]*storage.Document, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		doc, err := s.Get(ctx, hit.ID)
		if err != nil {
			continue
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

func (s *BleveStorage) Close() error {
	return s.index.Close()
}

func (s *BleveStorage) BatchStore(ctx context.Context, docs []*storage.Document) error {
	batch := s.index.NewBatch()
	for _, doc := range docs {
		bleveDoc := BleveDocument{
			URL:       doc.URL,
			Title:     doc.Title,
			Content:   doc.Content,
			Type:      doc.Type,
			Metadata:  doc.Metadata,
			CreatedAt: doc.CreatedAt,
		}
		if err := batch.Index(doc.URL, bleveDoc); err != nil {
			return fmt.Errorf("failed to add document to batch: %w", err)
		}
	}
	return s.index.Batch(batch)
}

func (s *BleveStorage) Delete(ctx context.Context, id string) error {
	err := s.index.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

func (s *BleveStorage) GetAll(ctx context.Context) ([]*storage.Document, error) {
	// Create a match all query
	query := bleve.NewMatchAllQuery()
	searchRequest := bleve.NewSearchRequest(query)
	// Set a high size to get all documents
	searchRequest.Size = 10000 // Consider implementing pagination for large datasets

	searchResult, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get all documents: %w", err)
	}

	docs := make([]*storage.Document, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		doc, err := s.Get(ctx, hit.ID)
		if err != nil {
			// Log error but continue with other documents
			continue
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// Clear implements the StorageAdapter interface
func (s *BleveStorage) Clear(ctx context.Context) error {
	// Close the current index
	if err := s.index.Close(); err != nil {
		return fmt.Errorf("failed to close index: %w", err)
	}

	// Delete the index directory
	if err := os.RemoveAll(s.path); err != nil {
		return fmt.Errorf("failed to remove index directory: %w", err)
	}

	// Create a new empty index
	mapping := createMapping()
	index, err := bleve.New(s.path, mapping)
	if err != nil {
		return fmt.Errorf("failed to create new index: %w", err)
	}
	s.index = index

	return nil
}
