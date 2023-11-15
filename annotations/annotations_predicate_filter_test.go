package annotations

import (
	"reflect"
	"sort"
	"testing"
)

const (
	MENTIONS                = "http://www.ft.com/ontology/annotation/mentions"
	MAJORMENTIONS           = "http://www.ft.com/ontology/annotation/majormentions"
	ABOUT                   = "http://www.ft.com/ontology/annotation/about"
	HASBRAND                = "http://www.ft.com/ontology/classification/isclassifiedby"
	ISCLASSIFIEDBY          = "http://www.ft.com/ontology/classification/isclassifiedby"
	IMPLICITLYCLASSIFIEDBY  = "http://www.ft.com/ontology/implicitlyclassifiedby"
	ISPRIMARILYCLASSIFIEDBY = "http://www.ft.com/ontology/classification/isprimarilyclassifiedby"
	HASAUTHOR               = "http://www.ft.com/ontology/annotation/hasauthor"
	ConceptA                = "1a2359b1-9326-4b80-9b97-2a91ccd68d23"
	ConceptB                = "2f1fead1-5e99-4e92-b23d-fb3cee7f17f2"
)

// Test case definitions taken from https://www.lucidchart.com/documents/edit/df1fead1-5e99-4e92-b23d-fb3cee7f17f2/1?kme=Clicked%20E-mail%20Link&kmi=julia.fernee@ft.com&km_Link=DocInviteButton&km_DocInviteUserArm=T-B
var tests = map[string]struct {
	input          []Annotation
	expectedOutput []Annotation
}{

	"1. Returns one occurrence of Mentions for this concept": {
		[]Annotation{
			{Predicate: MENTIONS, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: MENTIONS, ID: ConceptA},
		},
	},
	"2. Returns one occurrence of Major Mentions for this concept": {
		[]Annotation{
			{Predicate: MAJORMENTIONS, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: MAJORMENTIONS, ID: ConceptA},
		},
	},
	"3. Returns one occurrence of About for this concept": {
		[]Annotation{
			{Predicate: MAJORMENTIONS, ID: ConceptA},
			{Predicate: ABOUT, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
		},
	},
	"4. Returns one occurrence of About for this concept": {
		[]Annotation{
			{Predicate: MENTIONS, ID: ConceptA},
			{Predicate: MAJORMENTIONS, ID: ConceptA},
			{Predicate: ABOUT, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
		},
	},
	"5. Returns one occurrence of Is Classified By for this concept": {
		[]Annotation{
			{Predicate: ISCLASSIFIEDBY, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ISCLASSIFIEDBY, ID: ConceptA},
		},
	},
	"6. Returns one occurrence of Is Primarily Classified By for this concept": {
		[]Annotation{
			{Predicate: ISPRIMARILYCLASSIFIEDBY, ID: ConceptA},
			{Predicate: ISCLASSIFIEDBY, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ISPRIMARILYCLASSIFIEDBY, ID: ConceptA},
		},
	},
	"7. Returns Has Author & Major Mentions for this concept": {
		[]Annotation{

			{Predicate: MAJORMENTIONS, ID: ConceptA},
			{Predicate: HASAUTHOR, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: MAJORMENTIONS, ID: ConceptA},
			{Predicate: HASAUTHOR, ID: ConceptA},
		},
	},
	"8. Returns Has Author & About for this concept": {
		[]Annotation{

			{Predicate: ABOUT, ID: ConceptA},
			{Predicate: MAJORMENTIONS, ID: ConceptA},
			{Predicate: HASAUTHOR, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
			{Predicate: HASAUTHOR, ID: ConceptA},
		},
	},
	"9. Returns About for this concept": {
		[]Annotation{

			{Predicate: ABOUT, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
		},
	},
	"10. Returns About for this concept": {
		[]Annotation{
			{Predicate: MENTIONS, ID: ConceptA},
			{Predicate: ABOUT, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
		},
	},
	"11. Returns one occurrence of Is Primarily Classified By for this concept": {
		[]Annotation{
			{Predicate: ISPRIMARILYCLASSIFIEDBY, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ISPRIMARILYCLASSIFIEDBY, ID: ConceptA},
		},
	},
	"12. Returns About annotation for one concept and Mentions annotations for another": {
		[]Annotation{
			{Predicate: MAJORMENTIONS, ID: ConceptA},
			{Predicate: ABOUT, ID: ConceptA},
			{Predicate: MENTIONS, ID: ConceptB},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
			{Predicate: MENTIONS, ID: ConceptB},
		},
	},
	"13. Returns Is Primarily Classified By annotation for one concept and Is Classified By annotations for another": {
		[]Annotation{
			{Predicate: ISCLASSIFIEDBY, ID: ConceptA},
			{Predicate: ISPRIMARILYCLASSIFIEDBY, ID: ConceptA},
			{Predicate: ISCLASSIFIEDBY, ID: ConceptB},
		},
		[]Annotation{
			{Predicate: ISPRIMARILYCLASSIFIEDBY, ID: ConceptA},
			{Predicate: ISCLASSIFIEDBY, ID: ConceptB},
		},
	},
	"14. IsClassifiedBy should be with highest priority": {
		input: []Annotation{
			{Predicate: ISCLASSIFIEDBY, ID: ConceptA},
			{Predicate: IMPLICITLYCLASSIFIEDBY, ID: ConceptA},
			{Predicate: HASBRAND, ID: ConceptA},
		},
		expectedOutput: []Annotation{
			{Predicate: ISCLASSIFIEDBY, ID: ConceptA},
		},
	},
	"15. HasBrand should be with higher priority than ImplicitlyClassifiedBy": {
		input: []Annotation{
			{Predicate: HASBRAND, ID: ConceptA},
			{Predicate: IMPLICITLYCLASSIFIEDBY, ID: ConceptA},
		},
		expectedOutput: []Annotation{
			{Predicate: HASBRAND, ID: ConceptA},
		},
	},
	"16. Returns one occurrence of Implicitly Classified By for one concept": {
		input: []Annotation{
			{Predicate: IMPLICITLYCLASSIFIEDBY, ID: ConceptA},
		},
		expectedOutput: []Annotation{
			{Predicate: IMPLICITLYCLASSIFIEDBY, ID: ConceptA},
		},
	},
}

func TestFilterForBasicSingleConcept(t *testing.T) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			filter := NewAnnotationsPredicateFilter()
			chain := newAnnotationsFilterChain(filter)
			actualOutput := chain.doNext(test.input)

			By(byUUID).Sort(test.expectedOutput)
			By(byUUID).Sort(actualOutput)

			if !reflect.DeepEqual(test.expectedOutput, actualOutput) {
				t.Fatalf("Expected %d annotations but returned %d.", len(test.expectedOutput), len(actualOutput))
			}
		})
	}
}

// Tests support for sort needed by other tests in order to compare 2 arrays of annotations
func TestSortAnnotations(t *testing.T) {
	expected := []Annotation{
		{Predicate: ISCLASSIFIEDBY, ID: "1"},
		{Predicate: ISPRIMARILYCLASSIFIEDBY, ID: "2"},
	}
	test := []Annotation{
		{Predicate: ISPRIMARILYCLASSIFIEDBY, ID: "2"},
		{Predicate: ISCLASSIFIEDBY, ID: "1"},
	}

	By(byUUID).Sort(test)
	if !reflect.DeepEqual(expected, test) {
		t.Fatal("Expected input to be equal to output")
	}
}

// Implementation of sort for an array of structs in order to compare equality of 2 arrays of annotations
type By func(p1, p2 *Annotation) bool

type AnnotationSorter struct {
	annotations []Annotation
	by          func(a1, a2 *Annotation) bool
}

func (by By) Sort(unsorted []Annotation) {
	sorter := &AnnotationSorter{
		annotations: unsorted,
		by:          by,
	}
	sort.Sort(sorter)
}

func (s *AnnotationSorter) Len() int {
	return len(s.annotations)
}

func (s *AnnotationSorter) Swap(i, j int) {
	s.annotations[i], s.annotations[j] = s.annotations[j], s.annotations[i]
}

func (s *AnnotationSorter) Less(i, j int) bool {
	return s.by(&s.annotations[i], &s.annotations[j])
}

func byUUID(a1, a2 *Annotation) bool {
	return a1.ID < a2.ID
}
