package models

import (
	"log"

	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortData struct {
	PersonId int `gorm:"column:sample.id"`
	Age      int
	Gender   string
	Hare     string
	CDWrace  string `gorm:"column:CDW_race"`
	Height   float32
}

type PersonConceptAndValue struct {
	PersonId             int
	ConceptId            int
	ConceptName          string
	ConceptValueAsString string
	ConceptValueAsNumber float32
}

func (h CohortData) GetCohortData(datasourcename string) ([]*CohortPhenotypeData, error) {
	var dataSourceModel = new(Source)
	dataSource, _ := dataSourceModel.GetSourceByNameWithConnection(datasourcename)

	sourceConnectionString := dataSource.SourceConnection
	dbSchema := "MISC_OMOP."
	omopDataSource := utils.GetDataSourceDB(sourceConnectionString, dbSchema)

	var cohortDataPhenotype []*CohortPhenotypeData
	omopDataSource.Model(&CohortPhenotypeData{}).Select("[sample.id], age, gender, Hare, CDW_race, Height").Scan(&cohortDataPhenotype)
	return cohortDataPhenotype, nil
}

func (h CohortData) RetrieveDataBySourceIdAndCohortIdAndConceptIdsOrderedByPersonId(sourceId int, cohortDefinitionId int, conceptIds []int) ([]*PersonConceptAndValue, error) {
	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, "OMOP")

	resultsDataSource := dataSourceModel.GetDataSource(sourceId, "RESULTS")
	var cohortSubjectIds []int
	resultsDataSource.Model(&Cohort{}).
		Select("subject_id").
		Where("cohort_definition_id = ?", cohortDefinitionId).
		Scan(&cohortSubjectIds)
	log.Printf("Querying cohort of size %d", len(cohortSubjectIds))

	// get the observations for the subjects and the concepts, to build up the data rows to return:
	var cohortData []*PersonConceptAndValue
	omopDataSource.Model(&Observation{}).
		Select("observation.person_id, concept.concept_id, concept.concept_name, observation.value_as_string as concept_value_as_string, observation.value_as_number as concept_value_as_number").
		Joins("INNER JOIN OMOP.concept as concept ON observation.observation_concept_id = concept.concept_id").
		Where("concept.concept_id in (?)", conceptIds).
		Where("observation.person_id in (?)", cohortSubjectIds).
		Order("observation.person_id asc"). // this order is important!
		Scan(&cohortData)

	return cohortData, nil
}
