package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
	"net/http"
)

type CohortPhenotypeData struct{}

var cohortPhenotypeDataModel = new(models.CohortPhenotypeData)

func (u CohortPhenotypeData) Retrieve(c *gin.Context) {
	//cohortName := c.Param("cohortname")
	sourceName := c.Param("sourcename")

	if sourceName != "" {
		cohort, err := cohortPhenotypeDataModel.GetCohortDataPhenotype(sourceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve cohort", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"Cohort": cohort})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
	return
}
