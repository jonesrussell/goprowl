package engine

type BasicQuery struct {
	terms      []string
	filters    map[string]interface{}
	pagination *Page
	sort       []SortField
}

func NewBasicQuery(searchTerms []string) *BasicQuery {
	return &BasicQuery{
		terms:   searchTerms,
		filters: make(map[string]interface{}),
		pagination: &Page{
			Number: 1,
			Size:   10,
		},
		sort: []SortField{},
	}
}

func (q *BasicQuery) Terms() []string                 { return q.terms }
func (q *BasicQuery) Filters() map[string]interface{} { return q.filters }
func (q *BasicQuery) Pagination() *Page               { return q.pagination }
func (q *BasicQuery) Sort() []SortField               { return q.sort }
