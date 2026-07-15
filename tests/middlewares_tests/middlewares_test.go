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
	succeedFirst bool
	statusCode   int
	nrCalls      int
}

func (h *dummyHttpClient) Do(req *http.Request) (*http.Response, error) {
	h.nrCalls++
	if h.nrCalls == 1 && h.succeedFirst {
		return &http.Response{StatusCode: 200}, nil
	}
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

func (h dummyCohortDefinitionDataModel) GetCohortDefinitionStatsByObservationWindow(sourceId int, cohortId int, observationWindow int) (*models.CohortDefinitionStats, error) {
	return nil, nil
}

func (h dummyCohortDefinitionDataModel) GetCohortDefinitionStatsByObservationWindow1stCohortAndOverlap2ndCohort(sourceId int, cohort1Id int, cohort2Id int, observationWindow1stCohort int) (*models.CohortDefinitionStats, error) {
	return nil, nil
}

func (h dummyCohortDefinitionDataModel) GetCohortDefinitionStatsByObservationWindow1stCohortAndOverlap2ndCohortAndOutcomeWindow2ndCohort(sourceId int, cohort1Id int, cohort2Id int, observationWindow1stCohort int, outcomeWindow2ndCohort int) (*models.CohortDefinitionStats, error) {
	return nil, nil
}

func (h dummyCohortDefinitionDataModel) GetCohortDefinitionStatsByObservationWindow1stCohortAndOverlap2ndCohortAnd2ndCohortEntryFirst(sourceId int, cohort1Id int, cohort2Id int, observationWindow1stCohort int) (*models.CohortDefinitionStats, error) {
	return nil, nil
}

type dummySourceModel struct{}

func (h dummySourceModel) GetSourceById(id int) (*models.Source, error) {
	return &models.Source{SourceId: id, SourceName: "dummy source"}, nil
}

func (h dummySourceModel) GetSourceByName(name string) (*models.Source, error) {
	return &models.Source{SourceId: 1, SourceName: name}, nil
}

func (h dummySourceModel) GetAllSources() ([]*models.Source, error) {
	return []*models.Source{{SourceId: 1, SourceName: "dummy source"}}, nil
}

func (h dummySourceModel) GetAllSourcesWithTeamProject(teamName string) ([]*models.Source, error) {
	// if dummyModelReturnError {
	// 	return nil, fmt.Errorf("fake model error!")
	// }
	return []*models.Source{
		{
			SourceId:                     1,
			SourceName:                   "source for " + teamName,
			CurrentTeamProjectAccessible: true,
		},
	}, nil
}

func (h dummySourceModel) GetAllRoleNamesWithSourceGeneratePermission(sourceId int) ([]string, error) {
	// dummy switch just to support two test scenarios:
	if sourceId == testSourceId {
		return []string{"dummy role"}, nil
	}
	return make([]string, 0), nil

}

func TestTeamProjectValidationForCohort(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel), *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidationForSourceIdAndCohort(requestContext, testSourceId, 1)
	if result == false {
		t.Errorf("Expected TeamProjectValidationForCohort result to be 'true'")
	}
	if dummyHttpClient.nrCalls != 2 {
		t.Errorf("Expected dummyHttpClient to have been called twice")
	}
}

func TestTeamProjectValidationForSourceIdAndCohortIdsList(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel), *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	// happy scenario:
	result := teamProjectAuthz.TeamProjectValidationForSourceIdAndCohortIdsList(requestContext, testSourceId, []int{1})
	if result == false {
		t.Errorf("Expected TeamProjectValidationForSourceIdAndCohortIdsList result to be 'true'")
	}
	if dummyHttpClient.nrCalls != 2 {
		t.Errorf("Expected dummyHttpClient to have been called twice")
	}
	// error scenario:
	result = teamProjectAuthz.TeamProjectValidationForSourceIdAndCohortIdsList(requestContext, -1, []int{1})
	if result == true {
		t.Errorf("Expected TeamProjectValidationForSourceIdAndCohortIdsList result to be 'false'")
	}
	// no extra calls when compared to above:
	if dummyHttpClient.nrCalls != 2 {
		t.Errorf("Expected no extra calls to dummyHttpClient")
	}
}

func TestTeamProjectValidationForCohortArborist401(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 401
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel), *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidationForSourceIdAndCohort(requestContext, testSourceId, 1)
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
	cohortsToCheck := []int{1, 2}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts}, *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, testSourceId, cohortsToCheck, nil)
	if result == false {
		t.Errorf("Expected TeamProjectValidation result to be 'true'")
	}
	if dummyHttpClient.nrCalls != 2 {
		t.Errorf("Expected dummyHttpClient to have been called twice")
	}
}

func TestTeamProjectValidationFullOverlapWithGlobalCohorts(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	globalCohorts := []int{1, 2}
	cohortsToCheck := []int{1, 2}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts}, *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, testSourceId, cohortsToCheck, nil)
	if result == false {
		t.Errorf("Expected TeamProjectValidation result to be 'true'")
	}
	if dummyHttpClient.nrCalls != 1 {
		t.Errorf("Expected dummyHttpClient to only have been called once, but got %d calls", dummyHttpClient.nrCalls)
	}
}

func TestTeamProjectValidationOnlyGlobalCohorts(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	globalCohorts := []int{1, 2}
	cohortsToCheck := []int{}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts}, *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, testSourceId, cohortsToCheck, nil)
	if result == true {
		t.Errorf("Expected TeamProjectValidation result to be 'false'")
	}
	if dummyHttpClient.nrCalls != 1 {
		t.Errorf("Expected dummyHttpClient to only have been called once, but got %d calls", dummyHttpClient.nrCalls)
	}
}

func TestTeamProjectValidationPartialOverlapWithGlobalCohorts(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	globalCohorts := []int{1}
	cohortsToCheck := []int{1, 2}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts}, *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, testSourceId, cohortsToCheck, nil)
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
	cohortsToCheck := []int{}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(&dummyCohortDefinitionDataModel{returnForGetCohortDefinitionIdsForTeamProject: globalCohorts}, *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, testSourceId, cohortsToCheck, nil)
	if result == true {
		t.Errorf("Expected TeamProjectValidation result to be 'false'")
	}
	if dummyHttpClient.nrCalls != 1 {
		t.Errorf("Expected dummyHttpClient to only have been called once, but got %d calls", dummyHttpClient.nrCalls)
	}
}

func TestTeamProjectValidationArborist401ForTeamProject(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 401
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode, succeedFirst: true}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel), *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, testSourceId, []int{1, 2}, nil)
	if result == true {
		t.Errorf("Expected TeamProjectValidation result to be 'false'")
	}
	// the studyId permission check will pass (see succeedFirst: true above), but then the
	// cohort permission checks will fail with 401. In total, we expect 3 calls:
	if dummyHttpClient.nrCalls != 3 {
		t.Errorf("Expected dummyHttpClient to have been called three times")
	}
}

func TestTeamProjectValidationNoTeamProjectMatchingAllCohortDefinitions(t *testing.T) {
	setUp(t)
	config.Init("mocktest")
	arboristAuthzResponseCode := 200
	dummyHttpClient := &dummyHttpClient{statusCode: arboristAuthzResponseCode}
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel), *new(dummySourceModel),
		dummyHttpClient)
	requestContext := new(gin.Context)
	requestContext.Request = new(http.Request)
	requestContext.Request.Header = map[string][]string{
		"Authorization": {"dummy_token_value"},
	}
	result := teamProjectAuthz.TeamProjectValidation(requestContext, testSourceId, []int{0}, nil)
	if result == true {
		t.Errorf("Expected TeamProjectValidation result to be 'false'")
	}
	if dummyHttpClient.nrCalls != 1 {
		t.Errorf("Expected dummyHttpClient to only have been called once, but got %d calls", dummyHttpClient.nrCalls)
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
	teamProjectAuthz := middlewares.NewTeamProjectAuthz(*new(dummyCohortDefinitionDataModel), *new(dummySourceModel),
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
