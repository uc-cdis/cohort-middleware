package models

import (
	"fmt"
	"log"

	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/gorm"
)

// Helper function that adds extra filter clauses to the query, joining on the right set of tables, excluding data where necessary, etc.
// * It was added here to make it reusable, given these filters need to be added to many of the queries that take in
//   a list of filters in the form of concept ids and cohort pairs. The one assumption it makes is that the given `query` object already contains
//   a basic query on a table or view that have been named or aliased as "observation" (see comments in code).
//   TODO - find a way to check this assumption here. If we have a good way to get the raw SQL from given `query`, we could check with `\bas observation\b`...
func QueryFilterByConceptIdsAndCohortPairsHelper(query *gorm.DB, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef, omopSchemaName string, resultSchemaName string) *gorm.DB {
	// iterate over the filterConceptIds, adding a new INNER JOIN and filters for each, so that the resulting set is the
	// set of persons that have a non-null value for each and every one of the concepts:
	for i, filterConceptId := range filterConceptIds {
		observationTableAlias := fmt.Sprintf("observation_filter_%d", i)
		log.Printf("Adding extra INNER JOIN with alias %s", observationTableAlias)
		query = query.Joins("INNER JOIN "+omopSchemaName+".observation_continuous as "+observationTableAlias+" ON "+observationTableAlias+".person_id = observation.person_id"). // assumption: there is a table or view named or aliased as "observation"
			Where(observationTableAlias+".observation_concept_id = ?", filterConceptId).
			Where("(" + observationTableAlias + ".value_as_string is not null or " + observationTableAlias + ".value_as_number is not null)") // TODO - improve performance by only filtering on type according to getConceptValueType()
	}
	// iterate over the list of filterCohortPairs, adding a new INNER JOIN to the UNION of each pair, so that the resulting set is the
	// set of persons that are part of the intersections above and of one of the cohorts in the filterCohortPairs:
	for i, filterCohortPair := range filterCohortPairs {
		cohortTableAlias1 := fmt.Sprintf("cohort_filter_1_%d", i)
		cohortTableAlias2 := fmt.Sprintf("cohort_filter_2_%d", i)
		cohortTableAlias3 := fmt.Sprintf("cohort_filter_3_%d", i)
		unionExceptAlias := fmt.Sprintf("union_%d", i)
		log.Printf("Adding extra INNER JOIN on UNION and EXCEPT with alias %s", unionExceptAlias)
		query = query.Joins(
			"INNER JOIN "+
				" (Select "+cohortTableAlias1+".subject_id FROM "+resultSchemaName+".cohort as "+cohortTableAlias1+
				"  where "+cohortTableAlias1+".cohort_definition_id in (?,?) "+ //the UNION of both cohorts
				"  EXCEPT "+ //now use EXCEPT to exclude the part where both cohorts INTERSECT
				"  Select "+cohortTableAlias2+".subject_id FROM "+resultSchemaName+".cohort as "+cohortTableAlias2+
				"  INNER JOIN "+resultSchemaName+".cohort as "+cohortTableAlias3+" ON "+cohortTableAlias3+".subject_id = "+cohortTableAlias2+".subject_id "+
				"  where "+cohortTableAlias2+".cohort_definition_id = ? AND "+cohortTableAlias3+".cohort_definition_id =? ) AS "+unionExceptAlias+" ON "+unionExceptAlias+".subject_id = observation.person_id", // assumption: there is a table or view named or aliased as "observation"
			filterCohortPair.CohortId1, filterCohortPair.CohortId2, filterCohortPair.CohortId1, filterCohortPair.CohortId2)
	}

	return query
}
