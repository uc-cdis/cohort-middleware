package tests

import (
	"log"
	"os"
	"testing"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
)

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
	//   populate the DB...now the tests assume a DB
	//   with specific data that is initialized elsewhere...
	config.Init("development")
	db.Init()
}

func tearDownSuite() {
	log.Println("teardown for suite")
	// nothing to do for now...
}

// TODO - move this to a generic test_utils
func RunTestWrapper(setUp func(), tearDown func()) func(*testing.T, func()) {
	return func(t *testing.T, testFunc func()) {
		setUp()
		t.Cleanup(func() {
			tearDown()
		})
		testFunc()
	}
}

var RunTest = RunTestWrapper(setUp, tearDown)

func setUp() {
	log.Println("setup for test")
}

func tearDown() {
	log.Println("teardown for test")
}

var conceptModel = new(models.Concept)

func TestGetConceptId(t *testing.T) {
	RunTest(t, func() {
		conceptId := conceptModel.GetConceptId("ID_12345")
		if conceptId != 12345 {
			t.Error()
		}
	})
}

func TestGetPrefixedConceptId(t *testing.T) {
	RunTest(t, func() {
		conceptId := conceptModel.GetPrefixedConceptId(12345)
		if conceptId != "ID_12345" {
			t.Error()
		}
	})
}

func TestRetriveAllBySourceId(t *testing.T) {
	RunTest(t, func() {
		sourceId := 1 // TODO - this should not be hard-coded here, but come from a central place that is also used for populating test DB in the first place...see also comment in setupSuite...
		concepts, _ := conceptModel.RetriveAllBySourceId(sourceId)
		if len(concepts) != 4 {
			t.Error()
		}
	})
}
