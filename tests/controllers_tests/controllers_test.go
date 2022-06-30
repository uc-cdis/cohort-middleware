package controllers_tests

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/controllers"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/tests"
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

	// ensure tearDown is called when test "t" is done:
	t.Cleanup(func() {
		tearDown()
	})
}

func tearDown() {
	log.Println("teardown for test")
}

var cohortDataController = controllers.NewCohortDataController(*new(dummyCohortDataModel))

// instance of the controller that talks to the regular model implementation (that needs a real DB):
var cohortDefinitionControllerNeedsDb = controllers.NewCohortDefinitionController(*new(models.CohortDefinition))

// instance of the controller that talks to a mock implementation of the model:
var cohortDefinitionController = controllers.NewCohortDefinitionController(*new(dummyCohortDefinitionDataModel))

var conceptController = controllers.NewConceptController(*new(dummyConceptDataModel))

type dummyCohortDataModel struct{}

func (h dummyCohortDataModel) RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*models.PersonConceptAndValue, error) {
	cohortData := []*models.PersonConceptAndValue{
		{PersonId: 1, ConceptId: 10, ConceptValueAsString: "abc", ConceptValueAsNumber: 0.0},
		{PersonId: 1, ConceptId: 22, ConceptValueAsString: "", ConceptValueAsNumber: 1.5},
		{PersonId: 2, ConceptId: 10, ConceptValueAsString: "A value with, comma!", ConceptValueAsNumber: 0.0},
	}
	return cohortData, nil
}

func (h dummyCohortDataModel) RetrieveCohortOverlapStats(sourceId int, caseCohortId int, controlCohortId int,
	filterConceptId int64, filterConceptValue int64, otherFilterConceptIds []int64) (models.CohortOverlapStats, error) {
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

func (h dummyCohortDefinitionDataModel) GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int) ([]*models.CohortDefinitionStats, error) {
	cohortDefinitionStats := []*models.CohortDefinitionStats{
		{Id: 1, CohortSize: 10, Name: "name1"},
		{Id: 2, CohortSize: 22, Name: "name2"},
		{Id: 3, CohortSize: 33, Name: "name3"},
	}
	return cohortDefinitionStats, nil
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

type dummyConceptDataModel struct{}

func (h dummyConceptDataModel) RetriveAllBySourceId(sourceId int) ([]*models.Concept, error) {
	return nil, nil
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
func (h dummyConceptDataModel) RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*models.ConceptStats, error) {
	// dummy data with _some_ of the relevant fields:
	conceptStats := []*models.ConceptStats{
		{ConceptId: 1234, CohortSize: 11, NmissingRatio: 0.56},
		{ConceptId: 4567, CohortSize: 22, NmissingRatio: 0.67},
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("fake model error!")
	}
	return conceptStats, nil
}
func (h dummyConceptDataModel) RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId int, cohortDefinitionId int, breakdownConceptId int64) ([]*models.ConceptBreakdown, error) {
	conceptBreakdown := []*models.ConceptBreakdown{
		{ConceptValue: "value1", NpersonsInCohortWithValue: 5},
		{ConceptValue: "value2", NpersonsInCohortWithValue: 8},
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("error!")
	}
	return conceptBreakdown, nil
}
func (h dummyConceptDataModel) RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, filterConceptIds []int64, breakdownConceptId int64) ([]*models.ConceptBreakdown, error) {
	conceptBreakdown := []*models.ConceptBreakdown{
		{ConceptValue: "value1", NpersonsInCohortWithValue: 5},
		{ConceptValue: "value2", NpersonsInCohortWithValue: 8},
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("error!")
	}
	return conceptBreakdown, nil
}

func TestRetrieveDataBySourceIdAndCohortIdAndConceptIdsWrongParams(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "Abc", Value: "def"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDataController.RetrieveDataBySourceIdAndCohortIdAndConceptIds(requestContext)
	// Params above are wrong, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
}

func TestRetrieveDataBySourceIdAndCohortIdAndConceptIdsCorrectParams(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "1"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestBody := "{\"variables\":[{\"variable_type\": \"concept\", \"prefixed_concept_id\": \"ID_2000000324\"},{\"variable_type\": \"custom_dichotomous\", \"cohort_ids\": [1, 3]}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))
	cohortDataController.RetrieveDataBySourceIdAndCohortIdAndConceptIds(requestContext)
	// Params above are correct, so request should NOT abort:
	if requestContext.IsAborted() {
		t.Errorf("Did not expect this request to abort")
	}
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !strings.Contains(result.CustomResponseWriterOut, "sample.id,") {
		t.Errorf("Expected output starting with 'sample.id,...'")
	}
}

func TestRetrieveCohortOverlapStats(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "filterconceptid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "filtervalue", Value: "123"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "casecohortid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "controlcohortid", Value: "2"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"ConceptIds\":[2000000324,2000006885]}"))

	cohortDataController.RetrieveCohortOverlapStats(requestContext)
	// Params above are correct, so request should NOT abort:
	if requestContext.IsAborted() {
		t.Errorf("Did not expect this request to abort")
	}
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	if !strings.Contains(result.CustomResponseWriterOut, "case_control_overlap_after_filter") {
		t.Errorf("Expected output containing 'case_control_overlap_after_filter...'")
	}
}

func TestRetrieveCohortOverlapStatsBadRequest(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "filterconceptidwrong", Value: "1"})
	requestContext.Writer = new(tests.CustomResponseWriter)

	cohortDataController.RetrieveCohortOverlapStats(requestContext)
	// Params above are incorrect, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("Expected this request to abort")
	}
}

func TestGenerateCSV(t *testing.T) {
	setUp(t)
	cohortData := []*models.PersonConceptAndValue{
		{PersonId: 1, ConceptId: 10, ConceptValueAsString: "abc", ConceptValueAsNumber: 0.0},
		{PersonId: 1, ConceptId: 22, ConceptValueAsString: "", ConceptValueAsNumber: 1.5},
		{PersonId: 2789580123456, ConceptId: 10, ConceptValueAsString: "A value with, comma!", ConceptValueAsNumber: 0.0},
	}
	conceptIds := []int64{10, 22}

	csvLines := controllers.GeneratePartialCSV(
		testSourceId, cohortData, conceptIds)

	// the above should result in one header line and 2 data lines (2 persons)
	if len(csvLines) != 3 {
		t.Errorf("Expected 1 header line + 2 data lines, found %d lines in total",
			len(csvLines))
		t.Errorf("Lines: %s", csvLines)
	}
	expectedLines := [][]string{
		{"sample.id", "ID_10", "ID_22"},
		{"1", "abc", "1.50"},
		{"2789580123456", "A value with, comma!", "NA"},
	}

	for i, expectedLine := range expectedLines {
		if !reflect.DeepEqual(expectedLine, csvLines[i]) {
			t.Errorf("CSV line not as expected. \nExpected: \n%s \nFound: \n%s",
				expectedLine, csvLines[i])
		}

	}
}

func TestRetriveStatsBySourceIdWrongParams(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "Abc", Value: "def"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionController.RetriveStatsBySourceId(requestContext)
	// Params above are wrong, so request should abort:
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
}

func TestRetriveStatsBySourceIdDbPanic(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Writer = new(tests.CustomResponseWriter)

	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
			if err != "AtlasDB not initialized" {
				t.Errorf("Expected error")
			}
		}
	}()
	cohortDefinitionControllerNeedsDb.RetriveStatsBySourceId(requestContext)
	t.Errorf("Expected error")
}

func TestRetriveStatsBySourceId(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: strconv.Itoa(tests.GetTestSourceId())})
	requestContext.Writer = new(tests.CustomResponseWriter)
	cohortDefinitionController.RetriveStatsBySourceId(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	log.Printf("result: %s", result)
	// expect result with all of the dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "name1") ||
		!strings.Contains(result.CustomResponseWriterOut, "name2") ||
		!strings.Contains(result.CustomResponseWriterOut, "name3") {
		t.Errorf("Expected 3 rows in result")
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
	log.Printf("result: %s", result)
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

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "breakdownconceptid", Value: "1"})
	requestContext.Request = new(http.Request)
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"ConceptIds\":[1234,5678]}"))
	requestContext.Writer = new(tests.CustomResponseWriter)
	conceptController.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	log.Printf("result: %s", result)
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "persons_in_cohort_with_value") {
		t.Errorf("Expected data in result")
	}
}

func TestRetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsModelError(t *testing.T) {
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
	conceptController.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(requestContext)
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
	log.Printf("result: %s", result)
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
	log.Printf("result: %s", result)
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
	log.Printf("result: %s", result)
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
	log.Printf("result: %s", result)
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
	log.Printf("result: %s", result)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
	if !strings.Contains(result.CustomResponseWriterOut, "no request body") {
		t.Errorf("Expected 'no request body' error message")
	}
}

func TestRetrieveStatsBySourceIdAndCohortIdAndConceptIds(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "cohortid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "breakdownconceptid", Value: "1"})
	requestContext.Request = new(http.Request)
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"ConceptIds\":[1234,5678]}"))
	requestContext.Writer = new(tests.CustomResponseWriter)
	conceptController.RetrieveStatsBySourceIdAndCohortIdAndConceptIds(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	log.Printf("result: %s", result)
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "n_missing_ratio") ||
		!strings.Contains(result.CustomResponseWriterOut, "cohort_size") {
		t.Errorf("Expected data not found in result")
	}
}

func TestRetrieveStatsBySourceIdAndCohortIdAndConceptIdsModelError(t *testing.T) {
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
	conceptController.RetrieveStatsBySourceIdAndCohortIdAndConceptIds(requestContext)
	if !requestContext.IsAborted() {
		t.Errorf("Expected aborted request")
	}
	// error message:
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	log.Printf("result: %s", result)
	// expect specific error message:
	if !strings.Contains(result.CustomResponseWriterOut, "Error retrieving concept details") {
		t.Errorf("Expected error did not occur...other error occurred instead")
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
	sourceId := 1
	cohortId := 1
	var breakdownConceptId int64 = 1
	cohortName := "hello"

	result, _ := conceptController.GenerateHeaderAndNonFilteredRow(cohortName, sourceId, cohortId, breakdownConceptId)
	if len(result) != 2 {
		t.Errorf("Expected 1 header line + 1 data lines, found %d lines in total",
			len(result))
		t.Errorf("Lines: %s", result)
	}

	expectedLines := [][]string{
		{"Cohort", "Size", "value1", "value2"},
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

func TestGetFilteredConceptRows(t *testing.T) {
	setUp(t)
	sourceId := 1
	cohortId := 1
	var breakdownConceptId int64 = 1
	conceptIds := []int64{1234, 5678, 2090006880}
	sortedConceptValues := []string{"value1", "value2"}

	result, _ := conceptController.GetFilteredConceptRows(sourceId, cohortId, conceptIds, breakdownConceptId, sortedConceptValues)
	if len(result) != 3 {
		t.Errorf("Expected 3 data lines, found %d lines in total",
			len(result))
		t.Errorf("Lines: %s", result)
	}

	expectedLines := [][]string{
		{"Concept A", "13", "5", "8"},
		{"Concept B", "13", "5", "8"},
		{"Concept C", "13", "5", "8"},
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

func TestGenerateCompleteCSV(t *testing.T) {
	setUp(t)

	partialCsv := [][]string{
		{"sample.id", "ID_2000000324", "ID_2000006885", "ID_2000007027"},
		{"1", "F", "5.40", "HIS"},
		{"2", "A value with, comma!"},
	}

	personIdToCSVValues := map[int64]map[string]string{
		int64(1): {
			"2_3": "1",
		},
		int64(2): {
			"2_3": "NA",
		},
	}

	cohortPairs := [][]int{
		{2, 3},
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
	cohortPairs := [][]int{
		{2, 3},
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
			"2_3": "0",
		},
		int64(2): {
			"2_3": "1",
		},
		int64(3): {
			"2_3": "1",
		},
	}

	res, _ := cohortDataController.RetrievePeopleIdAndCohort(testSourceId, cohortId, cohortPairs, cohortData)
	for expectedPersonId, headerToCSVValue := range expectedResults {
		if res[expectedPersonId]["2_3"] != headerToCSVValue["2_3"] {
			t.Errorf("expected %v for csv value but instead got %v", headerToCSVValue["2_3"], res[expectedPersonId]["2_3"])
		}
	}
}
