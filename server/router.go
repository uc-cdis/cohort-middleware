package server

import (
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/controllers"
	"github.com/uc-cdis/cohort-middleware/middlewares"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	health := new(controllers.HealthController)
	r.GET("/_health", health.Status)

	version := new(controllers.VersionController)
	r.GET("/_version", version.Retrieve)

	authorized := r.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{
		//source := new(controllers.SourceController)
		//authorized.GET("/source/by-id/:id", source.RetriveById)
		//authorized.GET("/source/by-name/:name", source.RetriveByName)
		//authorized.GET("/sources", source.RetriveAll)

		cohortdefinitions := new(controllers.CohortDefinitionController)
		authorized.GET("/cohortdefinition/by-id/:id", cohortdefinitions.RetriveById)

		cohort := new(controllers.Cohort)
		authorized.GET("/cohort/by-name/:cohortname/source/by-name/:sourcename", cohort.RetrieveByName)
	}

	return r
}
