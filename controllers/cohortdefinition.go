package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
)

type CohortDefinitionController struct{}

var cohortDefinitionModel = new(models.CohortDefinition)

func (u CohortDefinitionController) RetriveById(c *gin.Context) {
	cohortDefinitionId := c.Param("id")

	if cohortDefinitionId != "" {
		cohortDefinitionId, _ := strconv.Atoi(cohortDefinitionId)
		cohortDefinition, err := cohortDefinitionModel.GetCohortDefinitionById(cohortDefinitionId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve cohortDefinition", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"cohort_definition": cohortDefinition})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
	return
}

func (u CohortDefinitionController) RetriveByName(c *gin.Context) {
	cohortDefinitionName := c.Param("name")

	if cohortDefinitionName != "" {
		cohortDefinition, err := cohortDefinitionModel.GetCohortDefinitionByName(cohortDefinitionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve cohortDefinition", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"CohortDefinition": cohortDefinition})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
	return
}

func (u CohortDefinitionController) RetriveAll(c *gin.Context) {
	cohortDefinitions, err := cohortDefinitionModel.GetAllCohortDefinitions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve cohortDefinition", "error": err})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"cohort_definitions": cohortDefinitions})
	return
}

func (u CohortDefinitionController) RetriveStatsBySourceId(c *gin.Context) {
	// This method returns ALL cohortdefinition entries with cohort size statistics (for a given source)

	sourceId := c.Param("sourceid")
	if sourceId != "" {
		sourceId, _ := strconv.Atoi(sourceId)
		cohortDefinitionsAndStats, err := cohortDefinitionModel.GetAllCohortDefinitionsAndStats(sourceId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve cohortDefinition", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"cohort_definitions_and_stats": cohortDefinitionsAndStats})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}
