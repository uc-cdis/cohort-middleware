package models

import (
	"log"
	"strconv"
	"strings"
)

type Concept struct {
	ConceptId   int    `json:"concept_id"`
	ConceptName string `json:"concept_name"`
	DomainId    string `json:"domain_id"`
	DomainName  string `json:"domain_name"`
}

type ConceptAndPersonsWithDataStats struct {
	ConceptId  int
	NpersonIds int
}

type ConceptStats struct {
	ConceptId         int     `json:"concept_id"`
	PrefixedConceptId string  `json:"prefixed_concept_id"`
	ConceptName       string  `json:"concept_name"`
	DomainId          string  `json:"domain_id"`
	DomainName        string  `json:"domain_name"`
	CohortSize        int     `json:"cohort_size"`
	NmissingRatio     float32 `json:"n_missing_ratio"`
}

type Observation struct {
}

func (h Concept) RetriveAllBySourceId(sourceId int) ([]*Concept, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var concept []*Concept
	result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name").
		Joins("INNER JOIN " + omopDataSource.Schema + ".domain as domain ON concept.domain_id = domain.domain_id").
		Order("concept_name").
		Scan(&concept)
	return concept, result.Error
}

func (h Concept) GetConceptBySourceIdAndConceptId(sourceId int, conceptId int) *Concept {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var concept *Concept
	omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name").
		Joins("INNER JOIN "+omopDataSource.Schema+".domain as domain ON concept.domain_id = domain.domain_id"). //TODO - this is crashing with Out of Memory...limit it?? Add paging?
		Where("concept_id = ?", conceptId).
		Scan(&concept)
	return concept
}

// Very simple function, just to add a prefix in front of the conceptId.
// It is a public method here, since it is needed in different places...
// ...so we need to keep it consistent:
func (h Concept) GetPrefixedConceptId(conceptId int) string {
	return "ID_" + strconv.Itoa(conceptId)
}

// The reverse of above function:
func (h Concept) GetConceptId(prefixedConceptId string) int {
	// validate: it should start with ID_
	if strings.Index(prefixedConceptId, "ID_") != 0 {
		log.Panicf("Prefixed concept id should start with ID_ . However, found this instead: %s", prefixedConceptId)
	}
	var conceptId = strings.Split(prefixedConceptId, "ID_")[1]
	var result, _ = strconv.Atoi(conceptId)
	return result
}

// DEPRECATED - too slow for larger cohorts
func (h Concept) RetrieveStatsBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int) ([]*ConceptStats, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptStats []*ConceptStats
	result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name, 0 as n_missing_ratio").
		Joins("INNER JOIN "+omopDataSource.Schema+".domain as domain ON concept.domain_id = domain.domain_id").
		Where("concept_id in (?)", conceptIds).
		Order("concept_name").
		Scan(&conceptStats)
	if result.Error != nil {
		return nil, result.Error
	}

	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var cohortSubjectIds []int
	result = resultsDataSource.Db.Model(&Cohort{}).
		Select("subject_id").
		Where("cohort_definition_id = ?", cohortDefinitionId).
		Scan(&cohortSubjectIds)
	if result.Error != nil {
		return nil, result.Error
	}

	var cohortSize = len(cohortSubjectIds)

	// find, for each concept, the ratio of persons in the given cohortId that have
	// an empty value for this concept:
	for _, conceptStat := range conceptStats {
		// since we are looping over items anyway, also set prefixed_concept_id and cohort_size:
		conceptStat.PrefixedConceptId = h.GetPrefixedConceptId(conceptStat.ConceptId)
		conceptStat.CohortSize = cohortSize
		if cohortSize == 0 {
			conceptStat.NmissingRatio = 0
		} else {
			// calculate missing ratio for cohorts that actually have a size:
			var nrPersonsWithData int
			omopDataSource.Db.Model(&Observation{}).
				Select("count(distinct(person_id))").
				Where("observation_concept_id = ?", conceptStat.ConceptId).
				Where("person_id in (?)", cohortSubjectIds).
				Where("(value_as_string is not null or value_as_number is not null)").
				Scan(&nrPersonsWithData)
			log.Printf("Found %d persons with data for concept_id %d", nrPersonsWithData, conceptStat.ConceptId)
			n_missing := cohortSize - nrPersonsWithData
			conceptStat.NmissingRatio = float32(n_missing) / float32(cohortSize)
		}
	}

	return conceptStats, nil
}

func getNrPersonsWithData(conceptId int, conceptsAndPersonsWithData []*ConceptAndPersonsWithDataStats) int {
	for _, conceptsAndDataInfo := range conceptsAndPersonsWithData {
		if conceptsAndDataInfo.ConceptId == conceptId {
			return conceptsAndDataInfo.NpersonIds
		}
	}
	return 0
}

// Same as above, but for LARGE cohorts/dbs
// Assumption is that both OMOP and RESULTS schemas
// are on same DB
func (h Concept) RetrieveStatsLargeBySourceIdAndCohortIdAndConceptIds(sourceId int, cohortDefinitionId int, conceptIds []int) ([]*ConceptStats, error) {
	log.Printf(">> Using inner join impl. for large cohorts")
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	var conceptStats []*ConceptStats
	result := omopDataSource.Db.Model(&Concept{}).
		Select("concept_id, concept_name, domain.domain_id, domain.domain_name, 0 as n_missing_ratio").
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
	// an empty value for this concept:
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
		conceptStat.PrefixedConceptId = h.GetPrefixedConceptId(conceptStat.ConceptId)
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
