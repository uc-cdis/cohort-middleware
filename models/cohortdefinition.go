package models

import (
	"github.com/uc-cdis/cohort-middleware/db"
)

type CohortDefinitionI interface {
	GetCohortDefinitionById(id int) (*CohortDefinition, error)
	GetCohortDefinitionByName(name string) (*CohortDefinition, error)
	GetAllCohortDefinitions() ([]*CohortDefinition, error)
	GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int) ([]*CohortDefinitionStats, error)
}

type CohortDefinition struct {
	Id             int    `json:"cohort_definition_id"`
	Name           string `json:"cohort_name"`
	Description    string `json:"cohort_description"`
	ExpressionType string `json:",omitempty"`
	CreatedById    int    `json:",omitempty"`
	ModifiedById   int    `json:",omitempty"`
}

type CohortDefinitionStats struct {
	Id         int    `json:"cohort_definition_id"`
	Name       string `json:"cohort_name"`
	CohortSize int    `json:"size"`
}

func (h CohortDefinition) GetCohortDefinitionById(id int) (*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition *CohortDefinition
	result := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("id = ?", id).
		Scan(&cohortDefinition)
	return cohortDefinition, result.Error
}

func (h CohortDefinition) GetCohortDefinitionByName(name string) (*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition *CohortDefinition
	result := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("name = ?", name).
		Scan(&cohortDefinition)
	return cohortDefinition, result.Error
}

func (h CohortDefinition) GetAllCohortDefinitions() ([]*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition []*CohortDefinition
	result := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Scan(&cohortDefinition)
	return cohortDefinition, result.Error
}

func (h CohortDefinition) GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int) ([]*CohortDefinitionStats, error) {

	// Connect to source db and gather stats:
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var cohortDefinitionStats []*CohortDefinitionStats
	result := resultsDataSource.Db.Model(&Cohort{}).
		Select("cohort_definition_id as id, '' as name, count(*) as cohort_size").
		Group("cohort_definition_id").
		Order("count(*) desc").
		Scan(&cohortDefinitionStats)

	// add name details:
	for _, cohortDefinitionStat := range cohortDefinitionStats {
		var cohortDefinition, _ = h.GetCohortDefinitionById(cohortDefinitionStat.Id)
		cohortDefinitionStat.Name = cohortDefinition.Name
	}
	return cohortDefinitionStats, result.Error
}
