package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Add table-driven tests with subtests
func TestSearch(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		want    *SearchResults
		wantErr error
	}{
		// ... test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			engine := newTestEngine(t)
			got, err := engine.Search(ctx, tt.query)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Search() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
