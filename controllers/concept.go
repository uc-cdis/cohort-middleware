package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type ConceptController struct {
	conceptModel models.ConceptI
}

func NewConceptController(conceptModel models.ConceptI) ConceptController {
	return ConceptController{conceptModel: conceptModel}
}

func (u ConceptController) RetriveAllBySourceId(c *gin.Context) {
	sourceId := c.Param("sourceid")

	if sourceId != "" {
		sourceId, _ := strconv.Atoi(sourceId)
		concepts, err := u.conceptModel.RetriveAllBySourceId(sourceId)
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
}

type ConceptIds struct {
	ConceptIds []int
}

func parseSourceIdAndConceptIds(c *gin.Context) (int, []int, error) {
	// parse and validate all parameters:
	sourceId, err1 := utils.ParseNumericArg(c, "sourceid")
	if err1 != nil {
		return -1, nil, err1
	}
	decoder := json.NewDecoder(c.Request.Body)
	var conceptIds ConceptIds
	err := decoder.Decode(&conceptIds)
	if err != nil {
		return -1, nil, errors.New("bad request - no request body")
	}
	log.Printf("Querying concept ids: %v", conceptIds.ConceptIds)

	return sourceId, conceptIds.ConceptIds, nil
}

func parseSourceIdAndCohortIdAndConceptIds(c *gin.Context) (int, int, []int, error) {
	// parse and validate all parameters:
	sourceId, conceptIds, err := parseSourceIdAndConceptIds(c)
	if err != nil {
		return -1, -1, nil, err
	}
	cohortIdStr := c.Param("cohortid")
	log.Printf("Querying cohort for cohort definition id: %s", cohortIdStr)
	if _, err := strconv.Atoi(cohortIdStr); err != nil {
		return -1, -1, nil, errors.New("bad request - cohort_definition_id should be a number")
	}
	cohortId, _ := strconv.Atoi(cohortIdStr)
	return sourceId, cohortId, conceptIds, nil
}

func (u ConceptController) RetrieveStatsBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {

	sourceId, cohortId, conceptIds, err := parseSourceIdAndCohortIdAndConceptIds(c)
	if err != nil {
		log.Printf("Error parsing request parameters")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		c.Abort()
		return
	}

	// call model method:
	conceptStats, err := u.conceptModel.RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId, cohortId, conceptIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"concepts": conceptStats})
}

func (u ConceptController) RetrieveInfoBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {

	sourceId, conceptIds, err := parseSourceIdAndConceptIds(c)
	if err != nil {
		log.Printf("Error parsing request parameters")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		c.Abort()
		return
	}

	// call model method:
	conceptInfo, err := u.conceptModel.RetrieveInfoBySourceIdAndConceptIds(sourceId, conceptIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"concepts": conceptInfo})
}

func (u ConceptController) RetrieveBreakdownStatsBySourceIdAndCohortId(c *gin.Context) {
	sourceId, err1 := utils.ParseNumericArg(c, "sourceid")
	cohortId, err2 := utils.ParseNumericArg(c, "cohortid")
	breakdownConceptId, err3 := utils.ParseNumericArg(c, "breakdownconceptid")
	if err1 == nil && err2 == nil && err3 == nil {
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId, cohortId, breakdownConceptId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"concept_breakdown": breakdownStats})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u ConceptController) RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {
	sourceId, cohortId, conceptIds, err1 := parseSourceIdAndCohortIdAndConceptIds(c)
	breakdownConceptId, err2 := utils.ParseNumericArg(c, "breakdownconceptid")
	valueFormat, err3 := utils.ParseStringArg(c, "valueformat")
	if err1 == nil && err2 == nil && err3 == nil {
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(sourceId, cohortId, conceptIds, breakdownConceptId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"concept_breakdown": breakdownStats})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}
