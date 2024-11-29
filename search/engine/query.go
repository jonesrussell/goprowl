package engine

import (
	"github.com/jonesrussell/goprowl/search/engine/query"
)

// Re-export query types and functions for backward compatibility
type (
	Query       = query.Query
	QueryTerm   = query.QueryTerm
	QueryType   = query.QueryType
	QueryParser = query.QueryProcessor
)

// Constants
const (
	TypeSimple  = query.TypeSimple
	TypePhrase  = query.TypePhrase
	TypeFuzzy   = query.TypeFuzzy
	TypeBoolean = query.TypeBoolean
)

// NewQuery creates a new Query instance
func NewQuery() *Query {
	return query.NewQuery()
}

// NewQueryParser creates a new query parser
func NewQueryParser() *QueryParser {
	return query.NewQueryProcessor()
}
