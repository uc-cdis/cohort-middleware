package controllers

import (
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/config"
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
	// This method returns ALL cohortdefinition entries for a teamProject with cohort size statistics (for a given source).
	// If the user has access to the default global reader role, the cohorts that are part of that role are also returned.
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
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinitions for 'team project' role", "error": err.Error()})
			c.Abort()
			return
		}
		// all users should be allowed to see the cohorts shared with the default global role,
		// so also include cohorts from there:
		conf := config.GetConfig()
		globalReaderRole := conf.GetString("global_reader_role")
		log.Printf("INFO: found %s as global_reader_role", globalReaderRole)
		globalCohortDefinitionsAndStats, err := u.cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId, globalReaderRole)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinition for 'global reader' role", "error": err.Error()})
			c.Abort()
			return
		}
		// remove overlaps (if any):
		combinedTeamAndGlobalCohorts := MakeUniqueListOfCohortStats(append(cohortDefinitionsAndStats, globalCohortDefinitionsAndStats...))
		// sort by CohortSize desc:
		sort.Slice(combinedTeamAndGlobalCohorts, func(i, j int) bool {
			return combinedTeamAndGlobalCohorts[i].CohortSize > combinedTeamAndGlobalCohorts[j].CohortSize
		})
		c.JSON(http.StatusOK, gin.H{"cohort_definitions_and_stats": combinedTeamAndGlobalCohorts})
		return

	}
	c.JSON(http.StatusBadRequest, gin.H{"message": err1.Error()})
	c.Abort()
}

func MakeUniqueListOfCohortStats(input []*models.CohortDefinitionStats) []*models.CohortDefinitionStats {
	uniqueMap := make(map[int]bool)
	var uniqueList []*models.CohortDefinitionStats

	for _, item := range input {
		if !uniqueMap[item.Id] {
			uniqueMap[item.Id] = true
			uniqueList = append(uniqueList, item)
		}
	}
	return uniqueList
}

func (u CohortDefinitionController) RetriveStatsBySourceIdAndCohortIdAndObservationWindow(c *gin.Context) {
	// This method returns the cohortdefinition details and filtered size for a
	// given cohort_definition Id and observation window (aka "look back window").

	sourceId, cohortId, err := utils.ParseSourceAndCohortId(c)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}
	validAccessRequest := u.teamProjectAuthz.TeamProjectValidationForCohort(c, cohortId)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
		c.Abort()
		return
	}

	observationWindow, err := utils.ParseNumericArg(c, "observationwindow")
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}

	cohortDefinitionAndStats, err := u.cohortDefinitionModel.GetCohortDefinitionStatsByObservationWindow(sourceId, cohortId, observationWindow)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinition stats/details", "error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"cohort_definition_and_stats": cohortDefinitionAndStats})
}
