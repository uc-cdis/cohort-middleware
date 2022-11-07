package server

import (
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/controllers"
	"github.com/uc-cdis/cohort-middleware/middlewares"
	"github.com/uc-cdis/cohort-middleware/models"
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

		cohortdefinitions := controllers.NewCohortDefinitionController(*new(models.CohortDefinition))
		authorized.GET("/cohortdefinition/by-id/:id", cohortdefinitions.RetriveById)
		authorized.GET("/cohortdefinition/by-name/:name", cohortdefinitions.RetriveByName)
		authorized.GET("/cohortdefinitions", cohortdefinitions.RetriveAll)
		authorized.GET("/cohortdefinition-stats/by-source-id/:sourceid", cohortdefinitions.RetriveStatsBySourceId)

		// concept endpoints:
		concepts := controllers.NewConceptController(*new(models.Concept), *new(models.CohortDefinition))
		authorized.GET("/concept/by-source-id/:sourceid", concepts.RetriveAllBySourceId)
		authorized.POST("/concept/by-source-id/:sourceid", concepts.RetrieveInfoBySourceIdAndConceptIds)
		authorized.POST("/concept/by-source-id/:sourceid/by-type", concepts.RetrieveInfoBySourceIdAndConceptTypes)

		authorized.POST("/concept-stats/by-source-id/:sourceid/by-cohort-definition-id/:cohortid", concepts.RetrieveStatsBySourceIdAndCohortIdAndConceptIds)
		authorized.GET("/concept-stats/by-source-id/:sourceid/by-cohort-definition-id/:cohortid/breakdown-by-concept-id/:breakdownconceptid", concepts.RetrieveBreakdownStatsBySourceIdAndCohortId)
		authorized.POST("/concept-stats/by-source-id/:sourceid/by-cohort-definition-id/:cohortid/breakdown-by-concept-id/:breakdownconceptid", concepts.RetrieveBreakdownStatsBySourceIdAndCohortIdAndVariables)
		authorized.POST("/concept-stats/by-source-id/:sourceid/by-cohort-definition-id/:cohortid/breakdown-by-concept-id/:breakdownconceptid/csv", concepts.RetrieveAttritionTable)

		// cohort stats and checks:
		cohortData := controllers.NewCohortDataController(*new(models.CohortData))
		// :casecohortid/:controlcohortidare just labels here and have no special meaning. Could also just be :cohortAId/:cohortBId here:
		authorized.POST("/cohort-stats/check-overlap/by-source-id/:sourceid/by-case-control-cohort-definition-ids/:casecohortid/:controlcohortid/filter-by-concept-id-and-value/:filterconceptid/:filtervalue", cohortData.RetrieveCohortOverlapStats)
		authorized.POST("/cohort-stats/check-overlap/by-source-id/:sourceid/by-cohort-definition-ids/:casecohortid/:controlcohortid", cohortData.RetrieveCohortOverlapStatsWithoutFilteringOnConceptValue)

		// full data endpoints:
		authorized.POST("/cohort-data/by-source-id/:sourceid/by-cohort-definition-id/:cohortid", cohortData.RetrieveDataBySourceIdAndCohortIdAndVariables)

		// histogram endpoint
		authorized.POST("/histogram/by-source-id/:sourceid/by-cohort-definition-id/:cohortid/by-histogram-concept-id/:histogramid", cohortData.RetrieveHistogramForCohortIdAndConceptId)
	}

	return r
}
