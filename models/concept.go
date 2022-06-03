package models

import (
	"fmt"
	"log"
)

type ConceptI interface {
	RetriveAllBySourceId(sourceId int) ([]*Concept, error)
	RetrieveInfoBySourceIdAndConceptIds(sourceId int, conceptIds []int64) ([]*ConceptSimple, error)
	RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*ConceptStats, error)
	RetrieveBreakdownStatsBySourceIdAndCohortId(sourceId int, cohortDefinitionId int, breakdownConceptId int64) ([]*ConceptBreakdown, error)
	RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, filterConceptIds []int64, breakdownConceptId int64) ([]*ConceptBreakdown, error)
}
type Concept struct {
	ConceptId   int         `json:"concept_id"`
	ConceptName string      `json:"concept_name"`
	DomainId    string      `json:"domain_id"`
	DomainName  string      `json:"domain_name"`
	ConceptType ConceptType `json:"concept_type"`
}

type ConceptAndPersonsWithDataStats struct {
	ConceptId  int64
	NpersonIds int
}

type ConceptStats struct {
	ConceptId         int64       `json:"concept_id"`
	PrefixedConceptId string      `json:"prefixed_concept_id"`
	ConceptName       string      `json:"concept_name"`
	DomainId          string      `json:"domain_id"`
	DomainName        string      `json:"domain_name"`
	ConceptType       ConceptType `json:"concept_type"`
	CohortSize        int         `json:"cohort_size"`
	NmissingRatio     float32     `json:"n_missing_ratio"`
}

type ConceptSimple struct {
	ConceptId         int64       `json:"concept_id"`
	PrefixedConceptId string      `json:"prefixed_concept_id"`
	ConceptName       string      `json:"concept_name"`
	DomainId          string      `json:"domain_id"`
	DomainName        string      `json:"domain_name"`
	ConceptType       ConceptType `json:"concept_type"`
}

type ConceptBreakdown struct {
	ConceptValue              string `json:"concept_value"`
	NpersonsInCohortWithValue int    `json:"persons_in_cohort_with_value"`
}

type Observation struct {
}

func (h Concept) RetriveAllBySourceId(sourceId int) ([]*Concept, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var concepts []*Concept
	result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name, concept_class_id as concept_type").
		Joins("INNER JOIN " + omopDataSource.Schema + ".domain as domain ON concept.domain_id = domain.domain_id").
		Order("concept_name").
		Scan(&concepts)
	return concepts, result.Error
}

func getNrPersonsWithData(conceptId int64, conceptsAndPersonsWithData []*ConceptAndPersonsWithDataStats) int {
	for _, conceptsAndDataInfo := range conceptsAndPersonsWithData {
		if conceptsAndDataInfo.ConceptId == conceptId {
			return conceptsAndDataInfo.NpersonIds
		}
	}
	return 0
}

// Retrieve just a simple list of concept names and domain info for given list of conceptIds.
func (h Concept) RetrieveInfoBySourceIdAndConceptIds(sourceId int, conceptIds []int64) ([]*ConceptSimple, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptItems []*ConceptSimple
	result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name, concept_class_id as concept_type").
		Joins("INNER JOIN "+omopDataSource.Schema+".domain as domain ON concept.domain_id = domain.domain_id").
		Where("concept_id in (?)", conceptIds).
		Order("concept_name").
		Scan(&conceptItems)
	if result.Error != nil {
		return nil, result.Error
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

// Retrieve concept name, domain and missing ratio statistics for given list of conceptIds.
// Assumption is that both OMOP and RESULTS schemas are on same DB.
func (h Concept) RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*ConceptStats, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptStats []*ConceptStats
	result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name, 0 as n_missing_ratio, concept_class_id as concept_type").
		Joins("INNER JOIN "+omopDataSource.Schema+".domain as domain ON concept.domain_id = domain.domain_id").
		Where("concept_id in (?)", conceptIds).
		Order("concept_name").
		Scan(&conceptStats)
	if result.Error != nil {
		return nil, result.Error
	}

	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var cohortSize int
	result = resultsDataSource.Db.Model(&Cohort{}).
		Select("count(*) as cohort_size").
		Where("cohort_definition_id = ?", cohortDefinitionId).
		Scan(&cohortSize)
	if result.Error != nil {
		return nil, result.Error
	}

	// find, for each concept, the ratio of persons in the given cohortId that have
	// no value for this concept by first finding the ones that do have some value and
	// then subtracting them from cohort size before dividing:
	var conceptsAndPersonsWithData []*ConceptAndPersonsWithDataStats
	omopDataSource.Db.Model(&Observation{}).
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

	return conceptStats, nil
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
	// TODO - this query is identical to query in function below if that function is called with empty filterConceptIds[]... use that instead...?
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// count persons, grouping by concept value:
	var valueFieldName = "value_as_" + getConceptValueType(breakdownConceptId)
	var result []*ConceptBreakdown
	omopDataSource.Db.Model(&Observation{}).
		Select(valueFieldName+" as concept_value, count(distinct(person_id)) as npersons_in_cohort_with_value").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation_concept_id = ?", breakdownConceptId).
		Group(valueFieldName).
		Scan(&result)
	return result, nil
}

// Similar to function above, but only count persons that have a non-null value for each of the ids in the given filterConceptIds.
// So, using the example documented in the function above, it will return something like:
//  {ConceptValue: "A", NPersonsInCohortWithValue: M-X},
//  {ConceptValue: "B", NPersonsInCohortWithValue: N-M-X},
// where X is the number of persons that have NO value or just a "null" value for one or more of the ids in the given filterConceptIds.
func (h Concept) RetrieveBreakdownStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, filterConceptIds []int64, breakdownConceptId int64) ([]*ConceptBreakdown, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// count persons, grouping by concept value:
	var breakdownValueFieldName = "observation.value_as_" + getConceptValueType(breakdownConceptId)
	var result []*ConceptBreakdown
	query := omopDataSource.Db.Model(&Observation{}).
		Select(breakdownValueFieldName+" as concept_value, count(distinct(observation.person_id)) as npersons_in_cohort_with_value").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation.observation_concept_id = ?", breakdownConceptId)

	// iterate over the filterConceptIds, adding a new INNER JOIN and filters for each, so that it becomes an intersection of all:
	for _, filterConceptId := range filterConceptIds {
		observationTableAlias := fmt.Sprintf("observation_filter_%d", filterConceptId)
		log.Printf("Adding extra INNER JOIN with alias %s", observationTableAlias)
		query = query.Joins("INNER JOIN "+omopDataSource.Schema+".observation as "+observationTableAlias+" ON "+observationTableAlias+".person_id = observation.person_id").
			Where(observationTableAlias+".observation_concept_id = ?", filterConceptId).
			Where("(" + observationTableAlias + ".value_as_string is not null or " + observationTableAlias + ".value_as_number is not null)") // TODO - improve performance by only filtering on type according to getConceptValueType()
	}
	query.Group(breakdownValueFieldName).
		Scan(&result)
	return result, nil
}
