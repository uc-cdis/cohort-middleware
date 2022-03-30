package models

// DEPRECATED!
import (
	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortPhenotypeData struct {
	SampleId string `gorm:"column:sample.id"`
	Age      int
	Gender   string
	Hare     string
	CDWrace  string `gorm:"column:CDW_race"`
	Height   float32
}

func (h CohortPhenotypeData) GetCohortDataPhenotype(datasourcename string) ([]*CohortPhenotypeData, error) {
	var dataSourceModel = new(Source)
	dataSource, _ := dataSourceModel.GetSourceByNameWithConnection(datasourcename)

	sourceConnectionString := dataSource.SourceConnection
	dbSchema := "MISC_OMOP."
	omopDataSource := utils.GetDataSourceDB(sourceConnectionString, dbSchema)

	var cohortDataPhenotype []*CohortPhenotypeData
	omopDataSource.Model(&CohortPhenotypeData{}).Select("[sample.id], age, gender, Hare, CDW_race, Height").Scan(&cohortDataPhenotype)
	return cohortDataPhenotype, nil
}
