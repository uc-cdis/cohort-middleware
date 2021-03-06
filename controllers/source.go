package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
)

type SourceController struct{}

var sourceModel = new(models.Source)

func (u SourceController) RetriveById(c *gin.Context) {
	sourceId := c.Param("id")
	if sourceId != "" {
		sourceId, _ := strconv.Atoi(c.Param("id"))
		source, err := sourceModel.GetSourceById(sourceId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve source", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"source": source})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u SourceController) RetriveByName(c *gin.Context) {
	sourceName := c.Param("name")
	if sourceName != "" {
		source, err := sourceModel.GetSourceByName(sourceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve source", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"source": source})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u SourceController) RetriveAll(c *gin.Context) {
	source, err := sourceModel.GetAllSources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve source", "error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"sources": source})
}
