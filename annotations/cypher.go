package annotations

import (
	"fmt"

	"errors"

	"github.com/Financial-Times/neo-model-utils-go/mapper"
	"github.com/Financial-Times/neo-utils-go/neoutils"
	log "github.com/Sirupsen/logrus"
	"github.com/jmcvetta/neoism"
)

// Driver interface
type driver interface {
	read(id string) (anns annotations, found bool, err error)
	filteredRead(id string, platformVersion string) (anns annotations, found bool, err error)
	checkConnectivity() error
}

// CypherDriver struct
type cypherDriver struct {
	conn neoutils.NeoConnection
	env  string
}

func NewCypherDriver(conn neoutils.NeoConnection, env string) cypherDriver {
	return cypherDriver{conn, env}
}

func (cd cypherDriver) checkConnectivity() error {
	return neoutils.Check(cd.conn)
}

type neoAnnotation struct {
	Predicate string   `json:"predicate"`
	ID        string   `json:"id"`
	APIURL    string   `json:"apiUrl"`
	Types     []string `json:"types"`
	LeiCode   string   `json:"leiCode,omitempty"`
	PrefLabel string   `json:"prefLabel,omitempty"`

	// Canonical information
	PrefUUID           string   `json:"prefUUID"`
	CanonicalTypes     []string `json:"canonicalTypes"`
	CanonicalLeiCode   string   `json:"canonicalLei,omitempty"`
	CanonicalPrefLabel string   `json:"canonicalPrefLabel,omitempty"`

	//the fields below are populated only for the /content/{uuid}/annotations/{plaformVersion} endpoint
	FactsetIDs      []string `json:"factsetID,omitempty"`
	TmeIDs          []string `json:"tmeIDs,omitempty"`
	UUIDs           []string `json:"uuids,omitempty"`
	PlatformVersion string   `json:"platformVersion,omitempty"`
}

func (cd cypherDriver) read(contentUUID string) (anns annotations, found bool, err error) {
	results := []neoAnnotation{}

	query := &neoism.CypherQuery{
		Statement: `
			MATCH (c:Thing{uuid:{contentUUID}})-[rel]->(cc:Concept)
			OPTIONAL MATCH (cc)-[:EQUIVALENT_TO]->(canonical:Concept)
			OPTIONAL MATCH (cc)<-[iden:IDENTIFIES]-(i:LegalEntityIdentifier)
			WITH c, collect({id: cc.uuid, predicate: type(rel), types: labels(cc), prefLabel:cc.prefLabel, leiCode:i.value, prefUUID: canonical.prefUUID, canonicalPrefLabel: canonical.prefLabel, canonicalLEI: canonical.LEICode, canonicalTypes: labels(canonical)}) as rows
			OPTIONAL MATCH (canonicalBrand:Brand)-[:EQUIVALENT_TO]-(sourceBrand:Brand)<-[rel]-(c)
            OPTIONAL MATCH (canonicalBrand)-[:EQUIVALENT_TO]-(treeBrand:Brand)-[r:HAS_PARENT*0..]->(cd:Concept)
            OPTIONAL MATCH (cd)-[:EQUIVALENT_TO]-(canon:Brand)
			WITH collect({id: cd.uuid, predicate:'IS_CLASSIFIED_BY', types: labels(cd), prefLabel:cd.prefLabel,leiCode:null, canonicalPrefLabel: canon.prefLabel, prefUUID: canon.prefUUID, canonicalTypes: labels(canon), canonicalLEI: null}) + rows as allRows
			UNWIND allRows as row
			WITH DISTINCT(row) as drow
			RETURN distinct drow.id as id, drow.predicate as predicate, drow.types as types, drow.prefLabel as prefLabel, drow.leiCode as leiCode, drow.canonicalPrefLabel as canonicalPrefLabel, drow.prefUUID as prefUUID, drow.canonicalTypes as canonicalTypes,  drow.canonicalLEI as canonicalLei
			`,
		Parameters: neoism.Props{"contentUUID": contentUUID},
		Result:     &results,
	}

	err = cd.conn.CypherBatch([]*neoism.CypherQuery{query})
	if err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v", contentUUID, query.Statement, err)
		return annotations{}, false, fmt.Errorf("Error accessing Annotations datastore for uuid: %s", contentUUID)
	}
	log.Debugf("Found %d Annotations for uuid: %s", len(results), contentUUID)
	if (len(results)) == 0 {
		return annotations{}, false, nil
	}

	mappedAnnotations := []annotation{}
	filter := NewAnnotationsFilter()
	found = false

	for idx := range results {
		annotation, err := mapToResponseFormat(results[idx], cd.env)

		if err == nil {
			filter.Add(annotation)
			found = true
		}
	}
	mappedAnnotations = filter.Filter()
	return mappedAnnotations, found, nil
}

// Returns all the annotations with the specified platformVersion enriched with all the existing concept IDs for a given content
func (cd cypherDriver) filteredRead(contentUUID string, platformVersion string) (anns annotations, found bool, err error) {
	results := []neoAnnotation{}

	query := &neoism.CypherQuery{
		Statement: `
			MATCH (c:Thing{uuid:{contentUUID}})-[rel{platformVersion:{platformVersion}}]->(cc:Concept)
			OPTIONAL MATCH (cc)-[:EQUIVALENT_TO]-(canonicalNode:Concept)
			OPTIONAL MATCH (canonicalNode)-[:EQUIVALENT_TO]-(allSources:Concept)
			OPTIONAL MATCH (cc)<-[:IDENTIFIES]-(upp:UPPIdentifier)
			optional MATCH (cc)<-[:IDENTIFIES]-(lei:LegalEntityIdentifier)
			optional MATCH (cc)<-[:IDENTIFIES]-(tme:TMEIdentifier)
			optional MATCH (cc)<-[:IDENTIFIES]-(fs:FactsetIdentifier)
			OPTIONAL MATCH (allSources)<-[:IDENTIFIES]-(sourceUPP:UPPIdentifier)
			optional MATCH (allSources)<-[:IDENTIFIES]-(sourceLEI:LegalEntityIdentifier)
			optional MATCH (allSources)<-[:IDENTIFIES]-(sourceTME:TMEIdentifier)
			optional MATCH (allSources)<-[:IDENTIFIES]-(sourceFS:FactsetIdentifier)
			WITH c, cc, rel, lei, collect(distinct fs.value) + collect(distinct sourceFS.value) as fs,  collect(distinct tme.value) + collect(distinct sourceTME.value) as tme, collect(distinct upp.value) + collect(distinct sourceUPP.value) as upp
			WITH c, collect({id: cc.uuid, predicate: type(rel), types: labels(cc), prefLabel:cc.prefLabel, uuids:upp, tmeIDs:tme, leiCode:lei.value, factsetID:fs, platformVersion:rel.platformVersion}) as rows
			UNWIND rows as row
			WITH DISTINCT(row) as drow
			RETURN drow.id as id, drow.predicate as predicate, drow.types as types, drow.prefLabel as prefLabel, drow.leiCode as leiCode, drow.factsetID as factsetID, drow.uuids as uuids, drow.tmeIDs as tmeIDs, drow.platformVersion as platformVersion
			`,
		Parameters: neoism.Props{"contentUUID": contentUUID, "platformVersion": platformVersion},
		Result:     &results,
	}

	err = cd.conn.CypherBatch([]*neoism.CypherQuery{query})
	if err != nil {
		log.Errorf("Error looking up uuid %s with query %s from neoism: %+v", contentUUID, query.Statement, err)
		return annotations{}, false, fmt.Errorf("Error accessing Annotations datastore for uuid: %s", contentUUID)
	}
	log.Debugf("Found %d Annotations for uuid: %s with platformVersion: %s", len(results), contentUUID, platformVersion)
	if (len(results)) == 0 {
		return annotations{}, false, nil
	}

	mappedAnnotations := []annotation{}

	found = false

	for idx := range results {
		annotation, err := mapToResponseFormat(results[idx], cd.env)
		if err == nil {
			mappedAnnotations = append(mappedAnnotations, annotation)
			found = true
		}
	}

	return mappedAnnotations, found, nil
}

func mapToResponseFormat(neoAnn neoAnnotation, env string) (annotation, error) {
	var ann annotation

	// New concordance model
	if neoAnn.PrefUUID != "" {
		ann.PrefLabel = neoAnn.CanonicalPrefLabel
		ann.LeiCode = neoAnn.CanonicalLeiCode
		ann.APIURL = mapper.APIURL(neoAnn.PrefUUID, neoAnn.Types, env)
		ann.ID = mapper.IDURL(neoAnn.PrefUUID)
		types := mapper.TypeURIs(neoAnn.CanonicalTypes)
		if types == nil || len(types) == 0 {
			log.Debugf("Could not map type URIs for ID %s with types %s", ann.ID, neoAnn.CanonicalTypes)
			return ann, errors.New("Concept not found")
		}
		ann.Types = types
	} else {
		ann.PrefLabel = neoAnn.PrefLabel
		ann.LeiCode = neoAnn.LeiCode
		ann.APIURL = mapper.APIURL(neoAnn.ID, neoAnn.Types, env)
		ann.ID = mapper.IDURL(neoAnn.ID)
		types := mapper.TypeURIs(neoAnn.Types)
		if types == nil || len(types) == 0 {
			log.Debugf("Could not map type URIs for ID %s with types %s", ann.ID, ann.Types)
			return ann, errors.New("Concept not found")
		}
		ann.Types = types
	}

	predicate, err := getPredicateFromRelationship(neoAnn.Predicate)
	if err != nil {
		log.Warnf("Could not find predicate for ID %s for relationship %s", ann.ID, ann.Predicate)
		return ann, err
	}
	ann.Predicate = predicate
	ann.PlatformVersion = neoAnn.PlatformVersion
	if len(neoAnn.TmeIDs) > 0 {
		ann.TmeIDs = deduplicateList(neoAnn.TmeIDs)
	}
	if len(neoAnn.FactsetIDs) > 0 {
		ann.FactsetIDs = deduplicateList(neoAnn.FactsetIDs)
	}
	if len(neoAnn.UUIDs) > 0 {
		ann.UUIDs = deduplicateList(neoAnn.UUIDs)
	}
	return ann, nil
}

func deduplicateList(inList []string) []string {
	outList := []string{}
	deduped := map[string]bool{}
	for _, v := range inList {
		deduped[v] = true
	}
	for k, o := range deduped {
		if o {
			outList = append(outList, k)
		}
	}
	return outList
}

func getPredicateFromRelationship(relationship string) (predicate string, err error) {
	predicate = predicates[relationship]
	if predicate == "" {
		return "", errors.New("Not a valid annotation type")
	}
	return predicate, nil
}
