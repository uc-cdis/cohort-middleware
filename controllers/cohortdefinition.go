package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortDefinitionController struct {
	cohortDefinitionModel models.CohortDefinitionI
}

func NewCohortDefinitionController(cohortDefinitionModel models.CohortDefinitionI) CohortDefinitionController {
	return CohortDefinitionController{cohortDefinitionModel: cohortDefinitionModel}
}

func (u CohortDefinitionController) RetriveById(c *gin.Context) {
	cohortDefinitionId := c.Param("id")

	if cohortDefinitionId != "" {
		cohortDefinitionId, _ := strconv.Atoi(cohortDefinitionId)
		cohortDefinition, err := u.cohortDefinitionModel.GetCohortDefinitionById(cohortDefinitionId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinition", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"cohort_definition": cohortDefinition})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u CohortDefinitionController) RetriveStatsBySourceIdAndTeamProject(c *gin.Context) {
	// This method returns ALL cohortdefinition entries with cohort size statistics (for a given source)
	sourceId, err1 := utils.ParseNumericArg(c, "sourceid")
	teamProject := c.Query("team-project")
	if teamProject == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while parsing request", "error": "team-project is a mandatory parameter but was found to be empty!"})
		c.Abort()
		return
	}
	if err1 == nil {
		cohortDefinitionsAndStats, err := u.cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId, teamProject)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinition", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"cohort_definitions_and_stats": cohortDefinitionsAndStats})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": err1.Error()})
	c.Abort()
}
