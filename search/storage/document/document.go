package document

import (
	"time"
)

// Document implements types.Document interface
type Document struct {
	URL       string
	Title     string
	Content   string
	Type      string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}

// NewDocument creates a new Document
func NewDocument(url, title, content, docType string, createdAt time.Time, metadata map[string]interface{}) *Document {
	return &Document{
		URL:       url,
		Title:     title,
		Content:   content,
		Type:      docType,
		CreatedAt: createdAt,
		Metadata:  metadata,
	}
}

// Interface implementation methods
func (d *Document) GetURL() string                      { return d.URL }
func (d *Document) GetTitle() string                    { return d.Title }
func (d *Document) GetContent() string                  { return d.Content }
func (d *Document) GetType() string                     { return d.Type }
func (d *Document) GetCreatedAt() time.Time             { return d.CreatedAt }
func (d *Document) GetMetadata() map[string]interface{} { return d.Metadata }
