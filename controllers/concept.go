package controllers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
)

var cohortDefinitionModel = new(models.CohortDefinition)

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
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"concepts": concepts})
		return
	}
	log.Printf("Error: bad request")
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u ConceptController) RetrieveStatsBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {

	sourceId, cohortId, conceptIds, err := utils.ParseSourceIdAndCohortIdAndConceptIds(c)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		c.Abort()
		return
	}

	// call model method:
	conceptStats, err := u.conceptModel.RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId, cohortId, conceptIds)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"concepts": conceptStats})
}

func (u ConceptController) RetrieveInfoBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {

	sourceId, conceptIds, err := utils.ParseSourceIdAndConceptIds(c)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		c.Abort()
		return
	}

	// call model method:
	conceptInfo, err := u.conceptModel.RetrieveInfoBySourceIdAndConceptIds(sourceId, conceptIds)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"concepts": conceptInfo})
}

func (u ConceptController) RetrieveBreakdownStatsBySourceIdAndCohortId(c *gin.Context) {
	sourceId, err1 := utils.ParseNumericArg(c, "sourceid")
	cohortId, err2 := utils.ParseNumericArg(c, "cohortid")
	breakdownConceptId, err3 := utils.ParseBigNumericArg(c, "breakdownconceptid")
	if err1 == nil && err2 == nil && err3 == nil {
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId, cohortId, breakdownConceptId)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"concept_breakdown": breakdownStats})
		return
	}
	log.Printf("Error: bad request")
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func (u ConceptController) RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {
	sourceId, cohortId, conceptIds, err1 := utils.ParseSourceIdAndCohortIdAndConceptIds(c)
	breakdownConceptId, err2 := utils.ParseBigNumericArg(c, "breakdownconceptid")
	if err1 == nil && err2 == nil {
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(sourceId, cohortId, conceptIds, breakdownConceptId)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err.Error()})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, gin.H{"concept_breakdown": breakdownStats})
		return
	}
	log.Printf("Error: bad request")
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func getConceptValueToPeopleCount(breakdownStats []*models.ConceptBreakdown) map[string]int {
	concept_values_to_people_count := make(map[string]int)
	for _, breakdownStat := range breakdownStats {
		concept_value := breakdownStat.ConceptValue
		if concept_value == "" {
			concept_value = "empty string"
		}
		concept_values_to_people_count[concept_value] = breakdownStat.NpersonsInCohortWithValue
	}

	return concept_values_to_people_count
}

func generateConceptRow(cohortName string, conceptValuesToPeopleCount map[string]int, sortedConceptValues []string) []string {
	cohortSize := 0
	for _, peopleCount := range conceptValuesToPeopleCount {
		cohortSize += peopleCount
	}

	row := []string{cohortName, strconv.Itoa(cohortSize)}
	for _, concept := range sortedConceptValues {
		row = append(row, strconv.Itoa(conceptValuesToPeopleCount[concept]))
	}

	return row
}

func (u ConceptController) GetFilteredConceptRows(sourceId int, cohortId int, conceptIds []int64, breakdownConceptId int64, sortedConceptValues []string) ([][]string, error) {
	conceptIdToName := make(map[int64]string)
	conceptInformations, err := u.conceptModel.RetrieveInfoBySourceIdAndConceptIds(sourceId, conceptIds)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve concept informations due to error: %s", err.Error())
	}
	for _, conceptInformation := range conceptInformations {
		conceptIdToName[conceptInformation.ConceptId] = conceptInformation.ConceptName
	}

	var rows [][]string
	for idx, conceptId := range conceptIds {
		// run each query with a longer list of filterConceptIds, until the last query is run with them all:
		filterConceptIds := conceptIds[0 : idx+1]
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(sourceId, cohortId, filterConceptIds, breakdownConceptId)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve concept Breakdown for concepts %v due to error: %s", filterConceptIds, err.Error())
		}

		conceptValuesToPeopleCount := getConceptValueToPeopleCount(breakdownStats)
		conceptName := conceptIdToName[conceptId]
		log.Printf("Generating row for concept name %s", conceptName)
		generatedRow := generateConceptRow(conceptName, conceptValuesToPeopleCount, sortedConceptValues)
		rows = append(rows, generatedRow)
	}

	return rows, nil
}

func (u ConceptController) RetrieveAttritionTable(c *gin.Context) {
	sourceId, cohortId, conceptIds, err1 := utils.ParseSourceIdAndCohortIdAndConceptIds(c)
	breakdownConceptId, err2 := utils.ParseBigNumericArg(c, "breakdownconceptid")

	if err1 == nil && err2 == nil {
		cohortName, err := cohortDefinitionModel.GetCohortName(cohortId)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohort name", "error": err.Error()})
			c.Abort()
			return
		}

		headerAndNonFilteredRow, err := u.GenerateHeaderAndNonFilteredRow(cohortName, sourceId, cohortId, breakdownConceptId)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown with filtered conceptIds", "error": err.Error()})
			c.Abort()
			return
		}

		header := headerAndNonFilteredRow[0]
		sortedConceptValues := header[2:]
		filteredRows, err := u.GetFilteredConceptRows(sourceId, cohortId, conceptIds, breakdownConceptId, sortedConceptValues)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown with filtered conceptIds", "error": err.Error()})
			c.Abort()
			return
		}

		b := GenerateAttritionCSV(headerAndNonFilteredRow, filteredRows)
		c.String(http.StatusOK, b.String())
		return
	}
	log.Printf("Error: bad request")
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
}

func getSortedConceptValues(conceptValuesToPeopleCount map[string]int) []string {
	concepts := []string{}
	for concept := range conceptValuesToPeopleCount {
		concepts = append(concepts, concept)
	}
	sort.Strings(concepts)
	return concepts
}

func (u ConceptController) GenerateHeaderAndNonFilteredRow(cohortName string, sourceId int, cohortId int, breakdownConceptId int64) ([][]string, error) {
	breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId, cohortId, breakdownConceptId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving stats due to error: %s", err.Error())
	}
	conceptValuesToPeopleCount := getConceptValueToPeopleCount(breakdownStats)

	cohortSize := 0
	for _, peopleCount := range conceptValuesToPeopleCount {
		cohortSize += peopleCount
	}

	conceptValues := getSortedConceptValues(conceptValuesToPeopleCount)
	header := append([]string{"Cohort", "Size"}, conceptValues...)
	row := []string{cohortName, strconv.Itoa(cohortSize)}
	for _, concept := range conceptValues {
		row = append(row, strconv.Itoa(conceptValuesToPeopleCount[concept]))
	}

	return [][]string{
		header,
		row,
	}, nil
}

func GenerateAttritionCSV(headerAndNonFilteredRow [][]string, filteredRows [][]string) *bytes.Buffer {
	var rows [][]string
	rows = append(rows, headerAndNonFilteredRow...)
	rows = append(rows, filteredRows...)

	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.Comma = ',' // CSV

	err := w.WriteAll(rows)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
