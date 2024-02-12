package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/middlewares"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortDefinitionController struct {
	cohortDefinitionModel models.CohortDefinitionI
	teamProjectAuthz      middlewares.TeamProjectAuthzI
}

func NewCohortDefinitionController(cohortDefinitionModel models.CohortDefinitionI, teamProjectAuthz middlewares.TeamProjectAuthzI) CohortDefinitionController {
	return CohortDefinitionController{
		cohortDefinitionModel: cohortDefinitionModel,
		teamProjectAuthz:      teamProjectAuthz,
	}
}

func (u CohortDefinitionController) RetriveById(c *gin.Context) {
	cohortDefinitionId := c.Param("id")

	if cohortDefinitionId != "" {
		cohortDefinitionId, _ := strconv.Atoi(cohortDefinitionId)
		// validate teamproject access permission for cohort:
		validAccessRequest := u.teamProjectAuthz.TeamProjectValidationForCohort(c, cohortDefinitionId)
		if !validAccessRequest {
			log.Printf("Error: invalid request")
			c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
			c.Abort()
			return
		}
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
	// validate teamproject access permission:
	validAccessRequest := u.teamProjectAuthz.HasAccessToTeamProject(c, teamProject)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
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
