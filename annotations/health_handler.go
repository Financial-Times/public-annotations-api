package annotations

import (
	fthealth "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/service-status-go/gtg"
)

const (
	runbookURL = "https://runbooks.in.ft.com/annotationsapi"
)

func HealthCheck(hctx *HandlerCtx) fthealth.Check {
	return fthealth.Check{
		ID:               "neo4j-cluster-health",
		BusinessImpact:   "Unable to respond to Public Annotations api requests",
		Name:             "Check connectivity to Neo4j",
		PanicGuide:       runbookURL,
		Severity:         1,
		TechnicalSummary: `Cannot connect to Neo4j. If this check fails, check that Neo4j instance is up and running. You can find the neoUrl as a parameter for this service.`,
		Checker:          Neo4jChecker(hctx.AnnotationsDriver),
	}
}

func Neo4jChecker(annDriver driver) func() (string, error) {
	return func() (string, error) {
		err := annDriver.checkConnectivity()
		if err != nil {
			return "Error connecting to neo4j", err
		}

		return "Connectivity to neo4j is ok", nil
	}
}

func GoodToGo(hctx *HandlerCtx) func() gtg.Status {
	return func() gtg.Status {
		if _, err := Neo4jChecker(hctx.AnnotationsDriver)(); err != nil {
			return gtg.Status{GoodToGo: false, Message: err.Error()}
		}
		return gtg.Status{GoodToGo: true}
	}
}
