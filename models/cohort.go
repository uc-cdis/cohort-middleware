package models

import (
	"fmt"
	"github.com/uc-cdis/cohort-middleware/db"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/url"
	"strings"
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

	sourceConnectionParts := strings.FieldsFunc(sourceConnectionString, func(r rune) bool {
		separators := ":/;="
		if strings.ContainsRune(separators, r) {
			return true
		}
		return false
	})
	host := sourceConnectionParts[2]
	port := sourceConnectionParts[3]
	dbname := sourceConnectionParts[5]
	username := sourceConnectionParts[7]
	password := sourceConnectionParts[9]

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		username,
		url.QueryEscape(password),
		host,
		port,
		dbname)

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
