package models

import (
	"fmt"
	"log"

	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/gorm"
)

func QueryFilterByConceptIdsAndCohortPairsHelper(query *gorm.DB, filterConceptIds []int64, filterCohortPairs []utils.CustomDichotomousVariableDef, omopSchemaName string, resultSchemaName string) *gorm.DB {
	// iterate over the filterConceptIds, adding a new INNER JOIN and filters for each, so that the resulting set is the
	// set of persons that have a non-null value for each and every one of the concepts:
	for i, filterConceptId := range filterConceptIds {
		observationTableAlias := fmt.Sprintf("observation_filter_%d", i)
		log.Printf("Adding extra INNER JOIN with alias %s", observationTableAlias)
		query = query.Joins("INNER JOIN "+omopSchemaName+".observation as "+observationTableAlias+" ON "+observationTableAlias+".person_id = observation.person_id").
			Where(observationTableAlias+".observation_concept_id = ?", filterConceptId).
			Where("(" + observationTableAlias + ".value_as_string is not null or " + observationTableAlias + ".value_as_number is not null)") // TODO - improve performance by only filtering on type according to getConceptValueType()
	}
	// iterate over the list of filterCohortPairs, adding a new INNER JOIN to the UNION of each pair, so that the resulting set is the
	// set of persons that are part of the intersections above and of one of the cohorts in the filterCohortPairs:
	for i, filterCohortPair := range filterCohortPairs {
		cohortTableAlias1 := fmt.Sprintf("cohort_filter_1_%d", i)
		cohortTableAlias2 := fmt.Sprintf("cohort_filter_2_%d", i)
		unionAlias := "union_" + cohortTableAlias1 + "_" + cohortTableAlias2
		log.Printf("Adding extra INNER JOIN on UNION with alias %s", unionAlias)
		query = query.Joins("INNER JOIN (Select "+cohortTableAlias1+".subject_id,"+cohortTableAlias1+".cohort_definition_id FROM "+resultSchemaName+".cohort as "+cohortTableAlias1+
			" UNION ALL Select "+cohortTableAlias2+".subject_id,"+cohortTableAlias2+".cohort_definition_id FROM "+resultSchemaName+".cohort as "+cohortTableAlias2+
			") AS "+unionAlias+" ON "+unionAlias+".subject_id = observation.person_id").
			Where(unionAlias+".cohort_definition_id in (?,?)", filterCohortPair.CohortId1, filterCohortPair.CohortId2)
	}

	return query
}
