package models_tests

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

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
var largestCohort *models.CohortDefinitionStats
var secondLargestCohort *models.CohortDefinitionStats
var extendedCopyOfSecondLargestCohort *models.CohortDefinitionStats
var thirdLargestCohort *models.CohortDefinitionStats
var allConceptIds []int64
var dummyContinuousConceptId = tests.GetTestDummyContinuousConceptId()
var hareConceptId = tests.GetTestHareConceptId()
var histogramConceptId = tests.GetTestHistogramConceptId()
var defaultTeamProject = "defaultteamproject"

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
	// load test seed data, including test cohorts referenced below:
	tests.ExecSQLScript("../setup_local_db/test_data_results_and_cdm.sql", testSourceId)

	// initialize some handy variables to use in tests below:
	// (see also tests/setup_local_db/test_data_results_and_cdm.sql for these test cohort details)
	allCohortDefinitions, _ = cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, defaultTeamProject)
	largestCohort = allCohortDefinitions[0]
	secondLargestCohort = allCohortDefinitions[2]
	extendedCopyOfSecondLargestCohort = allCohortDefinitions[1]
	thirdLargestCohort = allCohortDefinitions[3]
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
var dataDictionaryModel = new(models.DataDictionary)

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

func TestGetConceptValueNotNullCheckBasedOnConceptTypeError(t *testing.T) {
	setUp(t)
	// the call below should result in panic/error:
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	models.GetConceptValueNotNullCheckBasedOnConceptType("observation", testSourceId, -1)
}

func TestGetConceptValueNotNullCheckBasedOnConceptTypeError2(t *testing.T) {
	setUp(t)
	// add dummy concept:
	conceptId := tests.AddInvalidTypeConcept(models.Omop)

	// the call below should result in a specific panic/error on the concept type not being supported:
	defer func() {
		r := recover()
		if r == nil || !strings.HasPrefix(r.(string), "error: concept type not supported") {
			t.Errorf("The code did not panic with expected error")
		}
		// cleanup:
		tests.RemoveConcept(models.Omop, conceptId)
	}()
	models.GetConceptValueNotNullCheckBasedOnConceptType("observation", testSourceId, conceptId)
}

func TestGetConceptValueNotNullCheckBasedOnConceptTypeSuccess(t *testing.T) {
	setUp(t)
	// check success scenarios:
	result := models.GetConceptValueNotNullCheckBasedOnConceptType("observation", testSourceId, hareConceptId)
	if result != "observation.value_as_concept_id is not null and observation.value_as_concept_id != 0" {
		t.Errorf("Unexpected result. Found %s", result)
	}
	result = models.GetConceptValueNotNullCheckBasedOnConceptType("observation", testSourceId, histogramConceptId)
	if result != "observation.value_as_number is not null" {
		t.Errorf("Unexpected result. Found %s", result)
	}
}

func TestRetriveAllBySourceId(t *testing.T) {
	setUp(t)
	concepts, _ := conceptModel.RetriveAllBySourceId(testSourceId)
	if len(concepts) != 10 {
		t.Errorf("Found %d", len(concepts))
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
	conceptInfo, error := conceptModel.RetrieveInfoBySourceIdAndConceptId(testSourceId,
		-1)
	if conceptInfo != nil {
		t.Errorf("Did not expect to find data")
	}
	if error == nil {
		t.Errorf("Expected error")
	}
}

func TestRetrieveInfoBySourceIdAndConceptId(t *testing.T) {
	setUp(t)
	// get all concepts:
	conceptInfo, _ := conceptModel.RetrieveInfoBySourceIdAndConceptId(testSourceId,
		hareConceptId)
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

// Tests various scenarios for QueryFilterByCohortPairsHelper.
// These tests currently use pre-defined cohorts (see .sql loaded in  setupSuite()).
// A possible improvement / TODO could be to write more test utility functions
// to add specific test data on the fly. This could make some of the test code (like the tests here)
// more readable by having the test data and the test assertions close together. For now,
// consider reading these tests together with the .sql file that is loaded in setupSuite()
// to understand how the test cohorts relate to each other.
func TestQueryFilterByCohortPairsHelper(t *testing.T) {
	setUp(t)

	type SubjectId struct {
		SubjectId int
	}
	// smallestCohort and largestCohort do not overlap...
	filterCohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	resultsDataSource := tests.GetResultsDataSource()
	var subjectIds []*SubjectId
	population := largestCohort
	query := models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// ...so we expect overlap the size of the largestCohort:
	if len(subjectIds) != largestCohort.CohortSize {
		t.Errorf("Expected %d overlap, found %d", largestCohort.CohortSize, len(subjectIds))
	}

	// now add a pair that overlaps with largestCohort:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
		{
			CohortDefinitionId1: extendedCopyOfSecondLargestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	subjectIds = []*SubjectId{}
	population = largestCohort
	resultsDataSource = tests.GetResultsDataSource()
	query = models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// in this case we expect overlap the size of the largestCohort-6 (where 6 is the size of the overlap between extendedCopyOfSecondLargestCohort and largestCohort):
	if len(subjectIds) != (largestCohort.CohortSize - 6) {
		t.Errorf("Expected %d overlap, found %d", largestCohort.CohortSize-5, len(subjectIds))
	}

	// order doesn't matter:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: extendedCopyOfSecondLargestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	subjectIds = []*SubjectId{}
	population = largestCohort
	resultsDataSource = tests.GetResultsDataSource()
	query = models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// in this case we expect same as previous test above:
	if len(subjectIds) != (largestCohort.CohortSize - 6) {
		t.Errorf("Expected %d overlap, found %d", largestCohort.CohortSize-5, len(subjectIds))
	}

	// now test with two other cohorts that overlap:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: secondLargestCohort.Id,
			CohortDefinitionId2: extendedCopyOfSecondLargestCohort.Id,
			ProvidedName:        "test"},
	}
	subjectIds = []*SubjectId{}
	population = extendedCopyOfSecondLargestCohort
	resultsDataSource = tests.GetResultsDataSource()
	query = models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// in this case we expect overlap the size of the extendedCopyOfSecondLargestCohort.CohortSize - secondLargestCohort.CohortSize:
	if len(subjectIds) != (extendedCopyOfSecondLargestCohort.CohortSize - secondLargestCohort.CohortSize) {
		t.Errorf("Expected %d overlap, found %d", extendedCopyOfSecondLargestCohort.CohortSize-secondLargestCohort.CohortSize, len(subjectIds))
	}

	// now add in the largestCohort as a pair of extendedCopyOfSecondLargestCohort to the mix above:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: secondLargestCohort.Id,
			CohortDefinitionId2: extendedCopyOfSecondLargestCohort.Id,
			ProvidedName:        "test"},
		{
			CohortDefinitionId1: largestCohort.Id,
			CohortDefinitionId2: extendedCopyOfSecondLargestCohort.Id,
			ProvidedName:        "test"},
	}
	subjectIds = []*SubjectId{}
	population = extendedCopyOfSecondLargestCohort
	resultsDataSource = tests.GetResultsDataSource()
	query = models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// in this case we expect overlap the size to be 0, since all items remaining from first pair happen to overlap with largestCohort and are therefore excluded (pair overlap is excluded):
	if len(subjectIds) != 0 {
		t.Errorf("Expected 0 overlap, found %d", len(subjectIds))
	}

	// now if the population is largestCohort, for the same pairs above, we expect the overlap to be 0 as well, as the first pair restricts the set for every other following pair (i.e. attrition at work):
	subjectIds = []*SubjectId{}
	population = largestCohort
	resultsDataSource = tests.GetResultsDataSource()
	query = models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// in this case we expect overlap the size to be 0 as explained in comment above:
	if len(subjectIds) != 0 {
		t.Errorf("Expected 0 overlap, found %d", len(subjectIds))
	}

	// should return all in cohort:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{}
	subjectIds = []*SubjectId{}
	population = largestCohort
	resultsDataSource = tests.GetResultsDataSource()
	query = models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// in this case we expect overlap the size to be the size of the cohort, since there are no filtering pairs:
	if len(subjectIds) != largestCohort.CohortSize {
		t.Errorf("Expected 0 overlap, found %d", len(subjectIds))
	}

	// should return 0:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: largestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	subjectIds = []*SubjectId{}
	population = largestCohort
	resultsDataSource = tests.GetResultsDataSource()
	query = models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// in this case we expect overlap the size to be 0, since the pair is composed of the same cohort in CohortDefinitionId1 and CohortDefinitionId2 and their overlap is excluded:
	if len(subjectIds) != 0 {
		t.Errorf("Expected 0 overlap, found %d", len(subjectIds))
	}

	// should return 0:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: thirdLargestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	subjectIds = []*SubjectId{}
	population = smallestCohort
	resultsDataSource = tests.GetResultsDataSource()
	query = models.QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, population.Id, "unionAndIntersect").
		Select("subject_id")
	_ = query.Scan(&subjectIds)
	// in this case we expect overlap the size to be 0, since the cohorts in the pair do not overlap with the population:
	if len(subjectIds) != 0 {
		t.Errorf("Expected 0 overlap, found %d", len(subjectIds))
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndTwoCohortPairsWithResults(t *testing.T) {
	setUp(t)
	filterIds := []int64{hareConceptId}
	populationCohort := largestCohort
	// setting the largest and smallest cohorts here as a pair:
	filterCohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	breakdownConceptId := hareConceptId // not normally the case...but we'll use the same here just for the test...
	stats, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId,
		populationCohort.Id, filterIds, filterCohortPairs, breakdownConceptId)
	// we expect results, and we expect the total of persons to be 6, since only 6 of the persons
	// in largestCohort have a HARE value (and smallestCohort does not overlap with largest):
	countPersons := 0
	for _, stat := range stats {
		countPersons += stat.NpersonsInCohortWithValue
	}
	if countPersons != 6 {
		t.Errorf("Expected 6 persons, found %d", countPersons)
	}
	// now if we add another cohort pair with secondLargestCohort and extendedCopyOfSecondLargestCohort,
	// then we should expect a reduction in the number of persons found. The reduction in this case
	// will take place because of a smaller intersection of the new cohorts with the population cohort,
	// and because of an overlaping person found in the two cohorts of the new pair.
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
		{
			CohortDefinitionId1: secondLargestCohort.Id,
			CohortDefinitionId2: extendedCopyOfSecondLargestCohort.Id,
			ProvidedName:        "test2"},
	}
	stats, _ = conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId,
		populationCohort.Id, filterIds, filterCohortPairs, breakdownConceptId)
	countPersons = 0
	for _, stat := range stats {
		countPersons += stat.NpersonsInCohortWithValue
	}
	if countPersons != 4 {
		t.Errorf("Expected 4 persons, found %d", countPersons)
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairsWithResults(t *testing.T) {
	setUp(t)
	filterIds := []int64{hareConceptId}
	// setting the same cohort id here (artificial...but just to check if that returns the same value as when this filter is not there):
	filterCohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: secondLargestCohort.Id,
			CohortDefinitionId2: extendedCopyOfSecondLargestCohort.Id,
			ProvidedName:        "test"},
	}
	breakdownConceptId := hareConceptId // not normally the case...but we'll use the same here just for the test...
	stats, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId,
		extendedCopyOfSecondLargestCohort.Id, filterIds, filterCohortPairs, breakdownConceptId)
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
		extendedCopyOfSecondLargestCohort.Id, filterIds, filterCohortPairs, breakdownConceptId)
	// very rough check (ideally we would check the individual stats as well...TODO?):
	if len(stats) > len(stats2) {
		t.Errorf("First query is more restrictive, so its stats should not be larger than stats2 of second query. Got %d and %d", len(stats), len(stats2))
	}

	// test filtering with secondLargestCohort, smallest and largestCohort.
	// Lenght of result set should be 2 persons (one HIS, one ASN), since there is a overlap of 1 between secondLargestCohort and smallest cohort,
	// and overlap of 2 between secondLargestCohort and largestCohort, BUT only 1 has a HARE value:
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	stats3, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId,
		secondLargestCohort.Id, filterIds, filterCohortPairs, breakdownConceptId)
	if len(stats3) != 2 {
		t.Errorf("Expected only two items in resultset, found %d", len(stats3))
	}
	countPersons := 0
	for _, stat := range stats3 {
		countPersons = countPersons + stat.NpersonsInCohortWithValue
	}
	if countPersons != 2 {
		t.Errorf("Expected only two persons in resultset, found %d", countPersons)
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
// of the heavy queries in the RetrieveBreakdownStats methods... Idea: run this check as a QC query for each cohort
// at startup and write an ERROR to the log (with cohort id and name information) if it detects such data issues.
func TestRetrieveBreakdownStatsBySourceIdAndCohortIdWithResultsWithOnePersonTwoHare(t *testing.T) {
	setUp(t)
	breakdownConceptId := hareConceptId
	statsthirdLargestCohort, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(testSourceId,
		thirdLargestCohort.Id,
		breakdownConceptId)

	totalPersonInthirdLargestCohortWithValue := 0

	for _, statSecondLargest := range statsthirdLargestCohort {
		totalPersonInthirdLargestCohortWithValue += statSecondLargest.NpersonsInCohortWithValue
	}

	if totalPersonInthirdLargestCohortWithValue != thirdLargestCohort.CohortSize+1 {
		t.Errorf("Expected total peope in return data to be 1 larger than cohort size, but total people was %d and cohort size is %d", totalPersonInthirdLargestCohortWithValue, thirdLargestCohort.CohortSize)
	}

	statssecondLargestCohort, _ := conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(testSourceId,
		secondLargestCohort.Id,
		breakdownConceptId)

	totalPersonInsecondLargestCohortWithValue := 0

	for _, statLargeCohort := range statssecondLargestCohort {
		totalPersonInsecondLargestCohortWithValue += statLargeCohort.NpersonsInCohortWithValue
	}

	expectedWithValueInSecondLargest := secondLargestCohort.CohortSize - 1 // because 2nd largest has one person that has a NULL HARE entry...
	if totalPersonInsecondLargestCohortWithValue != expectedWithValueInSecondLargest+1 {
		t.Errorf("Expected total peope in return data to be 1 larger than nr distinct persons with HARE, but total was %d and nr distinct persons with HARE +1 is %d", totalPersonInsecondLargestCohortWithValue, expectedWithValueInSecondLargest+1)
	}
}

func TestGetTeamProjectsThatMatchAllCohortDefinitionIdsOnlyDefaultMatch(t *testing.T) {
	setUp(t)
	cohortDefinitionId := 2 // 'Medium cohort' in test_data_atlas.sql
	filterCohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	uniqueCohortDefinitionIdsList := utils.GetUniqueCohortDefinitionIdsList([]int{cohortDefinitionId}, filterCohortPairs)
	if len(uniqueCohortDefinitionIdsList) != 3 {
		t.Errorf("Expected uniqueCohortDefinitionIdsList length to be 3")
	}
	teamProjects, _ := cohortDefinitionModel.GetTeamProjectsThatMatchAllCohortDefinitionIds(uniqueCohortDefinitionIdsList)
	if len(teamProjects) != 1 || teamProjects[0] != "defaultteamproject" {
		t.Errorf("Expected to find only defaultteamproject")
	}

	// Should also hold true if the uniqueCohortDefinitionIdsList is length 2 (which matches teamprojectX's cohort
	// list length but not in contents):
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: 2,
			CohortDefinitionId2: largestCohort.Id,
			ProvidedName:        "test"},
	}
	uniqueCohortDefinitionIdsList = utils.GetUniqueCohortDefinitionIdsList([]int{cohortDefinitionId}, filterCohortPairs)
	if len(uniqueCohortDefinitionIdsList) != 2 {
		t.Errorf("Expected uniqueCohortDefinitionIdsList length to be 2")
	}
	teamProjects, _ = cohortDefinitionModel.GetTeamProjectsThatMatchAllCohortDefinitionIds(uniqueCohortDefinitionIdsList)
	if len(teamProjects) != 1 || teamProjects[0] != "defaultteamproject" {
		t.Errorf("Expected to find only defaultteamproject")
	}
}

func TestGetTeamProjectsThatMatchAllCohortDefinitionIds(t *testing.T) {
	setUp(t)
	cohortDefinitionId := 2 // 'Medium cohort' in test_data_atlas.sql
	filterCohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: 2,
			CohortDefinitionId2: 32,
			ProvidedName:        "test"},
	}
	uniqueCohortDefinitionIdsList := utils.GetUniqueCohortDefinitionIdsList([]int{cohortDefinitionId}, filterCohortPairs)
	if len(uniqueCohortDefinitionIdsList) != 2 {
		t.Errorf("Expected uniqueCohortDefinitionIdsList length to be 2")
	}
	teamProjects, _ := cohortDefinitionModel.GetTeamProjectsThatMatchAllCohortDefinitionIds(uniqueCohortDefinitionIdsList)
	if len(teamProjects) != 2 {
		t.Errorf("Expected to find two 'team projects' matching the cohort list, found %s", teamProjects)
	}
	if !utils.ContainsString(teamProjects, "defaultteamproject") {
		t.Errorf("Expected to find 'defaultteamproject' in the results, found %s", teamProjects)
	}
	if !utils.ContainsString(teamProjects, "teamprojectX") {
		t.Errorf("Expected to find 'teamprojectX' in the results, found %s", teamProjects)
	}
}

func TestGetCohortDefinitionIdsForTeamProject(t *testing.T) {
	setUp(t)
	testTeamProject := "teamprojectY"
	allowedCohortDefinitionIds, _ := cohortDefinitionModel.GetCohortDefinitionIdsForTeamProject(testTeamProject)
	if len(allowedCohortDefinitionIds) != 1 {
		t.Errorf("Expected teamProject '%s' to have one cohort, but found %d",
			testTeamProject, len(allowedCohortDefinitionIds))
	}
	// test data is crafted in such a way that the default "team project" has access to all
	// the cohorts. Check if this is indeed the case:
	testTeamProject = defaultTeamProject
	allowedCohortDefinitionIds, _ = cohortDefinitionModel.GetCohortDefinitionIdsForTeamProject(testTeamProject)
	allCohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitions()
	if len(allCohortDefinitions) != len(allowedCohortDefinitionIds) && len(allCohortDefinitions) > 1 {
		t.Errorf("Found %d, expected %d", len(allowedCohortDefinitionIds), len(allCohortDefinitions))
	}
}

func TestGetAllCohortDefinitionsAndStatsOrderBySizeDesc(t *testing.T) {
	setUp(t)
	cohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, defaultTeamProject)
	if len(cohortDefinitions) != len(allCohortDefinitions) {
		t.Errorf("Found %d, expected %d", len(cohortDefinitions), len(allCohortDefinitions))
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

	// some extra tests to cover also the teamProject option for this method:
	testTeamProject := "teamprojectY"
	allowedCohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, testTeamProject)
	if len(allowedCohortDefinitions) != 1 {
		t.Errorf("Expected teamProject '%s' to have one cohort, but found %d",
			testTeamProject, len(allowedCohortDefinitions))
	}
	testTeamProject = "teamprojectX"
	allowedCohortDefinitions, _ = cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, testTeamProject)
	if len(allowedCohortDefinitions) != 2 {
		t.Errorf("Expected teamProject '%s' to have 2 cohorts, but found %d",
			testTeamProject, len(allowedCohortDefinitions))
	}
	if len(cohortDefinitions) <= len(allowedCohortDefinitions) {
		t.Errorf("Expected list of projects for '%s' to be larger than for %s",
			defaultTeamProject, testTeamProject)
	}
	testTeamProject = "teamprojectNonExisting"
	allowedCohortDefinitions, _ = cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, testTeamProject)
	if len(allowedCohortDefinitions) != 0 {
		t.Errorf("Expected teamProject '%s' to have NO cohort, but found %d",
			testTeamProject, len(allowedCohortDefinitions))
	}
}

// Tests whether the code deals correctly with the (error) situation where
// the `cohort_definition` and `cohort` tables are not in sync (more specifically
// the situation where a cohort still exists in `cohort` table but not in `cohort_definition`).
func TestGetAllCohortDefinitionsAndStatsOrderBySizeDescWhenCohortDefinitionIsMissing(t *testing.T) {
	setUp(t)
	cohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, defaultTeamProject)
	if len(cohortDefinitions) != len(allCohortDefinitions) {
		t.Errorf("Found %d", len(cohortDefinitions))
	}

	// remove one cohort_definition record and verify that the list is now indeed smaller:
	firstCohort := cohortDefinitions[0]
	tests.ExecAtlasSQLString(fmt.Sprintf("delete from %s.cohort_definition where id = %d",
		db.GetAtlasDB().Schema, firstCohort.Id))
	cohortDefinitions, _ = cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, defaultTeamProject)
	if len(cohortDefinitions) != len(allCohortDefinitions)-1 {
		t.Errorf("Number of cohor_definition records expected to be %d, found %d",
			len(allCohortDefinitions)-1, len(cohortDefinitions))
	}
	// restore:
	tests.ExecAtlasSQLString(fmt.Sprintf("insert into %s.cohort_definition (id,name,description) "+
		"values (%d, '%s', '%s')",
		db.GetAtlasDB().Schema,
		firstCohort.Id,
		firstCohort.Name,
		firstCohort.Name))
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
	filterConceptIds := []int64{}
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	data, _ := cohortDataModel.RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId, largestCohort.Id, histogramConceptId, filterConceptIds, filterCohortPairs)
	// everyone in the largestCohort has the histogramConceptId, but one person has NULL in the value_as_number:
	if len(data) != largestCohort.CohortSize-1 {
		t.Errorf("expected %d histogram data but got %d", largestCohort.CohortSize, len(data))
	}

	// now filter on the extendedCopyOfSecondLargestCohort
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: extendedCopyOfSecondLargestCohort.Id,
			ProvidedName:        "test"},
	}
	// then we expect histogram data for the overlapping population only (which is 5 for extendedCopyOfSecondLargestCohort and largestCohort):
	data, _ = cohortDataModel.RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(testSourceId, largestCohort.Id, histogramConceptId, filterConceptIds, filterCohortPairs)
	if len(data) != 5 {
		t.Errorf("expected 5 histogram data but got %d", len(data))
	}

}

func TestQueryFilterByConceptIdsHelper(t *testing.T) {
	// This test checks whether the query succeeds when the mainObservationTableAlias
	// argument passed to QueryFilterByConceptIdsHelper (last argument)
	// matches the alias used in the main query, and whether it fails otherwise.

	setUp(t)
	omopDataSource := tests.GetOmopDataSource()
	filterConceptIds := []int64{allConceptIds[0], allConceptIds[1], allConceptIds[2]}
	var personIds []struct {
		PersonId int64
	}

	// Subtest1: correct alias "observation":
	query := omopDataSource.Db.Table(omopDataSource.Schema + ".observation_continuous as observation" + omopDataSource.GetViewDirective()).
		Select("observation.person_id")
	query = models.QueryFilterByConceptIdsHelper(query, testSourceId, filterConceptIds, omopDataSource, "", "observation.person_id")
	meta_result := query.Scan(&personIds)
	if meta_result.Error != nil {
		t.Errorf("Did NOT expect an error")
	}
	// Subtest2: incorrect alias "observation"...should fail:
	query = omopDataSource.Db.Table(omopDataSource.Schema + ".observation_continuous as observationWRONG").
		Select("*")
	query = models.QueryFilterByConceptIdsHelper(query, testSourceId, filterConceptIds, omopDataSource, "", "observation.person_id")
	meta_result = query.Scan(&personIds)
	if meta_result.Error == nil {
		t.Errorf("Expected an error")
	}
}

func TestRetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(t *testing.T) {
	setUp(t)
	cohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, defaultTeamProject)
	var sumNumeric float32 = 0
	textConcat := ""
	classIdConcat := ""
	foundConceptValueAsNumberAsNil := false
	for _, cohortDefinition := range cohortDefinitions {

		cohortData, _ := cohortDataModel.RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(
			testSourceId, cohortDefinition.Id, allConceptIds)

		// count nr observation records for cohort through an independent simpler query:
		totalObservationsCohort := tests.GetCountWhere(tests.GetOmopDataSourceForSourceId(tests.GetTestSourceId()), "observation",
			fmt.Sprintf("observation_concept_id > 0 and person_id in (Select b.subject_id from %s.cohort b where b.cohort_definition_id = %d)",
				tests.GetResultsDataSource().Schema, cohortDefinition.Id))

		// 1- the number of items in cohortsData and totalObservationsCohort should match:
		if int64(len(cohortData)) != totalObservationsCohort {
			t.Errorf("Expected %d observations, found %d", totalObservationsCohort, len(cohortData))
		}

		// 2- cohortData items > 0, assuming each cohort has a person wit at least one observation
		if len(cohortData) <= 0 {
			t.Errorf("Expected some cohort data")
		}
		// 3- check data size. Max size, if all persons have data for each concept, is cohort size x len(allConceptIds):
		if len(cohortData) > cohortDefinition.CohortSize*len(allConceptIds) {
			t.Errorf("Cohort data size larger than expected. Expected < %d, got %d",
				cohortDefinition.CohortSize*len(allConceptIds), len(cohortData))
		}

		var previousPersonId int64 = -1
		defaultHareValue := map[string]bool{"non-Hispanic Asian": true, "non-Hispanic Black": true, "non-Hispanic White": true, "Hispanic": true}
		emptyDataCounter := 0
		for _, cohortDatum := range cohortData {
			// check for order: person_id is not smaller than previous person_id
			if cohortDatum.PersonId < previousPersonId {
				t.Errorf("Data not ordered by person_id!")
			}

			if cohortDatum.ConceptId == 2000007027 {
				//Only one row has empty data
				if !defaultHareValue[cohortDatum.ObservationValueAsConceptName] {
					if emptyDataCounter == 0 && cohortDatum.ObservationValueAsConceptName == "" {
						emptyDataCounter++
					} else {
						t.Errorf("Did not get concept value name correctly!")
					}
				}
			}

			previousPersonId = cohortDatum.PersonId
			if cohortDatum.ConceptValueAsNumber != nil {
				sumNumeric += *cohortDatum.ConceptValueAsNumber
			} else {
				foundConceptValueAsNumberAsNil = true
			}
			textConcat += cohortDatum.ObservationValueAsConceptName
			classIdConcat += cohortDatum.ConceptClassId
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
	// check for data: some concepts have class id, so this should not be empty
	if classIdConcat == "" {
		t.Errorf("Expected query to return concept class id information")
	}
	// check if some numeric values were nil as expected:
	if foundConceptValueAsNumberAsNil == false {
		t.Errorf("Expected query to return some nil values for ConceptValueAsNumber")
	}

}

func TestErrorForRetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(t *testing.T) {
	// Tests if the method returns an error when query fails.

	cohortDefinitions, _ := cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(testSourceId, defaultTeamProject)

	// break something in the Results schema to cause a query failure in the next method:
	tests.BreakSomething(models.Results, "cohort", "cohort_definition_id")
	// set last action to restore back:
	// run test:
	_, error := cohortDataModel.RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(
		testSourceId, cohortDefinitions[0].Id, allConceptIds)
	if error == nil {
		t.Errorf("Expected error")
	}
	// revert the broken part:
	tests.FixSomething(models.Results, "cohort", "cohort_definition_id")
}

func TestRetrieveCohortOverlapStats(t *testing.T) {
	// Tests if we get the expected overlap
	setUp(t)
	caseCohortId := secondLargestCohort.Id
	controlCohortId := secondLargestCohort.Id // to ensure we get some overlap, just repeat the same here...
	otherFilterConceptIds := []int64{}
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	stats, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		otherFilterConceptIds, filterCohortPairs)
	// basic test:
	if stats.CaseControlOverlap != int64(secondLargestCohort.CohortSize) {
		t.Errorf("Expected nr persons to be %d, found %d", secondLargestCohort.CohortSize, stats.CaseControlOverlap)
	}

	// now use largestCohort as background and filter on the extendedCopyOfSecondLargestCohort
	caseCohortId = largestCohort.Id
	controlCohortId = largestCohort.Id // to ensure we get largestCohort as initial overlap, just repeat the same here...
	filterCohortPairs = []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: smallestCohort.Id,
			CohortDefinitionId2: extendedCopyOfSecondLargestCohort.Id,
			ProvidedName:        "test"},
	}
	// then we expect overlap of 6 for extendedCopyOfSecondLargestCohort and largestCohort:
	stats, _ = cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		otherFilterConceptIds, filterCohortPairs)
	if stats.CaseControlOverlap != 6 {
		t.Errorf("Expected nr persons to be %d, found %d", 6, stats.CaseControlOverlap)
	}

	// extra test: different parameters that should return the same as above ^:
	caseCohortId = largestCohort.Id
	controlCohortId = extendedCopyOfSecondLargestCohort.Id
	filterCohortPairs = []utils.CustomDichotomousVariableDef{}
	otherFilterConceptIds = []int64{histogramConceptId} // extra filter, to cover this part of the code...
	// then we expect overlap of 5 for extendedCopyOfSecondLargestCohort and largestCohort (the filter on histogramConceptId should not matter
	// since all in largestCohort have an observation for this concept id except one person who has it but has value_as_number as NULL):
	stats2, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		otherFilterConceptIds, filterCohortPairs)
	if stats2.CaseControlOverlap != stats.CaseControlOverlap-1 {
		t.Errorf("Expected nr persons to be %d, found %d", stats.CaseControlOverlap, stats2.CaseControlOverlap)
	}

	// test for otherFilterConceptIds by filtering above on dummyContinuousConceptId, which is NOT
	// found in any observations of the largestCohort:
	otherFilterConceptIds = []int64{histogramConceptId, dummyContinuousConceptId}
	// all other arguments are the same as test above, and we expect overlap of 0, showing the otherFilterConceptIds
	// had the expected effect:
	stats3, _ := cohortDataModel.RetrieveCohortOverlapStats(testSourceId, caseCohortId, controlCohortId,
		otherFilterConceptIds, filterCohortPairs)
	if stats3.CaseControlOverlap != 0 {
		t.Errorf("Expected nr persons to be 0, found %d", stats3.CaseControlOverlap)
	}
}

func TestValidateObservationData(t *testing.T) {
	// Tests if we get the expected validation results
	setUp(t)
	var cohortDataModel = new(models.CohortData)
	nrIssues, error := cohortDataModel.ValidateObservationData([]int64{hareConceptId})
	// we know that the test dataset has at least one patient with more than one HARE:
	if error != nil {
		t.Errorf("Did not expect an error, but got %v", error)
	}
	if nrIssues == 0 {
		t.Errorf("Expected validation issues")
	}
	nrIssues2, error := cohortDataModel.ValidateObservationData([]int64{456789999}) // some random concept id not in db
	// we expect no results for a concept that does not exist:
	if error != nil {
		t.Errorf("Did not expect an error, but got %v", error)
	}
	if nrIssues2 != 0 {
		t.Errorf("Expected NO validation issues")
	}
	nrIssues3, error := cohortDataModel.ValidateObservationData([]int64{})
	// we expect no results for an empty concept list:
	if error != nil {
		t.Errorf("Did not expect an error, but got %v", error)
	}
	if nrIssues3 != -1 {
		t.Errorf("Expected result to be -1")
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
	originalCohortSize := thirdLargestCohort.CohortSize
	originalCohortId := thirdLargestCohort.Id
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

func TestAddTimeoutToQuery(t *testing.T) {
	setUp(t)

	// take a simple query, run with short timeout, and expect error:
	db2 := db.GetAtlasDB().Db
	var dataSource []*models.Source
	query := db2.Model(&models.Source{}).
		Select("source_id, source_name")
	query, cancel := utils.AddSpecificTimeoutToQuery(query, 2*time.Nanosecond)
	defer cancel()
	meta_result := query.Scan(&dataSource)
	if meta_result.Error == nil || len(dataSource) > 0 {
		t.Errorf("Expected timeout error and NO data")
	}

	// then switch to default (longer) timeout and expect a result:
	query2, cancel2 := utils.AddTimeoutToQuery(query)
	defer cancel2()
	meta_result2 := query2.Scan(&dataSource)

	if meta_result2.Error != nil || len(dataSource) == 0 {
		t.Errorf("Expected result and NO error")
	}
}

func TestPersonConceptAndCountString(t *testing.T) {
	a := models.PersonConceptAndCount{
		PersonId:  1,
		ConceptId: 2,
		Count:     3,
	}

	expected := "(person_id=1, concept_id=2, count=3)"
	if a.String() != expected {
		t.Errorf("Expected %s, found %s", expected, a.String())
	}

}

func TestGetDataDictionaryFail(t *testing.T) {
	setUp(t)

	data, _ := dataDictionaryModel.GetDataDictionary()
	//Pre generation cache should be empty
	if data != nil {
		t.Errorf("Get Data Dictionary should have failed.")
	}
}

func TestCheckIfDataDictionaryIsFilled(t *testing.T) {
	setUp(t)
	var source = new(models.Source)
	sources, _ := source.GetAllSources()
	var dataSourceModel = new(models.Source)
	miscDataSource := dataSourceModel.GetDataSource(sources[0].SourceId, models.Misc)

	filled := dataDictionaryModel.CheckIfDataDictionaryIsFilled(miscDataSource)
	if filled != false {
		t.Errorf("Flag should be false")
	}
	dataDictionaryModel.GenerateDataDictionary()
	filled = dataDictionaryModel.CheckIfDataDictionaryIsFilled(miscDataSource)
	if filled != true {
		t.Errorf("Flag should be true")
	}
}

func TestGenerateDataDictionary(t *testing.T) {
	setUp(t)
	dataDictionaryModel.GenerateDataDictionary()
	//Update this with read
	data, _ := dataDictionaryModel.GetDataDictionary()
	if data == nil || data.Total != 18 || data.Data == nil {
		t.Errorf("Get Data Dictionary should have succeeded.")
	}
}

func TestWriteToDB(t *testing.T) {
	setUp(t)
	var source = new(models.Source)
	sources, _ := source.GetAllSources()
	var dataSourceModel = new(models.Source)
	miscDataSource := dataSourceModel.GetDataSource(sources[0].SourceId, models.Misc)

	resultList := append([]*models.DataDictionaryResult{}, &models.DataDictionaryResult{ConceptID: 123})
	success := dataDictionaryModel.WriteResultToDB(miscDataSource, resultList)
	//Write succeeded without panicking
	if success != true {
		t.Errorf("Write failed")
	}
}
