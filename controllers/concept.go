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

type ConceptController struct {
	conceptModel          models.ConceptI
	cohortDefinitionModel models.CohortDefinitionI
}

func NewConceptController(conceptModel models.ConceptI, cohortDefinitionModel models.CohortDefinitionI) ConceptController {
	return ConceptController{conceptModel: conceptModel, cohortDefinitionModel: cohortDefinitionModel}
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
	sourceId, cohortId, err := utils.ParseSourceAndCohortId(c)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}
	breakdownConceptId, err := utils.ParseBigNumericArg(c, "breakdownconceptid")
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}
	breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId, cohortId, breakdownConceptId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"concept_breakdown": breakdownStats})
}

func (u ConceptController) RetrieveBreakdownStatsBySourceIdAndCohortIdAndVariables(c *gin.Context) {
	sourceId, cohortId, conceptIds, cohortPairs, err := utils.ParseSourceIdAndCohortIdAndVariablesList(c)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}
	breakdownConceptId, err := utils.ParseBigNumericArg(c, "breakdownconceptid")
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}
	breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId, cohortId, conceptIds, cohortPairs, breakdownConceptId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"concept_breakdown": breakdownStats})
}

func getConceptValueToPeopleCount(breakdownStats []*models.ConceptBreakdown) map[string]int {
	conceptValuesToPeopleCount := make(map[string]int)
	for _, breakdownStat := range breakdownStats {
		concept_value := breakdownStat.ConceptValue
		if concept_value == "" {
			concept_value = "empty string"
		}
		conceptValuesToPeopleCount[concept_value] = breakdownStat.NpersonsInCohortWithValue
	}

	return conceptValuesToPeopleCount
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
		filterCohortPairs := []utils.CustomDichotomousVariableDef{}
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

func (u ConceptController) GetCustomDichotomousVariablesAttritionRows(sourceId int, cohortId int, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef, breakdownConceptId int64, sortedConceptValues []string) ([][]string, error) {
	// TODO - this function is very similar to GetConceptVariablesAttritionRows above and they can probably be merged.
	var rows [][]string
	for idx, cohortPair := range filterCohortPairs {
		// run each query with the full list of filterConceptIds and an increasingly longer list of filterCohortPairs, until the last query is run with them all:
		filterCohortPairs := filterCohortPairs[0 : idx+1]
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId, cohortId, filterConceptIds, filterCohortPairs, breakdownConceptId)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve concept Breakdown for dichotomous variables %v due to error: %s", filterConceptIds, err.Error())
		}

		conceptValuesToPeopleCount := getConceptValueToPeopleCount(breakdownStats)
		variableName := models.GetCohortPairKey(cohortPair.CohortId1, cohortPair.CohortId2)
		log.Printf("Generating row for variable...")
		generatedRow := generateRowForVariable(variableName, conceptValuesToPeopleCount, sortedConceptValues)
		rows = append(rows, generatedRow)
	}

	return rows, nil
}

func (u ConceptController) RetrieveAttritionTable(c *gin.Context) {
	sourceId, cohortId, conceptIds, cohortPairs, err := utils.ParseSourceIdAndCohortIdAndVariablesList(c)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}
	breakdownConceptId, err := utils.ParseBigNumericArg(c, "breakdownconceptid")
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}
	cohortName, err := u.cohortDefinitionModel.GetCohortName(cohortId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohort name", "error": err.Error()})
		c.Abort()
		return
	}

	breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId, cohortId, breakdownConceptId)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown with filtered conceptIds", "error": err.Error()})
		c.Abort()
		return
	}

	sortedConceptValues := getSortedConceptValues(breakdownStats)

	headerAndNonFilteredRow, err := u.GenerateHeaderAndNonFilteredRow(breakdownStats, sortedConceptValues, cohortName)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown with filtered conceptIds", "error": err.Error()})
		c.Abort()
		return
	}

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
	var allVariablesAttritionRows = append(conceptVariablesAttritionRows, customDichotomousVariablesAttritionRows...)
	b := GenerateAttritionCSV(headerAndNonFilteredRow, allVariablesAttritionRows)
	c.String(http.StatusOK, b.String())
}

func getSortedConceptValues(breakdownStats []*models.ConceptBreakdown) []string {
	concepts := []string{}
	for _, breakdownStat := range breakdownStats {
		concepts = append(concepts, breakdownStat.ConceptValue)
	}
	sort.Strings(concepts)
	return concepts
}

func getConceptValueToConceptName(breakdownStats []*models.ConceptBreakdown) map[string]string {
	conceptValuesToConceptName := make(map[string]string)
	for _, breakdownStat := range breakdownStats {
		conceptValue := breakdownStat.ConceptValue
		conceptName := breakdownStat.ValueName
		if conceptValue == "" {
			conceptValue = "empty string"
			conceptName = "empty string"
		}

		conceptValuesToConceptName[conceptValue] = conceptName
	}

	return conceptValuesToConceptName
}

func getConceptNamesFromConceptValues(conceptValues []string, conceptValuesToConceptName map[string]string) []string {
	conceptNames := []string{}

	for _, conceptValue := range conceptValues {
		conceptNames = append(conceptNames, conceptValuesToConceptName[conceptValue])
	}

	return conceptNames
}

func (u ConceptController) GenerateHeaderAndNonFilteredRow(breakdownStats []*models.ConceptBreakdown, sortedConceptValues []string, cohortName string) ([][]string, error) {
	conceptValuesToPeopleCount := getConceptValueToPeopleCount(breakdownStats)
	conceptValuesToConceptName := getConceptValueToConceptName(breakdownStats)

	cohortSize := 0
	for _, peopleCount := range conceptValuesToPeopleCount {
		cohortSize += peopleCount
	}

	conceptNames := getConceptNamesFromConceptValues(sortedConceptValues, conceptValuesToConceptName)
	header := append([]string{"Cohort", "Size"}, conceptNames...)
	row := []string{cohortName, strconv.Itoa(cohortSize)}
	for _, concept := range sortedConceptValues {
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
