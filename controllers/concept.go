package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
)

type ConceptController struct{}

var conceptModel = new(models.Concept)

func (u ConceptController) RetriveAllBySourceId(c *gin.Context) {
	sourceId := c.Param("sourceid")

	if sourceId != "" {
		sourceId, _ := strconv.Atoi(sourceId)
		concepts, err := conceptModel.RetriveAllBySourceId(sourceId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"concepts": concepts})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
	return
}
