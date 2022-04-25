package tests

import (
	"log"
	"os"
	"testing"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/tests"
)

var testSourceId = 1 // TODO - this should also be used when populating "source" tables in test Atlas DB in the first place...see also comment in setupSuite...
var firstCohort *models.CohortDefinition
var allConceptIds []int

func TestMain(m *testing.M) {
	setupSuite()
	retCode := m.Run()
	tearDownSuite()
	os.Exit(retCode)
}

func setupSuite() {
	log.Println("setup for suite")
	// connect to db with test data:
	// TODO - this needs to be improved to also
	//   populate the Atlas DB...now the tests assume an Atlas DB
	//   with specific data that is initialized elsewhere...
	config.Init("development")
	db.Init()
	// ensure we start w/ empty db:
	tearDownSuite()
	// load test seed data:
	tests.ExecSQLScript("../setup_local_db/test_data_results_and_cdm.sql", testSourceId)

	// initialize some handy variables to use in tests below:
	cohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitions()
	firstCohort = cohortDefinitions[0]
	concepts, _ := conceptModel.RetriveAllBySourceId(testSourceId)
	allConceptIds = tests.MapIntAttr(concepts, "ConceptId")
}

func tearDownSuite() {
	log.Println("teardown for suite")
	tests.ExecAtlasSQLScript("../setup_local_db/ddl_atlas.sql")
	// we need some basic atlas data in "source" table to be able to connect to results DB:  TODO - make this a minimal data .sql
	tests.ExecAtlasSQLScript("../setup_local_db/test_data_atlas.sql")
	tests.ExecSQLScript("../setup_local_db/ddl_results_and_cdm.sql", testSourceId)
}

func setUp(t *testing.T) {
	log.Println("setup for test")

	// ensure tearDown is called when test "t" is done:
	t.Cleanup(func() {
		tearDown()
	})
}

func tearDown() {
	log.Println("teardown for test")
}

var conceptModel = new(models.Concept)
var cohortDefinitionModel = new(models.CohortDefinition)

func TestGetConceptId(t *testing.T) {
	setUp(t)
	conceptId := conceptModel.GetConceptId("ID_12345")
	if conceptId != 12345 {
		t.Error()
	}
}

func TestGetPrefixedConceptId(t *testing.T) {
	setUp(t)
	conceptId := conceptModel.GetPrefixedConceptId(12345)
	if conceptId != "ID_12345" {
		t.Error()
	}
}

func TestRetriveAllBySourceId(t *testing.T) {
	setUp(t)
	concepts, _ := conceptModel.RetriveAllBySourceId(testSourceId)
	if len(concepts) != 4 {
		t.Errorf("Found %d", len(concepts))
	}
}

func TestRetrieveStatsBySourceIdAndCohortIdAndConceptIds(t *testing.T) {
	setUp(t)
	conceptsStats, _ := conceptModel.RetrieveStatsBySourceIdAndCohortIdAndConceptIds(testSourceId,
		firstCohort.Id,
		allConceptIds)
	// simple test: we expect stats for each valid conceptId, therefore the lists are
	//  expected to have the same lenght here:
	if len(conceptsStats) != len(allConceptIds) {
		t.Errorf("Found %d", len(conceptsStats))
	}
}

func TestRetrieveInfoBySourceIdAndConceptIds(t *testing.T) {
	setUp(t)
	conceptsInfo, _ := conceptModel.RetrieveInfoBySourceIdAndConceptIds(testSourceId,
		allConceptIds)
	// simple test: we expect info for each valid conceptId, therefore the lists are
	//  expected to have the same lenght here:
	if len(conceptsInfo) != len(allConceptIds) {
		t.Errorf("Found %d", len(conceptsInfo))
	}
}
