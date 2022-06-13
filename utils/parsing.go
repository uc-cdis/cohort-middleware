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

func ParseStringArg(c *gin.Context, paramName string) (string, error) {
	// parse and validate:
	stringArgValue := c.Param(paramName)
	log.Printf("Querying %s: ", paramName)
	if stringArgValue == "" {
		log.Printf("bad request - %s should be set", paramName)
		return "", fmt.Errorf("bad request - %s should set", paramName)
	} else {
		return stringArgValue, nil
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
	sourceId, conceptIds, err := ParseSourceIdAndConceptIds(c)
	if err != nil {
		return -1, -1, nil, err
	}
	cohortIdStr := c.Param("cohortid")
	log.Printf("Querying cohort for cohort definition id...")
	if _, err := strconv.Atoi(cohortIdStr); err != nil {
		return -1, -1, nil, errors.New("bad request - cohort_definition_id should be a number")
	}
	cohortId, _ := strconv.Atoi(cohortIdStr)
	return sourceId, cohortId, conceptIds, nil
}
