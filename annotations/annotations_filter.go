package annotations

type annotationsFilter interface {
	filter(ann []annotation, chain *annotationsFilterChain) []annotation
}

type annotationsFilterChain struct {
	index   int
	filters []annotationsFilter
}

func newAnnotationsFilterChain(filters ...annotationsFilter) *annotationsFilterChain {
	size := len(filters)
	f := make([]annotationsFilter, size+1)
	copy(f, filters)
	f[size] = defaultDedupFilter
	return &annotationsFilterChain{0, f}
}

func (chain *annotationsFilterChain) doNext(ann []annotation) []annotation {
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

func (f *dedupFilter) filter(in []annotation, chain *annotationsFilterChain) []annotation {
	var out []annotation

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
