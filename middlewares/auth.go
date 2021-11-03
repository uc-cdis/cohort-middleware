package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/config"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorization := ctx.Request.Header.Get("Authorization")
		if authorization == "" {
			ctx.AbortWithStatus(401)
		}

		c := config.GetConfig()

		resourcePath := fmt.Sprintf("/cohort-middleware%s", ctx.Request.URL.Path)
		arboristAuth := fmt.Sprintf("%s/auth/proxy?resource=%s&service=%s&method=%s",
			c.GetString("arborist_endpoint"),
			resourcePath,
			"cohort-middleware",
			"access")

		req, err := http.NewRequest("GET", arboristAuth, nil)
		if err != nil {
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
