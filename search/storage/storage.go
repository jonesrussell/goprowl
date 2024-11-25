package storage

// StorageAdapter defines the interface for storage implementations
type StorageAdapter interface {
	// Store saves a document to storage
	Store(id string, data map[string]interface{}) error

	// Get retrieves a document by ID
	Get(id string) (map[string]interface{}, error)

	// Delete removes a document from storage
	Delete(id string) error

	// List returns all document IDs
	List() ([]string, error)

	// Search performs a basic search operation
	Search(query map[string]interface{}) ([]map[string]interface{}, error)
}
