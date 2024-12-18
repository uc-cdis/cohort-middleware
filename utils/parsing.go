package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseNumericArg(c *gin.Context, paramName string) (int, error) {
	// parse and validate:
	numericArgValue := c.Param(paramName)
	log.Printf("Querying %s: ", paramName)
	if numericId, err := strconv.Atoi(numericArgValue); err != nil {
		log.Printf("bad request - %s should be a number", paramName)
		return -1, fmt.Errorf("bad request - %s should be a number", paramName)
	} else {
		return numericId, nil
	}
}

func ParseBigNumericArg(c *gin.Context, paramName string) (int64, error) {
	// parse and validate:
	numericArgValue := c.Param(paramName)
	log.Printf("Querying %s: ", paramName)
	if numericId, err := strconv.ParseInt(numericArgValue, 10, 64); err != nil {
		log.Printf("bad request - %s should be a number", paramName)
		return -1, fmt.Errorf("bad request - %s should be a number", paramName)
	} else {
		return numericId, nil
	}
}

func Pos(value int64, list []int64) int {
	for p, v := range list {
		if v == value {
			return p
		}
	}
	return -1
}

func ContainsString(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func Contains(list []int, value int) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func ParseInt64(strValue string) int64 {
	value, error := strconv.ParseInt(strValue, 10, 64)
	if error != nil {
		panic(fmt.Sprintf("Invalid numeric value. Error: %v", error))
	}
	return value
}

func ContainsNonNil(errors []error) bool {
	for _, item := range errors {
		if item != nil {
			return true
		}
	}
	return false
}

// Takes in a list of strings and attempts to transform them to int64,
// returning the resulting list of int64 values. Will fail if
// any of the strings cannot be parsed as a number.
func SliceAtoi(stringValues []string) ([]int64, error) {
	intValues := make([]int64, 0, len(stringValues))
	for _, a := range stringValues {
		i, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			return intValues, err
		}
		intValues = append(intValues, i)
	}
	return intValues, nil
}

type ConceptIds struct {
	ConceptIds []int64
}

type ConceptTypes struct {
	ConceptTypes []string
}

// fields that define a custom dichotomous variable:
type CustomDichotomousVariableDef struct {
	CohortDefinitionId1 int
	CohortDefinitionId2 int
	ProvidedName        string
}

type CustomConceptVariableDef struct {
	ConceptId     int64
	ConceptValues []int64
}

func GetCohortPairKey(firstCohortDefinitionId int, secondCohortDefinitionId int) string {
	return fmt.Sprintf("ID_%v_%v", firstCohortDefinitionId, secondCohortDefinitionId)
}

// This method expects a request body with a payload similar to the following example:
// {"variables": [
//
//	{variable_type: "concept", concept_id: 2000000324},
//	{variable_type: "concept", concept_id: 2000006885},
//	{variable_type: "custom_dichotomous", provided_name: "name1", cohort_ids: [cohortX_id, cohortY_id]},
//	{variable_type: "custom_dichotomous", provided_name: "name2", cohort_ids: [cohortM_id, cohortN_id]},
//	    ...
//
// ]}
// It returns the list with all concept_id values and custom dichotomous variable definitions.
func ParseConceptIdsAndDichotomousDefsAsSingleList(c *gin.Context) ([]interface{}, error) {
	if c.Request == nil || c.Request.Body == nil {
		return nil, errors.New("bad request - no request body")
	}
	decoder := json.NewDecoder(c.Request.Body)
	request := make(map[string][]map[string]interface{})
	err := decoder.Decode(&request)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}

	variables := request["variables"]
	conceptIdsAndCohortPairs := make([]interface{}, 0)

	// TODO - this parsing will throw a lot of "null pointer" errors since it does not validate if specific entries are found in the json before
	// accessing them...needs to be fixed to throw better errors:
	for _, variable := range variables {
		if variable["variable_type"] == "concept" {
			convertedConceptValues := []int64{}
			values, ok := variable["values"].([]interface{})
			// If the values are passed as parameter, add to list
			if ok {
				for _, val := range values {
					convertedConceptValues = append(convertedConceptValues, int64(val.(float64)))
				}
			}
			conceptVariableDef := CustomConceptVariableDef{
				ConceptId:     int64(variable["concept_id"].(float64)),
				ConceptValues: convertedConceptValues,
			}
			conceptIdsAndCohortPairs = append(conceptIdsAndCohortPairs, conceptVariableDef)
		}
		if variable["variable_type"] == "custom_dichotomous" {
			cohortPair := []int{}
			convertedCohortIds := variable["cohort_ids"].([]interface{})
			for _, convertedCohortId := range convertedCohortIds {
				cohortPair = append(cohortPair, int(convertedCohortId.(float64)))
			}
			providedName := GetCohortPairKey(cohortPair[0], cohortPair[1])
			if variable["provided_name"] != nil {
				providedName = variable["provided_name"].(string)
			}
			customDichotomousVariableDef := CustomDichotomousVariableDef{
				CohortDefinitionId1: cohortPair[0],
				CohortDefinitionId2: cohortPair[1],
				ProvidedName:        providedName,
			}
			conceptIdsAndCohortPairs = append(conceptIdsAndCohortPairs, customDichotomousVariableDef)
		}
	}
	return conceptIdsAndCohortPairs, nil
}

// deprecated: for backwards compatibility
func ParseConceptIdsAndDichotomousDefs(c *gin.Context) ([]CustomConceptVariableDef, []CustomDichotomousVariableDef, error) {
	conceptIdsAndCohortPairs, err := ParseConceptIdsAndDichotomousDefsAsSingleList(c)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, nil, err
	}
	conceptIds, cohortPairs := GetConceptIdsAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs)
	return conceptIds, cohortPairs, nil
}

func ParseSourceIdAndConceptIds(c *gin.Context) (int, []int64, error) {
	// parse and validate all parameters:
	sourceId, err1 := ParseNumericArg(c, "sourceid")
	if err1 != nil {
		return -1, nil, err1
	}
	if c.Request == nil || c.Request.Body == nil {
		return -1, nil, errors.New("bad request - no request body")
	}
	decoder := json.NewDecoder(c.Request.Body)
	var conceptIds ConceptIds
	err := decoder.Decode(&conceptIds)
	if err != nil {
		return -1, nil, errors.New("bad request - no request body")
	}
	log.Printf("Querying concept ids...")
	if len(conceptIds.ConceptIds) == 0 {
		return -1, nil, errors.New("bad request - no concept ids in body")
	}

	return sourceId, conceptIds.ConceptIds, nil
}

func ParseSourceIdAndConceptTypes(c *gin.Context) (int, []string, error) {
	// parse and validate all parameters:
	sourceId, err1 := ParseNumericArg(c, "sourceid")
	if err1 != nil {
		return -1, nil, err1
	}
	if c.Request == nil || c.Request.Body == nil {
		return -1, nil, errors.New("bad request - no request body")
	}
	decoder := json.NewDecoder(c.Request.Body)
	var conceptTypes ConceptTypes
	err := decoder.Decode(&conceptTypes)
	if err != nil {
		return -1, nil, errors.New("bad request - no request body")
	}
	log.Printf("Querying concept types...")
	if len(conceptTypes.ConceptTypes) == 0 {
		return -1, nil, errors.New("bad request - no concept types in body")
	}

	return sourceId, conceptTypes.ConceptTypes, nil
}

func ParseSourceIdAndCohortIdAndConceptIds(c *gin.Context) (int, int, []int64, error) {
	// parse and validate all parameters:
	sourceId, conceptIds, err1 := ParseSourceIdAndConceptIds(c)
	if err1 != nil {
		return -1, -1, nil, err1
	}
	cohortId, err2 := ParseNumericArg(c, "cohortid")
	if err2 != nil {
		return -1, -1, nil, err2
	}
	return sourceId, cohortId, conceptIds, nil
}

func ParseSourceAndCohortId(c *gin.Context) (int, int, error) {
	// parse and validate all parameters:
	sourceId, err := ParseNumericArg(c, "sourceid")
	if err != nil {
		return -1, -1, err
	}
	cohortId, err := ParseNumericArg(c, "cohortid")
	if err != nil {
		return -1, -1, err
	}
	return sourceId, cohortId, nil
}

// separates a conceptIdsAndCohortPairs into a conceptIds list and a cohortPairs list
func GetConceptIdsAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs []interface{}) ([]CustomConceptVariableDef, []CustomDichotomousVariableDef) {
	conceptIdsAndValues := []CustomConceptVariableDef{}
	cohortPairs := []CustomDichotomousVariableDef{}
	for _, item := range conceptIdsAndCohortPairs {
		switch convertedItem := item.(type) {
		case CustomConceptVariableDef:
			conceptIdsAndValues = append(conceptIdsAndValues, convertedItem)
		case CustomDichotomousVariableDef:
			cohortPairs = append(cohortPairs, convertedItem)
		}
	}
	return conceptIdsAndValues, cohortPairs
}

// deprecated: returns the conceptIds and cohortPairs as separate lists (for backwards compatibility)
func ParseSourceIdAndCohortIdAndVariablesList(c *gin.Context) (int, int, []int64, []CustomDichotomousVariableDef, error) {
	sourceId, cohortId, conceptIdsAndCohortPairs, err := ParseSourceIdAndCohortIdAndVariablesAsSingleList(c)
	if err != nil {
		return -1, -1, nil, nil, err
	}
	conceptIdsAndValues, cohortPairs := GetConceptIdsAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs)
	conceptIds := ExtractConceptIdsFromCustomConceptVariablesDef(conceptIdsAndValues)
	return sourceId, cohortId, conceptIds, cohortPairs, nil
}

// returns sourceid, cohortid, list of variables (formed by concept ids and/or of cohort tuples which are also known as custom dichotomous variables)
func ParseSourceIdAndCohortIdAndVariablesAsSingleList(c *gin.Context) (int, int, []interface{}, error) {
	// parse and validate all parameters:
	sourceId, cohortId, err := ParseSourceAndCohortId(c)
	if err != nil {
		return -1, -1, nil, err
	}
	conceptIdsAndCohortPairs, err := ParseConceptIdsAndDichotomousDefsAsSingleList(c)
	if err != nil {
		return -1, -1, nil, err
	}
	return sourceId, cohortId, conceptIdsAndCohortPairs, nil
}

func MakeUnique(input []int) []int {
	uniqueMap := make(map[int]bool)
	var uniqueList []int

	for _, num := range input {
		if !uniqueMap[num] {
			uniqueMap[num] = true
			uniqueList = append(uniqueList, num)
		}
	}
	return uniqueList
}

// Utility function to parse out the cohort definition ids from a specific structure in which
// they can be found (in this case deep inside CustomDichotomousVariableDef items) and concatenate them
// with another given list of ids, removing duplicate ids (if any).
func GetUniqueCohortDefinitionIdsList(cohortDefinitionIds []int, filterCohortPairs []CustomDichotomousVariableDef) []int {
	var idsList []int
	idsList = append(idsList, cohortDefinitionIds...)
	if len(filterCohortPairs) > 0 {
		for _, filterCohortPair := range filterCohortPairs {
			idsList = append(idsList, filterCohortPair.CohortDefinitionId1, filterCohortPair.CohortDefinitionId2)
		}
	}
	uniqueIdsList := MakeUnique(idsList)
	return uniqueIdsList
}

func Intersect(list1 []int, list2 []int) []int {
	results := map[int]bool{}
	list2Map := map[int]bool{}

	// add each list2 item to a map, deduplicating any copies:
	for _, list2Item := range list2 {
		list2Map[list2Item] = true
	}

	// check list1 item against list2Map, add to result map if it is found in list2Map,
	// again deduplicating any copies in list1:
	for _, list1Item := range list1 {
		if _, found := list2Map[list1Item]; found {
			results[list1Item] = true
		}
	}

	// collect results into a []int list:
	intersectItems := []int{}
	for resultItem := range results {
		intersectItems = append(intersectItems, resultItem)
	}
	return intersectItems
}

// subtract list2 from list1
func Subtract(list1 []int, list2 []int) []int {
	result := []int{}
	for _, list1Item := range list1 {
		if !Contains(list2, list1Item) {
			result = append(result, list1Item)
		}
	}
	return result
}

func ConvertConceptIdToCustomConceptVariablesDef(conceptIds []int64) []CustomConceptVariableDef {
	result := []CustomConceptVariableDef{}
	for _, val := range conceptIds {
		variable := CustomConceptVariableDef{ConceptId: val, ConceptValues: []int64{}}
		result = append(result, variable)
	}

	return result
}

func ExtractConceptIdsFromCustomConceptVariablesDef(conceptIdsAndValues []CustomConceptVariableDef) []int64 {
	result := []int64{}
	for _, val := range conceptIdsAndValues {
		result = append(result, val.ConceptId)
	}

	return result
}
