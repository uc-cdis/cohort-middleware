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
	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortDataController struct {
	cohortDataModel models.CohortDataI
}

type PrefixedConceptIds struct {
	PrefixedConceptIds []string
}

func NewCohortDataController(cohortDataModel models.CohortDataI) CohortDataController {
	return CohortDataController{cohortDataModel: cohortDataModel}
}

func (u CohortDataController) RetrieveDataBySourceIdAndCohortIdAndConceptIds(c *gin.Context) {
	// TODO - add some validation to ensure that only calls from Argo are allowed through since it outputs FULL data?

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
	var prefixedConceptIds PrefixedConceptIds
	err := decoder.Decode(&prefixedConceptIds)
	if err != nil {
		log.Printf("Error: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request - no valid request body"})
		c.Abort()
		return
	}
	log.Printf("Querying concept ids: %v", prefixedConceptIds.PrefixedConceptIds)

	sourceId, _ := strconv.Atoi(sourceIdStr)
	cohortId, _ := strconv.Atoi(cohortIdStr)
	conceptIds := getConceptIdsFromPrefixedConceptIds(prefixedConceptIds.PrefixedConceptIds)

	// call model method:
	cohortData, err := u.cohortDataModel.RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId, cohortId, conceptIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err.Error()})
		c.Abort()
		return
	}
	b := GenerateCSV(sourceId, cohortData, conceptIds)
	c.String(http.StatusOK, b.String())
}

func getConceptIdsFromPrefixedConceptIds(ids []string) []int64 {
	var result []int64
	for _, id := range ids {
		idAsNumber := models.GetConceptId(id)
		result = append(result, idAsNumber)
	}
	return result
}

// This function will take the given cohort data and transform it into a matrix
// that contains the person id as the first column and the concept values found
// for this person in the subsequent columns. The transformation is necessary
// since the cohortData list contains one row per person-concept combination.
// E.g. the following (simplified version of the) data:
//   {PersonId:1, ConceptId:1, ConceptValue: A},
//   {PersonId:1, ConceptId:2, ConceptValue: B},
//   {PersonId:2, ConceptId:1, ConceptValue: C},
// will be transformed to a CSV table like:
//   sample.id,ID_concept_id1,ID_concept_id2
//   1,"A value with, comma!",B
//   2,Simple value,NA
// where "NA" means that the person did not have a data element for that concept
// or that the data element had a NULL/empty value.
func GenerateCSV(sourceId int, cohortData []*models.PersonConceptAndValue, conceptIds []int64) *bytes.Buffer {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.Comma = ',' // CSV

	var rows [][]string
	var header []string
	header = append(header, "sample.id")
	header = addConceptsToHeader(sourceId, header, conceptIds)
	rows = append(rows, header)

	var currentPersonId int64 = -1
	var row []string
	for _, cohortDatum := range cohortData {
		// if new person, start new row:
		if cohortDatum.PersonId != currentPersonId {
			if currentPersonId != -1 {
				rows = append(rows, row)
			}
			row = []string{}
			row = append(row, strconv.FormatInt(cohortDatum.PersonId, 10))
			row = appendInitEmptyConceptValues(row, len(conceptIds))
			currentPersonId = cohortDatum.PersonId
		}
		row = populateConceptValue(row, *cohortDatum, conceptIds)
	}
	// append last person row:
	rows = append(rows, row)

	// TODO - is there a way to write as the rows are produced? Building up all rows in memory
	// could cause issues if the cohort vs concepts matrix gets very large...or will the number of concepts
	// queried at the same time never be very large? Should we restrict the number of concepts to
	// a max here in this method?
	err := w.WriteAll(rows)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func addConceptsToHeader(sourceId int, header []string, conceptIds []int64) []string {
	for i := 0; i < len(conceptIds); i++ {
		//var conceptName = getConceptName(sourceId, conceptIds[i]) // instead of name, we now prefer ID_concept_id...below:
		var conceptPrefixedId = models.GetPrefixedConceptId(conceptIds[i])
		header = append(header, conceptPrefixedId)
	}
	return header
}

func appendInitEmptyConceptValues(row []string, nrConceptIds int) []string {
	for i := 0; i < nrConceptIds; i++ {
		row = append(row, "NA")
	}
	return row
}

func populateConceptValue(row []string, cohortItem models.PersonConceptAndValue, conceptIds []int64) []string {
	var conceptIdIdx int = utils.Pos(cohortItem.ConceptId, conceptIds)
	if conceptIdIdx != -1 {
		// conceptIdIdx+1 because first column is sample.id:
		conceptIdxInRow := conceptIdIdx + 1
		if cohortItem.ConceptValueAsString != "" {
			row[conceptIdxInRow] = cohortItem.ConceptValueAsString
		} else if cohortItem.ConceptValueAsNumber != 0.0 {
			row[conceptIdxInRow] = strconv.FormatFloat(float64(cohortItem.ConceptValueAsNumber), 'f', 2, 64)
		}
	}
	return row
}

func (u CohortDataController) RetrieveCohortOverlapStats(c *gin.Context) {
	errors := make([]error, 5)
	var sourceId, caseCohortId, controlCohortId int
	var filterConceptId int64
	var filterConceptValue string
	var conceptIds []int64
	sourceId, conceptIds, errors[0] = utils.ParseSourceIdAndConceptIds(c)
	filterConceptId, errors[1] = utils.ParseBigNumericArg(c, "filterconceptid")
	filterConceptValue, errors[2] = utils.ParseStringArg(c, "filtervalue")
	caseCohortId, errors[3] = utils.ParseNumericArg(c, "casecohortid")
	controlCohortId, errors[4] = utils.ParseNumericArg(c, "controlcohortid")
	if utils.ContainsNonNil(errors) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		c.Abort()
		return
	}
	breakdownStats, err := u.cohortDataModel.RetrieveCohortOverlapStats(sourceId, caseCohortId, controlCohortId,
		filterConceptId, filterConceptValue, conceptIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"cohort_overlap": breakdownStats})
}
