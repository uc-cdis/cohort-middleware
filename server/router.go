package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/controllers"
	"github.com/uc-cdis/cohort-middleware/middlewares"
	"github.com/uc-cdis/cohort-middleware/models"
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
		source := new(controllers.SourceController)
		authorized.GET("/source/by-id/:id", source.RetriveById)
		authorized.GET("/source/by-name/:name", source.RetriveByName)
		authorized.GET("/sources", source.RetriveAll)

		cohortdefinitions := controllers.NewCohortDefinitionController(*new(models.CohortDefinition))
		authorized.GET("/cohortdefinition-stats/by-source-id/:sourceid/by-team-project/:teamproject", cohortdefinitions.RetriveStatsBySourceIdAndTeamProject)

		// concept endpoints:
		concepts := controllers.NewConceptController(*new(models.Concept), *new(models.CohortDefinition),
			middlewares.NewTeamProjectAuthz(*new(models.CohortDefinition), &http.Client{}))
		authorized.GET("/concept/by-source-id/:sourceid", concepts.RetriveAllBySourceId)
		authorized.POST("/concept/by-source-id/:sourceid", concepts.RetrieveInfoBySourceIdAndConceptIds)
		authorized.POST("/concept/by-source-id/:sourceid/by-type", concepts.RetrieveInfoBySourceIdAndConceptTypes)

		authorized.GET("/concept-stats/by-source-id/:sourceid/by-cohort-definition-id/:cohortid/breakdown-by-concept-id/:breakdownconceptid", concepts.RetrieveBreakdownStatsBySourceIdAndCohortId)
		authorized.POST("/concept-stats/by-source-id/:sourceid/by-cohort-definition-id/:cohortid/breakdown-by-concept-id/:breakdownconceptid", concepts.RetrieveBreakdownStatsBySourceIdAndCohortIdAndVariables)
		authorized.POST("/concept-stats/by-source-id/:sourceid/by-cohort-definition-id/:cohortid/breakdown-by-concept-id/:breakdownconceptid/csv", concepts.RetrieveAttritionTable)

		// cohort stats and checks:
		cohortData := controllers.NewCohortDataController(*new(models.CohortData), middlewares.NewTeamProjectAuthz(*new(models.CohortDefinition), &http.Client{}))
		// :casecohortid/:controlcohortid are just labels here and have no special meaning. Could also just be :cohortAId/:cohortBId here:
		authorized.POST("/cohort-stats/check-overlap/by-source-id/:sourceid/by-cohort-definition-ids/:casecohortid/:controlcohortid", cohortData.RetrieveCohortOverlapStatsWithoutFilteringOnConceptValue)

		// full data endpoints:
		authorized.POST("/cohort-data/by-source-id/:sourceid/by-cohort-definition-id/:cohortid", cohortData.RetrieveDataBySourceIdAndCohortIdAndVariables)

		// histogram endpoint
		authorized.POST("/histogram/by-source-id/:sourceid/by-cohort-definition-id/:cohortid/by-histogram-concept-id/:histogramid", cohortData.RetrieveHistogramForCohortIdAndConceptId)
	}

	return r
}
