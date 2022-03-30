package controllers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
)

type CohortDataController struct {
}

var cohortDataModel = new(models.CohortData)

func (u CohortDataController) RetrieveDataBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {

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
	cohortData, err := cohortDataModel.RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId, cohortId, conceptIds.ConceptIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err})
		c.Abort()
		return
	}
	b := GenerateTSV(cohortData, conceptIds.ConceptIds)
	c.String(http.StatusOK, b.String())
	return
}

func GenerateTSV(cohort []*models.PersonConceptAndValue, conceptIds []int) *bytes.Buffer {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.Comma = '\t'

	var rows [][]string
	var header []string
	header = append(header, "sample.id")
	header = addConceptsToHeader(header, conceptIds)
	rows = append(rows, header)

	var currentPersonId = -1
	var row []string
	for _, cohortItem := range cohort {
		// if new person, start new row:
		if cohortItem.PersonId != currentPersonId {
			if currentPersonId != -1 {
				rows = append(rows, row)
			}
			row = []string{}
			row = append(row, strconv.Itoa(cohortItem.PersonId))
			row = appendInitEmptyConceptValues(row, len(conceptIds))
			currentPersonId = cohortItem.PersonId
		}
		row = populateConceptValue(row, *cohortItem, conceptIds)
	}
	// append last person row:
	rows = append(rows, row)

	err := w.WriteAll(rows)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func addConceptsToHeader(header []string, conceptIds []int) []string {
	for i := 0; i < len(conceptIds); i++ {
		header = append(header, strconv.Itoa(conceptIds[i]))
	}
	return header
}

func appendInitEmptyConceptValues(row []string, nrConceptIds int) []string {
	for i := 0; i < nrConceptIds; i++ {
		row = append(row, "NA")
	}
	return row
}

func populateConceptValue(row []string, cohortItem models.PersonConceptAndValue, conceptIds []int) []string {
	var conceptIdIdx int = pos(cohortItem.ConceptId, conceptIds)
	if conceptIdIdx != -1 {
		log.Printf("Value as string: %s", cohortItem.ConceptValueAsString)
		log.Printf("Value as number: %f", cohortItem.ConceptValueAsNumber)
		if cohortItem.ConceptValueAsString != "" {
			row[conceptIdIdx+1] = cohortItem.ConceptValueAsString // +1 because first column is sample.id
		} else if cohortItem.ConceptValueAsNumber != 0.0 {
			s := strconv.FormatFloat(float64(cohortItem.ConceptValueAsNumber), 'f', 2, 64)
			row[conceptIdIdx+1] = s
		}
	}
	return row
}

func pos(value int, list []int) int {
	for p, v := range list {
		if v == value {
			return p
		}
	}
	return -1
}
