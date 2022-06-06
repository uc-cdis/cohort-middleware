package models

import (
	"fmt"
	"log"
)

type CohortDataI interface {
	RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*PersonConceptAndValue, error)
	RetrieveCohortOverlapStats(sourceId int, caseCohortId int, controlCohortId int, filterConceptId int64, filterConceptValue string, otherFilterConceptIds []int64) (CohortOverlapStats, error)
}

type CohortData struct{}

type PersonConceptAndValue struct {
	PersonId             int
	ConceptId            int64
	ConceptValueAsString string
	ConceptValueAsNumber float32
}

type CohortOverlapStats struct {
	CaseControlOverlapAfterFilter int64 `json:"case_control_overlap_after_filter"`
}

// Retrieves observation data.
// Assumption is that both OMOP and RESULTS schemas
// are on same DB.
func (h CohortData) RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*PersonConceptAndValue, error) {
	log.Printf(">> Using inner join impl. for large cohorts")
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// get the observations for the subjects and the concepts, to build up the data rows to return:
	var cohortData []*PersonConceptAndValue
	meta_result := omopDataSource.Db.Model(&Observation{}).
		Select("observation.person_id, observation.observation_concept_id as concept_id, observation.value_as_string as concept_value_as_string, observation.value_as_number as concept_value_as_number").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation.observation_concept_id in (?)", conceptIds).
		Order("observation.person_id asc"). // this order is important!
		Scan(&cohortData)
	return cohortData, meta_result.Error
}

// Assesses the overlap between case and control cohorts. It does this after filtering the cohorts and keeping only
// the persons that have data for each of the selected conceptIds and match the filterConceptId/filterConceptValue criterion.
func (h CohortData) RetrieveCohortOverlapStats(sourceId int, caseCohortId int, controlCohortId int,
	filterConceptId int64, filterConceptValue string, otherFilterConceptIds []int64) (CohortOverlapStats, error) {

	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// count persons that are in the intersection of both case and control cohorts, filtering on filterConceptValue:
	var breakdownValueFieldName = "observation.value_as_" + getConceptValueType(filterConceptId)
	var cohortOverlapStats CohortOverlapStats
	query := omopDataSource.Db.Model(&Observation{}).
		Select("count(distinct(observation.person_id)) as case_control_overlap_after_filter").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as case_cohort ON case_cohort.subject_id = observation.person_id").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as control_cohort ON control_cohort.subject_id = case_cohort.subject_id"). // this one allows for the intersection between case and control and the assessment of the overlap
		Where("case_cohort.cohort_definition_id = ?", caseCohortId).
		Where("control_cohort.cohort_definition_id = ?", controlCohortId).
		Where("observation.observation_concept_id = ?", filterConceptId).
		Where(breakdownValueFieldName+" = ?", filterConceptValue)

	// iterate over the otherFilterConceptIds, adding a new INNER JOIN and filters for each, so that the resulting set is the
	// set of persons that have a non-null value for each and every one of the concepts:
	for i, filterConceptId := range otherFilterConceptIds {
		observationTableAlias := fmt.Sprintf("observation_filter_%d", i)
		log.Printf("Adding extra INNER JOIN with alias %s", observationTableAlias)
		query = query.Joins("INNER JOIN "+omopDataSource.Schema+".observation as "+observationTableAlias+" ON "+observationTableAlias+".person_id = observation.person_id").
			Where(observationTableAlias+".observation_concept_id = ?", filterConceptId).
			Where("(" + observationTableAlias + ".value_as_string is not null or " + observationTableAlias + ".value_as_number is not null)") // TODO - improve performance by only filtering on type according to getConceptValueType()
	}
	meta_result := query.Scan(&cohortOverlapStats)
	return cohortOverlapStats, meta_result.Error
}
