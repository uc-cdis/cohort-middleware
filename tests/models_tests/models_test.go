package models_tests

import (
	"log"
	"os"
	"testing"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/tests"
	"github.com/uc-cdis/cohort-middleware/utils"
	"github.com/uc-cdis/cohort-middleware/version"
)

var testSourceId = tests.GetTestSourceId()
var allCohortDefinitions []*models.CohortDefinitionStats
var smallestCohort *models.CohortDefinitionStats
var secondLargestCohort *models.CohortDefinitionStats
var thirdsecondLargestCohort *models.CohortDefinitionStats
var allConceptIds []int64
var genderConceptId = tests.GetTestGenderConceptId()
var hareConceptId = tests.GetTestHareConceptId()
var asnHareConceptId = tests.GetTestAsnHareConceptId()
var histogramConceptId = tests.GetTestHistogramConceptId()

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
	secondLargestCohort = allCohortDefinitions[1]
	thirdsecondLargestCohort = allCohortDefinitions[2]
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
	filterIds := []int64{genderConceptId}
	conceptsStats, _ := conceptModel.RetrieveStatsBySourceIdAndCohortIdAndConceptIds(testSourceId,
		secondLargestCohort.Id,
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
	conceptTypes := []string{conceptsInfo[0].ConceptType}
	conceptsInfo, _ = conceptModel.RetrieveInfoBySourceIdAndConceptTypes(testSourceId,
		conceptTypes)
	if !(1 <= len(conceptsInfo) && len(conceptsInfo) < len(allConceptIds)) {
		t.Errorf("Found %d", len(conceptsInfo))
	}
}

func TestRetrieveInfoBySourceIdAndConceptIdNotFound(t *testing.T) {
	setUp(t)
	// get all concepts:
	conceptInfo, _ := conceptModel.RetrieveInfoBySourceIdAndConceptId(testSourceId,
		-1)
	if conceptInfo != nil {
		t.Errorf("Did not expected to find data")
	}
}

func TestRetrieveInfoBySourceIdAndConceptId(t *testing.T) {
	setUp(t)
	// get all concepts:
	conceptInfo, _ := conceptModel.RetrieveInfoBySourceIdAndConceptId(testSourceId,
		genderConceptId)
	if conceptInfo == nil {
		t.Errorf("Expected to find data")
	}
}

func TestRetrieveInfoBySourceIdAndConceptTypesWrongType(t *testing.T) {
	setUp(t)
	// simple test: invalid/non-existing type should return an empty list:
	conceptTypes := []string{"invalid type"}
	conceptsInfo, _ := conceptModel.RetrieveInfoBySourceIdAndConceptTypes(testSourceId,
		conceptTypes)
	if len(conceptsInfo) != 0 {
		t.Errorf("Found %d", len(conceptsInfo))
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairsNoResults(t *testing.T) {
	setUp(t)
	// empty:
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	stats, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId,
		smallestCohort.Id,
		allConceptIds, filterCohortPairs, allConceptIds[0])
	// none of the subjects has a value in all the concepts, so we expect len==0 here:
	if len(stats) != 0 {
		t.Errorf("Expected no results, found %d", len(stats))
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairsWithResults(t *testing.T) {
	setUp(t)
	filterIds := []int64{hareConceptId}
	// setting the same cohort id here (artificial...but just to check if that returns the same value as when this filter is not there):
	filterCohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortId1:    secondLargestCohort.Id,
			CohortId2:    secondLargestCohort.Id,
			ProvidedName: "test"},
	}
	breakdownConceptId := hareConceptId // not normally the case...but we'll use the same here just for the test...
	stats, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId,
		secondLargestCohort.Id, filterIds, filterCohortPairs, breakdownConceptId)
	// we expect values since secondLargestCohort has multiple subjects with hare info:
	if len(stats) < 4 {
		t.Errorf("Expected at least 4 results, found %d", len(stats))
	}
	prevName := ""
	for _, stat := range stats {
		// some very basic checks, making sure fields are not empty, repeated in next row, etc:
		if len(stat.ConceptValue) == len(stat.ValueName) ||
			len(stat.ConceptValue) == 0 ||
			len(stat.ValueName) == 0 ||
			stat.ValueAsConceptId == 0 ||
			stat.ValueName == prevName {
			t.Errorf("Invalid results")
		}
		prevName = stat.ValueName
	}
	// test without the filterCohortPairs, should return the same result:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{}
	stats2, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId,
		secondLargestCohort.Id, filterIds, filterCohortPairs, breakdownConceptId)
	// very rough check (ideally we would check the individual stats as well...TODO?):
	if len(stats) != len(stats2) {
		t.Errorf("Expected same result")
	}
	// test filtering with smallest cohort, lenght should be 1, since that's the size of the smallest cohort:
	// setting the same cohort id here (artificial...normally it should be two different ids):
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortId1:    smallestCohort.Id,
			CohortId2:    smallestCohort.Id,
			ProvidedName: "test"},
	}
	stats3, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId,
		secondLargestCohort.Id, filterIds, filterCohortPairs, breakdownConceptId)
	if len(stats3) != 1 {
		t.Errorf("Expected only one item in resultset")
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdWithResults(t *testing.T) {
	setUp(t)
	breakdownConceptId := hareConceptId
	stats, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(testSourceId,
		secondLargestCohort.Id,
		breakdownConceptId)
	// we expect 5-1 rows since the largest test cohort has all HARE values represented in its population, but has NULL in the "OTH" entry:
	if len(stats) != 4 {
		t.Errorf("Expected 4 results, found %d", len(stats))
	}
}

// Tests what happens when persons have more than 1 HARE. This is a "data error" and should not
// happen in practice. The ideal solution would be for cohort-middleware to throw an error
// when it detects such a situation in the RetrieveBreakdownStats methods. This test shows that
// the code does not "hide" the error but instead returns the extra hare as an
// extra count, making the cohort numbers inconsistent and hopefully making the "data error" easy
// to spot.
// TODO - adjust the code to detect the issue and return an error, ideally with minimized or no repetition
// of the heavy queries in the RetrieveBreakdownStats methods...
func TestRetrieveBreakdownStatsBySourceIdAndCohortIdWithResultsWithOnePersonTwoHare(t *testing.T) {
	setUp(t)
	breakdownConceptId := hareConceptId
	statsthirdsecondLargestCohort, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(testSourceId,
		thirdsecondLargestCohort.Id,
		breakdownConceptId)

	totalPersonInthirdsecondLargestCohortWithValue := 0

	for _, statSecondLargest := range statsthirdsecondLargestCohort {
		totalPersonInthirdsecondLargestCohortWithValue += statSecondLargest.NpersonsInCohortWithValue
	}

	if totalPersonInthirdsecondLargestCohortWithValue != thirdsecondLargestCohort.CohortSize+1 {
		t.Errorf("Expected total peope in return data to be 1 larger than cohort size, but total people was %d and cohort size is %d", totalPersonInthirdsecondLargestCohortWithValue, thirdsecondLargestCohort.CohortSize)
	}

	statssecondLargestCohort, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(testSourceId,
		secondLargestCohort.Id,
		breakdownConceptId)

	totalPersonInsecondLargestCohortWithValue := 0

	for _, statLargeCohort := range statssecondLargestCohort {
		totalPersonInsecondLargestCohortWithValue += statLargeCohort.NpersonsInCohortWithValue
	}

	if totalPersonInsecondLargestCohortWithValue != secondLargestCohort.CohortSize+1 {
		t.Errorf("Expected total peope in return data to be 1 larger than cohort size, but total people was %d and cohort size is %d", totalPersonInsecondLargestCohortWithValue, secondLargestCohort.CohortSize)
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

func TestRetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(t *testing.T) {
	setUp(t)
	largestCohort := allCohortDefinitions[0]
	filterConceptIds := []int64{}
	filterCohortIds := []utils.CustomDichotomousVariableDef{}
	data, _ := cohortDataModel.RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId, largestCohort.Id, histogramConceptId, filterConceptIds, filterCohortIds)
	if len(data) == 0 {
		t.Errorf("expected 1 or more histogram data but got 0")
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
func getNrPersonsWithHareConceptValue(sourceId int, cohortId int, hareConceptValue int64) int64 {
	conceptIds := []int64{hareConceptId}
	personLevelData, _ := cohortDataModel.RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId, cohortId, conceptIds)
	var count int64 = 0
	for _, personLevelDatum := range personLevelData {
		if personLevelDatum.ConceptValueAsConceptId == hareConceptValue {
			count++
		}
	}
	return count
}

func TestRetrieveCohortOverlapStats(t *testing.T) {
	// Tests if we get the expected overlap
	setUp(t)
	caseCohortId := secondLargestCohort.Id
	controlCohortId := secondLargestCohort.Id // to ensure we get some overlap, just repeat the same here...
	filterConceptId := hareConceptId
	filterConceptValue := asnHareConceptId
	otherFilterConceptIds := []int64{}
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds, filterCohortPairs)
	// get the number of persons in this cohort that have this filterConceptValue:
	nr_expected := getNrPersonsWithHareConceptValue(testSourceId, caseCohortId, filterConceptValue)
	if nr_expected == 0 {
		t.Errorf("Expected nr persons with HARE value should be > 0")
	}
	if stats.CaseControlOverlapAfterFilter != nr_expected {
		t.Errorf("Expected overlap of %d, but found %d", nr_expected, stats.CaseControlOverlapAfterFilter)
	}
}

func TestRetrieveCohortOverlapStatsScenario2(t *testing.T) {
	// Tests if we get the expected overlap
	setUp(t)
	caseCohortId := secondLargestCohort.Id
	controlCohortId := secondLargestCohort.Id // to ensure we get some overlap, just repeat the same here...
	filterConceptId := hareConceptId
	filterConceptValue := asnHareConceptId
	otherFilterConceptIds := []int64{hareConceptId} // repeat hare concept id here...Artificial, but will ensure overlap
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds, filterCohortPairs)
	// get the number of persons in this cohort that have this filterConceptValue:
	nr_expected := getNrPersonsWithHareConceptValue(testSourceId, caseCohortId, filterConceptValue)
	if nr_expected == 0 {
		t.Errorf("Expected nr persons with HARE value should be > 0")
	}
	if stats.CaseControlOverlapAfterFilter != nr_expected {
		t.Errorf("Expected overlap of %d, but found %d", nr_expected, stats.CaseControlOverlapAfterFilter)
	}
}

func TestRetrieveCohortOverlapStatsZeroOverlap(t *testing.T) {
	// Tests if a scenario where NO overlap is expected indeed results in 0
	setUp(t)
	caseCohortId := secondLargestCohort.Id
	controlCohortId := smallestCohort.Id
	filterConceptId := hareConceptId
	var filterConceptValue int64 = -1 // should result in 0 overlap
	otherFilterConceptIds := []int64{}
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds, filterCohortPairs)
	if stats.CaseControlOverlapAfterFilter != 0 {
		t.Errorf("Expected overlap of %d, but found %d", 0, stats.CaseControlOverlapAfterFilter)
	}
}

func TestRetrieveCohortOverlapStatsZeroOverlapScenario2(t *testing.T) {
	// Tests if a scenario where NO overlap is expected indeed results in 0
	setUp(t)
	caseCohortId := secondLargestCohort.Id
	controlCohortId := secondLargestCohort.Id // to ensure THIS part does not break it, just repeat the same here...
	filterConceptId := hareConceptId
	filterConceptValue := asnHareConceptId
	// set this list to some dummy non-existing ids:
	otherFilterConceptIds := []int64{-1, -2}
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds, filterCohortPairs)
	if stats.CaseControlOverlapAfterFilter != 0 {
		t.Errorf("Expected overlap of %d, but found %d", 0, stats.CaseControlOverlapAfterFilter)
	}
}

func TestRetrieveCohortOverlapStatsWithCohortPairs(t *testing.T) {
	// Tests if we get the expected overlap
	setUp(t)
	caseCohortId := secondLargestCohort.Id
	controlCohortId := secondLargestCohort.Id // to ensure we get some overlap, just repeat the same here...
	filterConceptId := hareConceptId
	filterConceptValue := asnHareConceptId          // the cohorts we use below both have persons with "ASN" HARE value
	otherFilterConceptIds := []int64{hareConceptId} // repeat hare concept id here...Artificial, but will ensure overlap
	filterCohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortId1:    smallestCohort.Id,
			CohortId2:    thirdsecondLargestCohort.Id,
			ProvidedName: "test"}, // pair1
		{
			CohortId1:    thirdsecondLargestCohort.Id,
			CohortId2:    smallestCohort.Id,
			ProvidedName: "test"}, // pair2 (same as above, but switched...artificial, but will ensure some data):
	}
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds, filterCohortPairs)
	// get the number of persons in the smaller cohorts that have this filterConceptValue (this can be the expected nr because
	// the secondLargestCohort in this case contains all other cohorts):
	nr_expected := getNrPersonsWithHareConceptValue(testSourceId, thirdsecondLargestCohort.Id, filterConceptValue)
	nr_expected = nr_expected + getNrPersonsWithHareConceptValue(testSourceId, smallestCohort.Id, filterConceptValue)
	if nr_expected == 0 {
		t.Errorf("Expected nr persons with HARE value should be > 0")
	}
	if stats.CaseControlOverlapAfterFilter != nr_expected {
		t.Errorf("Expected overlap of %d, but found %d", nr_expected, stats.CaseControlOverlapAfterFilter)
	}
	filterCohortPairs = []utils.CustomDichotomousVariableDef{}
	// without the restrictive filter on cohort pairs, the result should be bigger, as the largest cohort has more persons with
	// the asnHareConceptId than the ones used in the pairs above:
	stats2, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, otherFilterConceptIds, filterCohortPairs)
	if stats.CaseControlOverlapAfterFilter >= stats2.CaseControlOverlapAfterFilter {
		t.Errorf("Expected overlap in first query to be smaller than in second one")
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

func TestRetrieveDataByOriginalCohortAndNewCohort(t *testing.T) {
	setUp(t)
	originalCohortSize := thirdsecondLargestCohort.CohortSize
	originalCohortId := thirdsecondLargestCohort.Id
	cohortDefinitionId := secondLargestCohort.Id

	personIdAndCohortList, _ := cohortDataModel.RetrieveDataByOriginalCohortAndNewCohort(testSourceId, originalCohortId, cohortDefinitionId)
	if len(personIdAndCohortList) != originalCohortSize {
		t.Errorf("length of return data does not match number of people in cohort")
	}

	for _, personIdAndCohort := range personIdAndCohortList {
		if personIdAndCohort.CohortId != int64(cohortDefinitionId) {
			t.Errorf("cohort_id we retireved is not correct")
		}
		if personIdAndCohort.PersonId == int64(0) {
			t.Error("person id should be valid and not 0")
		}
	}
}
