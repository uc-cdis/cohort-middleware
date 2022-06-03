package controllers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
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
	if err1 == nil && err2 == nil {
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

func (u ConceptController) GetFilteredConceptRows(sourceId int, cohortId int, conceptIds []int, breakdownConceptId int, sortedConceptValues []string) ([][]string, error) {
	conceptIdToName := make(map[int]string)
	conceptInformations, err := u.conceptModel.RetrieveInfoBySourceIdAndConceptIds(sourceId, conceptIds)
	if err != nil {
		return nil, fmt.Errorf("Could not retrieve concept informations due to error: %s", err.Error())
	}
	for _, conceptInformation := range conceptInformations {
		conceptIdToName[conceptInformation.ConceptId] = conceptInformation.ConceptName
	}

	var rows [][]string
	for idx, conceptId := range conceptIds {
		filterConceptIds := conceptIds[0 : idx+1]
		breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(sourceId, cohortId, filterConceptIds, breakdownConceptId)
		if err != nil {
			return nil, fmt.Errorf("Could not retrieve concept Breakdown for concepts %v due to error: %s", filterConceptIds, err.Error())
		}

		conceptValuesToPeopleCount := getConceptValueToPeopleCount(breakdownStats)
		generatedRow := generateConceptRow(conceptIdToName[conceptId], conceptValuesToPeopleCount, sortedConceptValues)
		rows = append(rows, generatedRow)
	}

	return rows, nil
}

func (u ConceptController) RetrieveAttritionTable(c *gin.Context) {
	sourceId, cohortId, conceptIds, err1 := parseSourceIdAndCohortIdAndConceptIds(c)
	breakdownConceptId, err2 := utils.ParseNumericArg(c, "breakdownconceptid")

	if err1 == nil && err2 == nil {
		cohortName, err := cohortDefinitionModel.GetCohortName(cohortId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving cohort name", "error": err})
			c.Abort()
			return
		}

		headerAndNonFilteredRow, err := u.GenerateHeaderAndNonFilteredRow(cohortName, sourceId, cohortId, breakdownConceptId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown with filtered conceptIds", "error": err})
			c.Abort()
			return
		}

		header := headerAndNonFilteredRow[0]
		sortedConceptValues := header[2:]
		filteredRows, err := u.GetFilteredConceptRows(sourceId, cohortId, conceptIds, breakdownConceptId, sortedConceptValues)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept breakdown with filtered conceptIds", "error": err})
			c.Abort()
			return
		}

		b := GenerateAttritionCSV(headerAndNonFilteredRow, filteredRows)
		c.String(http.StatusOK, b.String())
		return
	}

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

func (u ConceptController) GenerateHeaderAndNonFilteredRow(cohortName string, sourceId int, cohortId int, breakdownConceptId int) ([][]string, error) {
	breakdownStats, err := u.conceptModel.RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId, cohortId, breakdownConceptId)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving stats due to error: %s", err.Error())
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
