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
	requestBody := "{\"variables\":[{\"variable_type\": \"concept\", \"prefixed_concept_id\": \"ID_2000000324\"},{\"variable_type\": \"concept\", \"prefixed_concept_id\": \"ID_2000000123\"},{\"variable_type\": \"custom_dichotomous\", \"cohort_ids\": [1, 3]}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))

	prefixedConceptIds, cohortPairs, _ := utils.ParsePrefixedConceptIdsAndDichotomousIds(requestContext)
	if requestContext.IsAborted() {
		t.Errorf("Did not expect this request to abort")
	}

	expectedPrefixedConceptIds := []string{"ID_2000000324", "ID_2000000123"}
	if !reflect.DeepEqual(prefixedConceptIds, expectedPrefixedConceptIds) {
		t.Errorf("Expected %s but found %s", expectedPrefixedConceptIds, prefixedConceptIds)
	}

	expectedCohortPairs := [][]int{
		{1, 3},
	}

	for i, cohortPair := range cohortPairs {
		if !reflect.DeepEqual(cohortPair, expectedCohortPairs[i]) {
			t.Errorf("Cohort Pairs not as expected. \nExpected: \n%v \nFound: \n%v",
				expectedCohortPairs[i], cohortPair)
		}

	}

}
