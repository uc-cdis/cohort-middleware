package models

import (
	"fmt"
	"log"

	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortDataI interface {
	RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int64) ([]*PersonConceptAndValue, error)
	RetrieveCohortOverlapStats(sourceId int, caseCohortId int, controlCohortId int, otherFilterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) (CohortOverlapStats, error)
	RetrieveDataByOriginalCohortAndNewCohort(sourceId int, originalCohortDefinitionId int, cohortDefinitionId int) ([]*PersonIdAndCohort, error)
	RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, histogramConceptId int64, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) ([]*PersonConceptAndValue, error)
	RetrieveBarGraphDataBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, histogramConceptId int64) ([]*OrdinalGroupData, error)
}

type CohortData struct{}

type PersonConceptAndValue struct {
	PersonId                int64
	ConceptId               int64
	ConceptClassId          string
	ConceptValueName        string
	ConceptValueAsNumber    *float32
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

type Person struct {
	PersonId int64
}

type OrdinalGroupData struct {
	Name             string `json:"name"`
	PersonCount      int64  `json:"personCount"`
	ValueAsString    string `json:"valueAsString"`
	ValueAsConceptID int64  `json:"valueAsConceptID"`
}

// This function returns the subjects that belong to both cohorts (the intersection of both cohorts)
// TODO - name this function as such
func (h CohortData) RetrieveDataByOriginalCohortAndNewCohort(sourceId int, originalCohortDefinitionId int, cohortDefinitionId int) ([]*PersonIdAndCohort, error) {
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var personData []*PersonIdAndCohort

	query := resultsDataSource.Db.Model(&Cohort{}).
		Select("cohort.subject_id as person_id, cohort.cohort_definition_id as cohort_id").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as original_cohort ON cohort.subject_id = original_cohort.subject_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("original_cohort.cohort_definition_id = ?", originalCohortDefinitionId)
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&personData)
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
	query := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation"+omopDataSource.GetViewDirective()).
		Select("observation.person_id, observation.observation_concept_id as concept_id, concept.concept_class_id, hare_concept.concept_name as concept_value_name, observation.value_as_number as concept_value_as_number, observation.value_as_concept_id as concept_value_as_concept_id").
		Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort ON cohort.subject_id = observation.person_id").
		Joins("INNER JOIN "+omopDataSource.Schema+".concept as concept ON concept.concept_id = observation.observation_concept_id").
		Joins("LEFT JOIN "+omopDataSource.Schema+".concept as hare_concept ON hare_concept.concept_id = observation.value_as_concept_id").
		Where("cohort.cohort_definition_id = ?", cohortDefinitionId).
		Where("observation.observation_concept_id in (?)", conceptIds).
		Order("observation.person_id asc") // this order is important!
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortData)
	return cohortData, meta_result.Error
}

func (h CohortData) RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sourceId int, cohortDefinitionId int, histogramConceptId int64, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) ([]*PersonConceptAndValue, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	// get the observations for the subjects and the concepts, to build up the data rows to return:
	var cohortData []*PersonConceptAndValue
	query := QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, cohortDefinitionId, "unionAndIntersect").
		Select("distinct(observation.person_id), observation.observation_concept_id as concept_id, observation.value_as_number as concept_value_as_number").
		Joins("INNER JOIN "+omopDataSource.Schema+".observation_continuous as observation"+omopDataSource.GetViewDirective()+" ON unionAndIntersect.subject_id = observation.person_id").
		Where("observation.observation_concept_id = ?", histogramConceptId).
		Where("observation.value_as_number is not null")

	query = QueryFilterByConceptIdsHelper(query, sourceId, filterConceptIds, omopDataSource, resultsDataSource.Schema, "unionAndIntersect.subject_id")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortData)
	return cohortData, meta_result.Error
}

func (h CohortData) RetrieveBarGraphDataBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptId int64) ([]*OrdinalGroupData, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	// get the observations for the subjects and the concepts, to build up the data rows to return:
	var cohortData []*OrdinalGroupData

	query := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation"+omopDataSource.GetViewDirective()).
		Select("c1.concept_name as name, count(distinct person_id) as person_count,observation.value_as_string as value_as_string, value_as_concept_id as value_as_concept_id").
		Joins("INNER JOIN "+omopDataSource.Schema+".concept as c ON c.concept_id = observation.observation_concept_id").
		Joins("LEFT JOIN "+omopDataSource.Schema+".concept as c1 ON c1.concept_id = observation.value_as_concept_id").
		Where("c.concept_class_id = ?", "MVP Ordinal").
		Where("c.concept_id = ?", conceptId).
		Group("observation.observation_concept_id, observation.value_as_string, observation.value_as_concept_id, c1.concept_name")

	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortData)
	return cohortData, meta_result.Error
}

// Basically the same as the method above, but without the extra filtering on filterConceptId and filterConceptValue:
func (h CohortData) RetrieveCohortOverlapStats(sourceId int, caseCohortId int, controlCohortId int,
	filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef) (CohortOverlapStats, error) {

	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	var cohortOverlapStats CohortOverlapStats
	query := QueryFilterByCohortPairsHelper(filterCohortPairs, resultsDataSource, caseCohortId, "case_cohort_unionedAndIntersectedWithFilters").
		Select("count(distinct(case_cohort_unionedAndIntersectedWithFilters.subject_id)) as case_control_overlap").
		Joins("INNER JOIN " + resultsDataSource.Schema + ".cohort as control_cohort ON control_cohort.subject_id = case_cohort_unionedAndIntersectedWithFilters.subject_id") // this one allows for the intersection between case and control and the assessment of the overlap

	if len(filterConceptIds) > 0 {
		query = QueryFilterByConceptIdsHelper(query, sourceId, filterConceptIds, omopDataSource, resultsDataSource.Schema, "control_cohort.subject_id")
	}
	query = query.Where("control_cohort.cohort_definition_id = ?", controlCohortId)
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
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
		query := omopDataSource.Db.Table(omopDataSource.Schema+".observation_continuous as observation"+omopDataSource.GetViewDirective()).
			Select("observation.person_id, observation.observation_concept_id as concept_id, count(*)").
			Where("observation.observation_concept_id in (?)", observationConceptIdsToCheck).
			Group("observation.person_id, observation.observation_concept_id").
			Having("count(*) > 1")

		query, cancel := utils.AddTimeoutToQuery(query)
		defer cancel()
		meta_result := query.Scan(&personConceptAndCount)
		if meta_result.Error != nil {
			return -1, meta_result.Error
		} else if len(personConceptAndCount) == 0 {
			log.Printf("INFO: no issues found in observation table of data source %d.", source.SourceId)
		} else {
			log.Printf("WARNING: !!! found a total of %d `person` records with duplicated `observation` entries for one or more concepts "+
				"where this is not expected (in data source=%d).",
				len(personConceptAndCount), source.SourceId)
			countIssues += len(personConceptAndCount)
		}
	}
	return countIssues, nil
}
