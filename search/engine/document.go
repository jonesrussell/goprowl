package engine

// BaseDocument provides a basic implementation of the Document interface
type BaseDocument struct {
	id         string
	docType    string
	content    map[string]interface{}
	metadata   map[string]interface{}
	permission *Permission
}

func NewDocument(id, docType string) *BaseDocument {
	return &BaseDocument{
		id:       id,
		docType:  docType,
		content:  make(map[string]interface{}),
		metadata: make(map[string]interface{}),
		permission: &Permission{
			Read:  []string{},
			Write: []string{},
		},
	}
}

// Implementation of Document interface
func (d *BaseDocument) ID() string                       { return d.id }
func (d *BaseDocument) Type() string                     { return d.docType }
func (d *BaseDocument) Content() map[string]interface{}  { return d.content }
func (d *BaseDocument) Metadata() map[string]interface{} { return d.metadata }
func (d *BaseDocument) Permission() *Permission          { return d.permission }

// Helper methods for document manipulation
func (d *BaseDocument) SetContent(key string, value interface{}) {
	d.content[key] = value
}

func (d *BaseDocument) SetMetadata(key string, value interface{}) {
	d.metadata[key] = value
}

func (d *BaseDocument) AddReadPermission(role string) {
	d.permission.Read = append(d.permission.Read, role)
}

func (d *BaseDocument) AddWritePermission(role string) {
	d.permission.Write = append(d.permission.Write, role)
}
