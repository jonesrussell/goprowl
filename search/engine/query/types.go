package query

// Query represents a search query
type Query struct {
	Terms          []*QueryTerm
	Page           int
	PageSize       int
	Filters        map[string]interface{}
	HasAndOperator bool // Indicates if terms are combined with AND
}

// QueryTerm represents a single term in a search query
type QueryTerm struct {
	Text     string
	Type     QueryType
	Required bool
}

// QueryType represents different types of query terms
type QueryType int

const (
	TypeSimple QueryType = iota
	TypePhrase
	TypeFuzzy
	TypeBoolean
)

// NewQuery creates a new Query instance with default values
func NewQuery() *Query {
	return &Query{
		Terms:          make([]*QueryTerm, 0),
		Page:           1,
		PageSize:       10,
		Filters:        make(map[string]interface{}),
		HasAndOperator: false,
	}
}

// AddTerm adds a new term to the query
func (q *Query) AddTerm(term *QueryTerm) {
	q.Terms = append(q.Terms, term)
	// Update HasAndOperator if this term is required and not the first term
	if term.Required && len(q.Terms) > 1 {
		q.HasAndOperator = true
	}
}

// SetPage sets the page number for pagination
func (q *Query) SetPage(page int) {
	if page < 1 {
		page = 1
	}
	q.Page = page
}

// SetPageSize sets the page size for pagination
func (q *Query) SetPageSize(size int) {
	if size < 1 {
		size = 10
	}
	q.PageSize = size
}

// AddFilter adds a filter to the query
func (q *Query) AddFilter(key string, value interface{}) {
	q.Filters[key] = value
}
