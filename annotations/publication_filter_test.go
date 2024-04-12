package annotations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	sv = "8e6c705e-1132-42a2-8db0-c295e29e8658"
	st = "6bbd0457-15ab-4ddc-ab82-0cd5b8d9ce19"
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

var annotationD = Annotation{
	ID:          "5f2584bf-7f40-4513-94b5-dd340e572996",
	Predicate:   ABOUT,
	Publication: nil,
}

func TestPublicationFiltering(t *testing.T) {
	tests := map[string]struct {
		publication []string
		expected    []Annotation
	}{
		"Filter by FT Pink publication": {
			publication: []string{ftPink},
			expected:    []Annotation{annotationB, annotationD},
		},
		"Filter by SV publication": {
			publication: []string{sv},
			expected:    []Annotation{annotationA, annotationC},
		},
		"Filter by SV and FT Pink publication": {
			publication: []string{sv, ftPink},
			expected:    []Annotation{annotationA, annotationB, annotationC, annotationD},
		},
		"No publication filter applied": {
			publication: []string{},
			expected:    []Annotation{annotationA, annotationB, annotationC, annotationD},
		},
		"Unknown publication filter applied": {
			publication: []string{"unknown"},
			expected:    nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			annotations := []Annotation{annotationA, annotationB, annotationC, annotationD}
			f := newPublicationFilter(withPublication(tc.publication, true))
			chain := newAnnotationsFilterChain(f)
			filtered := chain.doNext(annotations)

			assert.Len(t, filtered, len(tc.expected))
			assert.Equal(t, filtered, tc.expected)
		})
	}
}
