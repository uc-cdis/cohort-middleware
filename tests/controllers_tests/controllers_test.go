package controllers_tests

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func (h dummyCohortDataModel) RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int) ([]*models.PersonConceptAndValue, error) {
	cohortData := []*models.PersonConceptAndValue{
		{PersonId: 1, ConceptId: 10, ConceptValueAsString: "abc", ConceptValueAsNumber: 0.0},
		{PersonId: 1, ConceptId: 22, ConceptValueAsString: "", ConceptValueAsNumber: 1.5},
		{PersonId: 2, ConceptId: 10, ConceptValueAsString: "A value with, comma!", ConceptValueAsNumber: 0.0},
	}
	return cohortData, nil
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
		return nil, fmt.Errorf("error!")
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
func (h dummyConceptDataModel) RetrieveInfoBySourceIdAndConceptIds(sourceId int, conceptIds []int) ([]*models.ConceptSimple, error) {
	// dummy data with _some_ of the relevant fields:
	conceptSimple := []*models.ConceptSimple{
		{ConceptId: 1234, ConceptName: "Concept A"},
		{ConceptId: 5678, ConceptName: "Concept B"},
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("error!")
	}
	return conceptSimple, nil
}
func (h dummyConceptDataModel) RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int) ([]*models.ConceptStats, error) {
	// dummy data with _some_ of the relevant fields:
	conceptStats := []*models.ConceptStats{
		{ConceptId: 1234, CohortSize: 11, NmissingRatio: 0.56},
		{ConceptId: 4567, CohortSize: 22, NmissingRatio: 0.67},
	}
	if dummyModelReturnError {
		return nil, fmt.Errorf("error!")
	}
	return conceptStats, nil
}
func (h dummyConceptDataModel) RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId int, cohortDefinitionId int, breakdownConceptId int) ([]*models.ConceptBreakdown, error) {
	return nil, nil
}
func (h dummyConceptDataModel) RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, filterConceptIds []int, breakdownConceptId int) ([]*models.ConceptBreakdown, error) {
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
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"PrefixedConceptIds\":[\"ID_2000000324\",\"ID_2000006885\"]}"))
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
func TestGenerateCSV(t *testing.T) {
	setUp(t)
	cohortData := []*models.PersonConceptAndValue{
		{PersonId: 1, ConceptId: 10, ConceptValueAsString: "abc", ConceptValueAsNumber: 0.0},
		{PersonId: 1, ConceptId: 22, ConceptValueAsString: "", ConceptValueAsNumber: 1.5},
		{PersonId: 2, ConceptId: 10, ConceptValueAsString: "A value with, comma!", ConceptValueAsNumber: 0.0},
	}
	conceptIds := []int{10, 22}

	b := controllers.GenerateCSV(
		testSourceId, cohortData, conceptIds)
	csvLines := strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
	// the above should result in one header line and 2 data lines (2 persons)
	if len(csvLines) != 3 {
		t.Errorf("Expected 1 header line + 2 data lines, found %d lines in total",
			len(csvLines))
		t.Errorf("Lines: %s", csvLines)
	}
	expectedLines := []string{
		"sample.id,ID_10,ID_22",
		"1,abc,1.50",
		"2,\"A value with, comma!\",NA",
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

func TestRetrieveInfoBySourceIdAndCohortIdAndConceptIds(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "sourceid", Value: "1"})
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "breakdownconceptid", Value: "1"})
	requestContext.Request = new(http.Request)
	requestContext.Request.Body = io.NopCloser(strings.NewReader("{\"ConceptIds\":[1234,5678]}"))
	requestContext.Writer = new(tests.CustomResponseWriter)
	conceptController.RetrieveInfoBySourceIdAndCohortIdAndConceptIds(requestContext)
	result := requestContext.Writer.(*tests.CustomResponseWriter)
	log.Printf("result: %s", result)
	// expect result with dummy data:
	if !strings.Contains(result.CustomResponseWriterOut, "Concept A") ||
		!strings.Contains(result.CustomResponseWriterOut, "Concept B") {
		t.Errorf("Expected data in result")
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
