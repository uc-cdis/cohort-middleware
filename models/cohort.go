package models

import (
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type Cohort struct {
	CohortDefinitionId int
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

var omopDataSource *gorm.DB

var cohortDefinitionModel = new(CohortDefinition)

func (h Cohort) GetCohortByName(datasourcename string, cohortname string) ([]*Cohort, error) {
	var dataSourceModel = new(Source)
	dataSource, _ := dataSourceModel.GetSourceByNameWithConnection(datasourcename)

	sourceConnectionString := dataSource.SourceConnection
	dsn := utils.GenerateDsn(sourceConnectionString)

	omopDataSource, _ = gorm.Open(sqlserver.Open(dsn),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "RESULTS.",
				SingularTable: true,
			}})

	cohortDefinition, _ := cohortDefinitionModel.GetCohortDefinitionByName(cohortname)
	cohortDefinitionId := cohortDefinition.Id

	var cohort []*Cohort
	omopDataSource.Model(&Cohort{}).Select("cohort_definition_id, subject_id, cohort_start_date, cohort_end_date").Where("cohort_definition_id = ?", cohortDefinitionId).Scan(&cohort)
	return cohort, nil
}
