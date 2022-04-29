package models

import (
	"github.com/uc-cdis/cohort-middleware/db"
)

type CohortDefinitionI interface {
	GetCohortDefinitionById(id int) (*CohortDefinition, error)
	GetCohortDefinitionByName(name string) (*CohortDefinition, error)
	GetAllCohortDefinitions() ([]*CohortDefinition, error)
	GetAllCohortDefinitionsAndStats(sourceId int) ([]*CohortDefinitionStats, error)
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
	db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("id = ?", id).
		Scan(&cohortDefinition)
	return cohortDefinition, nil
}

func (h CohortDefinition) GetCohortDefinitionByName(name string) (*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition *CohortDefinition
	result := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("name = ?", name).
		Scan(&cohortDefinition)
	if result.Error != nil {
		return nil, result.Error
	}
	return cohortDefinition, nil
}

func (h CohortDefinition) GetAllCohortDefinitions() ([]*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition []*CohortDefinition
	db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Scan(&cohortDefinition)
	return cohortDefinition, nil
}

func (h CohortDefinition) GetAllCohortDefinitionsAndStats(sourceId int) ([]*CohortDefinitionStats, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinitions []*CohortDefinitionStats
	db2.Model(&CohortDefinition{}).
		Select("id, name, null as cohort_size").
		Scan(&cohortDefinitions)

	// Connect to source db and gather stats:
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	for _, cohortDefinition := range cohortDefinitions {
		var cohortSize int
		result := resultsDataSource.Db.Model(&Cohort{}).
			Select("count(*)").
			Where("cohort_definition_id = ?", cohortDefinition.Id).
			Scan(&cohortSize)
		if result.Error != nil {
			return nil, result.Error
		}
		cohortDefinition.CohortSize = cohortSize
	}
	return cohortDefinitions, nil
}
