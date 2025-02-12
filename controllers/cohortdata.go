package controllers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/middlewares"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortDataController struct {
	cohortDataModel     models.CohortDataI
	dataDictionaryModel models.DataDictionaryI
	teamProjectAuthz    middlewares.TeamProjectAuthzI
}

func NewCohortDataController(cohortDataModel models.CohortDataI, dataDictionaryModel models.DataDictionaryI, teamProjectAuthz middlewares.TeamProjectAuthzI) CohortDataController {
	return CohortDataController{
		cohortDataModel:     cohortDataModel,
		dataDictionaryModel: dataDictionaryModel,
		teamProjectAuthz:    teamProjectAuthz,
	}
}

func (u CohortDataController) RetrieveHistogramForCohortIdAndConceptId(c *gin.Context) {
	sourceId, cohortId, conceptIdsAndCohortPairs, err := utils.ParseSourceIdAndCohortIdAndVariablesAsSingleList(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}

	// parse cohortPairs separately as well, so we can validate permissions
	_, cohortPairs := utils.GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs)

	validAccessRequest := u.teamProjectAuthz.TeamProjectValidation(c, []int{cohortId}, cohortPairs)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
		c.Abort()
		return
	}

	cohortData, err := u.cohortDataModel.RetrieveHistogramDataBySourceIdAndCohortIdAndConceptDefsPlusCohortPairs(sourceId, cohortId, conceptIdsAndCohortPairs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving histogram data", "error": err.Error()})
		c.Abort()
		return
	}

	conceptValues := []float64{}
	for _, personData := range cohortData {
		conceptValues = append(conceptValues, float64(*personData.ConceptValueAsNumber))
	}

	histogramData := utils.GenerateHistogramData(conceptValues)

	c.JSON(http.StatusOK, gin.H{"bins": histogramData})
}

func (u CohortDataController) RetrieveStatsForCohortIdAndConceptId(c *gin.Context) {
	sourceId, cohortId, conceptIdsAndCohortPairs, err := utils.ParseSourceIdAndCohortIdAndVariablesAsSingleList(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}

	// parse cohortPairs separately as well, so we can validate permissions
	_, cohortPairs := utils.GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs)

	validAccessRequest := u.teamProjectAuthz.TeamProjectValidation(c, []int{cohortId}, cohortPairs)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
		c.Abort()
		return
	}

	cohortData, err := u.cohortDataModel.RetrieveHistogramDataBySourceIdAndCohortIdAndConceptDefsPlusCohortPairs(sourceId, cohortId, conceptIdsAndCohortPairs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err.Error()})
		c.Abort()
		return
	}

	conceptValues := []float64{}
	for _, personData := range cohortData {
		conceptValues = append(conceptValues, float64(*personData.ConceptValueAsNumber))
	}
	conceptToStat, errGetLast := utils.CheckAndGetLastCustomConceptVariableDef(conceptIdsAndCohortPairs)
	if errGetLast != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error: last variable should be of numeric type", "error": errGetLast.Error()})
		c.Abort()
		return
	}

	statsData := utils.GenerateStatsData(cohortId, conceptToStat.ConceptId, conceptValues)

	c.JSON(http.StatusOK, gin.H{"statsData": statsData})
}

func (u CohortDataController) RetrieveDataBySourceIdAndCohortIdAndVariables(c *gin.Context) {
	// TODO - add some validation to ensure that only calls from Argo are allowed through since it outputs FULL data?
	//  -> this concern is considered to be addressed by https://github.com/uc-cdis/cloud-automation/pull/1884
	sourceId, cohortId, conceptDefsAndCohortPairs, err := utils.ParseSourceIdAndCohortIdAndVariablesAsSingleList(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request", "error": err.Error()})
		c.Abort()
		return
	}

	// parse cohortPairs separately as well, so we can validate permissions
	conceptDefs, cohortPairs := utils.GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptDefsAndCohortPairs)

	validAccessRequest := u.teamProjectAuthz.TeamProjectValidation(c, []int{cohortId}, cohortPairs)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
		c.Abort()
		return
	}

	// Iterate over conceptDefsAndCohortPairs and collect the concept values for each person:
	//	{PersonId:1, ConceptId:1, ConceptValue: "A value with, comma!"},
	//	{PersonId:1, ConceptId:2, ConceptValue: B},
	//	{PersonId:2, ConceptId:1, ConceptValue: C},
	var variablesToQuery []interface{}
	var finalConceptDataset []*models.PersonConceptAndValue
	for _, item := range conceptDefsAndCohortPairs {
		variablesToQuery = append(variablesToQuery, item)
		// if item is of type CustomConceptVariableDef, get the data:
		if _, ok := item.(utils.CustomConceptVariableDef); ok {
			// use variablesToQuery to query an increasingly tight set (simulating the attrition table that generated this query)
			cohortData, err := u.cohortDataModel.RetrieveHistogramDataBySourceIdAndCohortIdAndConceptDefsPlusCohortPairs(sourceId, cohortId, variablesToQuery)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving concept details", "error": err.Error()})
				c.Abort()
				return
			}
			// add to final concept data set:
			finalConceptDataset = append(finalConceptDataset, cohortData...)
		}
	}

	conceptIds := utils.ExtractConceptIdsFromCustomConceptVariablesDef(conceptDefs)
	partialCSV := GeneratePartialCSV(sourceId, finalConceptDataset, conceptIds) // use conceptdefs to improve column description? nah...no person is reading this table....just needs to be unique

	personIdToCSVValues, err := u.RetrievePeopleIdAndCohort(sourceId, cohortId, cohortPairs, finalConceptDataset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving people ID to csv value map", "error": err.Error()})
		c.Abort()
		return
	}

	b := GenerateCompleteCSV(partialCSV, personIdToCSVValues, cohortPairs)
	c.String(http.StatusOK, b.String())

}

func generateCohortPairsHeaders(cohortPairs []utils.CustomDichotomousVariableDef) []string {
	cohortPairsHeaders := []string{}

	for _, cohortPair := range cohortPairs {
		cohortPairsHeaders = append(cohortPairsHeaders, utils.GetCohortPairKey(cohortPair.CohortDefinitionId1, cohortPair.CohortDefinitionId2))
	}

	return cohortPairsHeaders
}

func GenerateCompleteCSV(partialCSV [][]string, personIdToCSVValues map[int64]map[string]string, cohortPairs []utils.CustomDichotomousVariableDef) *bytes.Buffer {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.Comma = ',' // CSV

	cohortPairHeaders := generateCohortPairsHeaders(cohortPairs)

	partialCSV[0] = append(partialCSV[0], cohortPairHeaders...)

	for i := 1; i < len(partialCSV); i++ {
		personId, _ := strconv.ParseInt(partialCSV[i][0], 10, 64)
		for _, cohortPair := range cohortPairHeaders {
			partialCSV[i] = append(partialCSV[i], personIdToCSVValues[personId][cohortPair])
		}
	}

	// TODO - is there a way to write as the rows are produced? Building up all rows in memory
	// could cause issues if the cohort vs concepts matrix gets very large...or will the number of concepts
	// queried at the same time never be very large? Should we restrict the number of concepts to
	// a max here in this method?
	err := w.WriteAll(partialCSV)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

// This function will take the given cohort data and transform it into a matrix
// that contains the person id as the first column and the concept values found
// for this person in the subsequent columns. The transformation is necessary
// since the cohortData list contains one row per person-concept combination.
// E.g. the following (simplified version of the) data:
//
//	{PersonId:1, ConceptId:1, ConceptValue: "A value with, comma!"},
//	{PersonId:1, ConceptId:2, ConceptValue: B},
//	{PersonId:2, ConceptId:1, ConceptValue: C},
//
// will be transformed to a CSV table like:
//
//	sample.id,ID_concept_id1,ID_concept_id2
//	1,"A value with, comma!",B
//	2,Simple value,NA
//
// where "NA" means that the person did not have a data element for that concept
// or that the data element had a NULL/empty value.
func GeneratePartialCSV(sourceId int, cohortData []*models.PersonConceptAndValue, conceptIds []int64) [][]string {
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
	return rows
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
		if cohortItem.ConceptClassId == "MVP Continuous" {
			if cohortItem.ConceptValueAsNumber != nil {
				row[conceptIdxInRow] = strconv.FormatFloat(float64(*cohortItem.ConceptValueAsNumber), 'f', 2, 64)
			}
		} else {
			// default to the ObservationValueAsConceptName for now:
			if cohortItem.ObservationValueAsConceptName != "" {
				row[conceptIdxInRow] = cohortItem.ObservationValueAsConceptName
			}
		}
	}
	return row
}

func (u CohortDataController) RetrieveCohortOverlapStats(c *gin.Context) {
	errors := make([]error, 4)
	var sourceId, caseCohortId, controlCohortId int
	var conceptDefsAndCohortPairs []interface{}
	sourceId, errors[0] = utils.ParseNumericArg(c, "sourceid")
	caseCohortId, errors[1] = utils.ParseNumericArg(c, "casecohortid")
	controlCohortId, errors[2] = utils.ParseNumericArg(c, "controlcohortid")
	conceptDefsAndCohortPairs, errors[3] = utils.ParseConceptDefsAndDichotomousDefsAsSingleList(c)
	_, cohortPairs := utils.GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptDefsAndCohortPairs)

	validAccessRequest := u.teamProjectAuthz.TeamProjectValidation(c, []int{caseCohortId, controlCohortId}, cohortPairs)
	if !validAccessRequest {
		log.Printf("Error: invalid request")
		c.JSON(http.StatusForbidden, gin.H{"message": "access denied"})
		c.Abort()
		return
	}

	if utils.ContainsNonNil(errors) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		c.Abort()
		return
	}
	overlapStats, err := u.cohortDataModel.RetrieveCohortOverlapStats(sourceId, caseCohortId,
		controlCohortId, conceptDefsAndCohortPairs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error retrieving stats", "error": err.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, gin.H{"cohort_overlap": overlapStats})
}

func convertCohortPeopleDataToMap(cohortPeopleData []*models.PersonIdAndCohort) map[int64]int64 {
	personIdToCohortDefinitionId := make(map[int64]int64)

	for _, cohortPersonData := range cohortPeopleData {
		personIdToCohortDefinitionId[cohortPersonData.PersonId] = cohortPersonData.CohortId
	}

	return personIdToCohortDefinitionId
}

func generateCohortPairCSVValue(personId int64, firstCohortValue int64, secondCohortValue int64) string {
	if firstCohortValue == 0 && secondCohortValue == 0 {
		return "NA" // the person is not in either cohort
	}

	if firstCohortValue > 0 && secondCohortValue > 0 {
		log.Printf("person with id %v has an overlap and is in cohort %v and cohort %v", personId, firstCohortValue, secondCohortValue)
		return "NA" // the person is overlapped
	}

	if firstCohortValue > 0 {
		return "0" // the person belongs to the first cohort
	}

	if secondCohortValue > 0 {
		return "1" // the person belongs to the second cohort
	}

	log.Printf("error with personId %v with first cohort value %v and second cohort value %v", personId, firstCohortValue, secondCohortValue)
	return "NA"
}

func getAllPeopleIdInCohortData(cohortData []*models.PersonConceptAndValue) []int64 {
	var personIds []int64
	for _, data := range cohortData {
		personIds = append(personIds, data.PersonId)
	}

	return personIds
}

func (u CohortDataController) RetrievePeopleIdAndCohort(sourceId int, cohortId int, cohortPairs []utils.CustomDichotomousVariableDef, cohortData []*models.PersonConceptAndValue) (map[int64]map[string]string, error) {
	peopleIds := getAllPeopleIdInCohortData(cohortData)

	/**
	makes a map of {
		"{person_id}" : {
			"{first_cohort}_{second_cohort}": "{csv_value}"
		}
	}
	*/
	personIdToCSVValues := make(map[int64]map[string]string)
	for _, cohortPair := range cohortPairs {
		firstCohortDefinitionId := cohortPair.CohortDefinitionId1
		secondCohortDefinitionId := cohortPair.CohortDefinitionId2
		cohortPairKey := utils.GetCohortPairKey(firstCohortDefinitionId, secondCohortDefinitionId)

		firstCohortPeopleData, err1 := u.cohortDataModel.RetrieveDataByOriginalCohortAndNewCohort(sourceId, cohortId, firstCohortDefinitionId)
		secondCohortPeopleData, err2 := u.cohortDataModel.RetrieveDataByOriginalCohortAndNewCohort(sourceId, cohortId, secondCohortDefinitionId)
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("getting cohort people data failed")
		}
		firstCohortPeopleMap := convertCohortPeopleDataToMap(firstCohortPeopleData)
		secondCohortPeopleMap := convertCohortPeopleDataToMap(secondCohortPeopleData)

		for _, personId := range peopleIds {
			CSVValue := generateCohortPairCSVValue(personId, firstCohortPeopleMap[personId], secondCohortPeopleMap[personId])
			_, exists := personIdToCSVValues[personId]
			if exists {
				personIdToCSVValues[personId][cohortPairKey] = CSVValue
			} else {
				personIdToCSVValues[personId] = map[string]string{cohortPairKey: CSVValue}
			}
		}
	}

	return personIdToCSVValues, nil
}

func (u CohortDataController) RetrieveDataDictionary(c *gin.Context) {

	var dataDictionary, error = u.dataDictionaryModel.GetDataDictionary()

	if dataDictionary == nil {
		c.JSON(http.StatusServiceUnavailable, error)
	} else {
		c.JSON(http.StatusOK, dataDictionary)
	}

}

func (u CohortDataController) GenerateDataDictionary(c *gin.Context) {
	log.Printf("Generating Data Dictionary...")
	go u.dataDictionaryModel.GenerateDataDictionary()
	c.JSON(http.StatusOK, "Data Dictionary Kicked Off")
}
