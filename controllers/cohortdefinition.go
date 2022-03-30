package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
	"net/http"
	"strconv"
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
		c.JSON(http.StatusOK, gin.H{"CohortDefinition": cohortDefinition})
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
	c.JSON(http.StatusOK, gin.H{"CohortDefinition": cohortDefinitions})
	return
}
