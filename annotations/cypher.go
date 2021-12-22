package annotations

import (
	"fmt"

	"errors"

	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
	"github.com/Financial-Times/neo-model-utils-go/mapper"
)

type driver interface {
	read(id string) (anns annotations, found bool, err error)
	checkConnectivity() error
}

type CypherDriver struct {
	driver *cmneo4j.Driver
	env    string
}

func NewCypherDriver(driver *cmneo4j.Driver, env string) CypherDriver {
	return CypherDriver{driver: driver, env: env}
}

func (cd CypherDriver) checkConnectivity() error {
	return cd.driver.VerifyConnectivity()
}

type neoAnnotation struct {
	Predicate       string
	ID              string
	APIURL          string
	Types           []string
	LeiCode         string
	FIGI            string
	NAICSIdentifier string
	NAICSPrefLabel  string
	NAICSRank       int
	PrefLabel       string
	Lifecycle       string
	IsDeprecated    bool

	// Canonical information
	PrefUUID           string
	CanonicalTypes     []string
	CanonicalLeiCode   string
	CanonicalPrefLabel string

	//the fields below are populated only for the /content/{uuid}/annotations/{plaformVersion} endpoint
	FactsetIDs      []string `json:"factsetID,omitempty"`
	TmeIDs          []string `json:"tmeIDs,omitempty"`
	UUIDs           []string `json:"uuids,omitempty"`
	PlatformVersion string   `json:"platformVersion,omitempty"`
}

func (cd CypherDriver) read(contentUUID string) (anns annotations, found bool, err error) {
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
			canonicalConcept.leiCode as leiCode,
			figi.figiCode as figi,
			naics.industryIdentifier as naicsIdentifier,
			naics.prefLabel as naicsPrefLabel,
			naicsRel.rank as naicsRank,
			rel.lifecycle as lifecycle
		UNION ALL
		MATCH (content:Content{uuid:$contentUUID})-[rel]-(:Concept)-[:EQUIVALENT_TO]->(canonicalBrand:Brand)
		OPTIONAL MATCH (canonicalBrand)<-[:EQUIVALENT_TO]-(leafBrand:Brand)-[r:HAS_PARENT*0..]->(parentBrand:Brand)-[:EQUIVALENT_TO]->(canonicalParent:Brand)
		RETURN 
			DISTINCT canonicalParent.prefUUID as id,
			canonicalParent.isDeprecated as isDeprecated,
			"IMPLICITLY_CLASSIFIED_BY" as predicate,
			labels(canonicalParent) as types,
			canonicalParent.prefLabel as prefLabel,
			null as leiCode,
			null as figi,
			null as naicsIdentifier,
			null as naicsPrefLabel,
			null as naicsRank,
			rel.lifecycle as lifecycle
		UNION ALL
		MATCH (content:Content{uuid:$contentUUID})-[rel:ABOUT]-(:Concept)-[:EQUIVALENT_TO]->(canonicalConcept:Concept)
		MATCH (canonicalConcept)<-[:EQUIVALENT_TO]-(leafConcept:Topic)<-[:IMPLIED_BY*1..]-(impliedByBrand:Brand)-[:EQUIVALENT_TO]->(canonicalBrand:Brand)
		RETURN 
			DISTINCT canonicalBrand.prefUUID as id,
			canonicalBrand.isDeprecated as isDeprecated,
			"IMPLICITLY_CLASSIFIED_BY" as predicate,
			labels(canonicalBrand) as types,
			canonicalBrand.prefLabel as prefLabel,
			null as leiCode,
			null as figi,
			null as naicsIdentifier,
			null as naicsPrefLabel,
			null as naicsRank,
			rel.lifecycle as lifecycle
		UNION ALL
		MATCH (content:Content{uuid:$contentUUID})-[rel:ABOUT]-(:Concept)-[:EQUIVALENT_TO]->(canonicalConcept:Concept)
		MATCH (canonicalConcept)<-[:EQUIVALENT_TO]-(leafConcept:Concept)-[:HAS_BROADER*1..]->(implicit:Concept)-[:EQUIVALENT_TO]->(canonicalImplicit)
		WHERE NOT (canonicalImplicit)<-[:EQUIVALENT_TO]-(:Concept)<-[:ABOUT]-(content) // filter out the original abouts
		RETURN 
			DISTINCT canonicalImplicit.prefUUID as id,
			canonicalImplicit.isDeprecated as isDeprecated,
			"IMPLICITLY_ABOUT" as predicate,
			labels(canonicalImplicit) as types,
			canonicalImplicit.prefLabel as prefLabel,
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

	err = cd.driver.Read(query)
	if errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return annotations{}, false, nil
	}
	if err != nil {
		return annotations{}, false,
			fmt.Errorf("failed looking up annotations for contentUUID %s: %w", contentUUID, err)
	}

	var mappedAnnotations []annotation
	found = false

	for idx := range results {
		annotation, err := mapToResponseFormat(results[idx], cd.env)
		if err == nil {
			found = true
			mappedAnnotations = append(mappedAnnotations, annotation)
		}
	}

	return mappedAnnotations, found, nil
}

func mapToResponseFormat(neoAnn neoAnnotation, env string) (annotation, error) {
	var ann annotation

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
	ann.APIURL = mapper.APIURL(neoAnn.ID, neoAnn.Types, env)
	ann.ID = mapper.IDURL(neoAnn.ID)
	types := mapper.TypeURIs(neoAnn.Types)
	if len(types) == 0 {
		return ann, fmt.Errorf("could not map type URIs for ID %s with types %s: concept not found", ann.ID, ann.Types)
	}
	ann.Types = types

	predicate, err := getPredicateFromRelationship(neoAnn.Predicate)
	if err != nil {
		return ann, fmt.Errorf("could not find predicate for ID %s for relationship %s: %w", ann.ID, ann.Predicate, err)
	}
	ann.Predicate = predicate
	ann.Lifecycle = neoAnn.Lifecycle
	ann.IsDeprecated = neoAnn.IsDeprecated

	return ann, nil
}

func getPredicateFromRelationship(relationship string) (predicate string, err error) {
	predicate = predicates[relationship]
	if predicate == "" {
		return "", errors.New("not a valid annotation type")
	}
	return predicate, nil
}
