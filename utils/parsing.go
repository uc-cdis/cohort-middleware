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

func ContainsNonNil(errors []error) bool {
	for _, v := range errors {
		if v != nil {
			return true
		}
	}
	return false
}

type ConceptIds struct {
	ConceptIds []int64
}

type ConceptTypes struct {
	ConceptTypes []string
}

// This method expects a request body with a payload similar to the following example:
// {"variables": [
//   {variable_type: "concept", concept_id: 2000000324},
//   {variable_type: "concept", concept_id: 2000006885},
//   {variable_type: "custom_dichotomous", cohort_ids: [cohortX_id, cohortY_id]},
//   {variable_type: "custom_dichotomous", cohort_ids: [cohortM_id, cohortN_id]},
//       ...
// ]}
// It returns the list of concept_id values and the list of cohort_id tuples.
func ParseConceptIdsAndDichotomousIds(c *gin.Context) ([]int64, [][]int, error) {
	if c.Request == nil || c.Request.Body == nil {
		return nil, nil, errors.New("bad request - no request body")
	}
	decoder := json.NewDecoder(c.Request.Body)
	request := make(map[string][]map[string]interface{})
	err := decoder.Decode(&request)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, nil, err
	}

	variables := request["variables"]
	conceptIds := []int64{}
	cohortPairs := [][]int{}
	// TODO - this parsing will throw a lot of "null pointer" errors since it does not validate if specific entries are found in the json before
	// accessing them...needs to be fixed to throw better errors:
	for _, variable := range variables {
		if variable["variable_type"] == "concept" {
			conceptIds = append(conceptIds, int64(variable["concept_id"].(float64)))
		}
		if variable["variable_type"] == "custom_dichotomous" {
			cohortPair := []int{}
			convertedCohortIds := variable["cohort_ids"].([]interface{})
			for _, convertedCohortId := range convertedCohortIds {
				cohortPair = append(cohortPair, int(convertedCohortId.(float64)))
			}

			cohortPairs = append(cohortPairs, cohortPair)
		}
	}
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

// returns sourceid, cohortid, list of variables (formed by concept ids and/or of cohort tuples which are also known as custom dichotomous variables)
func ParseSourceIdAndCohortIdAndVariablesList(c *gin.Context) (int, int, []int64, [][]int, error) {
	// parse and validate all parameters:
	sourceId, cohortId, err := ParseSourceAndCohortId(c)
	if err != nil {
		return -1, -1, nil, nil, err
	}
	conceptIds, cohortPairs, err := ParseConceptIdsAndDichotomousIds(c)
	if err != nil {
		return -1, -1, nil, nil, err
	}
	return sourceId, cohortId, conceptIds, cohortPairs, nil
}
