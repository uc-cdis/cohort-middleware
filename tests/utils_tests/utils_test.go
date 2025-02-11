package utils_tests

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"reflect"
	"sort"
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
		"{\"variable_type\": \"concept\", \"concept_id\": 2000000123, \"filters\": [{\"type\": \"in\", \"values_as_concept_ids\": [2000000237, 2000000238]}]}," +
		"{\"variable_type\": \"custom_dichotomous\", \"provided_name\": \"test\", \"cohort_ids\": [1, 3]}]}"
	requestContext.Request.Body = io.NopCloser(strings.NewReader(requestBody))

	conceptDefs, cohortPairs, _ := utils.ParseConceptDefsAndDichotomousDefs(requestContext)
	if requestContext.IsAborted() {
		t.Errorf("Did not expect this request to abort")
	}

	expectedPrefixedConceptDefs := []utils.CustomConceptVariableDef{
		{ConceptId: 2000000324},
		{ConceptId: 2000000123,
			Filters: []utils.Filter{
				{
					Type:               "in",
					ValuesAsConceptIds: []int64{2000000237, 2000000238},
				},
			},
		},
	}

	if !reflect.DeepEqual(conceptDefs, expectedPrefixedConceptDefs) {
		t.Errorf("Expected %v but found %v", expectedPrefixedConceptDefs, conceptDefs)
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

func TestIntersect(t *testing.T) {
	setUp(t)
	var expectedResult = []int{
		1234,
		5678,
	}
	var list1 = []int{
		567,
		1234,
		444,
		5678,
		29999999978,
		5678,
	}
	var list2 = []int{
		111,
		1234,
		2222,
		5678,
		2999999997,
	}
	result := utils.Intersect(list1, list2)
	sort.Ints(result)
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %v but found %v", expectedResult, result)
	}
	result = utils.Intersect(list2, list1)
	sort.Ints(result)
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %v but found %v", expectedResult, result)
	}

	list1 = []int{
		567,
		1234,
	}
	list2 = []int{
		111,
		222,
	}
	result = utils.Intersect(list1, list2)
	if len(result) > 0 {
		t.Errorf("Expected [] but found %v", result)
	}

	list1 = []int{}
	list2 = []int{
		111,
		222,
	}
	result = utils.Intersect(list1, list2)
	if len(result) > 0 {
		t.Errorf("Expected [] but found %v", result)
	}

	list1 = []int{
		111,
		222,
	}
	list2 = []int{}
	result = utils.Intersect(list1, list2)
	if len(result) > 0 {
		t.Errorf("Expected [] but found %v", result)
	}

	list1 = []int{}
	list2 = []int{}
	result = utils.Intersect(list1, list2)
	if len(result) > 0 {
		t.Errorf("Expected [] but found %v", result)
	}
}

func TestSubtract(t *testing.T) {
	setUp(t)
	var expectedResult = []int{
		567,
		444,
		29999999978,
	}
	var list1 = []int{
		567,
		1234,
		444,
		5678,
		29999999978,
		5678,
	}
	var list2 = []int{
		111,
		1234,
		2222,
		5678,
		2999999997,
	}
	result := utils.Subtract(list1, list2)
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %v but found %v", expectedResult, result)
	}

	list1 = []int{
		567,
		1234,
	}
	list2 = []int{
		111,
		222,
	}
	result = utils.Subtract(list1, list2)
	if !reflect.DeepEqual(list1, result) {
		t.Errorf("Expected %v but found %v", list1, result)
	}

	list1 = []int{}
	list2 = []int{
		111,
		222,
	}
	result = utils.Subtract(list1, list2)
	if len(result) > 0 {
		t.Errorf("Expected [] but found %v", result)
	}

	list1 = []int{
		111,
		222,
	}
	list2 = []int{}
	result = utils.Subtract(list1, list2)
	if !reflect.DeepEqual(list1, result) {
		t.Errorf("Expected %v but found %v", list1, result)
	}

	list1 = []int{}
	list2 = []int{}
	result = utils.Subtract(list1, list2)
	if len(result) > 0 {
		t.Errorf("Expected [] but found %v", result)
	}
}

func TestGenerateStatsData(t *testing.T) {
	setUp(t)

	var emptyData = []float64{}
	result := utils.GenerateStatsData(1, 1, emptyData)
	if result != nil {
		t.Errorf("Expected a nil result for an empty data set")
	}

	var expectedResult = &utils.ConceptStats{CohortId: 1, ConceptId: 1, NumberOfPeople: 11, Min: 6.0, Max: 49.0, Avg: 33.18181818181818, Sd: 15.134657288477642}
	result = utils.GenerateStatsData(1, 1, testData)
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %v but found %v", expectedResult, result)
	}
}

func TestConvertConceptIdToCustomConceptVariablesDef(t *testing.T) {
	setUp(t)

	expectedResult := []utils.CustomConceptVariableDef{{ConceptId: 1234}, {ConceptId: 5678}}
	conceptIds := []int64{1234, 5678}
	result := utils.ConvertConceptIdToCustomConceptVariablesDef(conceptIds)

	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %v but found %v", expectedResult, result)
	}
}

func TestExtractConceptIdsFromCustomConceptVariablesDef(t *testing.T) {
	setUp(t)

	var testData = []utils.CustomConceptVariableDef{
		{ConceptId: 1234,
			Filters: []utils.Filter{
				{
					Type:               "in",
					ValuesAsConceptIds: []int64{7890},
				},
			},
		},
		{ConceptId: 5678},
	}
	expectedResult := []int64{1234, 5678}
	result := utils.ExtractConceptIdsFromCustomConceptVariablesDef(testData)
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %v but found %v", expectedResult, result)
	}
}

func TestCheckAndGetLastCustomConceptVariableDef(t *testing.T) {
	setUp(t)
	testData := []interface{}{
		utils.CustomDichotomousVariableDef{
			CohortDefinitionId1: 123,
			CohortDefinitionId2: 456,
			ProvidedName:        "test",
		},
		utils.CustomConceptVariableDef{
			ConceptId: 1234,
		},
	}

	result, err := utils.CheckAndGetLastCustomConceptVariableDef(testData)

	if result == nil || err != nil {
		t.Errorf("Expected a valid result and no error. Result: %v, Error: %v", result, err)
	}

	testData = []interface{}{
		utils.CustomConceptVariableDef{
			ConceptId: 1234,
		},
		utils.CustomDichotomousVariableDef{
			CohortDefinitionId1: 123,
			CohortDefinitionId2: 456,
			ProvidedName:        "test",
		},
	}

	result, err = utils.CheckAndGetLastCustomConceptVariableDef(testData)

	if result != nil || err == nil {
		t.Errorf("Expected NO result and an error. Result: %v, Error: %v", result, err)
	}

	testData = []interface{}{}

	result, err = utils.CheckAndGetLastCustomConceptVariableDef(testData)

	if result != nil || err == nil {
		t.Errorf("Expected NO result and an error. Result: %v, Error: %v", result, err)
	}

}

func TestTempTableCache(t *testing.T) {
	setUp(t)
	var tempTableCache = &utils.CacheSimple{
		Data:    make(map[string]interface{}),
		MaxSize: 3,
	}
	tempTableCache.Set("abc_key", "abc_value")
	value, exists := tempTableCache.Get("abc_key")
	if value == nil || !exists || value != "abc_value" {
		t.Errorf("Expected abc_value but got: exists=%v, value=%s", exists, value)
	}

	// The cache has maxsize 3.
	// Scenario 1: After adding 3 new items, the .Get("abc_key") should result in empty,
	// as it will have been removed.
	// Scenario 2: The cache will keep items that were re-set (.Set) or recently retrieved (with .Get) "warm",
	// meaning that they get moved to the end of the queue of who gets dropped next from cache. So, in
	// this scenario, we want to add 2 new items, then .Get("abc_key"), then add one more item, and
	// assert that  "abc_key" is still in, while the older of the 2 new items got dropped.

	// (Scenario 1) Add 3 new items; "abc_key" should be removed
	tempTableCache.Set("key1", "value1")
	tempTableCache.Set("key2", "value2")
	tempTableCache.Set("key3", "value3")
	value, exists = tempTableCache.Get("abc_key")
	if exists {
		t.Errorf("Expected abc_key to be evicted, but it still exists with value: %v", value)
	}

	// (Scenario 2) Keep "abc_key" warm and verify eviction order
	tempTableCache.Set("abc_key", "abc_value")
	tempTableCache.Set("key4", "value4")
	tempTableCache.Set("key5", "value5")
	_, _ = tempTableCache.Get("abc_key") // Access abc_key to keep it warm
	tempTableCache.Set("key6", "value6") // Since abc_key is warm, key4 should be evicted at this point
	value, exists = tempTableCache.Get("abc_key")
	if !exists || value != "abc_value" {
		t.Errorf("Expected abc_key to be retained, but it was evicted")
	}
	value, exists = tempTableCache.Get("key4")
	if exists {
		t.Errorf("Expected key4 to be evicted, but it still exists with value: %v", value)
	}

	// Scenario 3: like scenario 2, but with Set. The .Set method
	// should be repeatable and also keeps the entry "warm"
	// (Scenario 2) Keep "abc_key" warm and verify eviction order
	tempTableCache.Set("abc_key", "abc_value")
	tempTableCache.Set("key4", "value4")
	tempTableCache.Set("key5", "value5")
	tempTableCache.Set("abc_key", "new_value1") // Reset abc_key to keep it warm, updating its value at the same time
	tempTableCache.Set("abc_key", "new_value2") // Reset abc_key to keep it warm, updating its value at the same time
	tempTableCache.Set("abc_key", "new_value3") // Reset abc_key to keep it warm, updating its value at the same time
	tempTableCache.Set("key6", "value6")        // Since abc_key is warm, key4 should be evicted at this point
	value, exists = tempTableCache.Get("abc_key")
	if !exists || value != "new_value3" {
		t.Errorf("Expected abc_key to be retained, but it was evicted, or expected value does not match. Value: %s", value)
	}
	value, exists = tempTableCache.Get("key4")
	if exists {
		t.Errorf("Expected key4 to be evicted, but it still exists with value: %v", value)
	}
	value, exists = tempTableCache.Get("key5")
	if !exists || value != "value5" {
		t.Errorf("Expected abc_key to be retained, but it was evicted, or expected value does not match. Value: %s", value)
	}
}
