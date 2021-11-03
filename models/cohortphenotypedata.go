package models

import (
	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

	dsn := utils.GenerateDsn(sourceConnectionString)

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
