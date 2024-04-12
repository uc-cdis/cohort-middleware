package middlewares_tests

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/middlewares"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/tests"
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

func TestPrepareNewArboristRequest(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "Authorization", Value: "dummy_token_value"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	u, _ := url.Parse("https://some-cohort-middl-server/api/abc/123")
	requestContext.Request.URL = u
	resultArboristRequest, error := middlewares.PrepareNewArboristRequest(requestContext)

	expectedResult := "resource=/cohort-middleware/api/abc/123&service=cohort-middleware&method=access"
	// check if expected result URL was produced:
	if error != nil || resultArboristRequest.URL.RawQuery != expectedResult {
		t.Errorf("Unexpected error or resource query is not as expected")
	}
}

func TestPrepareNewArboristRequestMissingToken(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	requestContext := new(gin.Context)
	requestContext.Params = append(requestContext.Params, gin.Param{Key: "Abc", Value: "def"})
	requestContext.Writer = new(tests.CustomResponseWriter)
	requestContext.Request = new(http.Request)
	u, _ := url.Parse("https://some-cohort-middl-server/api/abc/123")
	requestContext.Request.URL = u
	_, error := middlewares.PrepareNewArboristRequest(requestContext)

	// Params above are wrong, so request should abort:
	if error.Error() != "missing Authorization header" {
		t.Errorf("Expected error")
	}
}

type dummyHttpClient struct {
	statusCode int
	nrCalls    int
}

func (h *dummyHttpClient) Do(req *http.Request) (*http.Response, error) {
	h.nrCalls++
	return &http.Response{StatusCode: h.statusCode}, nil
}

type dummyCohortDefinitionDataModel struct {
	returnForGetCohortDefinitionIdsForTeamProject []int
}

func (h dummyCohortDefinitionDataModel) GetCohortDefinitionIdsForTeamProject(teamProject string) ([]int, error) {
	return h.returnForGetCohortDefinitionIdsForTeamProject, nil
}

func (h dummyCohortDefinitionDataModel) GetTeamProjectsThatMatchAllCohortDefinitionIds(uniqueCohortDefinitionIdsList []int) ([]string, error) {
	// dummy switch just to support three test scenarios:
	if len(uniqueCohortDefinitionIdsList) == 0 {
		return []string{}, nil
	} else if uniqueCohortDefinitionIdsList[0] == 0 { // simulate issue
		return nil, nil
	} else if len(uniqueCohortDefinitionIdsList) == 1 {
		return []string{"teamProject1"}, nil
	} else {
		return []string{"teamProject1", "teamProject2"}, nil
	}
}

func (h dummyCohortDefinitionDataModel) GetCohortName(cohortId int) (string, error) {
	return "dummy cohort name", nil
}

func (h dummyCohortDefinitionDataModel) GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int, teamProject string) ([]*models.CohortDefinitionStats, error) {
	return nil, nil
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

func TestTeamProjectValidationForCohort(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidationForCohort(requestContext, 1)
	if result == false {
		t.Errorf("Expected TeamProjectValidationForCohort result to be 'true'")
	}
	if dummyHttpClient.nrCalls != 1 {
		t.Errorf("Expected dummyHttpClient to have been only once")
	}
}

func TestTeamProjectValidationForCohortArborist401(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 401
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidationForCohort(requestContext, 1)
	if result == true {
		t.Errorf("Expected TeamProjectValidationForCohort result to be 'false'")
	}
	if dummyHttpClient.nrCalls != 1 {
		t.Errorf("Expected dummyHttpClient to have been only once")
	}
}

func TestTeamProjectValidationNoGlobalCohorts(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	globalCohorts := []int{}
	teamCohorts := []int{1, 2}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts},
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, teamCohorts, nil)
	if result == false {
		t.Errorf("Expected TeamProjectValidation result to be 'true'")
	}
	if dummyHttpClient.nrCalls != 1 {
		t.Errorf("Expected dummyHttpClient to have been called only once")
	}
}

func TestTeamProjectValidationFullOverlapWithGlobalCohorts(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	globalCohorts := []int{1, 2}
	teamCohorts := []int{1, 2}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts},
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, teamCohorts, nil)
	if result == false {
		t.Errorf("Expected TeamProjectValidation result to be 'true'")
	}
	if dummyHttpClient.nrCalls != 1 {
		t.Errorf("Expected dummyHttpClient to have been called only once")
	}
}

func TestTeamProjectValidationOnlyGlobalCohorts(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	globalCohorts := []int{1, 2}
	teamCohorts := []int{}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts},
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, teamCohorts, nil)
	if result == true {
		t.Errorf("Expected TeamProjectValidation result to be 'false'")
	}
	if dummyHttpClient.nrCalls != 0 {
		t.Errorf("Expected dummyHttpClient to not have been called")
	}
}

func TestTeamProjectValidationPartialOverlapWithGlobalCohorts(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	globalCohorts := []int{1}
	teamCohorts := []int{1, 2}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts},
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, teamCohorts, nil)
	if result == false {
		t.Errorf("Expected TeamProjectValidation result to be 'true'")
	}
	if dummyHttpClient.nrCalls != 2 {
		t.Errorf("Expected dummyHttpClient to have been called twice, but got %d", dummyHttpClient.nrCalls)
	}
}

func TestTeamProjectValidationNoCohorts(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	globalCohorts := []int{}
	teamCohorts := []int{}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts},
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, teamCohorts, nil)
	if result == true {
		t.Errorf("Expected TeamProjectValidation result to be 'false'")
	}
	if dummyHttpClient.nrCalls != 0 {
		t.Errorf("Expected dummyHttpClient to not have been called")
	}
}

func TestTeamProjectValidationArborist401(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 401
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, []int{1, 2}, nil)
	if result == true {
		t.Errorf("Expected TeamProjectValidation result to be 'false'")
	}
	if dummyHttpClient.nrCalls <= 1 {
		t.Errorf("Expected dummyHttpClient to have been called more than once")
	}
}

func TestTeamProjectValidationNoTeamProjectMatchingAllCohortDefinitions(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, []int{0}, nil)
	if result == true {
		t.Errorf("Expected TeamProjectValidation result to be 'false'")
	}
	if dummyHttpClient.nrCalls > 0 {
		t.Errorf("Expected dummyHttpClient to NOT have been called")
	}
}

func TestHasAccessToTeamProjectAbortOnArboristPrepError(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Writer = new(tests.CustomResponseWriter)
	// add empty header to force an error during PrepareNewArboristRequestForResourceAndService:
	requestContext.Request.Header = map[string][]string{
		"Authorization": {""},
	}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel),
		dummyHttpClient)

	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
			if err != "Error while preparing Arborist request" {
				t.Errorf("Expected error: 'Error while preparing Arborist request'")
			}
		}
	}()
	teamProjectAuthz.HasAccessToTeamProject(requestContext, "dummyTeam")
	t.Errorf("Expected error")
}
