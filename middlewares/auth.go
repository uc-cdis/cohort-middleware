package middlewares

import (
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
		authorization := ctx.Request.Header.Get("Authorization")
		if authorization == "" {
			ctx.AbortWithStatus(401)
		}

		resourcePath := fmt.Sprintf("/cohort-middleware%s", ctx.Request.URL.Path)
		arboristAuth := fmt.Sprintf("%s/auth/proxy?resource=%s&service=%s&method=%s",
			c.GetString("arborist_endpoint"),
			resourcePath,
			"cohort-middleware",
			"access")

		req, err := http.NewRequest("GET", arboristAuth, nil)
		if err != nil {
			log.Printf("arborist request: unhandled error in the middleware\n%s", err.Error())
		}

		req.Header.Set("Authorization", authorization)
		client := &http.Client{}
		resp, _ := client.Do(req)

		if resp.StatusCode != 200 {
			ctx.AbortWithStatus(401)
			return
		}

		ctx.Next()
	}
}
