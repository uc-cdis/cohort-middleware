package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
	"net/http"
)

type Cohort struct{}

var cohortModel = new(models.Cohort)

func (u Cohort) RetrieveByName(c *gin.Context) {
	cohortName := c.Param("cohortname")
	sourceName := c.Param("sourcename")

	if cohortName != "" || sourceName != "" {
		cohort, err := cohortModel.GetCohortByName(sourceName, cohortName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve cohort", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"cohort": cohort})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
	return
}
