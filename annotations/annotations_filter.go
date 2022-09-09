package annotations

type annotationsFilter interface {
	filter(ann []Annotation, chain *annotationsFilterChain) []Annotation
}

type annotationsFilterChain struct {
	index   int
	filters []annotationsFilter
}

func newAnnotationsFilterChain(filters ...annotationsFilter) *annotationsFilterChain {
	size := len(filters)
	f := make([]annotationsFilter, size)
	copy(f, filters)
	// f[size] = defaultDedupFilter
	return &annotationsFilterChain{0, f}
}

func (chain *annotationsFilterChain) doNext(ann []Annotation) []Annotation {
	if chain.index < len(chain.filters) {
		f := chain.filters[chain.index]
		chain.index++

		ann = f.filter(ann, chain)
	}

	return ann
}

type dedupFilter struct {
}

var defaultDedupFilter = &dedupFilter{}

func (f *dedupFilter) filter(in []Annotation, chain *annotationsFilterChain) []Annotation {
	var out []Annotation

OUTER:
	for _, ann := range in {
		for _, copied := range out {
			if copied.Predicate == ann.Predicate && copied.ID == ann.ID {
				continue OUTER
			}
		}

		out = append(out, ann)
	}

	return chain.doNext(out)
}
