package models

import (
	"github.com/uc-cdis/cohort-middleware/db"
)

type CohortDefinition struct {
	Id             int
	Name           string
	Description    string
	ExpressionType string `json:",omitempty"`
	CreatedById    int `json:",omitempty"`
	ModifiedById   int `json:",omitempty"`
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
