package annotations

import (
	"slices"
)

const (
	ftPink = "88fdde6c-2aa4-4f78-af02-9f680097cfd6"
)

type publicationFilter struct {
	publication     []string
	showPublication bool
}

func newPublicationFilter(opts ...func(*publicationFilter)) *publicationFilter {
	pf := publicationFilter{}
	for _, opt := range opts {
		opt(&pf)
	}

	return &pf
}

func withPublication(publication []string, showPublication bool) func(filter *publicationFilter) {
	return func(f *publicationFilter) {
		f.showPublication = showPublication
		f.publication = publication
		if len(f.publication) == 0 {
			f.publication = append(f.publication, ftPink)
		}
	}
}

func (f *publicationFilter) filter(in []Annotation, chain *annotationsFilterChain) []Annotation {
	return chain.doNext(f.filterByPublication(in))
}

func (f *publicationFilter) filterByPublication(annotations []Annotation) []Annotation {
	if len(f.publication) == 0 {
		return annotations
	}

	var filtered []Annotation
	for _, annotation := range annotations {
		for _, pub := range f.publication {
			if slices.Contains(annotation.Publication, pub) {
				filtered = append(filtered, annotation)
			}

			if pub == ftPink {
				if annotation.Publication == nil {
					if annotation.Lifecycle == pacLifecycle {
						annotation.Publication = append(annotation.Publication, ftPink)
					}
					filtered = append(filtered, annotation)
				}
			}
		}
	}

	if !f.showPublication {
		for i := range filtered {
			filtered[i].Publication = nil
		}
	}

	return filtered
}
