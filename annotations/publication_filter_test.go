package annotations

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	sv     = "8e6c705e-1132-42a2-8db0-c295e29e8658"
	ftPink = "88fdde6c-2aa4-4f78-af02-9f680097cfd6"
	st     = "6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce19"
)

var annotationA = Annotation{
	ID:          "6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce18",
	Predicate:   ABOUT,
	Publication: []string{sv},
}

var annotationB = Annotation{
	ID:          "0ab61bfc-a2b1-4b08-a864-4233fd72f250",
	Predicate:   MENTIONS,
	Publication: []string{ftPink},
}

var annotationC = Annotation{
	ID:          "a0076026-f2e5-414f-b7a0-419bc16c4c51",
	Predicate:   ABOUT,
	Publication: []string{sv, st},
}

func TestPublicationFiltering(t *testing.T) {
	tests := map[string]struct {
		publication []string
		expected    []Annotation
	}{
		"Filter by FT Pink publication": {
			publication: []string{ftPink},
			expected:    []Annotation{annotationB},
		},
		"Filter by SV publication": {
			publication: []string{sv},
			expected:    []Annotation{annotationA, annotationC},
		},
		"Filter by SV and FT Pink publication": {
			publication: []string{sv, ftPink},
			expected:    []Annotation{annotationA, annotationB, annotationC},
		},
		"No publication filter applied": {
			publication: []string{},
			expected:    []Annotation{annotationA, annotationB, annotationC},
		},
		"Unknown publication filter applied": {
			publication: []string{"unknown"},
			expected:    nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			annotations := []Annotation{annotationA, annotationB, annotationC}
			f := newPublicationFilter(withPublication(tc.publication))
			chain := newAnnotationsFilterChain(f)
			filtered := chain.doNext(annotations)

			assert.Len(t, filtered, len(tc.expected))
			assert.Equal(t, filtered, tc.expected)
		})
	}
}
