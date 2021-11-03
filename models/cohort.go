package models

import (
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/utils"
	"time"
)

type Cohort struct {
	CohortDefinitionId int `json:",omitempty"`
	SubjectId int64
	CohortStartDate time.Time
	CohortEndDate time.Time
}

func (h Cohort) GetCohortById(id int) ([]*Cohort, error) {
	db2 := db.GetAtlasDB()
	var cohort []*Cohort
	db2.Model(&Cohort{}).Select("cohort_definition_id, subject_id, cohort_start_date, cohort_end_date").Where("cohort_definition_id = ?", id).Scan(&cohort)
	return cohort, nil
}

var cohortDefinitionModel = new(CohortDefinition)

func (h Cohort) GetCohortByName(datasourcename string, cohortname string) ([]*Cohort, error) {
	var dataSourceModel = new(Source)
	dataSource, _ := dataSourceModel.GetSourceByNameWithConnection(datasourcename)

	sourceConnectionString := dataSource.SourceConnection
	dbSchema := "RESULTS."
	omopDataSource := utils.GetDataSourceDB(sourceConnectionString, dbSchema)

	cohortDefinition, _ := cohortDefinitionModel.GetCohortDefinitionByName(cohortname)
	cohortDefinitionId := cohortDefinition.Id

	var cohort []*Cohort
	omopDataSource.Model(&Cohort{}).Select("cohort_definition_id, subject_id, cohort_start_date, cohort_end_date").Where("cohort_definition_id = ?", cohortDefinitionId).Scan(&cohort)
	return cohort, nil
}
