package annotations

import (
	"reflect"
	"sort"
	"testing"
)

const (
	MENTIONS                   = "http://www.ft.com/ontology/annotation/mentions"
	MAJOR_MENTIONS             = "http://www.ft.com/ontology/annotation/majormentions"
	ABOUT                      = "http://www.ft.com/ontology/annotation/about"
	HAS_BRAND                  = "http://www.ft.com/ontology/classification/isclassifiedby"
	IS_CLASSIFIED_BY           = "http://www.ft.com/ontology/classification/isclassifiedby"
	IMPLICITLY_CLASSIFIED_BY   = "http://www.ft.com/ontology/implicitlyclassifiedby"
	IS_PRIMARILY_CLASSIFIED_BY = "http://www.ft.com/ontology/classification/isprimarilyclassifiedby"
	HAS_AUTHOR                 = "http://www.ft.com/ontology/annotation/hasauthor"
	ConceptA                   = "1a2359b1-9326-4b80-9b97-2a91ccd68d23"
	ConceptB                   = "2f1fead1-5e99-4e92-b23d-fb3cee7f17f2"
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
			{Predicate: MAJOR_MENTIONS, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: MAJOR_MENTIONS, ID: ConceptA},
		},
	},
	"3. Returns one occurrence of About for this concept": {
		[]Annotation{
			{Predicate: MAJOR_MENTIONS, ID: ConceptA},
			{Predicate: ABOUT, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
		},
	},
	"4. Returns one occurrence of About for this concept": {
		[]Annotation{
			{Predicate: MENTIONS, ID: ConceptA},
			{Predicate: MAJOR_MENTIONS, ID: ConceptA},
			{Predicate: ABOUT, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
		},
	},
	"5. Returns one occurrence of Is Classified By for this concept": {
		[]Annotation{
			{Predicate: IS_CLASSIFIED_BY, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: IS_CLASSIFIED_BY, ID: ConceptA},
		},
	},
	"6. Returns one occurrence of Is Primarily Classified By for this concept": {
		[]Annotation{
			{Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA},
			{Predicate: IS_CLASSIFIED_BY, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA},
		},
	},
	"7. Returns Has Author & Major Mentions for this concept": {
		[]Annotation{

			{Predicate: MAJOR_MENTIONS, ID: ConceptA},
			{Predicate: HAS_AUTHOR, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: MAJOR_MENTIONS, ID: ConceptA},
			{Predicate: HAS_AUTHOR, ID: ConceptA},
		},
	},
	"8. Returns Has Author & About for this concept": {
		[]Annotation{

			{Predicate: ABOUT, ID: ConceptA},
			{Predicate: MAJOR_MENTIONS, ID: ConceptA},
			{Predicate: HAS_AUTHOR, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: ABOUT, ID: ConceptA},
			{Predicate: HAS_AUTHOR, ID: ConceptA},
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
			{Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA},
		},
		[]Annotation{
			{Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA},
		},
	},
	"12. Returns About annotation for one concept and Mentions annotations for another": {
		[]Annotation{
			{Predicate: MAJOR_MENTIONS, ID: ConceptA},
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
			{Predicate: IS_CLASSIFIED_BY, ID: ConceptA},
			{Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA},
			{Predicate: IS_CLASSIFIED_BY, ID: ConceptB},
		},
		[]Annotation{
			{Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: ConceptA},
			{Predicate: IS_CLASSIFIED_BY, ID: ConceptB},
		},
	},
	"14. IsClassifiedBy should be with highest priority": {
		input: []Annotation{
			{Predicate: IS_CLASSIFIED_BY, ID: ConceptA},
			{Predicate: IMPLICITLY_CLASSIFIED_BY, ID: ConceptA},
			{Predicate: HAS_BRAND, ID: ConceptA},
		},
		expectedOutput: []Annotation{
			{Predicate: IS_CLASSIFIED_BY, ID: ConceptA},
		},
	},
	"15. HasBrand should be with higher priority than ImplicitlyClassifiedBy": {
		input: []Annotation{
			{Predicate: HAS_BRAND, ID: ConceptA},
			{Predicate: IMPLICITLY_CLASSIFIED_BY, ID: ConceptA},
		},
		expectedOutput: []Annotation{
			{Predicate: HAS_BRAND, ID: ConceptA},
		},
	},
	"16. Returns one occurrence of Implicitly Classified By for one concept": {
		input: []Annotation{
			{Predicate: IMPLICITLY_CLASSIFIED_BY, ID: ConceptA},
		},
		expectedOutput: []Annotation{
			{Predicate: IMPLICITLY_CLASSIFIED_BY, ID: ConceptA},
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

//Tests support for sort needed by other tests in order to compare 2 arrays of annotations
func TestSortAnnotations(t *testing.T) {
	expected := []Annotation{
		{Predicate: IS_CLASSIFIED_BY, ID: "1"},
		{Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: "2"},
	}
	test := []Annotation{
		{Predicate: IS_PRIMARILY_CLASSIFIED_BY, ID: "2"},
		{Predicate: IS_CLASSIFIED_BY, ID: "1"},
	}

	By(byUUID).Sort(test)
	if !reflect.DeepEqual(expected, test) {
		t.Fatal("Expected input to be equal to output")
	}
}

//Implementation of sort for an array of structs in order to compare equality of 2 arrays of annotations
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
