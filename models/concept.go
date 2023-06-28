package models

import (
	"fmt"

	"github.com/uc-cdis/cohort-middleware/utils"
)

type ConceptI interface {
	RetriveAllBySourceId(sourceId int) ([]*Concept, error)
	RetrieveInfoBySourceIdAndConceptId(sourceId int, conceptId int64) (*ConceptSimple, error)
	RetrieveInfoBySourceIdAndConceptIds(sourceId int, conceptIds []int64) ([]*ConceptSimple, error)
	RetrieveInfoBySourceIdAndConceptTypes(sourceId int, conceptTypes []string) ([]*ConceptSimple, error)
	RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId int, cohortDefinitionId int, breakdownConceptId int64) ([]*ConceptBreakdown, error)
	RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef, breakdownConceptId int64) ([]*ConceptBreakdown, error)
}
type Concept struct {
	ConceptId   int64  `json:"concept_id"`
	ConceptName string `json:"concept_name"`
	ConceptType string `json:"concept_type"`
}

type ConceptSimple struct {
	ConceptId         int64  `json:"concept_id"`
	PrefixedConceptId string `json:"prefixed_concept_id"`
	ConceptName       string `json:"concept_name"`
	ConceptCode       string `json:"concept_code"`
	ConceptType       string `json:"concept_type"`
}

type ConceptBreakdown struct {
	ConceptValue              string `json:"concept_value"`
	ValueAsConceptId          int64  `json:"concept_value_as_concept_id"`
	ValueName                 string `json:"concept_value_name"`
	NpersonsInCohortWithValue int    `json:"persons_in_cohort_with_value"`
}

type Observation struct {
	ObservationId int64
}

func (h Concept) RetriveAllBySourceId(sourceId int) ([]*Concept, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var concepts []*Concept
	query := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, concept_class_id as concept_type").
		Order("concept_name")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&concepts)
	return concepts, meta_result.Error
}

// Retrieve just a simple concept info for a given conceptId.
// Raises an error if concept is not found.
func (h Concept) RetrieveInfoBySourceIdAndConceptId(sourceId int, conceptId int64) (*ConceptSimple, error) {
	conceptIds := []int64{conceptId}
	result, err := h.RetrieveInfoBySourceIdAndConceptIds(sourceId, conceptIds)
	if err != nil {
		return nil, err
	} else if len(result) == 0 {
		// given concept_id not found, return error:
		return nil, fmt.Errorf("unexpected error: did not find concept")
	}
	return result[0], err
}

// Retrieve just a simple list of concept names and type info for given list of conceptIds.
// Raises an error if any of the concepts is not found.
func (h Concept) RetrieveInfoBySourceIdAndConceptIds(sourceId int, conceptIds []int64) ([]*ConceptSimple, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptItems []*ConceptSimple
	query := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, concept_code, concept_class_id as concept_type").
		Where("concept_id in (?)", conceptIds).
		Order("concept_name")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&conceptItems)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}
	for _, conceptItem := range conceptItems {
		// set prefixed_concept_id:
		conceptItem.PrefixedConceptId = GetPrefixedConceptId(conceptItem.ConceptId)
	}
	if len(conceptItems) != len(conceptIds) {
		return nil, fmt.Errorf("unexpected error: did not find the expected number of concepts")
	}
	return conceptItems, nil
}

func (h Concept) RetrieveInfoBySourceIdAndConceptTypes(sourceId int, conceptTypes []string) ([]*ConceptSimple, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptItems []*ConceptSimple
	query := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, concept_class_id as concept_type").
		Where("concept_class_id in (?)", conceptTypes).
		Order("concept_name")

	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&conceptItems)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}
	for _, conceptItem := range conceptItems {
		// set prefixed_concept_id:
		conceptItem.PrefixedConceptId = GetPrefixedConceptId(conceptItem.ConceptId)
	}
	return conceptItems, nil
}

// This function will return cohort size broken down over the different values
// of the given "breakdown concept" by querying, for each distinct concept value,
// how many persons in the cohort have that value in their observation records.
// E.g. if we have a cohort of size N and a concept that can have values "A" or "B",
// then it will return something like:
//  {ConceptValue: "A", NPersonsInCohortWithValue: M},
//  {ConceptValue: "B", NPersonsInCohortWithValue: N-M},
func (h Concept) RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId int, cohortDefinitionId int, breakdownConceptId int64) ([]*ConceptBreakdown, error) {
	// this is identical to the result of the function below if called with empty filterConceptIds[] and empty filterCohortPairs... so call that:
	filterConceptIds := []int64{}
	filterCohortPairs := []utils.CustomDichotomousVariableDef{}
	return h.RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId, cohortDefinitionId, filterConceptIds, filterCohortPairs, breakdownConceptId)
}

// Basically same goal as described in function above, but only count persons that have a non-null value for each
// of the ids in the given filterConceptIds. So, using the example documented in the function above, it will
// return something like:
//  {ConceptValue: "A", NPersonsInCohortWithValue: M-X},
//  {ConceptValue: "B", NPersonsInCohortWithValue: N-M-X},
// where X is the number of persons that have NO value or just a "null" value for one or more of the ids in the given filterConceptIds.
func (h Concept) RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef, breakdownConceptId int64) ([]*ConceptBreakdown, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// count persons, grouping by concept value:
	var conceptBreakdownList []*ConceptBreakdown
	query := QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, cohortDefinitionId, "unionAndIntersect").
		Select("observation.value_as_concept_id, count(distinct(observation.person_id)) as npersons_in_cohort_with_value").
		Joins("INNER JOIN "+omopDataSource.Schema+".observation_continuous as observation"+omopDataSource.GetViewDirective()+" ON unionAndIntersect.subject_id = observation.person_id").
		Where("observation.observation_concept_id = ?", breakdownConceptId).
		Where(GetConceptValueNotNullCheckBasedOnConceptType("observation", sourceId, breakdownConceptId))

	// note: here we pass empty []utils.CustomDichotomousVariableDef{} instead of filterCohortPairs, since we already use the SQL generated by QueryFilterByCohortPairsHelper above,
	// which is a better performing SQL in this particular scenario:
	query = QueryFilterByConceptIdsHelper(query, sourceId, filterConceptIds, omopDataSource, resultsDataSource.Schema, "observation")

	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Group("observation.value_as_concept_id").
		Scan(&conceptBreakdownList)

	// Add concept value (coded value) and concept name for each of the value_as_concept_id values:
	for _, conceptBreakdownItem := range conceptBreakdownList {
		conceptInfo, error := h.RetrieveInfoBySourceIdAndConceptId(sourceId, conceptBreakdownItem.ValueAsConceptId)
		if error != nil {
			return nil, error
		}
		conceptBreakdownItem.ConceptValue = conceptInfo.ConceptCode
		conceptBreakdownItem.ValueName = conceptInfo.ConceptName
	}
	return conceptBreakdownList, meta_result.Error
}
