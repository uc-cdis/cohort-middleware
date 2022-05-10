package controllers_tests

import (
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

	// ensure tearDown is called when test "t" is done:
	t.Cleanup(func() {
		tearDown()
	})
}

func tearDown() {
	log.Println("teardown for test")
}

var cohortDataController = controllers.NewCohortDataController(*new(dummyCohortDataModel))
var cohortDefinitionControllerNoDb = controllers.NewCohortDefinitionController(*new(models.CohortDefinition))
var cohortDefinitionController = controllers.NewCohortDefinitionController(*new(dummyCohortDefinitionDataModel))

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

func (h dummyCohortDefinitionDataModel) GetAllCohortDefinitionsAndStats(sourceId int) ([]*models.CohortDefinitionStats, error) {
	cohortDefinitionStats := []*models.CohortDefinitionStats{
		{Id: 1, CohortSize: 10, Name: "name1"},
		{Id: 2, CohortSize: 22, Name: "name2"},
		{Id: 3, CohortSize: 33, Name: "name3"},
	}
	return cohortDefinitionStats, nil
}
func (h dummyCohortDefinitionDataModel) GetCohortDefinitionById(id int) (*models.CohortDefinition, error) {
	return nil, nil
}
func (h dummyCohortDefinitionDataModel) GetCohortDefinitionByName(name string) (*models.CohortDefinition, error) {
	return nil, nil
}
func (h dummyCohortDefinitionDataModel) GetAllCohortDefinitions() ([]*models.CohortDefinition, error) {
	return nil, nil
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
	cohortDefinitionControllerNoDb.RetriveStatsBySourceId(requestContext)
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
