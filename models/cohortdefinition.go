package models

import (
	"github.com/uc-cdis/cohort-middleware/db"
)

type CohortDefinitionI interface {
	GetCohortDefinitionById(id int) (*CohortDefinition, error)
	GetCohortDefinitionByName(name string) (*CohortDefinition, error)
	GetAllCohortDefinitions() ([]*CohortDefinition, error)
	GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int) ([]*CohortDefinitionStats, error)
	GetCohortDefinitionsAndStatsBySourceIdAndCohortIdAndBreakDownOnConceptId(sourceId int, cohortId int, conceptId int) (*CohortDefinitionStatsBreakdown, error)
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

type CohortDefinitionStatsBreakdown struct {
	Id               int              `json:"cohort_definition_id"`
	Name             string           `json:"cohort_name"`
	CohortSize       int              `json:"size"`
	ConceptBreakdown ConceptBreakdown `json:"concept_breakdown"`
}

type ConceptBreakdown struct {
	ConceptId   int                 `json:"concept_id"`
	ConceptName string              `json:"concept_name"`
	Breakdown   []ConceptValueStats `json:"breakdown"`
}

type ConceptValueStats struct {
	ConceptValue              string `json:"concept_value"`
	NPersonsInCohortWithValue int    `json:"persons_in_cohort_with_value"`
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

func (h CohortDefinition) GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int) ([]*CohortDefinitionStats, error) {

	// Connect to source db and gather stats:
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var cohortDefinitionStats []*CohortDefinitionStats
	resultsDataSource.Db.Model(&Cohort{}).
		Select("cohort_definition_id as id, '' as name, count(*) as cohort_size").
		Group("cohort_definition_id").
		Order("count(*) desc").
		Scan(&cohortDefinitionStats)

	// add name details:
	for _, cohortDefinitionStat := range cohortDefinitionStats {
		var cohortDefinition, _ = h.GetCohortDefinitionById(cohortDefinitionStat.Id)
		cohortDefinitionStat.Name = cohortDefinition.Name
	}
	return cohortDefinitionStats, nil
}

func (h CohortDefinition) GetCohortDefinitionsAndStatsBySourceIdAndCohortIdAndBreakDownOnConceptId(sourceId int, cohortId int, conceptId int) (*CohortDefinitionStatsBreakdown, error) {

	cohortDefinitionStats := CohortDefinitionStatsBreakdown{
		Id:         1,
		CohortSize: 999999,
		Name:       "cohort name1",
		ConceptBreakdown: ConceptBreakdown{
			ConceptId:   1,
			ConceptName: "TEST CONCEPT",
			Breakdown: []ConceptValueStats{
				{ConceptValue: "ABC", NPersonsInCohortWithValue: 222},
				{ConceptValue: "DEF", NPersonsInCohortWithValue: 333},
			},
		},
	}
	return &cohortDefinitionStats, nil
}
