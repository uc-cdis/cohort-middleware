package models

import (
	"fmt"
	"log"

	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortDataI interface {
	RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*PersonConceptAndValue, error)
	RetrieveCohortOverlapStats(sourceId int, caseCohortId int, controlCohortId int, filterConceptId int64, filterConceptValue int64, otherFilterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) (CohortOverlapStats, error)
	RetrieveCohortOverlapStatsWithoutFilteringOnConceptValue(sourceId int, caseCohortId int, controlCohortId int, otherFilterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) (CohortOverlapStats, error)
	RetrieveDataByOriginalCohortAndNewCohort(sourceId int, originalCohortDefinitionId int, cohortDefinitionId int) ([]*PersonIdAndCohort, error)
	RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, histogramConceptId int64, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) ([]*PersonConceptAndValue, error)
}

type CohortData struct{}

type PersonConceptAndValue struct {
	PersonId                int64
	ConceptId               int64
	ConceptValueAsString    string
	ConceptValueAsNumber    float32
	ConceptValueAsConceptId int64
}

type PersonConceptAndCount struct {
	PersonId  int64
	ConceptId int64
	Count     int64
}

type CohortOverlapStats struct {
	CaseControlOverlap int64 `json:"case_control_overlap"`
}

type PersonIdAndCohort struct {
	PersonId int64
	CohortId int64
}

// This function returns the subjects that belong to both cohorts (the intersection of both cohorts)
// TODO - name this function as such
func (h CohortData) RetrieveDataByOriginalCohortAndNewCohort(sourceId int, originalCohortDefinitionId int, cohortDefinitionId int) ([]*PersonIdAndCohort, error) {
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var personData []*PersonIdAndCohort

	meta_result := resultsDataSource.Db.Model(&Cohort{}).
		Select("cohort.subject_id as person_id, cohort.cohort_definition_id as cohort_id").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as original_cohort ON cohort.subject_id = original_cohort.subject_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("original_cohort.cohort_definition_id = ?", originalCohortDefinitionId).
		Scan(&personData)
	return personData, meta_result.Error
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
	meta_result := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation").
		Select("observation.person_id, observation.observation_concept_id as concept_id, observation.value_as_string as concept_value_as_string, observation.value_as_number as concept_value_as_number, observation.value_as_concept_id as concept_value_as_concept_id").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation.observation_concept_id in (?)", conceptIds).
		Order("observation.person_id asc"). // this order is important!
		Scan(&cohortData)
	return cohortData, meta_result.Error
}

func (h CohortData) RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, histogramConceptId int64, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) ([]*PersonConceptAndValue, error) {
	log.Printf(">> Using inner join impl. for large cohorts")
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// get the observations for the subjects and the concepts, to build up the data rows to return:
	var cohortData []*PersonConceptAndValue
	query := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation").
		Select("distinct(observation.person_id), observation.observation_concept_id as concept_id, observation.value_as_number as concept_value_as_number").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation.observation_concept_id = ?", histogramConceptId).
		Where("observation.value_as_number is not null")

	query = QueryFilterByConceptIdsAndCohortPairsHelper(query, filterConceptIds, filterCohortPairs, omopDataSource.Schema, resultsDataSource.Schema)

	meta_result := query.Scan(&cohortData)
	return cohortData, meta_result.Error
}

// Assesses the overlap between case and control cohorts. It does this after filtering the cohorts and keeping only
// the persons that have data for each of the selected conceptIds and match the filterConceptId/filterConceptValue criteria.
func (h CohortData) RetrieveCohortOverlapStats(sourceId int, caseCohortId int, controlCohortId int,
	filterConceptId int64, filterConceptValue int64, otherFilterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) (CohortOverlapStats, error) {

	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// count persons that are in the intersection of both case and control cohorts, filtering on filterConceptValue:
	var cohortOverlapStats CohortOverlapStats
	query := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation").
		Select("count(distinct(observation.person_id)) as case_control_overlap").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as case_cohort ON case_cohort.subject_id = observation.person_id").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as control_cohort ON control_cohort.subject_id = case_cohort.subject_id"). // this one allows for the intersection between case and control and the assessment of the overlap
		Where("case_cohort.cohort_definition_id = ?", caseCohortId).
		Where("control_cohort.cohort_definition_id = ?", controlCohortId).
		Where("observation.observation_concept_id = ?", filterConceptId).
		Where("observation.value_as_concept_id = ?", filterConceptValue)

	query = QueryFilterByConceptIdsAndCohortPairsHelper(query, otherFilterConceptIds, filterCohortPairs, omopDataSource.Schema, resultsDataSource.Schema)

	meta_result := query.Scan(&cohortOverlapStats)
	return cohortOverlapStats, meta_result.Error
}

// Basically the same as the method above, but without the extra filtering on filterConceptId and filterConceptValue:
func (h CohortData) RetrieveCohortOverlapStatsWithoutFilteringOnConceptValue(sourceId int, caseCohortId int, controlCohortId int,
	otherFilterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) (CohortOverlapStats, error) {

	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// count persons that are in the intersection of both case and control cohorts, filtering on filterConceptValue:
	var cohortOverlapStats CohortOverlapStats
	query := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation").
		Select("count(distinct(observation.person_id)) as case_control_overlap").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as case_cohort ON case_cohort.subject_id = observation.person_id").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as control_cohort ON control_cohort.subject_id = case_cohort.subject_id"). // this one allows for the intersection between case and control and the assessment of the overlap
		Where("case_cohort.cohort_definition_id = ?", caseCohortId).
		Where("control_cohort.cohort_definition_id = ?", controlCohortId)

	query = QueryFilterByConceptIdsAndCohortPairsHelper(query, otherFilterConceptIds, filterCohortPairs, omopDataSource.Schema, resultsDataSource.Schema)

	meta_result := query.Scan(&cohortOverlapStats)
	return cohortOverlapStats, meta_result.Error
}

func (p *PersonConceptAndCount) String() string {
	return fmt.Sprintf("(person_id=%d, concept_id=%d, count=%d)",
		p.PersonId, p.ConceptId, p.Count)
}

// Some observations are only expected once for each person. This code implements a validation that
// checks if any person has a duplicated entry for any of these observations and prints out WARNINGS
// to the log if this is the case.
func (h CohortData) ValidateObservationData(observationConceptIdsToCheck []int64) (int, error) {
	if len(observationConceptIdsToCheck) == 0 {
		log.Print("WARNING: no concepts configured for validation. Skipping data integrity check...")
		return -1, nil
	}
	var sourceModel = new(Source)
	sources, _ := sourceModel.GetAllSources()
	// run this validation on all available data sources:
	countIssues := 0
	for _, source := range sources {
		var dataSourceModel = new(Source)
		omopDataSource := dataSourceModel.GetDataSource(source.SourceId, Omop)

		log.Printf("INFO: checking if no duplicate data is found for concept ids %v in `observation` table of data source %d...",
			observationConceptIdsToCheck, source.SourceId)
		var personConceptAndCount []*PersonConceptAndCount
		query := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation").
			Select("observation.person_id, observation.observation_concept_id as concept_id, count(*)").
			Where("observation.observation_concept_id in (?)", observationConceptIdsToCheck).
			Group("observation.person_id, observation.observation_concept_id").
			Having("count(*) > 1")

		meta_result := query.Scan(&personConceptAndCount)
		if meta_result.Error != nil {
			return -1, meta_result.Error
		} else if len(personConceptAndCount) == 0 {
			log.Printf("INFO: no issues found in observation table of data source %d.", source.SourceId)
		} else {
			log.Printf("WARNING: !!! found a total of %d `person` records with duplicated `observation` entries for one or more concepts "+
				"where this is not expected (in data source=%d). These are the entries found: %v !!!",
				len(personConceptAndCount), source.SourceId, personConceptAndCount)
			countIssues += len(personConceptAndCount)
		}
	}
	return countIssues, nil
}
