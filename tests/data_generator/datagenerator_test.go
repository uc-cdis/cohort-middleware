package main

import (
	"log"
	"os"
	"testing"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/tests"
)

func TestMain(m *testing.M) {
	setupSuite()
	retCode := m.Run()
	tearDownSuite()
	os.Exit(retCode)
}

var environment = "development"

func setupSuite() {
	log.Println("setup for suite")
	// connect to test db:
	config.Init(environment)
	db.Init()
	// ensure we start w/ empty db:
	tearDownSuite()
	// init other parts of the data generator tool:
	Init(environment, tests.GetTestSourceId())
}

func tearDownSuite() {
	log.Println("teardown for suite")
	tests.ExecAtlasSQLScript("../setup_local_db/ddl_atlas.sql")
	// we need some basic atlas data in "source" table to be able to connect to results DB, and this script has it:
	tests.ExecAtlasSQLScript("../setup_local_db/test_data_atlas.sql")
	// remove the rows in cohort_definition, created by script above: TODO - this will not be needed anymore once we cleanup these test_data .sql scripts...
	tests.EmptyTable(db.GetAtlasDB(), "cohort_definition")
	// create results and cdm tables:
	tests.ExecSQLScript("../setup_local_db/ddl_results_and_cdm.sql", tests.GetTestSourceId())
}

func setUp(t *testing.T) {
	log.Println("setup for test")
	// reset:
	Init(environment, tests.GetTestSourceId())

	// ensure tearDown is called when test "t" is done:
	t.Cleanup(func() {
		tearDown()
	})
}

func tearDown() {
	log.Println("teardown for test")
	// cleanup all tables:
	tests.EmptyTable(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "observation")
	tests.EmptyTable(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "concept")
	tests.EmptyTable(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "person")
	tests.EmptyTable(tests.GetResultsDataSourceForSourceId(tests.GetTestSourceId()), "cohort")
	tests.EmptyTable(db.GetAtlasDB(), "cohort_definition")
}

func TestRunDataGeneration(t *testing.T) {
	setUp(t)

	RunDataGeneration("models_tests_data_config")
	// assert on number of records per table:
	countObservations := tests.GetCount(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "observation")
	if countObservations != 35 {
		t.Errorf("Expected 35 observations, found %d", countObservations)
	}
	countConcepts := tests.GetCount(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "concept")
	if countConcepts != 7 {
		t.Errorf("Expected 7 concepts, found %d", countConcepts)
	}
	countPersons := tests.GetCount(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "person")
	if countPersons != 18 {
		t.Errorf("Expected 18 persons, found %d", countPersons)
	}
	// the name cohort is confusing...but it is one row per person x cohort_definition:
	totalCohortSize := tests.GetCount(tests.GetResultsDataSourceForSourceId(tests.GetTestSourceId()), "cohort")
	if totalCohortSize != 32 {
		t.Errorf("Expected total cohort size of 32, found %d", totalCohortSize)
	}
	countCohorts := tests.GetCount(db.GetAtlasDB(), "cohort_definition")
	if countCohorts != 5 {
		t.Errorf("Expected 5 cohort_definition records, found %d", countCohorts)
	}
}

func TestRunDataGeneration2(t *testing.T) {
	setUp(t)

	RunDataGeneration("example_test_data_config")

	countCohorts := tests.GetCount(db.GetAtlasDB(), "cohort_definition")
	if countCohorts != 3 {
		t.Errorf("Expected 3 cohort_definition records, found %d", countCohorts)
	}
	countPersons := tests.GetCount(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "person")
	if countPersons != 36 {
		t.Errorf("Expected 36 persons, found %d", countPersons)
	}
	countObservations := tests.GetCount(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "observation")
	if countObservations != 60 {
		t.Errorf("Expected 60 observations, found %d", countObservations)
	}
	countConcepts := tests.GetCount(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "concept")
	if countConcepts != 11 {
		t.Errorf("Expected 11 concepts, found %d", countConcepts)
	}
}
