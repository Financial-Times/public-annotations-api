package annotations

import (
	"strings"
)

// Defines all names of predicates that have to be considered by the annotation filter.
// Predicates that are not defined in FilteredPredicateNames are not filtered.

const (
	Mentions                = "http://www.ft.com/ontology/annotation/mentions"
	MajorMentions           = "http://www.ft.com/ontology/annotation/majormentions"
	About                   = "http://www.ft.com/ontology/annotation/about"
	IsClassifiedBy          = "http://www.ft.com/ontology/classification/isclassifiedby"
	IsPrimarilyClassifiedBy = "http://www.ft.com/ontology/classification/isprimarilyclassifiedby"
	ImplicitlyClassifiedBy  = "http://www.ft.com/ontology/implicitlyclassifiedby"
	HasBrand                = "http://www.ft.com/ontology/classification/isclassifiedby"
)

type PredicateFilter struct {
	// Definition of predicate groups to whom Rule of Importance should be applied.
	// Each group contains a list of predicate names in the order of increasing importance.
	ImportanceRuleConfig [][]string
	// Predicate names of annotations that should be considered for filtering
	enum []string
	// Stores annotations to be filtered keyed by concept ID (uuid).
	unfilteredAnnotations map[string][]Annotation
	// Stores annotations not to be filtered keyed by concept ID (uuid).
	filteredAnnotations map[string][]Annotation
}

func NewAnnotationsPredicateFilter() *PredicateFilter {
	return &PredicateFilter{
		enum: []string{
			Mentions,
			MajorMentions,
			About,
			IsClassifiedBy,
			HasBrand,
			ImplicitlyClassifiedBy,
			IsPrimarilyClassifiedBy,
		},
		// Configure groups of predicates that should be filtered according to their importance.
		ImportanceRuleConfig: [][]string{
			{
				Mentions,
				MajorMentions,
				About,
			},
			{
				ImplicitlyClassifiedBy,
				HasBrand,
				IsClassifiedBy,
				IsPrimarilyClassifiedBy,
			},
		},
		filteredAnnotations:   make(map[string][]Annotation),
		unfilteredAnnotations: make(map[string][]Annotation),
	}
}

func (f *PredicateFilter) FilterAnnotations(annotations []Annotation) {
	for _, ann := range annotations {
		f.Add(ann)
	}
}

func (f *PredicateFilter) Add(a Annotation) {
	pred := strings.ToLower(a.Predicate)
	for _, p := range f.enum {
		if p == pred {
			f.addFiltered(a)
			return
		}
	}

	f.addUnfiltered(a)
}

func (f *PredicateFilter) ProduceResponseList() []Annotation {
	out := []Annotation{}

	for _, allFiltered := range f.filteredAnnotations {
		for _, a := range allFiltered {
			if a.ID != "" {
				out = append(out, a)
			}
		}
	}

	for _, allUnfiltered := range f.unfilteredAnnotations {
		out = append(out, allUnfiltered...)
	}
	return out
}

func (f *PredicateFilter) addFiltered(a Annotation) {
	if f.filteredAnnotations[a.ID] == nil {
		// For each importance group we shell store 1 most important annotation
		f.filteredAnnotations[a.ID] = make([]Annotation, len(f.ImportanceRuleConfig))
	}
	grpID, pos := f.getGroupIDAndImportanceValue(strings.ToLower(a.Predicate))
	if grpID == -1 || pos == -1 {
		return
	}
	arr := f.filteredAnnotations[a.ID]
	prevAnnotation := arr[grpID]
	// Empty value indicates we have not seen annotations for this group before.
	if prevAnnotation.ID == "" {
		f.filteredAnnotations[a.ID][grpID] = a
	} else {
		prevPos := f.getImportanceValueForGroupID(strings.ToLower(prevAnnotation.Predicate), grpID)
		if prevPos < pos {
			f.filteredAnnotations[a.ID][grpID] = a
		}
	}
}

func (f *PredicateFilter) addUnfiltered(a Annotation) {
	if f.unfilteredAnnotations[a.ID] == nil {
		f.unfilteredAnnotations[a.ID] = []Annotation{}
	}
	f.unfilteredAnnotations[a.ID] = append(f.unfilteredAnnotations[a.ID], a)
}

func (f *PredicateFilter) getGroupIDAndImportanceValue(predicate string) (int, int) {
	for group, s := range f.ImportanceRuleConfig {
		for pos, val := range s {
			if val == predicate {
				return group, pos
			}
		}
	}
	//should not occur in normal circumstances
	return -1, -1
}

func (f *PredicateFilter) getImportanceValueForGroupID(predicate string, groupID int) int {
	for pos, val := range f.ImportanceRuleConfig[groupID] {
		if val == predicate {
			return pos
		}
	}
	//should not occur in normal circumstances
	return -1
}

func (f *PredicateFilter) filter(in []Annotation, chain *annotationsFilterChain) []Annotation {
	f.FilterAnnotations(in)
	return chain.doNext(f.ProduceResponseList())
}
