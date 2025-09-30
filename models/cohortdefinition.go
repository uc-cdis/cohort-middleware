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
	GetCohortDefinitionIdsForTeamProject(teamProject string) ([]int, error)
	GetTeamProjectsThatMatchAllCohortDefinitionIds(uniqueCohortDefinitionIdsList []int) ([]string, error)
	GetCohortDefinitionStatsByObservationWindow(sourceId int, cohortId int, observationWindow int) (*CohortDefinitionStats, error)
	GetCohortDefinitionStatsByObservationWindow1stCohortAndOverlap2ndCohort(sourceId int, cohort1Id int, cohort2Id int, observationWindow1stCohort int) (*CohortDefinitionStats, error)
	GetCohortDefinitionStatsByObservationWindow1stCohortAndOverlap2ndCohortAndOutcomeWindow2ndCohort(sourceId int, cohort1Id int, cohort2Id int, observationWindow1stCohort int, outcomeWindow2ndCohort int) (*CohortDefinitionStats, error)
}

type CohortDefinition struct {
	Id             int    `json:"cohort_definition_id"`
	Name           string `json:"cohort_name"`
	Description    string `json:"cohort_description"`
	ExpressionType string `json:",omitempty"`
	CreatedById    int    `json:",omitempty"`
	ModifiedById   int    `json:",omitempty"`
	Expression     string `json:",omitempty"`
}

type CohortDefinitionStats struct {
	Id         int    `json:"cohort_definition_id"`
	Name       string `json:"cohort_name"`
	CohortSize int    `json:"size"`
}

func (h CohortDefinition) GetCohortDefinitionById(id int) (*CohortDefinition, error) {
	atlasDb := db.GetAtlasDB()
	db2 := db.GetAtlasDB().Db
	var cohortDefinition *CohortDefinition
	query := db2.Model(&CohortDefinition{}).
		Select("cohort_definition.id, cohort_definition.name, cohort_definition.description, cohort_definition_details.expression").
		Where("cohort_definition.id = ?", id).
		Joins("INNER JOIN " + atlasDb.Schema + ".cohort_definition_details ON cohort_definition.id = cohort_definition_details.id")

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

// Returns any "team project" entries that are matched to _each and every one_ of the
// cohort definition ids found in uniqueCohortDefinitionIdsList.
func (h CohortDefinition) GetTeamProjectsThatMatchAllCohortDefinitionIds(uniqueCohortDefinitionIdsList []int) ([]string, error) {

	db2 := db.GetAtlasDB().Db
	var teamProjects []string
	// Find any roles that are paired to each and every one of the cohort_definition_id values.
	// Roles that ony match part of the values are filtered out by the having(count) clause:
	query := db2.Table(db.GetAtlasDB().Schema+".cohort_definition_sec_role").
		Select("sec_role_name").
		Where("cohort_definition_id in (?)", uniqueCohortDefinitionIdsList).
		Group("sec_role_name").
		Having("count(DISTINCT cohort_definition_id) = ?", len(uniqueCohortDefinitionIdsList)).
		Scan(&teamProjects)
	return teamProjects, query.Error
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

	// get the list of cohort_definition_ids that are allowed for the given teamProject:
	allowedCohortDefinitionIds, _ := h.GetCohortDefinitionIdsForTeamProject(teamProject)
	log.Printf("INFO: found %d cohorts for this team project", len(allowedCohortDefinitionIds))

	// Gather stats:
	atlasDb := db.GetAtlasDB().Db
	var cohortDefinitionStats []*CohortDefinitionStats
	query := atlasDb.Model(&CohortDefinition{}).
		Select("cohort_definition.id, cohort_definition.name, cohort_generation_info.person_count as cohort_size").
		Joins("INNER JOIN "+db.GetAtlasDB().Schema+".cohort_generation_info ON cohort_definition.id = cohort_generation_info.id").
		Where("cohort_generation_info.source_id = ?", sourceId).
		Where("cohort_generation_info.is_valid = true").
		Where("cohort_generation_info.is_canceled = false").
		Where("cohort_definition.id in (?)", allowedCohortDefinitionIds).
		Where("cohort_generation_info.person_count > 0").
		Order("cohort_generation_info.person_count desc")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortDefinitionStats)

	return cohortDefinitionStats, meta_result.Error
}

func (h CohortDefinition) GetCohortName(cohortId int) (string, error) {
	cohortDefinition, err := h.GetCohortDefinitionById(cohortId)
	if err != nil || cohortDefinition == nil {
		return "", fmt.Errorf("could not retrieve cohort name for cohortId=%d", cohortId)
	}

	return cohortDefinition.Name, nil
}

// Get the number of persons in a cohort that have an observation period equal or longer than
// the given observationWindow (aka "look back window").
func (h CohortDefinition) GetCohortDefinitionStatsByObservationWindow(sourceId int, cohortId int, observationWindow int) (*CohortDefinitionStats, error) {
	var cohortStats CohortDefinitionStats
	var dataSourceModel = new(Source)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)

	// Query to filter and count persons in cohort:
	query := QueryFilterByCohortIdAndObservationWindowHelper(resultsDataSource, omopDataSource, cohortId, observationWindow)

	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortStats)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}
	var err error
	cohortStats.Name, err = h.GetCohortName(cohortId)
	if err != nil {
		return nil, err
	}
	return &cohortStats, nil
}

// Get the number of persons in a cohort1 that have an observation period equal or longer than
// the given observationWindow (aka "look back window"), and are also present in cohort2.
func (h CohortDefinition) GetCohortDefinitionStatsByObservationWindow1stCohortAndOverlap2ndCohort(sourceId int, cohort1Id int, cohort2Id int, observationWindow1stCohort int) (*CohortDefinitionStats, error) {

	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	var cohortStats CohortDefinitionStats

	query := QueryFilterByCohortIdAndObservationWindowHelper(resultsDataSource, omopDataSource, cohort1Id, observationWindow1stCohort)
	query = query.Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort2 ON cohort2.subject_id = cohort.subject_id").
		Where("cohort2.cohort_definition_id = ?", cohort2Id)

	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortStats)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}
	var err error
	cohortStats.Name, err = h.GetCohortName(cohort1Id)
	if err != nil {
		return nil, err
	}
	return &cohortStats, nil
}

// Get the number of persons in a cohort1 that have an observation period equal or longer than
// the given observationWindow (aka "look back window"), and are also present in cohort2 with a cohort2 start date that is
// in the period between cohort1 start date and the given outcomeWindow2ndCohort.
func (h CohortDefinition) GetCohortDefinitionStatsByObservationWindow1stCohortAndOverlap2ndCohortAndOutcomeWindow2ndCohort(sourceId int, cohort1Id int, cohort2Id int, observationWindow1stCohort int, outcomeWindow2ndCohort int) (*CohortDefinitionStats, error) {

	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, Omop)
	resultsDataSource := dataSourceModel.GetDataSource(sourceId, Results)

	var cohortStats CohortDefinitionStats

	query := QueryFilterByCohortIdAndObservationWindowHelper(resultsDataSource, omopDataSource, cohort1Id, observationWindow1stCohort)
	query = query.Joins("INNER JOIN "+resultsDataSource.Schema+".cohort as cohort2 ON cohort2.subject_id = cohort.subject_id").
		Where("cohort2.cohort_definition_id = ?", cohort2Id)

	// Outcome must occur within "outcomeWindow2ndCohort days" of cohort1Id start:
	switch resultsDataSource.Db.Dialector.Name() {
	case "sqlserver":
		query = query.Where("cohort2.cohort_start_date < DATEADD(DAY, ?, cohort.cohort_start_date)", outcomeWindow2ndCohort).
			Where("cohort2.cohort_start_date > cohort.cohort_start_date") // TODO - plus 1 or more days when we later add "outcome observation window start, relative to cohort1 entry"

	case "postgres":
		query = query.Where("cohort2.cohort_start_date < ((INTERVAL '1 day' * ?) + cohort.cohort_start_date)", outcomeWindow2ndCohort).
			Where("cohort2.cohort_start_date > cohort.cohort_start_date") // TODO - plus 1 or more days when we later add "outcome observation window start, relative to cohort1 entry"
	default:
		log.Fatal("Unsupported dialect")
	}

	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&cohortStats)
	if meta_result.Error != nil {
		return nil, meta_result.Error
	}
	var err error
	cohortStats.Name, err = h.GetCohortName(cohort1Id)
	if err != nil {
		return nil, err
	}
	return &cohortStats, nil
}
