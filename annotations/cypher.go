package annotations

import (
	"errors"
	"fmt"
	"net/url"

	ontology "github.com/Financial-Times/cm-graph-ontology/v2"
	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
)

const IDPrefix = "http://api.ft.com/things/"

type driver interface {
	read(id string, bookmark string) (anns Annotations, found bool, err error)
	checkConnectivity() error
}

type CypherDriver struct {
	driver  *cmneo4j.Driver
	baseURL string
}

func NewCypherDriver(driver *cmneo4j.Driver, baseURL string) CypherDriver {
	return CypherDriver{driver: driver, baseURL: baseURL}
}

func (cd CypherDriver) checkConnectivity() error {
	return cd.driver.VerifyConnectivity()
}

type neoAnnotation struct {
	Predicate           string
	ID                  string
	APIURL              string
	Types               []string
	LeiCode             string
	FIGI                string
	NAICSIdentifier     string
	NAICSPrefLabel      string
	NAICSRank           int
	PrefLabel           string
	GeonamesFeatureCode string
	Lifecycle           string
	IsDeprecated        bool

	// Canonical information
	PrefUUID           string
	CanonicalTypes     []string
	CanonicalLeiCode   string
	CanonicalPrefLabel string

	// the fields below are populated only for the /content/{uuid}/annotations/{plaformVersion} endpoint
	FactsetIDs      []string `json:"factsetID,omitempty"`
	TmeIDs          []string `json:"tmeIDs,omitempty"`
	UUIDs           []string `json:"uuids,omitempty"`
	PlatformVersion string   `json:"platformVersion,omitempty"`
}

// read method reads the annotations for a given contentUUID from Neo4j.
// If non-empty bookmark is provided, it will be used in the session reading from Neo4j. The bookmark guarantees
// that the instance executing the read transaction is at least up to date to the point represented by the bookmark.
// If not existing bookmark is given but in correct format, the read will be successful.
// If bookmark in not valid format is provided, the read will fail. The format of the bookmarks is checked by the db.
func (cd CypherDriver) read(contentUUID string, bookmark string) (anns Annotations, found bool, err error) {
	var results []neoAnnotation

	query := &cmneo4j.Query{
		Cypher: `
		MATCH (content:Content{uuid:$contentUUID})-[rel]-(:Concept)-[:EQUIVALENT_TO]->(canonicalConcept:Concept)
		OPTIONAL MATCH (canonicalConcept)<-[:EQUIVALENT_TO]-(:Concept)<-[:ISSUED_BY]-(figi:FinancialInstrument)
		OPTIONAL MATCH (canonicalConcept)<-[:EQUIVALENT_TO]-(:Concept)-[naicsRel:HAS_INDUSTRY_CLASSIFICATION{rank:1}]->(NAICSIndustryClassification)-[:EQUIVALENT_TO]->(naics:NAICSIndustryClassification)
		RETURN
			canonicalConcept.prefUUID as id,
			canonicalConcept.isDeprecated as isDeprecated,
			type(rel) as predicate,
			labels(canonicalConcept) as types,
			canonicalConcept.prefLabel as prefLabel,
			canonicalConcept.geonamesFeatureCode as geonamesFeatureCode,
			canonicalConcept.leiCode as leiCode,
			figi.figiCode as figi,
			naics.industryIdentifier as naicsIdentifier,
			naics.prefLabel as naicsPrefLabel,
			naicsRel.rank as naicsRank,
			rel.lifecycle as lifecycle
		UNION
		MATCH (content:Content{uuid:$contentUUID})-[rel]-(:Concept)-[:EQUIVALENT_TO]->(canonicalBrand:Brand)
		OPTIONAL MATCH (canonicalBrand)<-[:EQUIVALENT_TO]-(leafBrand:Brand)-[r:HAS_PARENT*0..]->(parentBrand:Brand)-[:EQUIVALENT_TO]->(canonicalParent:Brand)
		RETURN 
			DISTINCT canonicalParent.prefUUID as id,
			canonicalParent.isDeprecated as isDeprecated,
			"IMPLICITLY_CLASSIFIED_BY" as predicate,
			labels(canonicalParent) as types,
			canonicalParent.prefLabel as prefLabel,
			null as geonamesFeatureCode,
			null as leiCode,
			null as figi,
			null as naicsIdentifier,
			null as naicsPrefLabel,
			null as naicsRank,
			rel.lifecycle as lifecycle
		UNION
		MATCH (content:Content{uuid:$contentUUID})-[rel:ABOUT]-(:Concept)-[:EQUIVALENT_TO]->(canonicalConcept:Concept)
		MATCH (canonicalConcept)<-[:EQUIVALENT_TO]-(leafConcept:Topic)<-[:IMPLIED_BY*1..]-(impliedByBrand:Brand)-[:EQUIVALENT_TO]->(canonicalBrand:Brand)
		RETURN 
			DISTINCT canonicalBrand.prefUUID as id,
			canonicalBrand.isDeprecated as isDeprecated,
			"IMPLICITLY_CLASSIFIED_BY" as predicate,
			labels(canonicalBrand) as types,
			canonicalBrand.prefLabel as prefLabel,
			null as geonamesFeatureCode,
			null as leiCode,
			null as figi,
			null as naicsIdentifier,
			null as naicsPrefLabel,
			null as naicsRank,
			rel.lifecycle as lifecycle
		UNION
		MATCH (content:Content{uuid:$contentUUID})-[rel:ABOUT]-(:Concept)-[:EQUIVALENT_TO]->(canonicalConcept:Concept)
		MATCH (canonicalConcept)<-[:EQUIVALENT_TO]-(leafConcept:Concept)-[:HAS_BROADER*1..]->(implicit:Concept)-[:EQUIVALENT_TO]->(canonicalImplicit)
		WHERE NOT (canonicalImplicit)<-[:EQUIVALENT_TO]-(:Concept)<-[:ABOUT]-(content) // filter out the original abouts
		RETURN 
			DISTINCT canonicalImplicit.prefUUID as id,
			canonicalImplicit.isDeprecated as isDeprecated,
			"IMPLICITLY_ABOUT" as predicate,
			labels(canonicalImplicit) as types,
			canonicalImplicit.prefLabel as prefLabel,
			canonicalImplicit.geonamesFeatureCode as geonamesFeatureCode,
			null as leiCode,
			null as figi,
			null as naicsIdentifier,
			null as naicsPrefLabel,
			null as naicsRank,
			rel.lifecycle as lifecycle
		UNION
		MATCH (content:Content{uuid:$contentUUID})-[rel:ABOUT]-(:Concept)-[:EQUIVALENT_TO]->(canonicalConcept:Concept)
		MATCH (canonicalConcept)<-[:EQUIVALENT_TO]-(leafConcept:Location)-[:IS_PART_OF*1..]->(implicit:Concept)-[:EQUIVALENT_TO]->(canonicalImplicit)
		WHERE NOT (canonicalImplicit)<-[:EQUIVALENT_TO]-(:Concept)<-[:ABOUT]-(content) // filter out the original abouts
		RETURN 
			DISTINCT canonicalImplicit.prefUUID as id,
			canonicalImplicit.isDeprecated as isDeprecated,
			"IMPLICITLY_ABOUT" as predicate,
			labels(canonicalImplicit) as types,
			canonicalImplicit.prefLabel as prefLabel,
			canonicalImplicit.geonamesFeatureCode as geonamesFeatureCode,
			null as leiCode,
			null as figi,
			null as naicsIdentifier,
			null as naicsPrefLabel,
			null as naicsRank,
			rel.lifecycle as lifecycle
		`,
		Params: map[string]interface{}{"contentUUID": contentUUID},
		Result: &results,
	}

	bookmarks := make([]string, 0, 1)
	if len(bookmark) > 0 {
		bookmarks = append(bookmarks, bookmark)
	}

	_, err = cd.driver.ReadMultiple([]*cmneo4j.Query{query}, bookmarks)
	if errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return Annotations{}, false, nil
	}
	if err != nil {
		return Annotations{}, false,
			fmt.Errorf("failed looking up annotations for contentUUID %s: %w", contentUUID, err)
	}

	var mappedAnnotations []Annotation
	found = false

	for idx := range results {
		annotation, err := mapToResponseFormat(results[idx], cd.baseURL)
		if err == nil {
			found = true
			mappedAnnotations = append(mappedAnnotations, annotation)
		}
	}

	return mappedAnnotations, found, nil
}

func mapToResponseFormat(neoAnn neoAnnotation, baseURL string) (Annotation, error) {
	var ann Annotation

	ann.PrefLabel = neoAnn.PrefLabel
	ann.LeiCode = neoAnn.LeiCode
	ann.FIGI = neoAnn.FIGI
	if neoAnn.NAICSIdentifier != "" {
		ann.NAICS = []IndustryClassification{
			{
				Identifier: neoAnn.NAICSIdentifier,
				PrefLabel:  neoAnn.NAICSPrefLabel,
				Rank:       neoAnn.NAICSRank,
			},
		}
	}

	apiURL, err := ontology.APIURL(neoAnn.ID, neoAnn.Types, baseURL)
	if err != nil {
		return ann, fmt.Errorf("could not construct api url for uuid %s with types %s", neoAnn.ID, neoAnn.Types)
	}
	ann.APIURL = apiURL

	id, err := getIDURI(neoAnn.ID)
	if err != nil {
		return ann, fmt.Errorf("could not construct ID uri for uuid %s", neoAnn.ID)
	}
	ann.ID = id

	types, err := ontology.TypeURIs(neoAnn.Types)
	if err != nil || len(types) == 0 {
		return ann, fmt.Errorf("could not map type URIs for uuid %s with types %s: concept not found", neoAnn.ID, neoAnn.Types)
	}
	ann.Types = types

	predicate, err := getPredicateFromRelationship(neoAnn.Predicate)
	if err != nil {
		return ann, fmt.Errorf("could not find predicate for ID %s for relationship %s: %w", ann.ID, ann.Predicate, err)
	}
	ann.Predicate = predicate
	ann.Lifecycle = neoAnn.Lifecycle
	ann.IsDeprecated = neoAnn.IsDeprecated
	ann.GeonamesFeatureCode = neoAnn.GeonamesFeatureCode

	return ann, nil
}

func getIDURI(uuid string) (string, error) {
	return url.JoinPath(IDPrefix, uuid)
}

func getPredicateFromRelationship(relationship string) (predicate string, err error) {
	predicate = predicates[relationship]
	if predicate == "" {
		return "", errors.New("not a valid annotation type")
	}
	return predicate, nil
}
