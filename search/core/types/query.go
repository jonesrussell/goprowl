package types

// QueryType represents the type of query term
type QueryType int

const (
	TypeSimple QueryType = iota
	TypePhrase
	TypeFuzzy
	TypeBoolean
)

// QueryTerm represents a single term in a search query
type QueryTerm struct {
	Text          string
	Type          QueryType
	Field         string
	Required      bool
	Excluded      bool
	Optional      bool
	FuzzyDistance int
}

// Query represents a parsed search query
type Query struct {
	Terms          []*QueryTerm
	HasAndOperator bool
	Page           int
	PageSize       int
	Filters        map[string]interface{}
}

// NewQuery creates a new Query instance
func NewQuery() *Query {
	return &Query{
		Terms:    make([]*QueryTerm, 0),
		Filters:  make(map[string]interface{}),
		Page:     1,
		PageSize: 10,
	}
}

// AddTerm adds a term to the query
func (q *Query) AddTerm(term *QueryTerm) {
	q.Terms = append(q.Terms, term)
	if term.Required && len(q.Terms) > 1 {
		q.HasAndOperator = true
	}
}
