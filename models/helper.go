package models

import (
	"fmt"
	"log"

	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/gorm"
)

// Helper function that adds extra filter clauses to the query, joining on the right set of tables.
// * It was added here to make it reusable, given these filters need to be added to many of the queries that take in
//   a list of filters in the form of concept ids.
func QueryFilterByConceptIdsHelper(query *gorm.DB, sourceId int, filterConceptIds []int64,
	omopDataSource *utils.DbAndSchema, resultSchemaName string, mainObservationTableAlias string) *gorm.DB {
	// iterate over the filterConceptIds, adding a new INNER JOIN and filters for each, so that the resulting set is the
	// set of persons that have a non-null value for each and every one of the concepts:
	for i, filterConceptId := range filterConceptIds {
		observationTableAlias := fmt.Sprintf("observation_filter_%d", i)
		log.Printf("Adding extra INNER JOIN with alias %s", observationTableAlias)
		query = query.Joins("INNER JOIN "+omopDataSource.Schema+".observation_continuous as "+observationTableAlias+omopDataSource.GetViewDirective()+" ON "+observationTableAlias+".person_id = "+mainObservationTableAlias+".person_id").
			Where(observationTableAlias+".observation_concept_id = ?", filterConceptId).
			Where(GetConceptValueNotNullCheckBasedOnConceptType(observationTableAlias, sourceId, filterConceptId))
	}
	return query
}

// Helper function that adds extra filter clauses to the query, for the given filterCohortPairs, intersecting on the
// right set of tables, excluding data where necessary, etc.
// It basically iterates over the list of filterCohortPairs, adding relevant INTERSECT and EXCEPT clauses, so that the resulting set is the
// set of persons that are part of the intersections of cohortDefinitionId and of one of the cohorts in the filterCohortPairs. The EXCEPT
// clauses exclude the persons that are found in both cohorts of a filterCohortPair.
func QueryFilterByCohortPairsHelper(filterCohortPairs []utils.CustomDichotomousVariableDef, resultsDataSource *utils.DbAndSchema, cohortDefinitionId int, unionAndIntersectSQLAlias string) *gorm.DB {
	unionAndIntersectSQL := "(" +
		"SELECT subject_id FROM " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? "
	var idsList []interface{}
	idsList = append(idsList, cohortDefinitionId)
	if len(filterCohortPairs) > 0 {
		// INTERSECT UNIONs section:
		for _, filterCohortPair := range filterCohortPairs {
			unionAndIntersectSQL = unionAndIntersectSQL +
				"INTERSECT ( " +
				"SELECT subject_id FROM " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? " +
				"UNION " +
				"SELECT subject_id FROM " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? " +
				")"
			idsList = append(idsList, filterCohortPair.CohortId1, filterCohortPair.CohortId2)
		}
		// EXCEPTs section:
		for _, filterCohortPair := range filterCohortPairs {
			unionAndIntersectSQL = unionAndIntersectSQL +
				"EXCEPT ( " +
				"SELECT subject_id FROM  " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? " +
				"INTERSECT " +
				"SELECT subject_id FROM  " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? " +
				")"
			idsList = append(idsList, filterCohortPair.CohortId1, filterCohortPair.CohortId2)
		}
	}
	unionAndIntersectSQL = unionAndIntersectSQL +
		") "
	query := resultsDataSource.Db.Table(unionAndIntersectSQL+" as "+unionAndIntersectSQLAlias+" ", idsList...)
	return query
}

// This function will get the concept information for given conceptId, and
// return the best SQL to use for doing a "not null" check on its value in the
// observation table.
func GetConceptValueNotNullCheckBasedOnConceptType(observationTableAlias string, sourceId int, conceptId int64) string {
	conceptModel := *new(Concept)
	conceptInfo, error := conceptModel.RetrieveInfoBySourceIdAndConceptId(sourceId, conceptId)
	if error != nil {
		panic("error while trying to get information for conceptId, or conceptId not found")
	} else if conceptInfo.ConceptType == "MVP Continuous" {
		return observationTableAlias + ".value_as_number is not null"
	} else if conceptInfo.ConceptType == "MVP Ordinal" {
		return observationTableAlias + ".value_as_concept_id is not null and " + observationTableAlias + ".value_as_concept_id != 0"
	} else {
		panic(fmt.Sprintf("error: concept type not supported [%s]", conceptInfo.ConceptType))
	}
}
