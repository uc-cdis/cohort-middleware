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
	"github.com/uc-cdis/cohort-middleware/middlewares"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type ConceptController struct {
	conceptModel          models.ConceptI
	cohortDefinitionModel models.CohortDefinitionI
	teamProjectAuthz      middlewares.TeamProjectAuthzI
}

func NewConceptController(conceptModel models.ConceptI, cohortDefinitionModel models.CohortDefinitionI, teamProjectAuthz middlewares.TeamProjectAuthzI) ConceptController {
	return ConceptController{
		conceptModel:          conceptModel,
		cohortDefinitionModel: cohortDefinitionModel,
		teamProjectAuthz:      teamProjectAuthz,
	}
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
	validAccessRequest := u.teamProjectAuthz.TeamProjectValidationForCohort(c, cohortId)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
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
	sourceId, cohortId, conceptDefsAndCohortPairs, err := utils.ParseSourceIdAndCohortIdAndVariablesAsSingleList(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}

	// parse cohortPairs separately as well, so we can validate permissions
	_, cohortPairs := utils.GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptDefsAndCohortPairs)

	validAccessRequest := u.teamProjectAuthz.TeamProjectValidation(c, []int{cohortId}, cohortPairs)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
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
	breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptDefsPlusCohortPairs(sourceId, cohortId, conceptDefsAndCohortPairs, breakdownConceptId)
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
	// validate:
	if variableName == "" {
		panic("unexpected error: variableName should be set!")
	}
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

func (u ConceptController) RetrieveAttritionTable(c *gin.Context) {
	sourceId, cohortId, conceptDefsAndCohortPairs, err := utils.ParseSourceIdAndCohortIdAndVariablesAsSingleList(c)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}
	_, cohortPairs := utils.GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptDefsAndCohortPairs)
	validAccessRequest := u.teamProjectAuthz.TeamProjectValidation(c, []int{cohortId}, cohortPairs)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown for given cohortId", "error": err.Error()})
		c.Abort()
		return
	}

	sortedConceptValues := getSortedConceptValues(breakdownStats)

	headerAndNonFilteredRow, err := u.GenerateHeaderAndNonFilteredRow(breakdownStats, sortedConceptValues, cohortName)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error generating concept breakdown header and cohort rows", "error": err.Error()})
		c.Abort()
		return
	}
	otherAttritionRows, err := u.GetAttritionRowForConceptDefsAndCohortPairs(sourceId, cohortId, conceptDefsAndCohortPairs, breakdownConceptId, sortedConceptValues)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown rows for filter conceptIds and cohortPairs", "error": err.Error()})
		c.Abort()
		return
	}
	b := GenerateAttritionCSV(headerAndNonFilteredRow, otherAttritionRows)
	c.String(http.StatusOK, b.String())
}

func (u ConceptController) GetAttritionRowForConceptDefsAndCohortPairs(sourceId int, cohortId int, conceptDefsAndCohortPairs []interface{}, breakdownConceptId int64, sortedConceptValues []string) ([][]string, error) {
	var otherAttritionRows [][]string
	for idx, conceptIdOrCohortPair := range conceptDefsAndCohortPairs {
		// attrition filter: run each query with an increasingly longer list of filterConceptDefsAndCohortPairs, until the last query is run with them all:
		filterConceptDefsAndCohortPairs := conceptDefsAndCohortPairs[0 : idx+1]

		attritionRow, err := u.GetAttritionRowForConceptDefOrCohortPair(sourceId, cohortId, conceptIdOrCohortPair, filterConceptDefsAndCohortPairs, breakdownConceptId, sortedConceptValues)
		if err != nil {
			log.Printf("Error: %s", err.Error())
			return nil, err
		}
		otherAttritionRows = append(otherAttritionRows, attritionRow)
	}
	return otherAttritionRows, nil
}

func (u ConceptController) GetAttritionRowForConceptDefOrCohortPair(sourceId int, cohortId int, conceptIdOrCohortPair interface{}, filterConceptDefsAndCohortPairs []interface{}, breakdownConceptId int64, sortedConceptValues []string) ([]string, error) {
	filterConceptDefsAndValues, filterCohortPairs := utils.GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(filterConceptDefsAndCohortPairs)
	filterConceptIds := utils.ExtractConceptIdsFromCustomConceptVariablesDef(filterConceptDefsAndValues)
	breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptDefsPlusCohortPairs(sourceId, cohortId, filterConceptDefsAndCohortPairs, breakdownConceptId)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve concept Breakdown for concepts %v dichotomous variables %v due to error: %s", filterConceptIds, filterCohortPairs, err.Error())
	}
	conceptValuesToPeopleCount := getConceptValueToPeopleCount(breakdownStats)
	variableName := ""
	switch convertedItem := conceptIdOrCohortPair.(type) {
	case utils.CustomConceptVariableDef:
		conceptInformation, err := u.conceptModel.RetrieveInfoBySourceIdAndConceptId(sourceId, convertedItem.ConceptId)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve concept details for %v due to error: %s", convertedItem, err.Error())
		}
		variableName = conceptInformation.ConceptName
	case utils.CustomDichotomousVariableDef:
		variableName = convertedItem.ProvidedName
	}
	log.Printf("Generating row for variable with name %s", variableName)
	generatedRow := generateRowForVariable(variableName, conceptValuesToPeopleCount, sortedConceptValues)
	return generatedRow, nil
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
