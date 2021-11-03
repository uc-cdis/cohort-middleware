package models

import (
	"fmt"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/url"
	"strings"
)

type CohortPhenotypeData struct {
	SampleId int
	Age int
	Gender string
	Hare string
	CDWrace string
	Height float32
}

func (h CohortPhenotypeData) GetCohortDataPhenotype(datasourcename string) ([]*CohortPhenotypeData, error) {
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
				TablePrefix:   "MISC_OMOP.",
				SingularTable: true,
			}})

	var cohortDataPhenotype []*CohortPhenotypeData
	omopDataSource.Model(&CohortPhenotypeData{}).Select("[sample.id], age, gender, Hare, CDW_race, Height").Scan(&cohortDataPhenotype)
	return cohortDataPhenotype, nil
}
