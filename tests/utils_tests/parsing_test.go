package parsing_test

import (
	"io"
	"log"
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
			CohortId1:    1,
			CohortId2:    3,
			ProvidedName: "test"},
	}

	for i, cohortPair := range cohortPairs {
		if !reflect.DeepEqual(cohortPair, expectedCohortPairs[i]) {
			t.Errorf("Cohort Pairs not as expected. \nExpected: \n%v \nFound: \n%v",
				expectedCohortPairs[i], cohortPair)
		}

	}

}
