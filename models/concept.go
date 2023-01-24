package models

import (
	"fmt"
	"log"

	"github.com/uc-cdis/cohort-middleware/utils"
)

type ConceptI interface {
	RetriveAllBySourceId(sourceId int) ([]*Concept, error)
	RetrieveInfoBySourceIdAndConceptIds(sourceId int, conceptIds []int64) ([]*ConceptSimple, error)
	RetrieveInfoBySourceIdAndConceptTypes(sourceId int, conceptTypes []string) ([]*ConceptSimple, error)
	RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*ConceptStats, error)
	RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId int, cohortDefinitionId int, breakdownConceptId int64) ([]*ConceptBreakdown, error)
	RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef, breakdownConceptId int64) ([]*ConceptBreakdown, error)
}
type Concept struct {
	ConceptId   int    `json:"concept_id"`
	ConceptName string `json:"concept_name"`
	ConceptType string `json:"concept_type"`
}

type ConceptAndPersonsWithDataStats struct {
	ConceptId  int64
	NpersonIds int
}

type ConceptStats struct {
	ConceptId         int64   `json:"concept_id"`
	PrefixedConceptId string  `json:"prefixed_concept_id"`
	ConceptName       string  `json:"concept_name"`
	ConceptType       string  `json:"concept_type"`
	CohortSize        int     `json:"cohort_size"`
	NmissingRatio     float32 `json:"n_missing_ratio"`
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
}

func (h Concept) RetriveAllBySourceId(sourceId int) ([]*Concept, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var concepts []*Concept
	meta_result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, concept_class_id as concept_type").
		Order("concept_name").
		Scan(&concepts)
	return concepts, meta_result.Error
}

func getNrPersonsWithData(conceptId int64, conceptsAndPersonsWithData []*ConceptAndPersonsWithDataStats) int {
	for _, conceptsAndDataInfo := range conceptsAndPersonsWithData {
		if conceptsAndDataInfo.ConceptId == conceptId {
			return conceptsAndDataInfo.NpersonIds
		}
	}
	return 0
}

// Retrieve just a simple concept info for a given conceptId.
func (h Concept) RetrieveInfoBySourceIdAndConceptId(sourceId int, conceptId int64) (*ConceptSimple, error) {
	conceptIds := []int64{conceptId}
	result, err := h.RetrieveInfoBySourceIdAndConceptIds(sourceId, conceptIds)
	if err != nil {
		return nil, err
	} else if len(result) == 0 {
		// given concept_id not found, just return nil:
		return nil, nil
	}
	return result[0], err
}

// Retrieve just a simple list of concept names and type info for given list of conceptIds.
func (h Concept) RetrieveInfoBySourceIdAndConceptIds(sourceId int, conceptIds []int64) ([]*ConceptSimple, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptItems []*ConceptSimple
	meta_result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, concept_code, concept_class_id as concept_type").
		Where("concept_id in (?)", conceptIds).
		Order("concept_name").
		Scan(&conceptItems)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}
	for _, conceptItem := range conceptItems {
		// set prefixed_concept_id:
		conceptItem.PrefixedConceptId = GetPrefixedConceptId(conceptItem.ConceptId)
	}
	if len(conceptItems) != len(conceptIds) {
		return nil, fmt.Errorf("unexpected error: did not find all concepts")
	}
	return conceptItems, nil
}

func (h Concept) RetrieveInfoBySourceIdAndConceptTypes(sourceId int, conceptTypes []string) ([]*ConceptSimple, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptItems []*ConceptSimple
	meta_result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, concept_class_id as concept_type").
		Where("concept_class_id in (?)", conceptTypes).
		Order("concept_name").
		Scan(&conceptItems)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}
	for _, conceptItem := range conceptItems {
		// set prefixed_concept_id:
		conceptItem.PrefixedConceptId = GetPrefixedConceptId(conceptItem.ConceptId)
	}
	return conceptItems, nil
}

// Retrieve concept name, type and missing ratio statistics for given list of conceptIds.
// Assumption is that both OMOP and RESULTS schemas are on same DB.
func (h Concept) RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*ConceptStats, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptStats []*ConceptStats
	meta_result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, 0 as n_missing_ratio, concept_class_id as concept_type").
		Where("concept_id in (?)", conceptIds).
		Order("concept_name").
		Scan(&conceptStats)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}

	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var cohortSize int
	meta_result = resultsDataSource.Db.Model(&Cohort{}).
		Select("count(*) as cohort_size").
		Where("cohort_definition_id = ?", cohortDefinitionId).
		Scan(&cohortSize)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}

	// find, for each concept, the ratio of persons in the given cohortId that have
	// no value for this concept by first finding the ones that do have some value and
	// then subtracting them from cohort size before dividing:
	var conceptsAndPersonsWithData []*ConceptAndPersonsWithDataStats
	meta_result = omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation WITH (NOEXPAND)").
		Select("observation_concept_id as concept_id, count(distinct(person_id)) as nperson_ids").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation_concept_id in (?)", conceptIds).
		Where("(value_as_string is not null or value_as_number is not null)").
		Group("observation_concept_id").
		Scan(&conceptsAndPersonsWithData)

	for _, conceptStat := range conceptStats {
		// since we are looping over items anyway, also set prefixed_concept_id and cohort_size:
		conceptStat.PrefixedConceptId = GetPrefixedConceptId(conceptStat.ConceptId)
		conceptStat.CohortSize = cohortSize
		if cohortSize == 0 {
			conceptStat.NmissingRatio = 0
		} else {
			// calculate missing ratio for cohorts that actually have a size:
			var nrPersonsWithData = getNrPersonsWithData(conceptStat.ConceptId, conceptsAndPersonsWithData)
			log.Printf("Found %d persons with data for concept_id %d", nrPersonsWithData, conceptStat.ConceptId)
			n_missing := cohortSize - nrPersonsWithData
			conceptStat.NmissingRatio = float32(n_missing) / float32(cohortSize)
		}
	}

	return conceptStats, meta_result.Error
}

func getConceptValueType(conceptId int64) string {
	return "string" // TODO - add logic to return "string" or "number" depending on concept type
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
	var breakdownValueFieldName = "observation.value_as_" + getConceptValueType(breakdownConceptId)
	var conceptBreakdownList []*ConceptBreakdown
	query := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation WITH (NOEXPAND)").
		Select("observation.value_as_concept_id, count(distinct(observation.person_id)) as npersons_in_cohort_with_value").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation.observation_concept_id = ?", breakdownConceptId).
		Where(breakdownValueFieldName + " is not null").      // this one seems like a bit of a random constraint...but was a request from the business side: skip records where this field is null
		Where("observation.value_as_concept_id is not null"). // this is assuming that the breakdownConceptId has its values nicely stored as concepts as well and correctly used in observation table...
		Where("observation.value_as_concept_id != 0")

	query = QueryFilterByConceptIdsAndCohortPairsHelper(query, filterConceptIds, filterCohortPairs, omopDataSource.Schema, resultsDataSource.Schema)

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
