package models

import (
	"fmt"

	"log"

	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type CohortDefinitionI interface {
	GetCohortDefinitionById(id int) (*CohortDefinition, error)
	GetCohortDefinitionByName(name string) (*CohortDefinition, error)
	GetAllCohortDefinitions() ([]*CohortDefinition, error)
	GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int, teamProject string) ([]*CohortDefinitionStats, error)
	GetCohortName(cohortId int) (string, error)
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
	query := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("id = ?", id)
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortDefinition)
	return cohortDefinition, meta_result.Error
}

func (h CohortDefinition) GetCohortDefinitionByName(name string) (*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition *CohortDefinition
	query := db2.Model(&CohortDefinition{}).
		Select("id, name, description").
		Where("name = ?", name)
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortDefinition)
	return cohortDefinition, meta_result.Error
}

func (h CohortDefinition) GetAllCohortDefinitions() ([]*CohortDefinition, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinition []*CohortDefinition
	query := db2.Model(&CohortDefinition{}).
		Select("id, name, description")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortDefinition)
	return cohortDefinition, meta_result.Error
}

// Get the list of cohort_definition ids for a given "team project" (where "team project" is basically
// a security role name of one of the roles in Atlas/WebAPI database).
func (h CohortDefinition) GetCohortDefinitionIdsForTeamProject(teamProject string) ([]int, error) {
	db2 := db.GetAtlasDB().Db
	var cohortDefinitionIds []int
	query := db2.Table(db.GetAtlasDB().Schema+".cohort_definition_sec_role").
		Select("cohort_definition_id").
		Where("sec_role_name = ?", teamProject).
		Scan(&cohortDefinitionIds)
	return cohortDefinitionIds, query.Error
}

func (h CohortDefinition) GetAllCohortDefinitionsAndStatsOrderBySizeDesc(sourceId int, teamProject string) ([]*CohortDefinitionStats, error) {

	// Connect to source db and gather stats:
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	var cohortDefinitionStats []*CohortDefinitionStats
	query := resultsDataSource.Db.Model(&Cohort{}).
		Select("cohort_definition_id as id, '' as name, count(*) as cohort_size").
		Group("cohort_definition_id").
		Order("count(*) desc")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortDefinitionStats)

	// get (from separate Atlas DB - hence not using JOIN above) the list of cohort_definition_ids
	// that are allowed for the given teamProject:
	allowedCohortDefinitionIds, _ := h.GetCohortDefinitionIdsForTeamProject(teamProject)
	log.Printf("INFO: found %d cohorts for this team project", len(allowedCohortDefinitionIds))

	// add name details:
	finalList := []*CohortDefinitionStats{}
	for _, cohortDefinitionStat := range cohortDefinitionStats {
		var cohortDefinition, _ = h.GetCohortDefinitionById(cohortDefinitionStat.Id)
		if cohortDefinition == nil {
			// unexpected issue: cohortDefinition not found. Warn and skip:
			log.Printf("WARNING: found a cohort of size %d with missing cohort_definition record",
				cohortDefinitionStat.CohortSize)
			continue
		} else {
			cohortDefinitionStat.Name = cohortDefinition.Name
			finalList = append(finalList, cohortDefinitionStat)
		}
	}
	// filter to keep only the allowed ones:
	filteredFinalList := []*CohortDefinitionStats{}
	for _, cohortDefinitionStat := range finalList {
		if utils.Contains(allowedCohortDefinitionIds, cohortDefinitionStat.Id) {
			filteredFinalList = append(filteredFinalList, cohortDefinitionStat)
		}
	}

	return filteredFinalList, meta_result.Error
}

func (h CohortDefinition) GetCohortName(cohortId int) (string, error) {
	cohortDefinition, err := h.GetCohortDefinitionById(cohortId)
	if err != nil {
		return "", fmt.Errorf("could not retrieve cohort name due to error: %s", err.Error())
	}

	return cohortDefinition.Name, nil
}
