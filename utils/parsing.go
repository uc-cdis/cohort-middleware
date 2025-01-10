package utils

import (
	"crypto/sha256"
	"encoding/hex"
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

type Variable struct {
	VariableType   string   `json:"variable_type"`
	ConceptId      *int64   `json:"concept_id,omitempty"`
	Filters        []Filter `json:"filters,omitempty"`
	Transformation string   `json:"transformation,omitempty"`
	ProvidedName   *string  `json:"provided_name,omitempty"`
	CohortIds      []int    `json:"cohort_ids,omitempty"`
}

type Filter struct {
	Type               string    `json:"type"`
	Value              *float64  `json:"value,omitempty"`
	Values             []float64 `json:"values,omitempty"`
	ValueAsConceptId   *int64    `json:"value_as_concept_id,omitempty"`
	ValuesAsConceptIds []int64   `json:"values_as_concept_ids,omitempty"`
}

// fields that define a custom dichotomous variable:
type CustomDichotomousVariableDef struct {
	CohortDefinitionId1 int
	CohortDefinitionId2 int
	ProvidedName        string
}

type CustomConceptVariableDef struct {
	ConceptId      int64
	Filters        []Filter
	Transformation string
}

func GetCohortPairKey(firstCohortDefinitionId int, secondCohortDefinitionId int) string {
	return fmt.Sprintf("ID_%v_%v", firstCohortDefinitionId, secondCohortDefinitionId)
}

// This method expects a request body with a payload similar to the following example:
// {"variables": [
//
//		{variable_type: "concept", concept_id: 2000000324},  // <- simple concept
//		{variable_type: "concept", concept_id: 2000006885    // <- complex concept item where optional filters and transformation are filled in
//		 "filters": [{"type": ">=", "value": 0.5},
//	 			 {"type": "<=", "value": 12.5}
//					],
//		 "transformation": "inverse_normal" },
//	    {"variable_type":"concept", "concept_id":2000007027,
//	     "filters": [{"type": "in", "values_as_concept_id" :[2000007028, 2000007029]}] // <- complex concept item where values_as_concept_id filter is specified
//	    },
//		{variable_type: "custom_dichotomous", provided_name: "name1", cohort_ids: [cohortX_id, cohortY_id]},
//		{variable_type: "custom_dichotomous", provided_name: "name2", cohort_ids: [cohortM_id, cohortN_id]},
//		    ...
//
// ]}
//
// The possible filter "type" values are: ">", "<", ">=", "<=", "=", "!=", "in".
// The second part of the filter can be specified as: "value", "values" (a list), "value_as_concept_id", "values_as_concept_id (a list)".
// The possible "transformation" values are: "log", "inverse_normal_rank", "z_score", "box_cox".
//
// It returns the list with all concept_id values and custom dichotomous variable definitions.
func ParseConceptDefsAndDichotomousDefsAsSingleList(c *gin.Context) ([]interface{}, error) {
	if c.Request == nil || c.Request.Body == nil {
		return nil, errors.New("bad request - no request body")
	}

	var requestBody struct {
		Variables []Variable `json:"variables"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&requestBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		return nil, errors.New("failed to parse JSON request body")
	}

	conceptDefsAndDichotomousDefs := make([]interface{}, 0)

	for _, variable := range requestBody.Variables {
		switch variable.VariableType {
		case "concept":
			conceptDef, err := parseConceptVariable(variable)
			if err != nil {
				log.Printf("Error parsing concept variable: %v", err)
				return nil, err
			}
			conceptDefsAndDichotomousDefs = append(conceptDefsAndDichotomousDefs, conceptDef)
		case "custom_dichotomous":
			dichotomousDef, err := parseCustomDichotomousVariable(variable)
			if err != nil {
				log.Printf("Error parsing custom dichotomous variable: %v", err)
				return nil, err
			}
			conceptDefsAndDichotomousDefs = append(conceptDefsAndDichotomousDefs, dichotomousDef)
		default:
			log.Printf("Unsupported variable type: %s", variable.VariableType)
			return nil, errors.New("unsupported variable type in request body")
		}
	}

	return conceptDefsAndDichotomousDefs, nil
}

func parseConceptVariable(variable Variable) (CustomConceptVariableDef, error) {
	if variable.ConceptId == nil {
		return CustomConceptVariableDef{}, errors.New("concept variable missing concept_id")
	}
	// Validate filters:
	var validFilterTypes = map[string]bool{
		">": true, "<": true, ">=": true, "<=": true, "=": true, "!=": true, "in": true,
	}
	for _, filter := range variable.Filters {
		if !validFilterTypes[filter.Type] {
			return CustomConceptVariableDef{}, errors.New("invalid filter type: " + filter.Type)
		}
		// Ensure at least one value field is provided:
		if filter.Value == nil && filter.Values == nil && filter.ValueAsConceptId == nil && len(filter.ValuesAsConceptIds) == 0 {
			return CustomConceptVariableDef{}, errors.New("filter must specify at least one of: value, values, value_as_concept_id, or values_as_concept_id")
		}
	}

	// Validate transformation:
	var validTransformationTypes = map[string]bool{
		"log": true, "inverse_normal_rank": true, "z_score": true, "box_cox": true,
	}
	if variable.Transformation != "" && !validTransformationTypes[variable.Transformation] {
		return CustomConceptVariableDef{}, errors.New("invalid transformation type: " + variable.Transformation)
	}

	return CustomConceptVariableDef{
		ConceptId:      *variable.ConceptId,
		Filters:        variable.Filters,
		Transformation: variable.Transformation,
	}, nil
}

func parseCustomDichotomousVariable(variable Variable) (CustomDichotomousVariableDef, error) {
	if len(variable.CohortIds) != 2 {
		return CustomDichotomousVariableDef{}, errors.New("custom dichotomous variable must have exactly 2 cohort_ids")
	}

	providedName := "default_name"
	if variable.ProvidedName != nil {
		providedName = *variable.ProvidedName
	}

	return CustomDichotomousVariableDef{
		CohortDefinitionId1: variable.CohortIds[0],
		CohortDefinitionId2: variable.CohortIds[1],
		ProvidedName:        providedName,
	}, nil
}

// deprecated: for backwards compatibility
func ParseConceptDefsAndDichotomousDefs(c *gin.Context) ([]CustomConceptVariableDef, []CustomDichotomousVariableDef, error) {
	conceptDefsAndCohortPairs, err := ParseConceptDefsAndDichotomousDefsAsSingleList(c)
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, nil, err
	}
	conceptDefs, cohortPairs := GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptDefsAndCohortPairs)
	return conceptDefs, cohortPairs, nil
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
func GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs []interface{}) ([]CustomConceptVariableDef, []CustomDichotomousVariableDef) {
	conceptDefsAndValues := []CustomConceptVariableDef{}
	cohortPairs := []CustomDichotomousVariableDef{}
	for _, item := range conceptIdsAndCohortPairs {
		switch convertedItem := item.(type) {
		case CustomConceptVariableDef:
			conceptDefsAndValues = append(conceptDefsAndValues, convertedItem)
		case CustomDichotomousVariableDef:
			cohortPairs = append(cohortPairs, convertedItem)
		}
	}
	return conceptDefsAndValues, cohortPairs
}

// deprecated: returns the conceptIds and cohortPairs as separate lists (for backwards compatibility)
func ParseSourceIdAndCohortIdAndVariablesList(c *gin.Context) (int, int, []int64, []CustomDichotomousVariableDef, error) {
	sourceId, cohortId, conceptIdsAndCohortPairs, err := ParseSourceIdAndCohortIdAndVariablesAsSingleList(c)
	if err != nil {
		return -1, -1, nil, nil, err
	}
	conceptIdsAndValues, cohortPairs := GetConceptDefsAndValuesAndCohortPairsAsSeparateLists(conceptIdsAndCohortPairs)
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
	conceptDefsAndCohortPairs, err := ParseConceptDefsAndDichotomousDefsAsSingleList(c)
	if err != nil {
		return -1, -1, nil, err
	}
	return sourceId, cohortId, conceptDefsAndCohortPairs, nil
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
		variable := CustomConceptVariableDef{ConceptId: val}
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

// returns the pointer to a float value
func Float64Ptr(v float64) *float64 {
	return &v
}

// Creates a unique hash from a given string
func GenerateHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:]) // Convert to hexadecimal string
}
