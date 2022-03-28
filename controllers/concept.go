package controllers

import (
	"encoding/json"
	"log"
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

type ConceptIds struct {
	ConceptIds []int
}

func (u ConceptController) RetriveStatsBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {

	// parse and validate all parameters:
	sourceIdStr := c.Param("sourceid")
	log.Printf("Querying source: %s", sourceIdStr)
	cohortIdStr := c.Param("cohortid")
	log.Printf("Querying cohort for cohort definition id: %s", cohortIdStr)
	if sourceIdStr == "" || cohortIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		c.Abort()
		return
	}
	decoder := json.NewDecoder(c.Request.Body)
	var conceptIds ConceptIds
	err := decoder.Decode(&conceptIds)
	if err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request - no request body"})
		c.Abort()
		return
	}
	log.Printf("Querying concept ids: %v", conceptIds.ConceptIds)

	sourceId, _ := strconv.Atoi(sourceIdStr)
	cohortId, _ := strconv.Atoi(cohortIdStr)

	// call model method:
	conceptStats, err := conceptModel.RetriveStatsBySourceIdAndCohortIdAndConceptIds(sourceId, cohortId, conceptIds.ConceptIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"concepts": conceptStats})
	return
}
