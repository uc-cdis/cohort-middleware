package tests

import (
	"log"
	"os"
	"testing"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/models"
)

//var testSourceId = 1 // TODO - this should also be used when populating "source" tables in test Atlas DB in the first place...see also comment in setupSuite...

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
	//db.Init()
	// ensure we start w/ empty db:
	//tearDownSuite()
	// load test seed data:
	//tests.ExecSQLScript("../setup_local_db/test_data_results_and_cdm.sql", testSourceId)
}

func tearDownSuite() {
	log.Println("teardown for suite")
	//tests.ExecSQLScript("../setup_local_db/ddl_results_and_cdm.sql", testSourceId)
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
	t.Skip() // skipping for now...
	setUp(t)
	sourceId := 1 //testSourceId
	concepts, _ := conceptModel.RetriveAllBySourceId(sourceId)
	if len(concepts) != 4 {
		t.Errorf("Found %d", len(concepts))
	}
}
