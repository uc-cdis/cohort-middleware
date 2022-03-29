package models

import (
	"github.com/uc-cdis/cohort-middleware/db"
)

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
	db2 := db.GetAtlasDB()
	var cohortDefinition *CohortDefinition
	db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("id = ?", id).
		Scan(&cohortDefinition)
	return cohortDefinition, nil
}

func (h CohortDefinition) GetCohortDefinitionByName(name string) (*CohortDefinition, error) {
	db2 := db.GetAtlasDB()
	var cohortDefinition *CohortDefinition
	db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("name = ?", name).
		Scan(&cohortDefinition)
	return cohortDefinition, nil
}

func (h CohortDefinition) GetAllCohortDefinitions() ([]*CohortDefinition, error) {
	db2 := db.GetAtlasDB()
	var cohortDefinition []*CohortDefinition
	db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Scan(&cohortDefinition)
	return cohortDefinition, nil
}

func (h CohortDefinition) GetAllCohortDefinitionsAndStats(sourceId int) ([]*CohortDefinitionStats, error) {
	db2 := db.GetAtlasDB()
	var cohortDefinitions []*CohortDefinitionStats
	db2.Model(&CohortDefinition{}).
		Select("id, name, null as cohort_size").
		Scan(&cohortDefinitions)

	// Connect to source db and gather stats:
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, "RESULTS")
	for _, cohortDefinition := range cohortDefinitions {
		var cohortSize int
		resultsDataSource.Model(&Cohort{}).
			Select("count(*)").
			Where("cohort_definition_id = ?", cohortDefinition.Id).
			Scan(&cohortSize)
		cohortDefinition.CohortSize = cohortSize
	}
	return cohortDefinitions, nil
}
