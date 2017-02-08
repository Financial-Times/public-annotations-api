package people

import (
	"encoding/json"
	"fmt"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/Financial-Times/up-rw-app-api-go/rwapi"
	"github.com/jmcvetta/neoism"
)

type service struct {
	conn neoutils.NeoConnection
}

// NewCypherPeopleService provides functions for create, update, delete operations on people in Neo4j,
// plus other utility functions needed for a service
func NewCypherPeopleService(cypherRunner neoutils.NeoConnection) service {
	return service{cypherRunner}
}

func (s service) Initialise() error {

	err := s.conn.EnsureIndexes(map[string]string{
		"Identifier": "value",
	})

	if err != nil {
		return err
	}

	return s.conn.EnsureConstraints(map[string]string{
		"Thing":             "uuid",
		"Concept":           "uuid",
		"Person":            "uuid",
		"FactsetIdentifier": "value",
		"TMEIdentifier":     "value",
		"UPPIdentifier":     "value"})
}

func (s service) Read(uuid string) (interface{}, bool, error) {
	results := []person{}

	readQuery := &neoism.CypherQuery{
		Statement: `MATCH (p:Person {uuid:{uuid}})
					OPTIONAL MATCH (upp:UPPIdentifier)-[:IDENTIFIES]->(p)
					OPTIONAL MATCH (factset:FactsetIdentifier)-[:IDENTIFIES]->(p)
					OPTIONAL MATCH (tme:TMEIdentifier)-[:IDENTIFIES]->(p)
					return p.uuid as uuid,
						p.name as name,
						p.emailAddress as emailAddress,
						p.twitterHandle as twitterHandle,
						p.facebookProfile as facebookProfile,
						p.linkedinProfile as linkedinProfile,
						p.description as description,
						p.descriptionXML as descriptionXML,
						p.prefLabel as prefLabel,
						p.birthYear as birthYear,
						p.salutation as salutation,
						p.aliases as aliases,
						p.imageURL as _imageUrl,
						labels(p) as types,
						{uuids:collect(distinct upp.value),
							TME:collect(distinct tme.value),
							factsetIdentifier:factset.value} as alternativeIdentifiers`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &results,
	}

	if err := s.conn.CypherBatch([]*neoism.CypherQuery{readQuery}); err != nil || len(results) == 0 {
		return person{}, false, err
	}

	if len(results) == 0 {
		return person{}, false, nil
	}
	result := results[0]

	p := person{
		UUID:                   result.UUID,
		Name:                   result.Name,
		PrefLabel:              result.PrefLabel,
		EmailAddress:           result.EmailAddress,
		TwitterHandle:          result.TwitterHandle,
		FacebookProfile:        result.FacebookProfile,
		LinkedinProfile:        result.LinkedinProfile,
		Description:            result.Description,
		DescriptionXML:         result.DescriptionXML,
		BirthYear:              result.BirthYear,
		Salutation:             result.Salutation,
		ImageURL:               result.ImageURL,
		AlternativeIdentifiers: result.AlternativeIdentifiers,
		Aliases:                result.Aliases,
		Types:                  result.Types,
	}

	return p, true, nil

}

func (s service) IDs(f func(id rwapi.IDEntry) (bool, error)) error {
	batchSize := 4096

	for skip := 0; ; skip += batchSize {
		results := []rwapi.IDEntry{}
		readQuery := &neoism.CypherQuery{
			Statement: `MATCH (p:Person) RETURN p.uuid as id, p.hash as hash SKIP {skip} LIMIT {limit}`,
			Parameters: map[string]interface{}{
				"limit": batchSize,
				"skip":  skip,
			},
			Result: &results,
		}
		if err := s.conn.CypherBatch([]*neoism.CypherQuery{readQuery}); err != nil {
			return err
		}
		if len(results) == 0 {
			return nil
		}
		for _, result := range results {
			more, err := f(result)
			if !more || err != nil {
				return err
			}
		}
	}
}

func (s service) Write(thing interface{}) error {

	hash, err := writeHash(thing)
	if err != nil {
		return err
	}

	p := thing.(person)

	params := map[string]interface{}{
		"uuid": p.UUID,
		"hash": hash,
	}

	if p.Name != "" {
		params["name"] = p.Name
	}

	if p.PrefLabel != "" {
		params["prefLabel"] = p.PrefLabel
	}

	if p.BirthYear != 0 {
		params["birthYear"] = p.BirthYear
	}

	if p.Salutation != "" {
		params["salutation"] = p.Salutation
	}

	if p.EmailAddress != "" {
		params["emailAddress"] = p.EmailAddress
	}

	if p.TwitterHandle != "" {
		params["twitterHandle"] = p.TwitterHandle
	}

	if p.FacebookProfile != "" {
		params["facebookProfile"] = p.FacebookProfile
	}

	if p.FacebookProfile != "" {
		params["linkedinProfile"] = p.LinkedinProfile
	}

	if p.Description != "" {
		params["description"] = p.Description
	}

	if p.DescriptionXML != "" {
		params["descriptionXML"] = p.DescriptionXML
	}

	if p.ImageURL != "" {
		params["imageURL"] = p.ImageURL
	}

	var aliases []string

	for _, alias := range p.Aliases {
		aliases = append(aliases, alias)
	}

	if len(aliases) > 0 {
		params["aliases"] = aliases
	}

	deleteEntityRelationshipsQuery := &neoism.CypherQuery{
		Statement: `MATCH (i:Identifier)-[ir:IDENTIFIES]->(t:Thing {uuid:{uuid}})
				DELETE ir, i`,
		Parameters: map[string]interface{}{
			"uuid": p.UUID,
		},
	}

	queries := []*neoism.CypherQuery{deleteEntityRelationshipsQuery}

	writeQuery := &neoism.CypherQuery{
		Statement: `MERGE (n:Thing{uuid: {uuid}})
						set n={props}
						set n :Concept
						set n :Person `,
		Parameters: map[string]interface{}{
			"uuid":  p.UUID,
			"props": params,
		},
	}

	queries = append(queries, writeQuery)

	//ADD all the IDENTIFIER nodes and IDENTIFIES relationships
	for _, alternativeUUID := range p.AlternativeIdentifiers.TME {
		alternativeIdentifierQuery := createNewIdentifierQuery(p.UUID, tmeIdentifierLabel, alternativeUUID)
		queries = append(queries, alternativeIdentifierQuery)
	}

	for _, alternativeUUID := range p.AlternativeIdentifiers.UUIDS {
		alternativeIdentifierQuery := createNewIdentifierQuery(p.UUID, uppIdentifierLabel, alternativeUUID)
		queries = append(queries, alternativeIdentifierQuery)
	}

	if p.AlternativeIdentifiers.FactsetIdentifier != "" {
		queries = append(queries, createNewIdentifierQuery(p.UUID, factsetIdentifierLabel, p.AlternativeIdentifiers.FactsetIdentifier))
	}

	return s.conn.CypherBatch(queries)
}

func createNewIdentifierQuery(uuid string, identifierLabel string, identifierValue string) *neoism.CypherQuery {
	statementTemplate := fmt.Sprintf(`MERGE (t:Thing {uuid:{uuid}})
					CREATE (i:Identifier {value:{value}})-[:IDENTIFIES]->(t)
					set i : %s `, identifierLabel)
	query := &neoism.CypherQuery{
		Statement: statementTemplate,
		Parameters: map[string]interface{}{
			"uuid":  uuid,
			"value": identifierValue,
		},
	}
	return query
}

func (s service) Delete(uuid string) (bool, error) {
	clearNode := &neoism.CypherQuery{
		Statement: `
			MATCH (p:Thing {uuid: {uuid}})
			REMOVE p:Concept
			REMOVE p:Person
			SET p={props}
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
			"props": map[string]interface{}{
				"uuid": uuid,
			},
		},
		IncludeStats: true,
	}

	// Please note that this removes the Identifiers if there are no other relationships attached to this
	// as Identifiers are not a 'Thing' only an Identifier
	removeNodeIfUnused := &neoism.CypherQuery{
		Statement: `
			MATCH (thing:Thing {uuid: {uuid}})
			OPTIONAL MATCH (thing)-[ir:IDENTIFIES]-(id:Identifier)
			OPTIONAL MATCH (thing)-[a]-(x:Thing)
			WITH ir, id, thing, count(a) AS relCount
			WHERE relCount = 0
			DELETE ir, id, thing
		`,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
	}

	err := s.conn.CypherBatch([]*neoism.CypherQuery{clearNode, removeNodeIfUnused})

	s1, err := clearNode.Stats()

	if err != nil {
		return false, err
	}

	var deleted bool
	if s1.ContainsUpdates && s1.LabelsRemoved > 0 {
		deleted = true
	}

	return deleted, err
}

func (s service) DecodeJSON(dec *json.Decoder) (interface{}, string, error) {
	p := person{}
	err := dec.Decode(&p)
	return p, p.UUID, err
}

func (s service) Check() error {
	return neoutils.Check(s.conn)
}

func (s service) Count() (int, error) {

	results := []struct {
		Count int `json:"c"`
	}{}

	query := &neoism.CypherQuery{
		Statement: `MATCH (n:Person) return count(n) as c`,
		Result:    &results,
	}

	err := s.conn.CypherBatch([]*neoism.CypherQuery{query})

	if err != nil {
		return 0, err
	}

	return results[0].Count, nil
}

type requestError struct {
	details string
}

func (re requestError) Error() string {
	return "Invalid Request"
}

func (re requestError) InvalidRequestDetails() string {
	return re.details
}
