package utils_tests

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/tests"
	"github.com/uc-cdis/cohort-middleware/utils"
)

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

func TestParsePrefixedConceptIdsAndDichotomousIds(t *testing.T) {
	setUp(t)
	requestContext := new(gin.Context)
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestBody := "{\"variables\":[{\"variable_type\": \"concept\", \"concept_id\": 2000000324}," +
		"{\"variable_type\": \"concept\", \"concept_id\": 2000000123}," +
		"{\"variable_type\": \"custom_dichotomous\", \"provided_name\": \"test\", \"cohort_ids\": [1, 3]}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))

	conceptIds, cohortPairs, _ := utils.ParseConceptIdsAndDichotomousDefs(requestContext)
	if requestContext.IsAborted() {
		t.Errorf("Did not expect this request to abort")
	}

	expectedPrefixedConceptIds := []int64{2000000324, 2000000123}
	if !reflect.DeepEqual(conceptIds, expectedPrefixedConceptIds) {
		t.Errorf("Expected %d but found %d", expectedPrefixedConceptIds, conceptIds)
	}

	expectedCohortPairs := []utils.CustomDichotomousVariableDef{
		{
			CohortDefinitionId1: 1,
			CohortDefinitionId2: 3,
			ProvidedName:        "test"},
	}

	for i, cohortPair := range cohortPairs {
		if !reflect.DeepEqual(cohortPair, expectedCohortPairs[i]) {
			t.Errorf("Cohort Pairs not as expected. \nExpected: \n%v \nFound: \n%v",
				expectedCohortPairs[i], cohortPair)
		}

	}

}

var testData = []float64{
	47.0,
	6.0,
	49.0,
	15.0,
	42.0,
	41.0,
	7.0,
	39.0,
	43.0,
	40.0,
	36.0,
}

var testData2 = []float64{
	1,
	10,
	10,
	10,
	10,
	10,
	10,
	10,
	10,
	10,
	10,
	10,
	10,
	10,
	10,
}

func TestIQR(t *testing.T) {
	setUp(t)
	expectedResult := float64(28.0)
	result := utils.IQR(testData)
	if result != expectedResult {
		t.Errorf("IQR is incorrect, expected %v but got %v", expectedResult, result)
	}
}

func TestFreedmanDiaconis(t *testing.T) {
	setUp(t)
	expectedResult := float64(25.18)
	result := utils.FreedmanDiaconis(testData)
	result = math.Floor(result*100) / 100
	if result != expectedResult {
		t.Errorf("Freedman Diaconis value is incorrect, expected %v but got %v", expectedResult, result)
	}
}

func TestGenerateHistogramData(t *testing.T) {
	setUp(t)
	expectedresult := `[{"start":6,"end":31.18008152926611,"personCount":3},{"start":31.18008152926611,"end":56.36016305853222,"personCount":8}]`
	resultArray := utils.GenerateHistogramData(testData)
	resultJson, _ := json.Marshal(resultArray)
	resultString := string(resultJson)

	if resultString != expectedresult {
		t.Errorf("expected %v for histogram but got %v", expectedresult, resultString)
	}
}

func TestGenerateHistogramDataSingleBin(t *testing.T) {
	// Tests whether we get a single bin that includes all persons if data has no variation in Q1 Q3
	setUp(t)
	expectedresult := `[{"start":1,"end":11,"personCount":15}]`
	resultArray := utils.GenerateHistogramData(testData2)
	resultJson, _ := json.Marshal(resultArray)
	resultString := string(resultJson)

	if resultString != expectedresult {
		t.Errorf("expected %v for histogram but got %v", expectedresult, resultString)
	}
}

func TestSliceAtoi(t *testing.T) {
	setUp(t)
	var expectedResult = []int64{
		1234,
		5678,
		2999999997,
	}
	result, _ := utils.SliceAtoi([]string{"1234", "5678", "2999999997"})
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %v but found %v", expectedResult, result)
	}

	_, error := utils.SliceAtoi([]string{"1234", "5678abc", "2999999997"})
	if error == nil {
		t.Errorf("Expected an error")
	}
}
