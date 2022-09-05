//go:build integration
// +build integration

package annotations

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	annrw "github.com/Financial-Times/annotations-rw-neo4j/v4/annotations"
	"github.com/Financial-Times/base-ft-rw-app-go/v2/baseftrwapp"
	cmneo4j "github.com/Financial-Times/cm-neo4j-driver"
	"github.com/Financial-Times/concepts-rw-neo4j/concepts"
	"github.com/Financial-Times/content-rw-neo4j/v3/content"
	"github.com/Financial-Times/go-logger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1Logger "github.com/Financial-Times/go-logger"
)

const (
	//Generate uuids so there's no clash with real data
	contentUUID                        = "3fc9fe3e-af8c-4f7f-961a-e5065392bb31"
	contentWithNoAnnotationsUUID       = "3fc9fe3e-af8c-1a1a-961a-e5065392bb31"
	contentWithParentAndChildBrandUUID = "3fc9fe3e-af8c-2a2a-961a-e5065392bb31"
	contentWithThreeLevelsOfBrandUUID  = "3fc9fe3e-af8c-3a3a-961a-e5065392bb31"
	contentWithCircularBrandUUID       = "3fc9fe3e-af8c-4a4a-961a-e5065392bb31"
	contentWithOnlyFTUUID              = "3fc9fe3e-af8c-5a5a-961a-e5065392bb31"
	contentWithHasBrand                = "ae17012e-ad40-11e9-8030-530adfa879c2"
	alphavilleSeriesUUID               = "747894f8-a231-4efb-805d-753635eca712"

	brandParentUUID                = "dbb0bdae-1f0c-1a1a-b0cb-b2227cce2b54"
	brandChildUUID                 = "ff691bf8-8d92-1a1a-8326-c273400bff0b"
	brandGrandChildUUID            = "ff691bf8-8d92-2a2a-8326-c273400bff0b"
	brandCircularAUUID             = "ff691bf8-8d92-3a3a-8326-c273400bff0b"
	brandCircularBUUID             = "ff691bf8-8d92-4a4a-8326-c273400bff0b"
	brandWithHasBrandPredicateUUID = "2d3e16e0-61cb-4322-8aff-3b01c59f4daa"
	brandHubPageUUID               = "87645070-7d8a-492e-9695-bf61ac2b4d18"
	genreOpinionUUID               = "6da31a37-691f-4908-896f-2829ebe2309e"
	orgUUID                        = "bb3c006d-e999-3919-8fb2-4992ef7a2100"
	locationA                      = "82cba3ce-329b-3010-b29d-4282a215889f"
	locationB                      = "8d54e308-807c-4e9e-9981-8faab10b6f1c"
	locationC                      = "5895ee1e-d5de-39c1-93ab-03fcb7d36caf"
	locationD                      = "307c91ed-31f5-33b7-895a-1ffbeec514f4"
	locationE                      = "822e3c99-afc6-3c55-b497-2255ac546f35"

	contentWithBrandsDiffTypesUUID = "3fc9fe3e-af8c-6a6a-961a-e5065392bb31"
	financialInstrumentUUID        = "77f613ad-1470-422c-bf7c-1dd4c3fd1693"

	MSJConceptUUID         = "5d1510f8-2779-4b74-adab-0a5eb138fca6"
	FakebookConceptUUID    = "eac853f5-3859-4c08-8540-55e043719400"
	MetalMickeyConceptUUID = "0483bef8-5797-40b8-9b25-b12e492f63c6"
	JohnSmithConceptUUID   = "75e2f7e9-cb5e-40a5-a074-86d69fe09f69"
	brokenPacUUID          = "8d965e66-5454-4856-972d-f64cc1a18a5d"

	contentWithNAICSOrgUUID = "3fc9fe3e-af8c-7a7a-961a-e5065392bb31"
	NYTConceptUUID          = "0d9fbdfc-7d95-332b-b77b-1e69274b1b83"
	NAICSConceptUUID        = "38ee195d-ebdd-48a9-af4b-c8a322e7b04d"

	narrowerTopic = "7e22c8b8-b280-4e52-aa22-fa1c6dffd894"
	aboutTopic    = "ca982370-66cd-43bd-b2e3-7bfcb73efb1e"
	broaderTopicA = "fde5eee9-3260-4125-adb6-3d91a4888be5"
	broaderTopicB = "b6469cc2-f6ff-45aa-a9bb-3d1bb0f9a35d"
	cyclicTopicA  = "e404e3bd-beff-4324-83f4-beb044baf916"
	cyclicTopicB  = "77a410a3-6857-4654-80ef-6aae29be852a"

	v1PlatformVersion    = "v1"
	v2PlatformVersion    = "v2"
	emptyPlatformVersion = ""

	brandType        = "http://www.ft.com/ontology/product/Brand"
	topicType        = "http://www.ft.com/ontology/Topic"
	locationType     = "http://www.ft.com/ontology/Location"
	genreType        = "http://www.ft.com/ontology/Genre"
	organisationType = "http://www.ft.com/ontology/organisation/Organisation"
)

var (
	conceptLabels = map[string]string{
		brandGrandChildUUID:            "Child Business School video",
		brandChildUUID:                 "Business School video",
		brandParentUUID:                "Financial Times",
		brandCircularAUUID:             "Circular Business School video - A",
		brandCircularBUUID:             "Circular Business School video - B",
		aboutTopic:                     "Ashes 2017",
		broaderTopicA:                  "The Ashes",
		broaderTopicB:                  "Cricket",
		narrowerTopic:                  "England Ashes 2017 Victory",
		cyclicTopicA:                   "Dodgy Cyclic Topic A",
		cyclicTopicB:                   "Dodgy Cyclic Topic B",
		brandWithHasBrandPredicateUUID: "Lex",
		brandHubPageUUID:               "Moral Money",
		genreOpinionUUID:               "Opinion",
		locationA:                      "Bulgaria",
		locationB:                      "Balkans",
		locationC:                      "Eastern Europe",
		locationD:                      "Balkan Peninsula",
		locationE:                      "Europe",
	}

	geonamesFeatureCodes = map[string]string{
		locationA: "http://www.geonames.org/ontology#A.PCLI",
		locationB: "http://www.geonames.org/ontology#L.RGN",
		locationC: "http://www.geonames.org/ontology#L.RGN",
		locationD: "http://www.geonames.org/ontology#T.PEN",
		locationE: "http://www.geonames.org/ontology#L.CONT",
	}

	conceptTypes = map[string][]string{
		brandType: {
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			brandType,
		},
		topicType: {
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			topicType,
		},
		genreType: {
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			genreType,
		},
		organisationType: {
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			organisationType,
		},
		locationType: {
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			locationType,
		},
	}

	conceptApiUrlTemplates = map[string]string{
		brandType:        "http://api.ft.com/brands/%s",
		topicType:        "http://api.ft.com/things/%s",
		genreType:        "http://api.ft.com/things/%s",
		locationType:     "http://api.ft.com/things/%s",
		organisationType: "http://api.ft.com/organisations/%s",
	}
)

func init() {
	// Used by concepts-rw-neo4j and if it's not initialized it
	// pollutes the output with a lot of useless log messages
	v1Logger.InitLogger("annotations_integration_tests", "PANIC")
}

type cypherDriverTestSuite struct {
	suite.Suite
	driver *cmneo4j.Driver
}

var allUUIDs = []string{contentUUID, contentWithNoAnnotationsUUID, contentWithParentAndChildBrandUUID,
	contentWithThreeLevelsOfBrandUUID, contentWithCircularBrandUUID, contentWithOnlyFTUUID, alphavilleSeriesUUID,
	brandParentUUID, brandChildUUID, brandGrandChildUUID, brandCircularAUUID, brandCircularBUUID, contentWithBrandsDiffTypesUUID,
	FakebookConceptUUID, MSJConceptUUID, MetalMickeyConceptUUID, brokenPacUUID, financialInstrumentUUID, JohnSmithConceptUUID,
	aboutTopic, broaderTopicA, broaderTopicB, narrowerTopic, cyclicTopicA, cyclicTopicB, brandWithHasBrandPredicateUUID,
	brandHubPageUUID, genreOpinionUUID, contentWithHasBrand, orgUUID, contentWithNAICSOrgUUID, NYTConceptUUID, NAICSConceptUUID,
}

func TestCypherDriverSuite(t *testing.T) {
	suite.Run(t, newCypherDriverTestSuite())
}

func newCypherDriverTestSuite() *cypherDriverTestSuite {
	return &cypherDriverTestSuite{}
}

func (s *cypherDriverTestSuite) SetupTest() {
	log := logger.NewUPPLogger("public-annotations-api-cm-neo4j", "PANIC")
	s.driver = getNeo4jDriver(s.T())
	writeAllDataToDB(s.T(), s.driver, log)
}

func (s *cypherDriverTestSuite) TearDownTest() {
	cleanDB(s.T(), s.driver)
}

func getNeo4jDriver(t *testing.T) *cmneo4j.Driver {
	if testing.Short() {
		t.Skip("Skipping Neo4j integration tests.")
		return nil
	}

	l := logger.NewUPPLogger("public-annotations-api-cm-neo4j", "PANIC")
	url := os.Getenv("NEO4J_TEST_URL")
	if url == "" {
		url = "bolt://localhost:7687"
	}

	driver, err := cmneo4j.NewDefaultDriver(url, l)
	require.NoError(t, err, "could not create a new cm-neo4j-driver")
	return driver
}

func (s *cypherDriverTestSuite) TestRetrieveMultipleAnnotations() {
	expectedAnnotations := Annotations{
		getExpectedMentionsFakebookAnnotation(),
		getExpectedMallStreetJournalAnnotation(),
		getExpectedMetalMickeyAnnotation(v1Lifecycle),
		getExpectedAlphavilleSeriesAnnotation(v1Lifecycle),
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrievePacAndV2AnnotationsAsPriority() {
	expectedAnnotations := Annotations{
		getExpectedMetalMickeyAnnotation(pacLifecycle),
		getExpectedHasDisplayTagFakebookAnnotation(pacLifecycle),
		getExpectedAboutFakebookAnnotation(pacLifecycle),
		getExpectedJohnSmithAnnotation(pacLifecycle),
		getExpectedMallStreetJournalAnnotation(),
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], pacLifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], pacLifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], pacLifecycle),
	}
	annotationsDriver := NewCypherDriver(s.driver, "prod")
	writePacAnnotations(s.T(), s.driver)
	//assert data for filtering
	numOfV1Annotations, _ := count(v1Lifecycle, s.driver)
	numOfV2Annotations, _ := count(v2Lifecycle, s.driver)
	numOfPACAnnotations, _ := count(pacLifecycle, s.driver)
	assert.True(s.T(), numOfV1Annotations > 0)
	assert.True(s.T(), numOfV2Annotations > 0)
	assert.True(s.T(), numOfPACAnnotations > 0)

	anns := getAndCheckAnnotations(annotationsDriver, contentUUID, s.T())

	assert.Len(s.T(), anns, len(expectedAnnotations), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveImplicitAbouts() {
	expectedAnnotations := Annotations{
		expectedAnnotation(aboutTopic, topicType, predicates["ABOUT"], pacLifecycle),
		expectedAnnotation(locationA, locationType, predicates["ABOUT"], pacLifecycle),
		expectedAnnotation(broaderTopicA, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(broaderTopicB, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(locationB, locationType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(locationC, locationType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(locationD, locationType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(locationE, locationType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		getExpectedMallStreetJournalAnnotation(),
		getExpectedMentionsFakebookAnnotation(),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	writeAboutAnnotations(s.T(), s.driver)

	anns := getAndCheckAnnotations(annotationsDriver, contentUUID, s.T())

	assert.Len(s.T(), anns, len(expectedAnnotations), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveCyclicImplicitAbouts() {
	expectedAnnotations := Annotations{
		expectedAnnotation(narrowerTopic, topicType, predicates["ABOUT"], pacLifecycle),
		expectedAnnotation(aboutTopic, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(broaderTopicA, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(broaderTopicB, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(cyclicTopicA, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		expectedAnnotation(cyclicTopicB, topicType, predicates["IMPLICITLY_ABOUT"], pacLifecycle),
		getExpectedMentionsFakebookAnnotation(),
		getExpectedMallStreetJournalAnnotation(),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	writeCyclicAboutAnnotations(s.T(), s.driver)

	anns := getAndCheckAnnotations(annotationsDriver, contentUUID, s.T())
	d, _ := json.MarshalIndent(anns, "", "   ")
	s.T().Log(string(d))

	assert.Len(s.T(), anns, len(expectedAnnotations), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveMultipleAnnotationsIfPacAnnotationCannotBeMapped() {
	expectedAnnotations := Annotations{
		getExpectedMentionsFakebookAnnotation(),
		getExpectedMallStreetJournalAnnotation(),
		getExpectedMetalMickeyAnnotation(v1Lifecycle),
		getExpectedAlphavilleSeriesAnnotation(v1Lifecycle),
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	writeBrokenPacAnnotations(s.T(), s.driver)
	//assert data for filtering
	numOfV1Annotations, _ := count(v1Lifecycle, s.driver)
	numOfv2Annotations, _ := count(v2Lifecycle, s.driver)
	numOfPacAnnotations, _ := count(pacLifecycle, s.driver)
	assert.True(s.T(), (numOfV1Annotations+numOfv2Annotations) > 0)
	assert.Equal(s.T(), numOfPacAnnotations, 1)

	anns := getAndCheckAnnotations(annotationsDriver, contentUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveContentWithParentBrand() {
	expectedAnnotations := Annotations{
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentWithParentAndChildBrandUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveContentWithGrandParentBrand() {
	expectedAnnotations := Annotations{
		expectedAnnotation(brandGrandChildUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandChildUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentWithThreeLevelsOfBrandUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveContentWithCircularBrand() {
	expectedAnnotations := Annotations{
		expectedAnnotation(brandCircularAUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandCircularBUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentWithCircularBrandUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveContentWithJustParentBrand() {
	expectedAnnotations := Annotations{
		expectedAnnotation(brandParentUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentWithOnlyFTUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

//Tests filtering Annotations where content is related to Brand A as isClassifiedBy and to Brand B as isPrimarilyClassifiedBy
// and Brands A and B have a circular relation HasParent
func (s *cypherDriverTestSuite) TestRetrieveContentBrandsOfDifferentTypes() {
	expectedAnnotations := Annotations{
		expectedAnnotation(brandCircularAUUID, brandType, predicates["IS_CLASSIFIED_BY"], v1Lifecycle),
		expectedAnnotation(brandCircularBUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], v1Lifecycle),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentWithCircularBrandUUID, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestRetrieveAnnotationWithHasBrand() {
	writeHasBrandAnnotations(s.T(), s.driver)

	expectedAnnotations := Annotations{
		expectedAnnotation(brandHubPageUUID, brandType, predicates["IS_CLASSIFIED_BY"], pacLifecycle),
		expectedAnnotation(brandWithHasBrandPredicateUUID, brandType, predicates["IS_CLASSIFIED_BY"], pacLifecycle),
		expectedAnnotation(genreOpinionUUID, genreType, predicates["IS_CLASSIFIED_BY"], pacLifecycle),
		expectedAnnotation(brandParentUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], pacLifecycle),
	}

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentWithHasBrand, s.T())
	assert.Equal(s.T(), len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(s.T(), anns, expectedAnnotations)
}

func (s *cypherDriverTestSuite) TestTransitivePropertyOfImpliedBy() {
	t := s.T()
	driver := s.driver

	contentRW := content.NewContentService(driver)
	assert.NoError(t, contentRW.Initialise())

	log := logger.NewUPPLogger("public-annotations-api-test", "PANIC")
	conceptRW := concepts.NewConceptService(driver, log)
	assert.NoError(t, conceptRW.Initialise())

	annotationRW := annrw.NewCypherAnnotationsService(s.driver)
	assert.NoError(t, annotationRW.Initialise())

	writeContent := func(fixture string) string {
		writeJSONToBaseService(contentRW, fixture, t)
		data := readJSONFile(t, fixture)
		uuid, _ := data["uuid"].(string)
		return uuid
	}
	writeConcept := func(fixture string) (string, string) {
		writeJSONToService(conceptRW, fixture, t)
		data := readJSONFile(t, fixture)
		uuid, _ := data["prefUUID"].(string)
		label, _ := data["prefLabel"].(string)
		return uuid, label
	}
	removeUUIDs := []string{}
	expected := []Annotation{}

	contentID := writeContent("./testdata/testImplicitlyClassifiedBy/content.json")
	removeUUIDs = append(removeUUIDs, contentID)

	concepts := []struct {
		Fixture   string // concept fixture
		Type      string // expected concept type
		Predicate string // expected annotations predicate
	}{
		{Fixture: "./testdata/testImplicitlyClassifiedBy/topic2-about.json", Type: topicType, Predicate: "ABOUT"},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/topic1-mentions.json", Type: topicType, Predicate: "MENTIONS"},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/organisation1-about.json", Type: organisationType, Predicate: "ABOUT"},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/brand1-isClassifiedBy.json", Type: brandType, Predicate: "IS_CLASSIFIED_BY"},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/topic3-broader-topic2.json", Type: topicType, Predicate: "IMPLICITLY_ABOUT"},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/topic4-impliedBy-organisation1.json", Type: topicType},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/brand6-impliedBy-organisation2.json", Type: organisationType},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/brand2-impliedBy-topic2.json", Type: brandType, Predicate: "IMPLICITLY_CLASSIFIED_BY"},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/brand5-parent-brand1.json", Type: brandType, Predicate: "IMPLICITLY_CLASSIFIED_BY"},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/topic5-broader-topic3.json", Type: topicType, Predicate: "IMPLICITLY_ABOUT"},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/brand4-parent-brand2.json", Type: brandType},
		{Fixture: "./testdata/testImplicitlyClassifiedBy/brand3-impliedBy-topic3.json", Type: brandType},
	}

	for _, c := range concepts {
		UUID, prefLabel := writeConcept(c.Fixture)
		removeUUIDs = append(removeUUIDs, UUID)
		if c.Predicate == "" {
			continue
		}
		expected = append(expected, expectedAnnotationWithPrefLabel(UUID, c.Type, predicates[c.Predicate], prefLabel))
	}

	writeJSONToAnnotationsService(t, annotationRW, "pac", "annotations-pac", contentID, "./testdata/testImplicitlyClassifiedBy/annotations.json")

	annotationsDriver := NewCypherDriver(s.driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentID, t)
	assert.Equal(t, len(expected), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expected)
	deleteUUIDs(t, s.driver, removeUUIDs)
}

func (s *cypherDriverTestSuite) TestRetrieveAnnotationsWithImpliedBy() {

	//setup
	t := s.T()
	driver := s.driver

	contentRW := content.NewContentService(driver)
	assert.NoError(t, contentRW.Initialise())

	log := logger.NewUPPLogger("public-annotations-api-test", "PANIC")
	conceptRW := concepts.NewConceptService(driver, log)
	assert.NoError(t, conceptRW.Initialise())

	annotationRW := annrw.NewCypherAnnotationsService(s.driver)
	assert.NoError(t, annotationRW.Initialise())

	writeConcept := func(fixture string) (string, string) {
		writeJSONToService(conceptRW, fixture, t)
		data := readJSONFile(t, fixture)

		uuid, ok := data["prefUUID"].(string)
		if !ok {
			t.Fatalf("in fixture %s prefUUID is not a string", fixture)
		}
		label, ok := data["prefLabel"].(string)
		if !ok {
			t.Fatalf("in fixture %s prefLabel is not a string", fixture)
		}
		return uuid, label
	}

	writeContent := func(fixture string) string {
		writeJSONToBaseService(contentRW, fixture, t)
		data := readJSONFile(t, fixture)
		uuid, ok := data["uuid"].(string)
		if !ok {
			t.Fatalf("in fixture %s uuid is not a string", fixture)
		}
		return uuid
	}

	contentID := writeContent("./testdata/impliedBy/content.json")
	brandUUID, brandLabel := writeConcept("./testdata/impliedBy/brand-hub-page.json")
	topicUUID, topicLabel := writeConcept("./testdata/impliedBy/topic-implies-brand.json")
	cleanUUIDs := []string{topicUUID, contentID, brandUUID}

	tests := map[string]struct {
		Annotations         string
		ExpectedAnnotations Annotations
	}{
		"implied by concept should return implicitly classified by": {
			Annotations: "./testdata/impliedBy/annotation-topic-about.json",
			ExpectedAnnotations: Annotations{
				expectedAnnotationWithPrefLabel(topicUUID, topicType, predicates["ABOUT"], topicLabel),
				expectedAnnotationWithPrefLabel(brandUUID, brandType, predicates["IMPLICITLY_CLASSIFIED_BY"], brandLabel),
			},
		},
		"direct isClassifiedBy annotations should override implicit ones": {
			Annotations: "./testdata/impliedBy/annotation-topic-and-brand-is-classified-by.json",
			ExpectedAnnotations: Annotations{
				expectedAnnotationWithPrefLabel(topicUUID, topicType, predicates["ABOUT"], topicLabel),
				expectedAnnotationWithPrefLabel(brandUUID, brandType, predicates["IS_CLASSIFIED_BY"], brandLabel),
			},
		},
		"direct hasBrand annotations should override implicit ones": {
			Annotations: "./testdata/impliedBy/annotation-topic-and-brand-has-brand.json",
			ExpectedAnnotations: Annotations{
				expectedAnnotationWithPrefLabel(topicUUID, topicType, predicates["ABOUT"], topicLabel),
				expectedAnnotationWithPrefLabel(brandUUID, brandType, predicates["HAS_BRAND"], brandLabel),
			},
		},
		"isClassifiedBy should be with greatest priority": {
			Annotations: "./testdata/impliedBy/annotation-topic-and-brand-multiple-ann.json",
			ExpectedAnnotations: Annotations{
				expectedAnnotationWithPrefLabel(topicUUID, topicType, predicates["ABOUT"], topicLabel),
				expectedAnnotationWithPrefLabel(brandUUID, brandType, predicates["IS_CLASSIFIED_BY"], brandLabel),
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			writeJSONToAnnotationsService(t, annotationRW, "pac", "annotations-pac", contentID, test.Annotations)

			annotationsDriver := NewCypherDriver(s.driver, "prod")
			anns := getAndCheckAnnotations(annotationsDriver, contentID, t)
			assert.Equal(t, len(test.ExpectedAnnotations), len(anns), "Didn't get the same number of annotations")
			assertListContainsAll(t, anns, test.ExpectedAnnotations)
		})
	}
	deleteUUIDs(t, s.driver, cleanUUIDs)
}

func TestRetrieveNoAnnotationsWhenThereAreNonePresentExceptBrands(t *testing.T) {
	assert := assert.New(t)
	driver := getNeo4jDriver(t)
	log := logger.NewUPPLogger("public-annotations-api-test", "PANIC")

	writeContent(t, driver)
	writeBrands(t, driver, log)

	defer cleanDB(t, driver)

	annotationsDriver := NewCypherDriver(driver, "prod")
	anns, found, err := annotationsDriver.read(contentWithNoAnnotationsUUID)
	anns = applyDefaultFilters(anns)
	assert.NoError(err, "Unexpected error for content %s", contentWithNoAnnotationsUUID)
	assert.False(found, "Found annotations for content %s", contentWithNoAnnotationsUUID)
	assert.Equal(0, len(anns), "Didn't get the same number of annotations") // Two brands, child and parent
}

func TestRetrieveAnnotationWithCorrectValues(t *testing.T) {
	assert := assert.New(t)
	d := getNeo4jDriver(t)
	log := logger.NewUPPLogger("public-annotations-api-test", "PANIC")

	writeContent(t, d)
	writeOrganisations(t, d, log)
	writeFinancialInstruments(t, d, log)
	writeV2Annotations(t, d)
	defer cleanDB(t, d)

	expectedAnnotations := Annotations{
		getExpectedMentionsFakebookAnnotation(),
		getExpectedMallStreetJournalAnnotation(),
	}

	annotationsDriver := NewCypherDriver(d, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentUUID, t)

	assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)

	for _, ann := range anns {
		for _, expected := range expectedAnnotations {
			if expected.ID == ann.ID {
				assert.Equal(expected.FIGI, ann.FIGI, "Didn't get the expected FIGI value")
				assert.Equal(expected.LeiCode, ann.LeiCode, "Didn't get the expected Leicode value")
				break
			}
		}
	}
}

func TestRetrieveNoAnnotationsWhenThereAreNoConceptsPresent(t *testing.T) {
	assert := assert.New(t)
	driver := getNeo4jDriver(t)

	writeContent(t, driver)
	writeV1Annotations(t, driver)
	writeV2Annotations(t, driver)

	defer cleanDB(t, driver)

	annotationsDriver := NewCypherDriver(driver, "prod")
	anns, found, err := annotationsDriver.read(contentUUID)
	anns = applyDefaultFilters(anns)
	assert.NoError(err, "Unexpected error for content %s", contentUUID)
	assert.False(found, "Found annotations for content %s", contentUUID)
	assert.Equal(0, len(anns), "Didn't get the same number of annotations, anns=%s", anns)
}

func TestRetrieveAnnotationsWithNAICSOrganisation(t *testing.T) {
	assert := assert.New(t)
	driver := getNeo4jDriver(t)
	annService := annrw.NewCypherAnnotationsService(driver)
	assert.NoError(annService.Initialise())
	log := logger.NewUPPLogger("public-annotations-api-test", "PANIC")

	writeContent(t, driver)
	writeOrganisations(t, driver, log)
	writeFinancialInstruments(t, driver, log)
	writeV2Annotations(t, driver)
	defer cleanDB(t, driver)

	annotationsDriver := NewCypherDriver(driver, "prod")
	anns := getAndCheckAnnotations(annotationsDriver, contentWithNAICSOrgUUID, t)

	expectedAnnotations := Annotations{
		getExpectedNewYorkshireTimesAnnotation(v2Lifecycle),
		getExpectedMentionsFakebookAnnotation(),
	}

	assert.Equal(len(expectedAnnotations), len(anns), "Didn't get the same number of annotations")
	assertListContainsAll(t, anns, expectedAnnotations)

	for _, ann := range anns {
		for _, expected := range expectedAnnotations {
			if expected.ID == ann.ID {
				assert.Equal(expected.NAICS, ann.NAICS, "Didn't get the expected NAICS details")
				break
			}
		}
	}
}

func getAndCheckAnnotations(driver CypherDriver, contentUUID string, t *testing.T) Annotations {
	anns, found, err := driver.read(contentUUID)
	anns = applyDefaultFilters(anns)
	assert.NoError(t, err, "Unexpected error for content %s", contentUUID)
	assert.True(t, found, "Found no annotations for content %s", contentUUID)
	return anns
}

// Utility functions
func writeAllDataToDB(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) {
	writeBrands(t, d, log)
	writeContent(t, d)
	writeOrganisations(t, d, log)
	writePeople(t, d, log)
	writeFinancialInstruments(t, d, log)
	writeSubjects(t, d, log)
	writeAlphavilleSeries(t, d, log)
	writeGenres(t, d, log)
	writeV1Annotations(t, d)
	writeV2Annotations(t, d)
	writeTopics(t, d, log)
	writeLocations(t, d, log)
}

func writeBrands(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) {
	brandRW := concepts.NewConceptService(d, log)
	assert.NoError(t, brandRW.Initialise())
	writeJSONToService(brandRW, "./testdata/Brand-dbb0bdae-1f0c-1a1a-b0cb-b2227cce2b54-parent.json", t)
	writeJSONToService(brandRW, "./testdata/Brand-ff691bf8-8d92-1a1a-8326-c273400bff0b-child.json", t)
	writeJSONToService(brandRW, "./testdata/Brand-ff691bf8-8d92-2a2a-8326-c273400bff0b-grand_child.json", t)
	writeJSONToService(brandRW, "./testdata/Brand-ff691bf8-8d92-3a3a-8326-c273400bff0b-circular_a.json", t)
	writeJSONToService(brandRW, "./testdata/Brand-ff691bf8-8d92-4a4a-8326-c273400bff0b-circular_b.json", t)
	writeJSONToService(brandRW, "./testdata/Brand-2d3e16e0-61cb-4322-8aff-3b01c59f4daa-true-brand.json", t)
	writeJSONToService(brandRW, "./testdata/Brand-87645070-7d8a-492e-9695-bf61ac2b4d18-hub-page.json", t)
}

func writeContent(t testing.TB, d *cmneo4j.Driver) {
	contentRW := content.NewContentService(d)
	assert.NoError(t, contentRW.Initialise())
	writeJSONToBaseService(contentRW, "./testdata/Content-3fc9fe3e-af8c-4f7f-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./testdata/Content-3fc9fe3e-af8c-1a1a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./testdata/Content-3fc9fe3e-af8c-2a2a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./testdata/Content-3fc9fe3e-af8c-3a3a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./testdata/Content-3fc9fe3e-af8c-4a4a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./testdata/Content-3fc9fe3e-af8c-5a5a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./testdata/Content-3fc9fe3e-af8c-6a6a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./testdata/Content-3fc9fe3e-af8c-7a7a-961a-e5065392bb31.json", t)
	writeJSONToBaseService(contentRW, "./testdata/Content-ae17012e-ad40-11e9-8030-530adfa879c2.json", t)
}

func writeTopics(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) concepts.ConceptService {
	topicsRW := concepts.NewConceptService(d, log)
	assert.NoError(t, topicsRW.Initialise())
	writeJSONToService(topicsRW, "./testdata/Topics-7e22c8b8-b280-4e52-aa22-fa1c6dffd894.json", t)
	writeJSONToService(topicsRW, "./testdata/Topics-b6469cc2-f6ff-45aa-a9bb-3d1bb0f9a35d.json", t)
	writeJSONToService(topicsRW, "./testdata/Topics-ca982370-66cd-43bd-b2e3-7bfcb73efb1e.json", t)
	writeJSONToService(topicsRW, "./testdata/Topics-fde5eee9-3260-4125-adb6-3d91a4888be5.json", t)
	writeJSONToService(topicsRW, "./testdata/Topics-77a410a3-6857-4654-80ef-6aae29be852a.json", t)
	writeJSONToService(topicsRW, "./testdata/Topics-e404e3bd-beff-4324-83f4-beb044baf916.json", t)
	return topicsRW
}

func writeOrganisations(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) {
	organisationRW := concepts.NewConceptService(d, log)
	assert.NoError(t, organisationRW.Initialise())
	writeJSONToService(organisationRW, "./testdata/Organisation-MSJ-5d1510f8-2779-4b74-adab-0a5eb138fca6.json", t)
	writeJSONToService(organisationRW, "./testdata/Organisation-Fakebook-eac853f5-3859-4c08-8540-55e043719400.json", t)
	writeJSONToService(organisationRW, "./testdata/NAICSIndustryClassification-38ee195d-ebdd-48a9-af4b-c8a322e7b04d.json", t)
	writeJSONToService(organisationRW, "./testdata/Organisation-NYT-0d9fbdfc-7d95-332b-b77b-1e69274b1b83.json", t)
}

func writeLocations(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) {
	locationRW := concepts.NewConceptService(d, log)
	assert.NoError(t, locationRW.Initialise())
	writeJSONToService(locationRW, "./testdata/Location-82cba3ce-329b-3010-b29d-4282a215889f.json", t)
	writeJSONToService(locationRW, "./testdata/Location-8d54e308-807c-4e9e-9981-8faab10b6f1c.json", t)
	writeJSONToService(locationRW, "./testdata/Location-5895ee1e-d5de-39c1-93ab-03fcb7d36caf.json", t)
	writeJSONToService(locationRW, "./testdata/Location-307c91ed-31f5-33b7-895a-1ffbeec514f4.json", t)
	writeJSONToService(locationRW, "./testdata/Location-822e3c99-afc6-3c55-b497-2255ac546f35.json", t)
}

func writePeople(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) concepts.ConceptService {
	peopleRW := concepts.NewConceptService(d, log)
	assert.NoError(t, peopleRW.Initialise())
	writeJSONToService(peopleRW, "./testdata/People-75e2f7e9-cb5e-40a5-a074-86d69fe09f69.json", t)
	return peopleRW
}

func writeFinancialInstruments(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) {
	fiRW := concepts.NewConceptService(d, log)
	assert.NoError(t, fiRW.Initialise())
	writeJSONToService(fiRW, "./testdata/FinancialInstrument-77f613ad-1470-422c-bf7c-1dd4c3fd1693.json", t)
}

func writeSubjects(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) concepts.ConceptService {
	subjectsRW := concepts.NewConceptService(d, log)
	assert.NoError(t, subjectsRW.Initialise())
	writeJSONToService(subjectsRW, "./testdata/Subject-MetalMickey-0483bef8-5797-40b8-9b25-b12e492f63c6.json", t)
	return subjectsRW
}

func writeAlphavilleSeries(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) concepts.ConceptService {
	alphavilleSeriesRW := concepts.NewConceptService(d, log)
	assert.NoError(t, alphavilleSeriesRW.Initialise())
	writeJSONToService(alphavilleSeriesRW, "./testdata/TestAlphavilleSeries.json", t)
	return alphavilleSeriesRW
}

func writeGenres(t testing.TB, d *cmneo4j.Driver, log *logger.UPPLogger) {
	genresRW := concepts.NewConceptService(d, log)
	assert.NoError(t, genresRW.Initialise())
	writeJSONToService(genresRW, "./testdata/Genre-6da31a37-691f-4908-896f-2829ebe2309e-opinion.json", t)
}

func writeV1Annotations(t testing.TB, driver *cmneo4j.Driver) {
	service := annrw.NewCypherAnnotationsService(driver)
	assert.NoError(t, service.Initialise())

	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentUUID, "./testdata/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithParentAndChildBrandUUID, "./testdata/Annotations-3fc9fe3e-af8c-2a2a-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithThreeLevelsOfBrandUUID, "./testdata/Annotations-3fc9fe3e-af8c-3a3a-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithCircularBrandUUID, "./testdata/Annotations-3fc9fe3e-af8c-4a4a-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithOnlyFTUUID, "./testdata/Annotations-3fc9fe3e-af8c-5a5a-961a-e5065392bb31-v1.json")
	writeJSONToAnnotationsService(t, service, v1PlatformVersion, v1Lifecycle, contentWithBrandsDiffTypesUUID, "./testdata/Annotations-3fc9fe3e-af8c-6a6a-961a-e5065392bb31-v1.json")
}

func writeV2Annotations(t testing.TB, driver *cmneo4j.Driver) {
	service := annrw.NewCypherAnnotationsService(driver)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, v2PlatformVersion, v2Lifecycle, contentUUID, "./testdata/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-v2.json")
	writeJSONToAnnotationsService(t, service, v2PlatformVersion, v2Lifecycle, contentWithNAICSOrgUUID, "./testdata/Annotations-3fc9fe3e-af8c-7a7a-961a-e5065392bb31-v2.json")
}

func writePacAnnotations(t testing.TB, driver *cmneo4j.Driver) {
	service := annrw.NewCypherAnnotationsService(driver)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, "pac", "annotations-pac", contentUUID, "./testdata/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-pac.json")
}

func writeHasBrandAnnotations(t testing.TB, driver *cmneo4j.Driver) {
	service := annrw.NewCypherAnnotationsService(driver)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, "pac", "annotations-pac", contentWithHasBrand, "./testdata/Annotations-ae17012e-ad40-11e9-8030-530adfa879c2-pac.json")
}

func writeAboutAnnotations(t testing.TB, driver *cmneo4j.Driver) {
	service := annrw.NewCypherAnnotationsService(driver)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, "pac", "annotations-pac", contentUUID, "./testdata/Annotations-ca982370-66cd-43bd-b2e3-7bfcb73efb1e-and-82cba3ce-329b-3010-b29d-4282a215889f-implicit-abouts.json")
}

func writeCyclicAboutAnnotations(t testing.TB, driver *cmneo4j.Driver) {
	service := annrw.NewCypherAnnotationsService(driver)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, "pac", "annotations-pac", contentUUID, "./testdata/Annotations-7e22c8b8-b280-4e52-aa22-fa1c6dffd894-cyclic-implicit-abouts.json")
}

func writeBrokenPacAnnotations(t testing.TB, driver *cmneo4j.Driver) {
	service := annrw.NewCypherAnnotationsService(driver)
	assert.NoError(t, service.Initialise())
	writeJSONToAnnotationsService(t, service, emptyPlatformVersion, pacLifecycle, contentUUID, "./testdata/Annotations-3fc9fe3e-af8c-4f7f-961a-e5065392bb31-broken-pac.json")
}

func writeJSONToBaseService(service baseftrwapp.Service, pathToJSONFile string, t testing.TB) {
	absPath, _ := filepath.Abs(pathToJSONFile)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	assert.NoError(t, err)
	err = service.Write(inst, "TEST_TRANS_ID")
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)
}

func writeJSONToService(service concepts.ConceptService, pathToJSONFile string, t testing.TB) {
	absPath, _ := filepath.Abs(pathToJSONFile)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, _, err := service.DecodeJSON(dec)
	assert.NoError(t, err)
	_, err = service.Write(inst, "TEST_TRANS_ID")
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)
}

func writeJSONToAnnotationsService(t testing.TB, service annrw.Service, platformVersion string, lifecycle string, contentUUID string, pathToJSONFile string) {
	absPath, _ := filepath.Abs(pathToJSONFile)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	dec := json.NewDecoder(f)
	inst, err := service.DecodeJSON(dec)
	assert.NoError(t, err, "Error parsing file %s", pathToJSONFile)
	err = service.Write(contentUUID, lifecycle, platformVersion, "TID_TEST", inst)
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)
}

func assertListContainsAll(t *testing.T, list interface{}, items ...interface{}) {
	if reflect.TypeOf(items[0]).Kind().String() == "slice" {
		expected := reflect.ValueOf(items[0])
		expectedLength := expected.Len()
		for i := 0; i < expectedLength; i++ {
			assert.Contains(t, list, expected.Index(i).Interface())
		}
	} else {
		for _, item := range items {
			assert.Contains(t, list, item)
		}
	}
}

func deleteUUIDs(t testing.TB, driver *cmneo4j.Driver, uuids []string) {
	qs := make([]*cmneo4j.Query, len(uuids))
	for i, uuid := range uuids {
		qs[i] = &cmneo4j.Query{
			Cypher: `
			    MATCH (a:Thing {uuid: $thingUUID})
			    OPTIONAL MATCH (a)-[:EQUIVALENT_TO]-(t:Thing)
			    DELETE t
			    DETACH DELETE a`,
			Params: map[string]interface{}{
				"thingUUID": uuid,
			},
		}
	}
	err := driver.Write(qs...)
	assert.NoError(t, err)
}

func cleanDB(t testing.TB, driver *cmneo4j.Driver) {
	deleteUUIDs(t, driver, allUUIDs)
}

func readJSONFile(t testing.TB, fixture string) map[string]interface{} {

	absPath, _ := filepath.Abs(fixture)
	f, err := os.Open(absPath)
	assert.NoError(t, err)
	data := map[string]interface{}{}
	err = json.NewDecoder(f).Decode(&data)
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)
	return data
}

func getExpectedMentionsFakebookAnnotation() Annotation {
	return Annotation{
		Predicate: "http://www.ft.com/ontology/annotation/mentions",
		ID:        "http://api.ft.com/things/eac853f5-3859-4c08-8540-55e043719400",
		APIURL:    "http://api.ft.com/organisations/eac853f5-3859-4c08-8540-55e043719400",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/organisation/Organisation",
			"http://www.ft.com/ontology/company/Company",
			"http://www.ft.com/ontology/company/PublicCompany",
		},
		LeiCode:   "BQ4BKCS1HXDV9TTTTTTTT",
		FIGI:      "BB8000C3P0-R2D2",
		PrefLabel: "Fakebook, Inc.",
		Lifecycle: "annotations-v2",
	}
}

func getExpectedAboutFakebookAnnotation(lifecycle string) Annotation {
	return Annotation{
		Predicate: "http://www.ft.com/ontology/annotation/about",
		ID:        "http://api.ft.com/things/eac853f5-3859-4c08-8540-55e043719400",
		APIURL:    "http://api.ft.com/organisations/eac853f5-3859-4c08-8540-55e043719400",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/organisation/Organisation",
			"http://www.ft.com/ontology/company/Company",
			"http://www.ft.com/ontology/company/PublicCompany",
		},
		LeiCode:   "BQ4BKCS1HXDV9TTTTTTTT",
		FIGI:      "BB8000C3P0-R2D2",
		PrefLabel: "Fakebook, Inc.",
		Lifecycle: lifecycle,
	}
}

func getExpectedMallStreetJournalAnnotation() Annotation {
	return Annotation{
		Predicate: "http://www.ft.com/ontology/annotation/mentions",
		ID:        "http://api.ft.com/things/5d1510f8-2779-4b74-adab-0a5eb138fca6",
		APIURL:    "http://api.ft.com/organisations/5d1510f8-2779-4b74-adab-0a5eb138fca6",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/organisation/Organisation",
		},
		PrefLabel: "The Mall Street Journal",
		Lifecycle: "annotations-v2",
	}
}

func getExpectedNewYorkshireTimesAnnotation(lifecycle string) Annotation {
	return Annotation{
		Predicate: "http://www.ft.com/ontology/annotation/mentions",
		ID:        "http://api.ft.com/things/" + NYTConceptUUID,
		APIURL:    "http://api.ft.com/organisations/" + NYTConceptUUID,
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/organisation/Organisation",
		},
		PrefLabel: "The New Yorkshire Times",
		Lifecycle: lifecycle,
		NAICS: []IndustryClassification{
			{
				PrefLabel:  "Newspaper, Periodical, Book, and Directory Publishers",
				Identifier: "5111-test",
				Rank:       1,
			},
		},
	}
}

func getExpectedMetalMickeyAnnotation(lifecycle string) Annotation {
	return Annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/0483bef8-5797-40b8-9b25-b12e492f63c6",
		APIURL:    "http://api.ft.com/things/0483bef8-5797-40b8-9b25-b12e492f63c6",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/Subject",
		},
		PrefLabel: "Metal Mickey",
		Lifecycle: lifecycle,
	}
}

func getExpectedHasDisplayTagFakebookAnnotation(lifecycle string) Annotation {
	return Annotation{
		Predicate: "http://www.ft.com/ontology/hasDisplayTag",
		ID:        "http://api.ft.com/things/eac853f5-3859-4c08-8540-55e043719400",
		APIURL:    "http://api.ft.com/organisations/eac853f5-3859-4c08-8540-55e043719400",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/organisation/Organisation",
			"http://www.ft.com/ontology/company/Company",
			"http://www.ft.com/ontology/company/PublicCompany",
		},
		PrefLabel:    "Fakebook, Inc.",
		Lifecycle:    lifecycle,
		LeiCode:      "BQ4BKCS1HXDV9TTTTTTTT",
		FIGI:         "BB8000C3P0-R2D2",
		IsDeprecated: false,
	}
}

func getExpectedJohnSmithAnnotation(lifecycle string) Annotation {
	return Annotation{
		Predicate: "http://www.ft.com/ontology/hasContributor",
		ID:        "http://api.ft.com/things/75e2f7e9-cb5e-40a5-a074-86d69fe09f69",
		APIURL:    "http://api.ft.com/people/75e2f7e9-cb5e-40a5-a074-86d69fe09f69",
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/person/Person",
		},
		PrefLabel:    "John Smith",
		Lifecycle:    lifecycle,
		IsDeprecated: true,
	}
}

func getExpectedAlphavilleSeriesAnnotation(lifecycle string) Annotation {
	return Annotation{
		Predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
		ID:        "http://api.ft.com/things/" + alphavilleSeriesUUID,
		APIURL:    "http://api.ft.com/things/" + alphavilleSeriesUUID,
		Types: []string{
			"http://www.ft.com/ontology/core/Thing",
			"http://www.ft.com/ontology/concept/Concept",
			"http://www.ft.com/ontology/classification/Classification",
			"http://www.ft.com/ontology/AlphavilleSeries",
		},
		PrefLabel: "Test Alphaville Series",
		Lifecycle: lifecycle,
	}
}

func expectedAnnotation(conceptUUID string, conceptType string, predicate string, lifecycle string) Annotation {
	return Annotation{
		Predicate:           predicate,
		ID:                  fmt.Sprintf("http://api.ft.com/things/%s", conceptUUID),
		APIURL:              fmt.Sprintf(conceptApiUrlTemplates[conceptType], conceptUUID),
		Types:               conceptTypes[conceptType],
		PrefLabel:           conceptLabels[conceptUUID],
		GeonamesFeatureCode: geonamesFeatureCodes[conceptUUID],
		Lifecycle:           lifecycle,
	}
}

func expectedAnnotationWithPrefLabel(conceptUUID string, conceptType string, predicate string, prefLabel string) Annotation {
	return Annotation{
		Predicate: predicate,
		ID:        fmt.Sprintf("http://api.ft.com/things/%s", conceptUUID),
		APIURL:    fmt.Sprintf(conceptApiUrlTemplates[conceptType], conceptUUID),
		Types:     conceptTypes[conceptType],
		PrefLabel: prefLabel,
		Lifecycle: "annotations-pac",
	}
}

func count(annotationLifecycle string, driver *cmneo4j.Driver) (int, error) {
	var results []struct {
		Count int `json:"c"`
	}
	query := &cmneo4j.Query{
		Cypher: `MATCH (c:Content)-[r]->( t:Thing)
				 WHERE r.lifecycle = $lifecycle
                 RETURN count(r) as c`,
		Params: map[string]interface{}{"lifecycle": annotationLifecycle},
		Result: &results,
	}

	err := driver.Read(query)
	if errors.Is(err, cmneo4j.ErrNoResultsFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return results[0].Count, nil
}

func applyDefaultFilters(anns []Annotation) []Annotation {
	lifecycleFilter := newLifecycleFilter()
	predicateFilter := NewAnnotationsPredicateFilter()
	chain := newAnnotationsFilterChain(lifecycleFilter, predicateFilter)
	return chain.doNext(anns)
}
