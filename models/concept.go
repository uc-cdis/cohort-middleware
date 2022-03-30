package models

import (
	"log"
)

type Concept struct {
	ConceptId   int    `json:"concept_id"`
	ConceptName string `json:"concept_name"`
	DomainId    string `json:"domain_id"`
	DomainName  string `json:"domain_name"`
}

type ConceptStats struct {
	ConceptId     int     `json:"concept_id"`
	ConceptName   string  `json:"concept_name"`
	NmissingRatio float32 `json:"n_missing_ratio"`
}

type Observation struct {
}

func (h Concept) RetriveAllBySourceId(sourceId int) ([]*Concept, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, "OMOP")

	var concept []*Concept
	omopDataSource.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name").
		Joins("INNER JOIN OMOP.domain as domain ON concept.domain_id = domain.domain_id").
		Order("concept_name").
		Scan(&concept)
	return concept, nil
}

func (h Concept) GetConceptBySourceIdAndConceptId(sourceId int, conceptId int) *Concept {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, "OMOP")

	var concept *Concept
	omopDataSource.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name").
		Joins("INNER JOIN OMOP.domain as domain ON concept.domain_id = domain.domain_id").
		Where("concept_id = ?", conceptId).
		Scan(&concept)
	return concept
}

func (h Concept) RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int) ([]*ConceptStats, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, "OMOP")

	var conceptStats []*ConceptStats
	omopDataSource.Model(&Concept{}).
		Select("concept_id, concept_name, domain_id, '' as domain_name, 0 as n_missing_ratio").
		Where("concept_id in (?)", conceptIds).
		Order("concept_name").
		Scan(&conceptStats)

	resultsDataSource := dataSourceModel.GetDataSource(sourceId, "RESULTS")
	var cohortSubjectIds []int
	resultsDataSource.Model(&Cohort{}).
		Select("subject_id").
		Where("cohort_definition_id = ?", cohortDefinitionId).
		Scan(&cohortSubjectIds)
	var cohortSize = len(cohortSubjectIds)

	// find, for each concept, the ratio of persons in the given cohortId that have
	// an empty value for this concept:
	for _, conceptStat := range conceptStats {

		var nrPersonsWithData int
		omopDataSource.Model(&Observation{}).
			Select("count(distinct(person_id))").
			Where("observation_concept_id = ?", conceptStat.ConceptId).
			Where("person_id in (?)", cohortSubjectIds).
			Where("(value_as_string is not null or value_as_number is not null)").
			Scan(&nrPersonsWithData)
		log.Printf("Found %d persons with data for concept_id %d", nrPersonsWithData, conceptStat.ConceptId)
		n_missing := cohortSize - nrPersonsWithData
		conceptStat.NmissingRatio = float32(n_missing) / float32(cohortSize)
	}

	return conceptStats, nil
}
