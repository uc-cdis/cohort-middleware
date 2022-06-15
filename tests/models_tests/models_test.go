package models_tests

import (
	"log"
	"os"
	"testing"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/tests"
	"github.com/uc-cdis/cohort-middleware/version"
)

var testSourceId = tests.GetTestSourceId()
var allCohortDefinitions []*models.CohortDefinitionStats
var smallestCohort *models.CohortDefinitionStats
var largestCohort *models.CohortDefinitionStats
var allConceptIds []int64
var genderConceptId = tests.GetTestGenderConceptId()
var hareConceptId = tests.GetTestHareConceptId()

func TestMain(m *testing.M) {
	setupSuite()
	retCode := m.Run()
	tearDownSuite()
	os.Exit(retCode)
}

func setupSuite() {
	log.Println("setup for suite")
	// connect to test db:
	config.Init("development")
	db.Init()
	// ensure we start w/ empty db:
	tearDownSuite()
	// load test seed data:
	tests.ExecSQLScript("../setup_local_db/test_data_results_and_cdm.sql", testSourceId)

	// initialize some handy variables to use in tests below:
	allCohortDefinitions, _ = cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId)
	largestCohort = allCohortDefinitions[0]
	smallestCohort = allCohortDefinitions[len(allCohortDefinitions)-1]
	concepts, _ := conceptModel.RetriveAllBySourceId(testSourceId)
	allConceptIds = tests.MapIntAttr(concepts, "ConceptId")
}

func tearDownSuite() {
	log.Println("teardown for suite")
	tests.ExecAtlasSQLScript("../setup_local_db/ddl_atlas.sql")
	// we need some basic atlas data in "source" table to be able to connect to results DB, and this script has it:
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
var cohortDataModel = new(models.CohortData)

var versionModel = new(models.Version)
var sourceModel = new(models.Source)

func TestGetConceptId(t *testing.T) {
	setUp(t)
	conceptId := models.GetConceptId("ID_12345")
	if conceptId != 12345 {
		t.Error()
	}
	// the GetConceptId below should result in panic/error:
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	models.GetConceptId("AD_12345")

}

func TestGetPrefixedConceptId(t *testing.T) {
	setUp(t)
	conceptId := models.GetPrefixedConceptId(12345)
	if conceptId != "ID_12345" {
		t.Error()
	}
}

func TestRetriveAllBySourceId(t *testing.T) {
	setUp(t)
	concepts, _ := conceptModel.RetriveAllBySourceId(testSourceId)
	if len(concepts) != 10 {
		t.Errorf("Found %d", len(concepts))
	}
}

func TestRetrieveStatsBySourceIdAndCohortIdAndConceptIds(t *testing.T) {
	setUp(t)
	conceptsStats, _ := conceptModel.RetrieveStatsBySourceIdAndCohortIdAndConceptIds(testSourceId,
		smallestCohort.Id,
		allConceptIds)
	// simple test: we expect stats for each valid conceptId, therefore the lists are
	//  expected to have the same lenght here:
	if len(conceptsStats) != len(allConceptIds) {
		t.Errorf("Found %d", len(conceptsStats))
	}
}

func TestRetrieveStatsBySourceIdAndCohortIdAndConceptIdsCheckRatio(t *testing.T) {
	setUp(t)
	filterIds := make([]int64, 1)
	filterIds[0] = genderConceptId
	conceptsStats, _ := conceptModel.RetrieveStatsBySourceIdAndCohortIdAndConceptIds(testSourceId,
		largestCohort.Id,
		filterIds)
	// simple test: in the test data we keep the gender concept *missing* at a ratio of 1/3 for the largest cohort. Here
	// we check if the missing ratio calculation is working correctly:
	if len(conceptsStats) != 1 {
		t.Errorf("Found %d", len(conceptsStats))
	}
	if conceptsStats[0].NmissingRatio != 1.0/3.0 {
		t.Errorf("Found wrong ratio %f", conceptsStats[0].NmissingRatio)
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

func TestRetrieveInfoBySourceIdAndConceptTypes(t *testing.T) {
	setUp(t)
	// get all concepts:
	conceptsInfo, _ := conceptModel.RetrieveInfoBySourceIdAndConceptIds(testSourceId,
		allConceptIds)
	// simple test: we know that not all concepts have the same type in our test db, so
	// if we query on the type of a single concept, the result should
	// be a list where 1 =< size < len(allConceptIds):
	conceptTypes := make([]string, 1)
	conceptTypes[0] = conceptsInfo[0].ConceptType
	conceptsInfo, _ = conceptModel.RetrieveInfoBySourceIdAndConceptTypes(testSourceId,
		conceptTypes)
	if !(1 <= len(conceptsInfo) && len(conceptsInfo) < len(allConceptIds)) {
		t.Errorf("Found %d", len(conceptsInfo))
	}
}

func TestRetrieveInfoBySourceIdAndConceptTypesWrongType(t *testing.T) {
	setUp(t)
	// simple test: invalid/non-existing type should return an empty list:
	conceptTypes := make([]string, 1)
	conceptTypes[0] = "invalid type"
	conceptsInfo, _ := conceptModel.RetrieveInfoBySourceIdAndConceptTypes(testSourceId,
		conceptTypes)
	if len(conceptsInfo) != 0 {
		t.Errorf("Found %d", len(conceptsInfo))
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsNoResults(t *testing.T) {
	setUp(t)
	stats, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(testSourceId,
		smallestCohort.Id,
		allConceptIds, allConceptIds[0])
	// none of the subjects has a value in all the concepts, so we expect len==0 here:
	if len(stats) != 0 {
		t.Errorf("Expected no results, found %d", len(stats))
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsWithResults(t *testing.T) {
	setUp(t)
	filterIds := make([]int64, 1)
	filterIds[0] = genderConceptId
	breakdownConceptId := genderConceptId // not normally the case...but we'll use the same here just for the test...
	stats, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(testSourceId,
		largestCohort.Id,
		filterIds, breakdownConceptId)
	// we expect values since all of the test cohorts have at least one subject with gender info:
	if len(stats) < 2 {
		t.Errorf("Expected at least two results, found %d", len(stats))
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdWithResults(t *testing.T) {
	setUp(t)
	breakdownConceptId := hareConceptId
	stats, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(testSourceId,
		largestCohort.Id,
		breakdownConceptId)
	// we expect 5-1 rows since the largest test cohort has all HARE values represented in its population, but has NULL in the "OTH" entry:
	if len(stats) != 4 {
		t.Errorf("Expected 4 results, found %d", len(stats))
	}
}

func TestGetAllCohortDefinitionsAndStatsOrderBySizeDesc(t *testing.T) {
	setUp(t)
	cohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId)
	if len(cohortDefinitions) != len(allCohortDefinitions) {
		t.Errorf("Found %d", len(cohortDefinitions))
	}
	// check if stats fields are filled and if order is as expected:
	previousSize := 1000000
	for _, cohortDefinition := range cohortDefinitions {
		if cohortDefinition.CohortSize <= 0 {
			t.Errorf("Expected positive value, found %d", cohortDefinition.CohortSize)
		}
		if cohortDefinition.CohortSize > previousSize {
			t.Errorf("Data not ordered by size descending!")
		}
		previousSize = cohortDefinition.CohortSize
	}
}

func TestGetCohortName(t *testing.T) {
	setUp(t)
	allCohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitions()
	firstCohortId := allCohortDefinitions[0].Id
	cohortName, _ := cohortDefinitionModel.GetCohortName(firstCohortId)
	if cohortName != allCohortDefinitions[0].Name {
		t.Errorf("Expected %s", allCohortDefinitions[0].Name)
	}
}

func TestGetCohortDefinitionByName(t *testing.T) {
	setUp(t)
	cohortDefinition, _ := cohortDefinitionModel.GetCohortDefinitionByName(smallestCohort.Name)
	if cohortDefinition == nil || cohortDefinition.Name != smallestCohort.Name {
		t.Errorf("Expected %s", smallestCohort.Name)
	}
}

func TestRetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(t *testing.T) {
	setUp(t)
	cohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId)
	var sumNumeric float32 = 0
	textConcat := ""
	for _, cohortDefinition := range cohortDefinitions {

		cohortData, _ := cohortDataModel.RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(
			testSourceId, cohortDefinition.Id, allConceptIds)

		// 1- cohortData items > 0, assuming each cohort has a person wit at least one observation
		if len(cohortData) <= 0 {
			t.Errorf("Expected some cohort data")
		}
		var previousPersonId int64 = -1
		for _, cohortDatum := range cohortData {
			// check for order: person_id is not smaller than previous person_id
			if cohortDatum.PersonId < previousPersonId {
				t.Errorf("Data not ordered by person_id!")
			}
			previousPersonId = cohortDatum.PersonId
			sumNumeric += cohortDatum.ConceptValueAsNumber
			textConcat += cohortDatum.ConceptValueAsString
		}
	}
	// check for data: sum of all numeric values > 0
	if sumNumeric == 0 {
		t.Errorf("Expected some numeric cohort data")
	}
	// check for data: concat of all string values != ""
	if textConcat == "" {
		t.Errorf("Expected some string cohort data")
	}
}

func TestErrorForRetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(t *testing.T) {
	// Tests if the method returns an error when query fails.

	// break something in the omop schema to cause a query failure in the next method:
	tests.BreakSomething(models.Omop, "observation", "person_id")
	// set last action to restore back:
	// run test:
	cohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId)
	_, error := cohortDataModel.RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(
		testSourceId, cohortDefinitions[0].Id, allConceptIds)
	if error == nil {
		t.Errorf("Expected error")
	}
	// revert the broken part:
	tests.FixSomething(models.Omop, "observation", "person_id")
}

// for given source and cohort, counts how many persons have the given HARE value
func getNrPersonsWithHareConceptValue(sourceId int, cohortId int, hareConceptValue string) int64 {
	conceptIds := make([]int64, 1)
	conceptIds[0] = hareConceptId
	personLevelData, _ := cohortDataModel.RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId, cohortId, conceptIds)
	var count int64 = 0
	for _, personLevelDatum := range personLevelData {
		if personLevelDatum.ConceptValueAsString == hareConceptValue {
			count++
		}
	}
	return count
}

func TestRetrieveCohortOverlapStats(t *testing.T) {
	// Tests if we get the expected overlap
	setUp(t)
	caseCohortId := largestCohort.Id
	controlCohortId := largestCohort.Id // to ensure we get some overlap, just repeat the same here...
	filterConceptId := hareConceptId
	filterConceptValue := "ASN"
	otherFilterConceptIds := make([]int64, 0)
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds)
	// get the number of persons in this cohort that have this filterConceptValue:
	nr_expected := getNrPersonsWithHareConceptValue(testSourceId, caseCohortId, filterConceptValue)
	if stats.CaseControlOverlapAfterFilter != nr_expected {
		t.Errorf("Expected overlap of %d, but found %d", nr_expected, stats.CaseControlOverlapAfterFilter)
	}
}

func TestRetrieveCohortOverlapStatsZeroOverlap(t *testing.T) {
	// Tests if a scenario where NO overlap is expected indeed results in 0
	setUp(t)
	caseCohortId := largestCohort.Id
	controlCohortId := smallestCohort.Id
	filterConceptId := hareConceptId
	filterConceptValue := "NON-EXISTING-VALUE" // should result in 0 overlap
	otherFilterConceptIds := make([]int64, 0)
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds)
	if stats.CaseControlOverlapAfterFilter != 0 {
		t.Errorf("Expected overlap of %d, but found %d", 0, stats.CaseControlOverlapAfterFilter)
	}
}

func TestRetrieveCohortOverlapStatsZeroOverlapScenario2(t *testing.T) {
	// Tests if a scenario where NO overlap is expected indeed results in 0
	setUp(t)
	caseCohortId := largestCohort.Id
	controlCohortId := largestCohort.Id // to ensure THIS part does not break it, just repeat the same here...
	filterConceptId := hareConceptId
	filterConceptValue := "ASN"
	// set this list to some dummy non-existing ids:
	otherFilterConceptIds := make([]int64, 2)
	otherFilterConceptIds[0] = -1
	otherFilterConceptIds[1] = -2
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds)
	if stats.CaseControlOverlapAfterFilter != 0 {
		t.Errorf("Expected overlap of %d, but found %d", 0, stats.CaseControlOverlapAfterFilter)
	}
}

func TestGetVersion(t *testing.T) {
	// mock values (in reality these are set at build time - see Dockerfile "go build" "-ldflags" argument):
	version.GitCommit = "abc"
	version.GitVersion = "def"
	v := versionModel.GetVersion()
	if v.GitCommit != version.GitCommit || v.GitVersion != version.GitVersion {
		t.Errorf("Wrong value")
	}
}

func TestGetSourceByName(t *testing.T) {
	allSources, _ := sourceModel.GetAllSources()
	foundSource, _ := sourceModel.GetSourceByName(allSources[0].SourceName)
	if allSources[0].SourceName != foundSource.SourceName {
		t.Errorf("Expected data not found")
	}
}

func TestGetSourceById(t *testing.T) {
	allSources, _ := sourceModel.GetAllSources()
	foundSource, _ := sourceModel.GetSourceById(allSources[0].SourceId)
	if allSources[0].SourceId != foundSource.SourceId {
		t.Errorf("Expected data not found")
	}
}

func TestGetCohortDefinitionById(t *testing.T) {
	allCohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitions()
	foundCohortDefinition, _ := cohortDefinitionModel.GetCohortDefinitionById(allCohortDefinitions[0].Id)
	if allCohortDefinitions[0].Id != foundCohortDefinition.Id {
		t.Errorf("Expected data not found")
	}
}
