package controllers_tests

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/controllers"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/tests"
	"github.com/uc-cdis/cohort-middleware/utils"
)

var testSourceId = tests.GetTestSourceId()

func TestMain(m *testing.M) {
	setupSuite()
	retCode := m.Run()
	tearDownSuite()
	os.Exit(retCode)
}

func setupSuite() {
	log.Println("setup for suite")
}

func tearDownSuite() {
	log.Println("teardown for suite")
}

func setUp(t *testing.T) {
	log.Println("setup for test")
	dummyModelReturnError = false
	config.Init("mocktest")

	// ensure tearDown is called when test "t" is done:
	t.Cleanup(func() {
		tearDown()
	})
}

func tearDown() {
	log.Println("teardown for test")
}

var cohortDataController = controllers.NewCohortDataController(*new(dummyCohortDataModel), *new(dummyDataDictionaryModel), *new(dummyTeamProjectAuthz))
var cohortDataControllerWithFailingTeamProjectAuthz = controllers.NewCohortDataController(*new(dummyCohortDataModel), *new(dummyDataDictionaryModel), &dummyFailingTeamProjectAuthz{failForGlobalOnly: false})
var cohortDataControllerWithFailingDataDictionary = controllers.NewCohortDataController(*new(dummyCohortDataModel), *new(dummyFailingDataDictionaryModel), *new(dummyTeamProjectAuthz))

// instance of the controller that talks to the regular model implementation (that needs a real DB):
var cohortDefinitionControllerNeedsDb = controllers.NewCohortDefinitionController(*new(models.CohortDefinition), *new(dummyTeamProjectAuthz))

// instance of the controller that talks to a mock implementation of the model:
var cohortDefinitionController = controllers.NewCohortDefinitionController(*new(dummyCohortDefinitionDataModel), *new(dummyTeamProjectAuthz))
var cohortDefinitionControllerWithFailingTeamProjectAuthz = controllers.NewCohortDefinitionController(*new(dummyCohortDefinitionDataModel), &dummyFailingTeamProjectAuthz{failForGlobalOnly: false})
var cohortDefinitionControllerWithNoTeamProjectAuthzForGlobalReaderRole = controllers.NewCohortDefinitionController(*new(dummyCohortDefinitionDataModel), &dummyFailingTeamProjectAuthz{failForGlobalOnly: true})

type dummyCohortDataModel struct{}

func (h dummyCohortDataModel) RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*models.PersonConceptAndValue, error) {
	value := float32(0.0)
	cohortData := []*models.PersonConceptAndValue{
		{PersonId: 1, ConceptId: 10, ConceptClassId: "something", ConceptValueAsString: "abc", ConceptValueAsNumber: &value},
	}
	return cohortData, nil
}

func (h dummyCohortDataModel) RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, histogramConceptId int64, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) ([]*models.PersonConceptAndValue, error) {

	cohortData := []*models.PersonConceptAndValue{}
	return cohortData, nil
}

func (h dummyCohortDataModel) RetrieveBarGraphDataBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, histogramConceptId int64) ([]*models.OrdinalGroupData, error) {
	cohortData := []*models.OrdinalGroupData{}
	return cohortData, nil
}

func (h dummyCohortDataModel) RetrieveCohortOverlapStats(sourceId int, caseCohortId int, controlCohortId int,
	otherFilterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) (models.CohortOverlapStats, error) {
	var zeroOverlap models.CohortOverlapStats
	return zeroOverlap, nil
}

func (h dummyCohortDataModel) RetrieveDataByOriginalCohortAndNewCohort(sourceId int, originalCohortDefinitionId int, cohortDefinitionId int) ([]*models.PersonIdAndCohort, error) {
	if cohortDefinitionId == 2 {
		return []*models.PersonIdAndCohort{
			{PersonId: 1, CohortId: int64(cohortDefinitionId)},
		}, nil
	}

	return []*models.PersonIdAndCohort{
		{PersonId: 2, CohortId: int64(cohortDefinitionId)},
		{PersonId: 3, CohortId: int64(cohortDefinitionId)},
	}, nil
}

type dummyCohortDefinitionDataModel struct{}

var dummyModelReturnError bool = false

func (h dummyCohortDefinitionDataModel) GetCohortDefinitionIdsForTeamProject(teamProject string) ([]int, error) {
	return []int{1}, nil
}

func (h dummyCohortDefinitionDataModel) GetTeamProjectsThatMatchAllCohortDefinitionIds(uniqueCohortDefinitionIdsList []int) ([]string, error) {
	return []string{"test"}, nil
}

func (h dummyCohortDefinitionDataModel) GetCohortName(cohortId int) (string, error) {
	return "dummy cohort name", nil
}

func (h dummyCohortDefinitionDataModel) GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int, teamProject string) ([]*models.CohortDefinitionStats, error) {
	conf := config.GetConfig()
	globalReaderRole := conf.GetString("global_reader_role")
	if teamProject == globalReaderRole {
		cohortDefinitionStats := []*models.CohortDefinitionStats{
			{Id: 4, CohortSize: 4, Name: "name4_" + teamProject}, // just concatenate teamProject here, so we can assert on it in a later test... teamprojects are otherwise not really part of cohort names
			{Id: 5, CohortSize: 55, Name: "name5_" + teamProject},
		}
		return cohortDefinitionStats, nil
	} else {
		cohortDefinitionStats := []*models.CohortDefinitionStats{
			{Id: 1, CohortSize: 10, Name: "name1_" + teamProject}, // just concatenate teamProject here, so we can assert on it in a later test... teamprojects are otherwise not really part of cohort names
			{Id: 2, CohortSize: 32, Name: "name2_" + teamProject},
			{Id: 3, CohortSize: 23, Name: "name3_" + teamProject},
		}
		return cohortDefinitionStats, nil
	} // when ordered by size descending, we get cohorts 5, 2, 3, 1, 4 (used in TestRetriveStatsBySourceIdAndTeamProject later on)
}

func (h dummyCohortDefinitionDataModel) GetCohortDefinitionById(id int) (*models.CohortDefinition, error) {
	cohortDefinition := models.CohortDefinition{
		Id:             1,
		Name:           "test 1",
		Description:    "test desc 1",
		ExpressionType: "?",
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("fake model error!")
	}
	return &cohortDefinition, nil
}
func (h dummyCohortDefinitionDataModel) GetCohortDefinitionByName(name string) (*models.CohortDefinition, error) {
	return nil, nil
}
func (h dummyCohortDefinitionDataModel) GetAllCohortDefinitions() ([]*models.CohortDefinition, error) {
	return nil, nil
}

type dummyTeamProjectAuthz struct{}

func (h dummyTeamProjectAuthz) TeamProjectValidationForCohort(ctx *gin.Context, cohortDefinitionId int) bool {
	return true
}

func (h dummyTeamProjectAuthz) TeamProjectValidation(ctx *gin.Context, cohortDefinitionIds []int, filterCohortPairs []utils.CustomDichotomousVariableDef) bool {
	return true
}

func (h dummyTeamProjectAuthz) TeamProjectValidationForCohortIdsList(ctx *gin.Context, uniqueCohortDefinitionIdsList []int) bool {
	return true
}

func (h dummyTeamProjectAuthz) HasAccessToTeamProject(ctx *gin.Context, teamProject string) bool {
	return true
}

type dummyFailingTeamProjectAuthz struct {
	failForGlobalOnly bool
}

func (h dummyFailingTeamProjectAuthz) TeamProjectValidationForCohort(ctx *gin.Context, cohortDefinitionId int) bool {
	return false
}

func (h dummyFailingTeamProjectAuthz) TeamProjectValidation(ctx *gin.Context, cohortDefinitionIds []int, filterCohortPairs []utils.CustomDichotomousVariableDef) bool {
	return false
}

func (h dummyFailingTeamProjectAuthz) TeamProjectValidationForCohortIdsList(ctx *gin.Context, uniqueCohortDefinitionIdsList []int) bool {
	return false
}

func (h dummyFailingTeamProjectAuthz) HasAccessToTeamProject(ctx *gin.Context, teamProject string) bool {
	conf := config.GetConfig()
	globalReaderRole := conf.GetString("global_reader_role")
	if h.failForGlobalOnly {
		if teamProject == globalReaderRole {
			return false
		} else {
			return true
		}
	} else {
		return false
	}
}

var conceptController = controllers.NewConceptController(*new(dummyConceptDataModel), *new(dummyCohortDefinitionDataModel), *new(dummyTeamProjectAuthz))
var conceptControllerWithFailingTeamProjectAuthz = controllers.NewConceptController(*new(dummyConceptDataModel), *new(dummyCohortDefinitionDataModel), *new(dummyFailingTeamProjectAuthz))

type dummyConceptDataModel struct{}

func (h dummyConceptDataModel) RetriveAllBySourceId(sourceId int) ([]*models.Concept, error) {
	return nil, nil
}

func (h dummyConceptDataModel) RetrieveInfoBySourceIdAndConceptId(sourceId int, conceptId int64) (*models.ConceptSimple, error) {
	conceptSimpleItems := []*models.ConceptSimple{
		{ConceptId: 1234, ConceptName: "Concept A"},
		{ConceptId: 5678, ConceptName: "Concept B"},
		{ConceptId: 2090006880, ConceptName: "Concept C"},
	}
	for _, conceptSimple := range conceptSimpleItems {
		if conceptSimple.ConceptId == conceptId {
			return conceptSimple, nil
		}
	}
	return nil, fmt.Errorf("concept id %d not found in mock data", conceptId)
}

func (h dummyConceptDataModel) RetrieveInfoBySourceIdAndConceptIds(sourceId int, conceptIds []int64) ([]*models.ConceptSimple, error) {
	// dummy data with _some_ of the relevant fields:
	conceptSimple := []*models.ConceptSimple{
		{ConceptId: 1234, ConceptName: "Concept A"},
		{ConceptId: 5678, ConceptName: "Concept B"},
		{ConceptId: 2090006880, ConceptName: "Concept C"},
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("fake model error!")
	}
	return conceptSimple, nil
}
func (h dummyConceptDataModel) RetrieveInfoBySourceIdAndConceptTypes(sourceId int, conceptTypes []string) ([]*models.ConceptSimple, error) {
	// dummy data with _some_ of the relevant fields:
	conceptSimple := []*models.ConceptSimple{
		{ConceptId: 1234, ConceptName: "Concept A"},
		{ConceptId: 5678, ConceptName: "Concept B"},
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("fake model error!")
	}
	return conceptSimple, nil
}
func (h dummyConceptDataModel) RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId int, cohortDefinitionId int, breakdownConceptId int64) ([]*models.ConceptBreakdown, error) {
	conceptBreakdown := []*models.ConceptBreakdown{
		{ConceptValue: "value1", NpersonsInCohortWithValue: 5, ValueName: "value1_name"},
		{ConceptValue: "value2", NpersonsInCohortWithValue: 8, ValueName: "value2_name"},
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("error!")
	}
	return conceptBreakdown, nil
}
func (h dummyConceptDataModel) RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef, breakdownConceptId int64) ([]*models.ConceptBreakdown, error) {
	conceptBreakdown := []*models.ConceptBreakdown{
		{ConceptValue: "value1", NpersonsInCohortWithValue: 4 - len(filterCohortPairs)}, // simulate decreasing numbers as filter increases - the use of filterCohortPairs instead of filterConceptIds is otherwise meaningless here...
		{ConceptValue: "value2", NpersonsInCohortWithValue: 7 - len(filterConceptIds)},  // simulate decreasing numbers as filter increases- the use of filterConceptIds instead of filterCohortPairs is otherwise meaningless here...
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("error!")
	}
	return conceptBreakdown, nil
}

type dummyDataDictionaryModel struct{}

func (h dummyDataDictionaryModel) GetDataDictionary() (*models.DataDictionaryModel, error) {
	data := new(models.DataDictionaryModel)
	data.Total = 2
	entries := []*models.DataDictionaryEntry{
		{VocabularyID: "Measurement", ConceptID: 2000006885, ConceptCode: "F", ConceptClassId: "MVP Continuous", NumberOfPeopleWithVariable: 16, NumberOfPeopleWhereValueIsFilled: 15, NumberOfPeopleWhereValueIsNull: 1, ValueStoredAs: "Number", MinValue: 1.16, MaxValue: 9.52, MeanValue: 5.9425, StandardDeviation: 2.3093216897320015, ValueSummary: nil},
		{VocabularyID: "Person", ConceptID: 2000007027, ConceptCode: "HARE_CODE", ConceptClassId: "MVP Ordinal", NumberOfPeopleWithVariable: 11, NumberOfPeopleWhereValueIsFilled: 10, NumberOfPeopleWhereValueIsNull: 1, ValueStoredAs: "Concept Id", MinValue: 0, MaxValue: 0, MeanValue: 0, StandardDeviation: 0, ValueSummary: nil},
	}
	valueSummary1, _ := json.Marshal("[{\"start\":1.159999966621399,\"end\":4.723933216866351,\"personCount\":4},{\"start\":4.723933216866351,\"end\":8.287866467111304,\"personCount\":7},{\"start\":8.287866467111304,\"end\":11.851799717356256,\"personCount\":2}]")
	valueSummary2, _ := json.Marshal("[{\"Name\":\"non-Hispanic Black\",\"PersonCount\":4,\"ValueAsString\":\"AFR\",\"ValueAsConceptID\":2000007030},{\"Name\":\"non-Hispanic Asian\",\"PersonCount\":3,\"ValueAsString\":\"ASN\",\"ValueAsConceptID\":2000007029},{\"Name\":\"non-Hispanic White\",\"PersonCount\":2,\"ValueAsString\":\"EUR\",\"ValueAsConceptID\":2000007031},{\"Name\":\"Hispanic\",\"PersonCount\":2,\"ValueAsString\":\"HIS\",\"ValueAsConceptID\":2000007028},{\"Name\":\"\",\"PersonCount\":1,\"ValueAsString\":\"\",\"ValueAsConceptID\":0}]")
	entries[0].ValueSummary = valueSummary1
	entries[1].ValueSummary = valueSummary2
	data.Data, _ = json.Marshal(entries)

	return data, nil
}

func (h dummyDataDictionaryModel) GenerateDataDictionary() (*models.DataDictionaryModel, error) {
	return nil, nil
}

type dummyFailingDataDictionaryModel struct{}

func (h dummyFailingDataDictionaryModel) GetDataDictionary() (*models.DataDictionaryModel, error) {
	return nil, errors.New("data dictionary is not available yet")
}

func (h dummyFailingDataDictionaryModel) GenerateDataDictionary() (*models.DataDictionaryModel, error) {
	return nil, nil
}
func TestRetrieveHistogramForCohortIdAndConceptIdWithWrongParams(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "4"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestBody := "{\"variables\":[{\"variable_type\": \"custom_dichotomous\", \"cohort_ids\": [1, 3]}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	//requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDataController.RetrieveHistogramForCohortIdAndConceptId(requestContext)
	// Params above are wrong, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("should have aborted")
	}
}

func TestRetrieveHistogramForCohortIdAndConceptIdWithCorrectParams(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "4"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "histogramid", Value: "2000006885"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestBody := "{\"variables\":[{\"variable_type\": \"concept\", \"concept_id\": 2000000324},{\"variable_type\": \"custom_dichotomous\", \"cohort_ids\": [1, 3]}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	cohortDataController.RetrieveHistogramForCohortIdAndConceptId(requestContext)
	// Params above are correct, so request should NOT abort:
	if requestContext.IsAborted() {
		t.Errorf("Did not expect this request to abort")
	}
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !strings.Contains(result.CustomResponseWriterOut, "bins") {
		t.Errorf("Expected output starting with 'bins,...'")
	}

	// the same request should fail if the teamProject authorization fails:
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	cohortDataControllerWithFailingTeamProjectAuthz.RetrieveHistogramForCohortIdAndConceptId(requestContext)
	result = requestContext.Writer.(*tests.CustomResponseWriter)
	// expect error:
	if !strings.Contains(result.CustomResponseWriterOut, "access denied") {
		t.Errorf("Expected 'access denied' as result")
	}
	if !requestContext.IsAborted() {
		t.Errorf("Expected request to be aborted")
	}
}

func TestRetrieveDataBySourceIdAndCohortIdAndVariablesWrongParams(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "Abc", Value: "def"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDataController.RetrieveDataBySourceIdAndCohortIdAndVariables(requestContext)
	// Params above are wrong, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
}

func TestRetrieveDataBySourceIdAndCohortIdAndVariablesCorrectParams(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "1"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestBody := "{\"variables\":[{\"variable_type\": \"concept\", \"concept_id\": 2000000324},{\"variable_type\": \"custom_dichotomous\", \"cohort_ids\": [1, 3]}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	cohortDataController.RetrieveDataBySourceIdAndCohortIdAndVariables(requestContext)
	// Params above are correct, so request should NOT abort:
	if requestContext.IsAborted() {
		t.Errorf("Did not expect this request to abort")
	}
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !strings.Contains(result.CustomResponseWriterOut, "sample.id,") {
		t.Errorf("Expected output starting with 'sample.id,...'")
	}

	// the same request should fail if the teamProject authorization fails:
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	cohortDataControllerWithFailingTeamProjectAuthz.RetrieveDataBySourceIdAndCohortIdAndVariables(requestContext)
	result = requestContext.Writer.(*tests.CustomResponseWriter)
	// expect error:
	if !strings.Contains(result.CustomResponseWriterOut, "access denied") {
		t.Errorf("Expected 'access denied' as result")
	}
	if !requestContext.IsAborted() {
		t.Errorf("Expected request to be aborted")
	}
}

func TestRetrieveCohortOverlapStats(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "casecohortid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "controlcohortid", Value: "2"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestBody := "{\"variables\":[{\"variable_type\": \"concept\", \"concept_id\": 2000000324},{\"variable_type\": \"concept\", \"concept_id\": 2000006885}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))

	cohortDataController.RetrieveCohortOverlapStats(requestContext)
	// Params above are correct, so request should NOT abort:
	if requestContext.IsAborted() {
		t.Errorf("Did not expect this request to abort")
	}
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !strings.Contains(result.CustomResponseWriterOut, "case_control_overlap") {
		t.Errorf("Expected output containing 'case_control_overlap...'")
	}

	// the same request should fail if the teamProject authorization fails:
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	cohortDataControllerWithFailingTeamProjectAuthz.RetrieveCohortOverlapStats(requestContext)
	result = requestContext.Writer.(*tests.CustomResponseWriter)
	// expect error:
	if !strings.Contains(result.CustomResponseWriterOut, "access denied") {
		t.Errorf("Expected 'access denied' as result")
	}
	if !requestContext.IsAborted() {
		t.Errorf("Expected request to be aborted")
	}
}

func TestRetrieveCohortOverlapStatsBadRequest(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Writer = new(tests.CustomResponseWriter)

	cohortDataController.RetrieveCohortOverlapStats(requestContext)
	// Params above are incorrect, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("Expected this request to abort")
	}
}

func TestGenerateCSV(t *testing.T) {
	setUp(t)
	value1 := float32(0.0)
	value2 := float32(1.5)

	cohortData := []*models.PersonConceptAndValue{
		{PersonId: 1, ConceptId: 10, ConceptClassId: "something else", ConceptValueAsString: "abc", ConceptValueAsNumber: &value1},
		{PersonId: 1, ConceptId: 22, ConceptClassId: "MVP Continuous", ConceptValueAsString: ">1", ConceptValueAsNumber: &value2},
		{PersonId: 2789580123456, ConceptId: 10, ConceptValueAsString: "A value with, comma!", ConceptValueAsNumber: &value1},
		{PersonId: 344567, ConceptId: tests.GetTestHareConceptId(), ConceptClassId: "MVP Ordinal", ConceptValueAsString: "HIS", ConceptValueAsConceptId: 2000007028, ConceptValueAsNumber: &value1},
		{PersonId: 344567, ConceptId: 22, ConceptClassId: "MVP Continuous", ConceptValueAsString: "", ConceptValueAsNumber: &value1},
		{PersonId: 789567, ConceptId: 22, ConceptClassId: "MVP Continuous"},
		{PersonId: 789567, ConceptId: 10, ConceptClassId: "something else", ConceptValueAsString: ""},
	}
	conceptIds := []int64{10, 22, tests.GetTestHareConceptId()}

	csvLines := controllers.GeneratePartialCSV(
		testSourceId, cohortData, conceptIds)

	// the above should result in one header line and 2 data lines (2 persons)
	if len(csvLines) != 5 {
		t.Errorf("Expected 1 header line + 4 data lines, found %d lines in total",
			len(csvLines))
		t.Errorf("Lines: %s", csvLines)
	}
	expectedLines := [][]string{
		{"sample.id", "ID_10", "ID_22", fmt.Sprintf("ID_%d", tests.GetTestHareConceptId())},
		{"1", "abc", "1.50", "NA"},
		{"2789580123456", "A value with, comma!", "NA", "NA"},
		{"344567", "NA", "0.00", "HIS"},
		{"789567", "NA", "NA", "NA"},
	}

	for i, expectedLine := range expectedLines {
		if !reflect.DeepEqual(expectedLine, csvLines[i]) {
			t.Errorf("CSV line not as expected. \nExpected: \n%s \nFound: \n%s",
				expectedLine, csvLines[i])
		}

	}
}

func TestRetriveStatsBySourceIdAndTeamProjectWrongParams(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "Abc", Value: "def"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionController.RetriveStatsBySourceIdAndTeamProject(requestContext)
	// Params above are wrong, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
}

func TestRetriveStatsBySourceIdAndTeamProjectDbPanic(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Request = &http.Request{URL: &url.URL{}}
	requestContext.Request.URL.RawQuery = "team-project=dummy-team-project"
	requestContext.Writer = new(tests.CustomResponseWriter)

	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
			if err != "AtlasDB not initialized" {
				t.Errorf("Expected error")
			}
		}
	}()
	cohortDefinitionControllerNeedsDb.RetriveStatsBySourceIdAndTeamProject(requestContext)
	t.Errorf("Expected error")
}

func TestRetriveStatsBySourceIdAndTeamProjectCheckMandatoryTeamProject(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionController.RetriveStatsBySourceIdAndTeamProject(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// Params above are wrong, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
	if !strings.Contains(result.CustomResponseWriterOut, "team-project is a mandatory parameter") {
		t.Errorf("Expected error about mandatory team-project")
	}
}

func TestRetriveStatsBySourceIdAndTeamProjectAuthorizationError(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Request = &http.Request{URL: &url.URL{}}
	teamProject := "/test/dummyname/dummy-team-project"
	requestContext.Request.URL.RawQuery = "team-project=" + teamProject
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionControllerWithFailingTeamProjectAuthz.RetriveStatsBySourceIdAndTeamProject(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
	if result.Status() != http.StatusForbidden {
		t.Errorf("Expected StatusForbidden, got %d", result.Status())
	}
	if !strings.Contains(result.CustomResponseWriterOut, "access denied") {
		t.Errorf("Expected 'access denied' in response")
	}
}

func TestRetriveStatsBySourceIdAndTeamProject(t *testing.T) {
	setUp(t)

	conf := config.GetConfig()
	globalReaderRole := conf.GetString("global_reader_role")
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Request = &http.Request{URL: &url.URL{}}
	teamProject := "/test/dummyname/dummy-team-project"
	requestContext.Request.URL.RawQuery = "team-project=" + teamProject
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionController.RetriveStatsBySourceIdAndTeamProject(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// expect result with all of the dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "name1_"+teamProject) ||
		!strings.Contains(result.CustomResponseWriterOut, "name2_"+teamProject) ||
		!strings.Contains(result.CustomResponseWriterOut, "name3_"+teamProject) ||
		!strings.Contains(result.CustomResponseWriterOut, "name4_"+globalReaderRole) ||
		!strings.Contains(result.CustomResponseWriterOut, "name5_"+globalReaderRole) {
		t.Errorf("Expected 5 specific rows in result, found %v", result.CustomResponseWriterOut)
	}
	// check if sorted descending. The name5_ (CohortSize=55) should be first, name4_ (CohortSize=4) last:
	index1 := strings.Index(result.CustomResponseWriterOut, "name1_"+teamProject)
	index2 := strings.Index(result.CustomResponseWriterOut, "name2_"+teamProject)
	index3 := strings.Index(result.CustomResponseWriterOut, "name3_"+teamProject)
	index4 := strings.Index(result.CustomResponseWriterOut, "name4_"+globalReaderRole)
	index5 := strings.Index(result.CustomResponseWriterOut, "name5_"+globalReaderRole)
	// we expect index5 < index2 < index3 < index1 < index4:
	if !(index5 < index2 && index2 < index3 && index3 < index1 && index1 < index4) {
		t.Errorf("Items in result are not sorted correctly: %v", result.CustomResponseWriterOut)
	}
}

func TestMakeUniqueListOfCohortStats(t *testing.T) {
	setUp(t)
	testInput := []*models.CohortDefinitionStats{}
	testInput = controllers.MakeUniqueListOfCohortStats(testInput)
	if len(testInput) > 0 {
		t.Errorf("Expected empty result")
	}
	testInput = []*models.CohortDefinitionStats{
		{Id: 123, Name: "a1", CohortSize: 123},
		{Id: 123, Name: "a1", CohortSize: 123},
		{Id: 456, Name: "a1", CohortSize: 123},
	}
	testInput = controllers.MakeUniqueListOfCohortStats(testInput)
	if len(testInput) != 2 {
		t.Errorf("Expected result of size 2")
	}
	if testInput[0].Id != 123 || testInput[1].Id != 456 {
		t.Errorf("Unexpected result")
	}
}

func TestRetriveStatsBySourceIdAndTeamProjectWithNoTeamProjectAuthzForGlobalReaderRole(t *testing.T) {
	setUp(t)

	conf := config.GetConfig()
	globalReaderRole := conf.GetString("global_reader_role")
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Request = &http.Request{URL: &url.URL{}}
	teamProject := "/test/dummyname/dummy-team-project"
	requestContext.Request.URL.RawQuery = "team-project=" + teamProject
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionControllerWithNoTeamProjectAuthzForGlobalReaderRole.RetriveStatsBySourceIdAndTeamProject(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// expect result with all of the dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "name1_"+teamProject) ||
		!strings.Contains(result.CustomResponseWriterOut, "name2_"+teamProject) ||
		!strings.Contains(result.CustomResponseWriterOut, "name3_"+teamProject) {
		t.Errorf("Expected 3 specific rows in result, found %v", result.CustomResponseWriterOut)
	}
	if strings.Contains(result.CustomResponseWriterOut, "name4_"+globalReaderRole) ||
		strings.Contains(result.CustomResponseWriterOut, "name5_"+globalReaderRole) {
		t.Errorf("Did not expect to find global roles in result, but found %v", result.CustomResponseWriterOut)
	}

}

func TestRetriveByIdWrongParam(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "Abc", Value: "def"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionController.RetriveById(requestContext)
	// Params above are wrong, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
}

func TestRetriveById(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "id", Value: "1"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionController.RetriveById(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "test 1") {
		t.Errorf("Expected data in result")
	}
}

func TestRetriveByIdModelError(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "id", Value: "1"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	// set flag to let mock model layer return error instead of mock data:
	dummyModelReturnError = true
	cohortDefinitionController.RetriveById(requestContext)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
}

func TestRetriveByIdAuthorizationError(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "id", Value: "1"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionControllerWithFailingTeamProjectAuthz.RetriveById(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
	if result.Status() != http.StatusForbidden {
		t.Errorf("Expected StatusForbidden, got %d", result.Status())
	}
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "access denied") {
		t.Errorf("Expected 'access denied' in response")
	}

}

func TestRetrieveBreakdownStatsBySourceIdAndCohortId(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "breakdownconceptid", Value: "1"})

	requestContext.Writer = new(tests.CustomResponseWriter)
	conceptController.RetrieveBreakdownStatsBySourceIdAndCohortId(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "persons_in_cohort_with_value") {
		t.Errorf("Expected data in result")
	}

	// the same request should fail if the teamProject authorization fails:
	conceptControllerWithFailingTeamProjectAuthz.RetrieveBreakdownStatsBySourceIdAndCohortId(requestContext)
	result = requestContext.Writer.(*tests.CustomResponseWriter)
	// expect error:
	if !strings.Contains(result.CustomResponseWriterOut, "access denied") {
		t.Errorf("Expected 'access denied' as result")
	}
	if !requestContext.IsAborted() {
		t.Errorf("Expected request to be aborted")
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndVariables(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "breakdownconceptid", Value: "1"})
	requestContext.Request = new(http.Request)
	requestBody := "{\"variables\":[{\"variable_type\": \"concept\", \"concept_id\": 1234},{\"variable_type\": \"concept\", \"concept_id\": 5678}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))

	requestContext.Writer = new(tests.CustomResponseWriter)
	conceptController.RetrieveBreakdownStatsBySourceIdAndCohortIdAndVariables(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "persons_in_cohort_with_value") {
		t.Errorf("Expected data in result")
	}

	// the same request should fail if the teamProject authorization fails:
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	conceptControllerWithFailingTeamProjectAuthz.RetrieveBreakdownStatsBySourceIdAndCohortIdAndVariables(requestContext)
	result = requestContext.Writer.(*tests.CustomResponseWriter)
	// expect error:
	if !strings.Contains(result.CustomResponseWriterOut, "access denied") {
		t.Errorf("Expected 'access denied' as result")
	}
	if !requestContext.IsAborted() {
		t.Errorf("Expected request to be aborted")
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndVariablesModelError(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "breakdownconceptid", Value: "1"})
	requestContext.Request = new(http.Request)
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"ConceptIds\":[1234,5678]}"))
	requestContext.Writer = new(tests.CustomResponseWriter)
	// set flag to let mock model layer return error instead of mock data:
	dummyModelReturnError = true
	conceptController.RetrieveBreakdownStatsBySourceIdAndCohortIdAndVariables(requestContext)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
}

func TestRetrieveInfoBySourceIdAndConceptIds(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Request = new(http.Request)
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"ConceptIds\":[1234,5678]}"))
	requestContext.Writer = new(tests.CustomResponseWriter)
	conceptController.RetrieveInfoBySourceIdAndConceptIds(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "Concept A") ||
		!strings.Contains(result.CustomResponseWriterOut, "Concept B") {
		t.Errorf("Expected data in result")
	}
}

func TestRetrieveInfoBySourceIdAndConceptTypes(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Request = new(http.Request)
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"ConceptTypes\":[\"Type A\",\"Type_B\"]}"))
	requestContext.Writer = new(tests.CustomResponseWriter)
	conceptController.RetrieveInfoBySourceIdAndConceptTypes(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "Concept A") ||
		!strings.Contains(result.CustomResponseWriterOut, "Concept B") {
		t.Errorf("Expected data in result")
	}
}

func TestRetrieveInfoBySourceIdAndConceptTypesModelError(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Request = new(http.Request)
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"ConceptTypes\":[\"Type A\",\"Type_B\"]}"))
	requestContext.Writer = new(tests.CustomResponseWriter)
	// set flag to let mock model layer return error instead of mock data:
	dummyModelReturnError = true
	conceptController.RetrieveInfoBySourceIdAndConceptTypes(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
	if !strings.Contains(result.CustomResponseWriterOut, "fake model error") {
		t.Errorf("Expected model error message")
	}
}

func TestRetrieveInfoBySourceIdAndConceptTypesArgsError(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "Xwrongourceid", Value: "1"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	// set flag to let mock model layer return error instead of mock data:
	dummyModelReturnError = true
	conceptController.RetrieveInfoBySourceIdAndConceptTypes(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
	if !strings.Contains(result.CustomResponseWriterOut, "bad request - sourceid") {
		t.Errorf("Expected error message on sourceid")
	}
}

func TestRetrieveInfoBySourceIdAndConceptTypesMissingBody(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	// set flag to let mock model layer return error instead of mock data:
	dummyModelReturnError = true
	conceptController.RetrieveInfoBySourceIdAndConceptTypes(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
	if !strings.Contains(result.CustomResponseWriterOut, "no request body") {
		t.Errorf("Expected 'no request body' error message")
	}
}

func TestGenerateAttritionCSV(t *testing.T) {
	setUp(t)
	headerAndNonFilteredRow := [][]string{
		{"Cohort", "Size", "Cohort_val_1", "Cohort_val_2"},
		{"test cohort", "5", "2", "3"},
	}
	filteredRows := [][]string{
		{"filtered_val", "4", "2", "2"},
	}
	b := controllers.GenerateAttritionCSV(headerAndNonFilteredRow, filteredRows)
	csvLines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	if len(csvLines) != 3 {
		t.Errorf("Expected 1 header line + 2 data lines, found %d lines in total",
			len(csvLines))
		t.Errorf("Lines: %s", csvLines)
	}

	expectedLines := []string{
		"Cohort,Size,Cohort_val_1,Cohort_val_2",
		"test cohort,5,2,3",
		"filtered_val,4,2,2",
	}
	i := 0
	for _, expectedLine := range expectedLines {
		if csvLines[i] != expectedLine {
			t.Errorf("CSV line not as expected. \nExpected: \n%s \nFound: \n%s",
				expectedLine, csvLines[i])
		}
		i++
	}

}

func TestGenerateHeaderAndNonFilterRow(t *testing.T) {
	setUp(t)
	cohortName := "hello"

	conceptBreakdown := []*models.ConceptBreakdown{
		{ConceptValue: "value1", NpersonsInCohortWithValue: 5, ValueName: "value1_name"},
		{ConceptValue: "value2", NpersonsInCohortWithValue: 8, ValueName: "value2_name"},
	}

	sortedConceptValues := []string{"value1", "value2"}

	result, _ := conceptController.GenerateHeaderAndNonFilteredRow(conceptBreakdown, sortedConceptValues, cohortName)
	if len(result) != 2 {
		t.Errorf("Expected 1 header line + 1 data lines, found %d lines in total",
			len(result))
		t.Errorf("Lines: %s", result)
	}

	expectedLines := [][]string{
		{"Cohort", "Size", "value1_name", "value2_name"},
		{"hello", "13", "5", "8"},
	}

	i := 0
	for _, expectedLine := range expectedLines {
		if !reflect.DeepEqual(expectedLine, result[i]) {
			t.Errorf("header or non filter row line not as expected. \nExpected: \n%s \nFound: \n%s",
				expectedLine, result[i])
		}
		i++
	}
}

func TestGetAttritionRowForConceptIdsAndCohortPairs(t *testing.T) {
	setUp(t)
	sourceId := 1
	cohortId := 1
	var breakdownConceptId int64 = 1
	sortedConceptValues := []string{"value1", "value2", "value3"}

	// mix of concept ids and CustomDichotomousVariableDef items:
	conceptIdsAndCohortPairs := []interface{}{
		int64(1234),
		int64(5678),
		utils.CustomDichotomousVariableDef{
			CohortDefinitionId1: 1,
			CohortDefinitionId2: 2,
			ProvidedName:        "testA12"},
		int64(2090006880),
		utils.CustomDichotomousVariableDef{
			CohortDefinitionId1: 3,
			CohortDefinitionId2: 4,
			ProvidedName:        "testB34"},
	}

	result, _ := conceptController.GetAttritionRowForConceptIdsAndCohortPairs(sourceId, cohortId, conceptIdsAndCohortPairs, breakdownConceptId, sortedConceptValues)
	if len(result) != len(conceptIdsAndCohortPairs) {
		t.Errorf("Expected %d data lines, found %d lines in total",
			len(conceptIdsAndCohortPairs),
			len(result))
		t.Errorf("Lines: %s", result)
	}

	expectedLines := [][]string{
		{"Concept A", "10", "4", "6", "0"},
		{"Concept B", "9", "4", "5", "0"},
		{"testA12", "8", "3", "5", "0"},
		{"Concept C", "7", "3", "4", "0"},
		{"testB34", "6", "2", "4", "0"},
	}

	for i, expectedLine := range expectedLines {
		if !reflect.DeepEqual(expectedLine, result[i]) {
			t.Errorf("header or non filter row line not as expected. \nExpected: \n%s \nFound: \n%s",
				expectedLine, result[i])
		}
	}
}

func TestGenerateCompleteCSV(t *testing.T) {
	setUp(t)

	partialCsv := [][]string{
		{"sample.id", "ID_2000000324", "ID_2000006885", "ID_2000007027"},
		{"1", "F", "5.40", "HIS"},
		{"2", "A value with, comma!"},
	}

	personIdToCSVValues := map[int64]map[string]string{
		int64(1): {
			"ID_2_3": "1",
		},
		int64(2): {
			"ID_2_3": "NA",
		},
	}

	cohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: 2,
			CohortDefinitionId2: 3,
			ProvidedName:        "test"},
	}

	b := controllers.GenerateCompleteCSV(partialCsv, personIdToCSVValues, cohortPairs)
	csvLines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")

	expectedLines := []string{
		"sample.id,ID_2000000324,ID_2000006885,ID_2000007027,ID_2_3",
		"1,F,5.40,HIS,1",
		"2,\"A value with, comma!\",NA",
	}
	i := 0
	for _, expectedLine := range expectedLines {
		if !reflect.DeepEqual(expectedLine, csvLines[i]) {
			t.Errorf("header or non filter row line not as expected. \nExpected: \n%s \nFound: \n%s",
				expectedLine, csvLines[i])
		}
		i++
	}
}

func TestRetrievePeopleIdAndCohort(t *testing.T) {
	cohortId := 1
	cohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: 2,
			CohortDefinitionId2: 3,
			ProvidedName:        "test"},
	}

	cohortData := []*models.PersonConceptAndValue{
		{
			PersonId: 1,
		},
		{
			PersonId: 2,
		},
		{
			PersonId: 3,
		},
	}

	expectedResults := map[int64]map[string]string{
		int64(1): {
			"ID_2_3": "0",
		},
		int64(2): {
			"ID_2_3": "1",
		},
		int64(3): {
			"ID_2_3": "1",
		},
	}

	res, _ := cohortDataController.RetrievePeopleIdAndCohort(testSourceId, cohortId, cohortPairs, cohortData)
	for expectedPersonId, headerToCSVValue := range expectedResults {
		if res[expectedPersonId]["2_3"] != headerToCSVValue["2_3"] {
			t.Errorf("expected %v for csv value but instead got %v", headerToCSVValue["2_3"], res[expectedPersonId]["2_3"])
		}
	}
}

func TestRetrievePeopleIdAndCohortNonExistingCohortPair(t *testing.T) {
	cohortId := 1
	cohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: 4,
			CohortDefinitionId2: 5,
			ProvidedName:        "test"},
	}

	cohortData := []*models.PersonConceptAndValue{
		{
			PersonId: 1,
		},
		{
			PersonId: 2,
		},
		{
			PersonId: 3,
		},
	}

	expectedResults := map[int64]map[string]string{
		int64(1): {
			"ID_4_5": "NA",
		},
		int64(2): {
			"ID_4_5": "NA",
		},
		int64(3): {
			"ID_4_5": "NA",
		},
	}

	res, _ := cohortDataController.RetrievePeopleIdAndCohort(testSourceId, cohortId, cohortPairs, cohortData)
	for expectedPersonId, headerToCSVValue := range expectedResults {
		if res[expectedPersonId]["4_5"] != headerToCSVValue["4_5"] {
			t.Errorf("expected %v for csv value but instead got %v", headerToCSVValue["4_5"], res[expectedPersonId]["4_5"])
		}
	}
}

func TestRetrievePeopleIdAndCohortOverlappingCohortPair(t *testing.T) {
	cohortId := 1
	cohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: 1,
			CohortDefinitionId2: 1,
			ProvidedName:        "test"},
	}

	cohortData := []*models.PersonConceptAndValue{
		{
			PersonId: 1,
		},
		{
			PersonId: 2,
		},
		{
			PersonId: 3,
		},
	}

	expectedResults := map[int64]map[string]string{
		int64(1): {
			"ID_1_1": "NA",
		},
		int64(2): {
			"ID_1_1": "NA",
		},
		int64(3): {
			"ID_1_1": "NA",
		},
	}

	res, _ := cohortDataController.RetrievePeopleIdAndCohort(testSourceId, cohortId, cohortPairs, cohortData)
	for expectedPersonId, headerToCSVValue := range expectedResults {
		if res[expectedPersonId]["1_1"] != headerToCSVValue["1_1"] {
			t.Errorf("expected %v for csv value but instead got %v", headerToCSVValue["1_1"], res[expectedPersonId]["1_1"])
		}
	}
}

func TestRetrieveAttritionTable(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "breakdownconceptid", Value: "2"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestBody := "{\"variables\":[{\"variable_type\": \"custom_dichotomous\", \"provided_name\": \"testABC\", \"cohort_ids\": [1, 3]}," +
		"{\"variable_type\": \"concept\", \"concept_id\": 2090006880}," +
		"{\"variable_type\": \"custom_dichotomous\", \"cohort_ids\": [4, 5]}]}" // this one with no provided name (to test auto generated one)
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	requestContext.Writer = new(tests.CustomResponseWriter)
	conceptController.RetrieveAttritionTable(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	// check result vs expect result:
	csvLines := strings.Split(strings.TrimRight(result.CustomResponseWriterOut, "\n"), "\n")
	expectedLines := []string{
		"Cohort,Size,value1_name,value2_name",
		"dummy cohort name,13,5,8",
		"testABC,10,3,7",
		"Concept C,9,3,6",
		"ID_4_5,8,2,6",
	}
	i := 0
	for _, expectedLine := range expectedLines {
		if !reflect.DeepEqual(expectedLine, csvLines[i]) {
			t.Errorf("header or data row line not as expected. \nExpected: \n%s \nFound: \n%s",
				expectedLine, csvLines[i])
		}
		i++
	}

	// the same request should fail if the teamProject authorization fails:
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	conceptControllerWithFailingTeamProjectAuthz.RetrieveAttritionTable(requestContext)
	result = requestContext.Writer.(*tests.CustomResponseWriter)
	// expect error:
	if !strings.Contains(result.CustomResponseWriterOut, "access denied") {
		t.Errorf("Expected 'access denied' as result")
	}
	if !requestContext.IsAborted() {
		t.Errorf("Expected request to be aborted")
	}
}

func TestRetrieveDataDictionary(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	cohortDataController.RetrieveDataDictionary(requestContext)

	result := requestContext.Writer.(*tests.CustomResponseWriter)

	if result.StatusCode != 200 {
		t.Errorf("Expected request to succeed")
	}

}

func TestFailingRetrieveDataDictionary(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	cohortDataControllerWithFailingDataDictionary.RetrieveDataDictionary(requestContext)

	result := requestContext.Writer.(*tests.CustomResponseWriter)

	if result.StatusCode != 503 {
		t.Errorf("Expected request to Fail with 503")
	}

}
