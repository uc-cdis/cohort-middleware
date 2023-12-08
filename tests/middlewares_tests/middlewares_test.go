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
