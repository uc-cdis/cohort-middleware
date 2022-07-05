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

func (u ConceptController) RetrieveInfoBySourceIdAndConceptIds(c *gin.Context) {

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

func (u ConceptController) RetrieveInfoBySourceIdAndConceptTypes(c *gin.Context) {

	sourceId, conceptTypes, err := utils.ParseSourceIdAndConceptTypes(c)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		c.Abort()
		return
	}

	// call model method:
	conceptInfo, err := u.conceptModel.RetrieveInfoBySourceIdAndConceptTypes(sourceId, conceptTypes)
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

func (u ConceptController) RetrieveBreakdownStatsBySourceIdAndCohortIdAndVariables(c *gin.Context) {
	sourceId, cohortId, conceptIds, cohortPairs, err1 := utils.ParseSourceIdAndCohortIdAndVariablesList(c)
	breakdownConceptId, err2 := utils.ParseBigNumericArg(c, "breakdownconceptid")
	if err1 == nil && err2 == nil {
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId, cohortId, conceptIds, cohortPairs, breakdownConceptId)
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

func generateRowForVariable(variableName string, breakdownConceptValuesToPeopleCount map[string]int, sortedBreakdownConceptValues []string) []string {
	cohortSize := 0
	for _, peopleCount := range breakdownConceptValuesToPeopleCount {
		cohortSize += peopleCount
	}

	row := []string{variableName, strconv.Itoa(cohortSize)}
	// make sure the numbers are printed out in the right order:
	for _, concept := range sortedBreakdownConceptValues {
		row = append(row, strconv.Itoa(breakdownConceptValuesToPeopleCount[concept]))
	}

	return row
}

func (u ConceptController) GetConceptVariablesAttritionRows(sourceId int, cohortId int, conceptIds []int64, breakdownConceptId int64, sortedConceptValues []string) ([][]string, error) {
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
		// use empty cohort pairs list:
		filterCohortPairs := make([][]int, 0)
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId, cohortId, filterConceptIds, filterCohortPairs, breakdownConceptId)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve concept Breakdown for concepts %v due to error: %s", filterConceptIds, err.Error())
		}

		conceptValuesToPeopleCount := getConceptValueToPeopleCount(breakdownStats)
		variableName := conceptIdToName[conceptId]
		log.Printf("Generating row for variable with name %s", variableName)
		generatedRow := generateRowForVariable(variableName, conceptValuesToPeopleCount, sortedConceptValues)
		rows = append(rows, generatedRow)
	}

	return rows, nil
}

func (u ConceptController) GetCustomDichotomousVariablesAttritionRows(sourceId int, cohortId int, filterConceptIds []int64, filterCohortPairs [][]int, breakdownConceptId int64, sortedConceptValues []string) ([][]string, error) {
	var rows [][]string
	for idx, cohortPair := range filterCohortPairs {
		// run each query with the full list of filterConceptIds and an increasingly longer list of filterCohortPairs, until the last query is run with them all:
		filterCohortPairs := filterCohortPairs[0 : idx+1]
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId, cohortId, filterConceptIds, filterCohortPairs, breakdownConceptId)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve concept Breakdown for concepts %v due to error: %s", filterConceptIds, err.Error())
		}

		conceptValuesToPeopleCount := getConceptValueToPeopleCount(breakdownStats)
		variableName := models.GetCohortPairKey(cohortPair[0], cohortPair[1])
		log.Printf("Generating row for variable...")
		generatedRow := generateRowForVariable(variableName, conceptValuesToPeopleCount, sortedConceptValues)
		rows = append(rows, generatedRow)
	}

	return rows, nil
}

func (u ConceptController) RetrieveAttritionTable(c *gin.Context) {
	sourceId, cohortId, conceptIds, cohortPairs, err1 := utils.ParseSourceIdAndCohortIdAndVariablesList(c)
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
		// append concepts to attrition table:
		conceptVariablesAttritionRows, err := u.GetConceptVariablesAttritionRows(sourceId, cohortId, conceptIds, breakdownConceptId, sortedConceptValues)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown with filtered conceptIds", "error": err.Error()})
			c.Abort()
			return
		}
		// append custom dichotomous items to attrition table:
		customDichotomousVariablesAttritionRows, err := u.GetCustomDichotomousVariablesAttritionRows(sourceId, cohortId, conceptIds, cohortPairs, breakdownConceptId, sortedConceptValues)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown with custom dichotomous variables (aka cohortpairs)", "error": err.Error()})
			c.Abort()
			return
		}

		// concat all rows:
		var allVariablesAttritionRows = conceptVariablesAttritionRows
		allVariablesAttritionRows = append(allVariablesAttritionRows, customDichotomousVariablesAttritionRows...)
		b := GenerateAttritionCSV(headerAndNonFilteredRow, allVariablesAttritionRows)
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
