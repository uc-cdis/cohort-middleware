package middlewares

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/config"
)

func AuthMiddleware() gin.HandlerFunc {

	c := config.GetConfig()

	// used in local DEV mode:
	if c.GetString("arborist_endpoint") == "NONE" {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}

	return func(ctx *gin.Context) {
		req, err := PrepareNewArboristRequest(ctx)
		if err != nil {
			ctx.AbortWithStatus(500)
			log.Printf("Error while preparing Arborist request: %s", err.Error())
			return
		}
		client := &http.Client{}
		// send the request to Arborist:
		resp, _ := client.Do(req)

		// arborist will return with 200 if the user has been granted access to the cohort-middleware URL in ctx:
		if resp.StatusCode != 200 {
			// return Unauthorized otherwise:
			log.Printf("Got response status %d from Arborist. Aborting this cohort-middleware request with 401...", resp.StatusCode)
			ctx.AbortWithStatus(401)
			return
		}

		ctx.Next()
	}
}

// this function will take the request from the given ctx, validated it for the presence of an "Authorization / Bearer" token
// and then return the URL that can be used to consult Arborist regarding cohort-middleware access permissions. This function
// returns an error if "Authorization / Bearer" token is missing in ctx
func PrepareNewArboristRequest(ctx *gin.Context) (*http.Request, error) {

	resourcePath := fmt.Sprintf("/cohort-middleware%s", ctx.Request.URL.Path)
	service := "cohort-middleware"

	return PrepareNewArboristRequestForResourceAndService(ctx, resourcePath, service)
}

// this function will take the request from the given ctx, validated it for the presence of an "Authorization / Bearer" token
// and then return the URL that can be used to consult Arborist regarding access permissions for the given
// resource path and service.
func PrepareNewArboristRequestForResourceAndService(ctx *gin.Context, resourcePath string, service string) (*http.Request, error) {
	c := config.GetConfig()
	arboristEndpoint := c.GetString("arborist_endpoint")
	// validate:
	authorization := ctx.Request.Header.Get("Authorization")
	if authorization == "" {
		return nil, errors.New("missing Authorization header")
	}

	// build up the request URL string:
	arboristAuth := fmt.Sprintf("%s/auth/proxy?resource=%s&service=%s&method=%s",
		arboristEndpoint,
		resourcePath,
		service,
		"access")

	// make request object / validate URL:
	req, err := http.NewRequest("GET", arboristAuth, nil)
	if err != nil {
		return nil, fmt.Errorf("unexpected error while assembling the Arborist request URL in cohort-middleware: %s", err.Error())
	}

	// make sure to pass on the auth/bearer token string in this new request:
	req.Header.Set("Authorization", authorization)
	return req, nil
}
