package annotations

type annotations []annotation

type annotation struct {
	Predicate string   `json:"predicate"`
	ID        string   `json:"id"`
	APIURL    string   `json:"apiUrl"`
	Types     []string `json:"types"`
	LeiCode   string   `json:"leiCode,omitempty"`
	FIGI      string   `json:"FIGI,omitempty"`
	PrefLabel string   `json:"prefLabel,omitempty"`
	//used for filtering, e.g. pac not exposed
	Lifecycle string `json:"-"`
}

var predicates = map[string]string{
	"MENTIONS":                   "http://www.ft.com/ontology/annotation/mentions",
	"MAJOR_MENTIONS":             "http://www.ft.com/ontology/annotation/majorMentions",
	"ABOUT":                      "http://www.ft.com/ontology/annotation/about",
	"HAS_AUTHOR":                 "http://www.ft.com/ontology/annotation/hasAuthor",
	"HAS_CONTRIBUTOR":            "http://www.ft.com/ontology/hasContributor",
	"HAS_DISPLAY_TAG":            "http://www.ft.com/ontology/hasDisplayTag",
	"IS_CLASSIFIED_BY":           "http://www.ft.com/ontology/classification/isClassifiedBy",
	"IS_PRIMARILY_CLASSIFIED_BY": "http://www.ft.com/ontology/classification/isPrimarilyClassifiedBy",
	"IMPLICITLY_CLASSIFIED_BY":   "http://www.ft.com/ontology/implicitlyClassifiedBy",
	"IMPLICITLY_ABOUT":           "http://www.ft.com/ontology/implicitlyAbout",
}
