package models

import (
	"fmt"

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
	meta_result := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("id = ?", id).
		Scan(&cohortDefinition)
	return cohortDefinition, meta_result.Error
}

func (h CohortDefinition) GetCohortDefinitionByName(name string) (*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition *CohortDefinition
	meta_result := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("name = ?", name).
		Scan(&cohortDefinition)
	return cohortDefinition, meta_result.Error
}

func (h CohortDefinition) GetAllCohortDefinitions() ([]*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition []*CohortDefinition
	meta_result := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Scan(&cohortDefinition)
	return cohortDefinition, meta_result.Error
}

func (h CohortDefinition) GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int) ([]*CohortDefinitionStats, error) {

	// Connect to source db and gather stats:
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var cohortDefinitionStats []*CohortDefinitionStats
	meta_result := resultsDataSource.Db.Model(&Cohort{}).
		Select("cohort_definition_id as id, '' as name, count(*) as cohort_size").
		Group("cohort_definition_id").
		Order("count(*) desc").
		Scan(&cohortDefinitionStats)

	// add name details:
	for _, cohortDefinitionStat := range cohortDefinitionStats {
		var cohortDefinition, _ = h.GetCohortDefinitionById(cohortDefinitionStat.Id)
		cohortDefinitionStat.Name = cohortDefinition.Name
	}
	return cohortDefinitionStats, meta_result.Error
}

func (h CohortDefinition) GetCohortName(cohortId int) (string, error) {
	cohortDefinition, err := h.GetCohortDefinitionById(cohortId)
	if err != nil {
		return "", fmt.Errorf("Could not retrieve cohort name due to error: %s", err.Error())
	}

	return cohortDefinition.Name, nil
}
