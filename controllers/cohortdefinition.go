package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
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
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinition", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"cohort_definition": cohortDefinition})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u CohortDefinitionController) RetriveByName(c *gin.Context) {
	cohortDefinitionName := c.Param("name")

	if cohortDefinitionName != "" {
		cohortDefinition, err := u.cohortDefinitionModel.GetCohortDefinitionByName(cohortDefinitionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinition", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"CohortDefinition": cohortDefinition})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u CohortDefinitionController) RetriveAll(c *gin.Context) {
	cohortDefinitions, err := u.cohortDefinitionModel.GetAllCohortDefinitions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinition", "error": err})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"cohort_definitions": cohortDefinitions})
}

func parseNumericId(c *gin.Context, paramName string) (int, error) {
	// parse and validate:
	numericIdStr := c.Param(paramName)
	log.Printf("Querying %s: %s", paramName, numericIdStr)
	if numericId, err := strconv.Atoi(numericIdStr); err != nil {
		log.Printf("bad request - %s should be a number", paramName)
		return -1, fmt.Errorf("bad request - %s should be a number", paramName)
	} else {
		return numericId, nil
	}
}

func (u CohortDefinitionController) RetriveStatsBySourceId(c *gin.Context) {
	// This method returns ALL cohortdefinition entries with cohort size statistics (for a given source)

	sourceId, err1 := parseNumericId(c, "sourceid")
	if err1 == nil {
		cohortDefinitionsAndStats, err := u.cohortDefinitionModel.GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohortDefinition", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"cohort_definitions_and_stats": cohortDefinitionsAndStats})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": err1.Error()})
	c.Abort()
}

func (u CohortDefinitionController) RetriveStatsBySourceIdAndCohortIdAndBreakDownOnConceptId(c *gin.Context) {
	sourceId, err1 := parseNumericId(c, "sourceid")
	cohortId, err2 := parseNumericId(c, "cohortid")
	conceptId, err3 := parseNumericId(c, "conceptid")

	if err1 == nil && err2 == nil && err3 == nil {
		cohortDefinitionsAndStats, err := u.cohortDefinitionModel.GetCohortDefinitionsAndStatsBySourceIdAndCohortIdAndBreakDownOnConceptId(sourceId, cohortId, conceptId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"cohort_definition_and_concept_breakdown_stats": cohortDefinitionsAndStats})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}
