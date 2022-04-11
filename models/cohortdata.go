package models

import (
	"log"
)

type CohortData struct{}

type PersonConceptAndValue struct {
	PersonId             int
	ConceptId            int
	ConceptValueAsString string
	ConceptValueAsNumber float32
}

// Retrieves observation data for LARGE cohorts/dbs.
// Assumption is that both OMOP and RESULTS schemas
// are on same DB.
func (h CohortData) RetrieveDataLargeBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int) ([]*PersonConceptAndValue, error) {
	log.Printf(">> Using inner join impl. for large cohorts")
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// get the observations for the subjects and the concepts, to build up the data rows to return:
	var cohortData []*PersonConceptAndValue
	omopDataSource.Db.Model(&Observation{}).
		Select("observation.person_id, observation.observation_concept_id as concept_id, observation.value_as_string as concept_value_as_string, observation.value_as_number as concept_value_as_number").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation.observation_concept_id in (?)", conceptIds).
		Order("observation.person_id asc"). // this order is important!
		Scan(&cohortData)

	return cohortData, nil
}
