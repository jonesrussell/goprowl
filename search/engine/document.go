package engine

import (
	"time"

	"github.com/jonesrussell/goprowl/search/core/types"
	"github.com/jonesrussell/goprowl/search/storage"
)

// BasicDocument implements both types.Document and engine.Document interfaces
type BasicDocument struct {
	url        string
	title      string
	content    string
	docType    string
	createdAt  time.Time
	metadata   map[string]interface{}
	permission *Permission
}

// NewBasicDocument creates a new BasicDocument instance
func NewBasicDocument(doc interface{}) *BasicDocument {
	switch d := doc.(type) {
	case types.Document:
		return &BasicDocument{
			url:       d.GetURL(),
			title:     d.GetTitle(),
			content:   d.GetContent(),
			docType:   d.GetType(),
			createdAt: d.GetCreatedAt(),
			metadata:  d.GetMetadata(),
			permission: &Permission{
				Read:  []string{"public"},
				Write: []string{"admin"},
			},
		}
	case *storage.StorageDocument:
		return &BasicDocument{
			url:     d.URL,
			title:   d.Title,
			content: d.Content,
			docType: d.Type,
			metadata: map[string]interface{}{
				"created_at": d.CreatedAt,
			},
			permission: &Permission{
				Read:  []string{"public"},
				Write: []string{"admin"},
			},
		}
	default:
		return nil
	}
}

// Implement types.Document interface
func (d *BasicDocument) GetURL() string                      { return d.url }
func (d *BasicDocument) GetTitle() string                    { return d.title }
func (d *BasicDocument) GetContent() string                  { return d.content }
func (d *BasicDocument) GetType() string                     { return d.docType }
func (d *BasicDocument) GetCreatedAt() time.Time             { return d.createdAt }
func (d *BasicDocument) GetMetadata() map[string]interface{} { return d.metadata }

// Implement engine.Document interface
func (d *BasicDocument) ID() string   { return d.url }
func (d *BasicDocument) Type() string { return d.docType }
func (d *BasicDocument) Content() map[string]interface{} {
	return map[string]interface{}{
		"title":   d.title,
		"content": d.content,
		"url":     d.url,
	}
}
func (d *BasicDocument) Metadata() map[string]interface{} { return d.metadata }
func (d *BasicDocument) Permission() *Permission          { return d.permission }

// Document option functions
func WithExpiration(t time.Time) types.DocumentOption {
	return func(opts *types.DocumentOptions) {
		opts.ExpiresAt = t
	}
}

func WithPriority(priority int) types.DocumentOption {
	return func(opts *types.DocumentOptions) {
		opts.Priority = priority
	}
}

func WithTags(tags []string) types.DocumentOption {
	return func(opts *types.DocumentOptions) {
		opts.Tags = tags
	}
}
