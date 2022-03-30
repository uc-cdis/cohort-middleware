package server

import (
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/controllers"
	"github.com/uc-cdis/cohort-middleware/middlewares"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(gintrace.Middleware("cohort-middleware"))

	health := new(controllers.HealthController)
	r.GET("/_health", health.Status)

	version := new(controllers.VersionController)
	r.GET("/_version", version.Retrieve)

	authorized := r.Group("/")
	authorized.Use(middlewares.AuthMiddleware())
	{
		source := new(controllers.SourceController)
		authorized.GET("/source/by-id/:id", source.RetriveById)
		authorized.GET("/source/by-name/:name", source.RetriveByName)
		authorized.GET("/sources", source.RetriveAll)

		cohortdefinitions := new(controllers.CohortDefinitionController)
		authorized.GET("/cohortdefinition/by-id/:id", cohortdefinitions.RetriveById)
		authorized.GET("/cohortdefinition/by-name/:name", cohortdefinitions.RetriveByName)
		authorized.GET("/cohortdefinitions", cohortdefinitions.RetriveAll)

		cohort := new(controllers.Cohort)
		authorized.GET("/cohort/by-name/:cohortname/source/by-name/:sourcename", cohort.RetrieveByName)

		cohortDataPhenotype := new(controllers.CohortPhenotypeData)
		authorized.GET("/cohort-data/:sourcename", cohortDataPhenotype.Retrieve)
	}

	return r
}
