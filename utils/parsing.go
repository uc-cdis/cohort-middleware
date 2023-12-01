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
	CohortId1    int
	CohortId2    int
	ProvidedName string
}

func GetCohortPairKey(firstCohortDefinitionId int, secondCohortDefinitionId int) string {
	return fmt.Sprintf("ID_%v_%v", firstCohortDefinitionId, secondCohortDefinitionId)
}

// This method expects a request body with a payload similar to the following example:
// {"variables": [
//   {variable_type: "concept", concept_id: 2000000324},
//   {variable_type: "concept", concept_id: 2000006885},
//   {variable_type: "custom_dichotomous", provided_name: "name1", cohort_ids: [cohortX_id, cohortY_id]},
//   {variable_type: "custom_dichotomous", provided_name: "name2", cohort_ids: [cohortM_id, cohortN_id]},
//       ...
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
			conceptIdsAndCohortPairs = append(conceptIdsAndCohortPairs, int64(variable["concept_id"].(float64)))
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
				CohortId1:    cohortPair[0],
				CohortId2:    cohortPair[1],
				ProvidedName: providedName,
			}
			conceptIdsAndCohortPairs = append(conceptIdsAndCohortPairs, customDichotomousVariableDef)
		}
	}
	return conceptIdsAndCohortPairs, nil
}

// deprecated: for backwards compatibility
func ParseConceptIdsAndDichotomousDefs(c *gin.Context) ([]int64, []CustomDichotomousVariableDef, error) {
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
func GetConceptIdsAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs []interface{}) ([]int64, []CustomDichotomousVariableDef) {
	conceptIds := []int64{}
	cohortPairs := []CustomDichotomousVariableDef{}
	for _, item := range conceptIdsAndCohortPairs {
		switch convertedItem := item.(type) {
		case int64:
			conceptIds = append(conceptIds, convertedItem)
		case CustomDichotomousVariableDef:
			cohortPairs = append(cohortPairs, convertedItem)
		}
	}
	return conceptIds, cohortPairs
}

// deprecated: returns the conceptIds and cohortPairs as separate lists (for backwards compatibility)
func ParseSourceIdAndCohortIdAndVariablesList(c *gin.Context) (int, int, []int64, []CustomDichotomousVariableDef, error) {
	sourceId, cohortId, conceptIdsAndCohortPairs, err := ParseSourceIdAndCohortIdAndVariablesAsSingleList(c)
	if err != nil {
		return -1, -1, nil, nil, err
	}
	conceptIds, cohortPairs := GetConceptIdsAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs)
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
