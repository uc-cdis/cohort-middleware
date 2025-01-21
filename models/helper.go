package models

import (
	"fmt"
	"log"
	"strings"

	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/gorm"
)

func QueryFilterByConceptDefsPlusCohortPairsHelper(sourceId int, mainCohortDefinitionId int, filterConceptDefsAndCohortPairs []interface{},
	omopDataSource *utils.DbAndSchema, resultsDataSource *utils.DbAndSchema, finalSetAlias string) (*gorm.DB, string) {
	// filterConceptDefsAndCohortPairs is a list of utils.CustomConceptVariableDef (concept definitions type of filter)
	// and utils.CustomDichotomousVariableDef (cohort pair type of filter) items.
	//
	// This method builds up the SQL for a list of filters. It iterates over the filters, building up the SQL according to
	// whether the filter is a concept definition or a cohort pair type of filter. For concept definition filters, in QueryFilterByConceptDefHelper,
	// if a transformation is present, it will make sure the respective data (filters are executed in order) is transformed and stored in a
	// temporary table before that can be used further as part of the SQL query.
	// Caching of temporary tables: for optimal performance, a temporary table dictionary / cache is updated, keeping a mapping
	// of existing temporary table names vs underlying subsets of items in filterConceptDefsAndCohortPairs that gave rise to these
	// tables.
	query := resultsDataSource.Db.Table(resultsDataSource.Schema+".cohort as "+finalSetAlias).
		Select("*").
		Where("cohort_definition_id=?", mainCohortDefinitionId)

	tmpTableName := ""
	for i, item := range filterConceptDefsAndCohortPairs {
		tableAlias := fmt.Sprintf("filter_%d", i)
		switch convertedItem := item.(type) {
		case utils.CustomConceptVariableDef:
			query, tmpTableName, _ = QueryFilterByConceptDefHelper(query, sourceId, convertedItem, omopDataSource, finalSetAlias+".subject_id", tableAlias) // TODO - improve error handling.
		case utils.CustomDichotomousVariableDef:
			query = QueryFilterByCohortPairHelper(query, convertedItem, resultsDataSource, mainCohortDefinitionId, finalSetAlias+".subject_id", tableAlias)
			tmpTableName = ""
		}
	}
	return query, tmpTableName
}

// DEPRECATED - USE QueryFilterByConceptDefsHelper
// Helper function that adds extra filter clauses to the query, joining on the right set of tables.
//   - It was added here to make it reusable, given these filters need to be added to many of the queries that take in
//     a list of filters in the form of concept ids.
func QueryFilterByConceptIdsHelper(query *gorm.DB, sourceId int, filterConceptIds []int64,
	omopDataSource *utils.DbAndSchema, resultSchemaName string, personIdFieldForObservationJoin string) *gorm.DB {
	// iterate over the filterConceptIds, adding a new INNER JOIN and filters for each, so that the resulting set is the
	// set of persons that have a non-null value for each and every one of the concepts:
	for i, filterConceptId := range filterConceptIds {
		observationTableAlias := fmt.Sprintf("observation_filter_%d", i)
		log.Printf("Adding extra INNER JOIN with alias %s", observationTableAlias)
		query = query.Joins("INNER JOIN "+omopDataSource.Schema+".observation_continuous as "+observationTableAlias+omopDataSource.GetViewDirective()+" ON "+observationTableAlias+".person_id = "+personIdFieldForObservationJoin).
			Where(observationTableAlias+".observation_concept_id = ?", filterConceptId).
			Where(GetConceptValueNotNullCheckBasedOnConceptType(observationTableAlias, sourceId, filterConceptId))
	}
	return query
}

func QueryFilterByConceptDefHelper(query *gorm.DB, sourceId int, filterConceptDef utils.CustomConceptVariableDef,
	omopDataSource *utils.DbAndSchema, personIdFieldForObservationJoin string, observationTableAlias string) (*gorm.DB, string, error) {
	// 1  - check if filterConceptDef has a transformation
	// 2a - if it does, transform the data into a new tmp table.
	//      Cache this by using the query definition so far as the key,
	//      and the temp table name as the value.
	//      Returns error if something fails.
	// 2b - if not, just use "observation_continuous"
	if filterConceptDef.Transformation != "" {
		// simple filterConceptDef with just the concept id
		simpleFilterConceptDef := utils.CustomConceptVariableDef{ConceptId: filterConceptDef.ConceptId}
		resultingQuery := QueryFilterByConceptDefHelper2(query, sourceId, simpleFilterConceptDef,
			omopDataSource, "", personIdFieldForObservationJoin, "observation_continuous", observationTableAlias)
		tmpTransformedTable, err := TransformDataIntoTempTable(omopDataSource, resultingQuery, filterConceptDef)
		// TODO - the resulting query should actually be Select * from temptable.... as this collapses all underlying queries. TODO2 - ensure the transform method also filters....
		resultingQuery = QueryFilterByConceptDefHelper2(query, sourceId, filterConceptDef, //TODO - turn around
			omopDataSource, "", personIdFieldForObservationJoin, tmpTransformedTable, observationTableAlias)
		return resultingQuery, tmpTransformedTable, err

	} else {
		// simple filterConceptDef with no transformation
		resultingQuery := QueryFilterByConceptDefHelper2(query, sourceId, filterConceptDef,
			omopDataSource, "", personIdFieldForObservationJoin, "observation_continuous", observationTableAlias)
		return resultingQuery, "", nil
	}
}

func QueryFilterByConceptDefHelper2(query *gorm.DB, sourceId int, filterConceptDef utils.CustomConceptVariableDef,
	omopDataSource *utils.DbAndSchema, resultSchemaName string, personIdFieldForObservationJoin string, observationDataSource string, observationTableAlias string) *gorm.DB {
	log.Printf("Adding extra INNER JOIN with alias %s", observationTableAlias)
	aliasedObservationDataSource := omopDataSource.Schema + "." + observationDataSource + " as " + observationTableAlias + omopDataSource.GetViewDirective()
	if strings.HasPrefix(observationDataSource, "tmp_") {
		aliasedObservationDataSource = observationDataSource
		observationTableAlias = aliasedObservationDataSource
	}
	query = query.Joins("INNER JOIN "+aliasedObservationDataSource+" ON "+observationTableAlias+".person_id = "+personIdFieldForObservationJoin).
		Where(observationTableAlias+".observation_concept_id = ?", filterConceptDef.ConceptId)

	valueExpression := fmt.Sprintf("%s.value_as_number", observationTableAlias)
	//If filters, add the value filtering clauses to the query
	if len(filterConceptDef.Filters) > 0 {
		for _, filter := range filterConceptDef.Filters {
			switch filter.Type {
			case ">", "<", ">=", "<=", "=", "!=":
				if filter.Value != nil {
					query = query.Where(fmt.Sprintf("%s %s ?", valueExpression, filter.Type), *filter.Value)
				}
			case "in":
				if len(filter.Values) > 0 {
					query = query.Where(fmt.Sprintf("%s.value_as_number IN (?)", observationTableAlias), filter.Values)
				} else if len(filter.ValuesAsConceptIds) > 0 {
					query = query.Where(fmt.Sprintf("%s.value_as_concept_id IN (?)", observationTableAlias), filter.ValuesAsConceptIds)
				}
			default:
				log.Printf("Unsupported filter type: %s", filter.Type)
			}
		}
	} else {
		query = query.Where(GetConceptValueNotNullCheckBasedOnConceptType(observationTableAlias, sourceId, filterConceptDef.ConceptId))
	}
	return query
}

// Transforms the data returned by query into a new temp table.
// Caches the temp table name by using the query definition + transformation method as the key,
// and the temp table name as the value. This allows the method to reuse a temp table if
// one has already been made for this combination.
// Returns the temp table name.
func TransformDataIntoTempTable(omopDataSource *utils.DbAndSchema, query *gorm.DB, filterConceptDef utils.CustomConceptVariableDef) (string, error) {
	// Generate a unique hash key based on the query and transformation
	querySQL := utils.ToSQL(query)
	queryKey := fmt.Sprintf("%s|%s", querySQL, filterConceptDef.Transformation)
	cacheKey := utils.GenerateHash(queryKey) // TODO - review

	// Check if the temporary table already exists in the cache
	if cachedTableName, exists := utils.TempTableCache.Get(cacheKey); exists {
		log.Printf("Reusing cached temp table: %s", cachedTableName)
		return cachedTableName.(string), nil
	}
	// Create a unique temporary table name
	tempTableName := fmt.Sprintf("tmp_transformed_%s", utils.GenerateSynchronizedTimestampID())

	finalTempTableName := CreateAndFillTempTable(omopDataSource, query, tempTableName, querySQL, filterConceptDef)

	// Cache the temp table name
	utils.TempTableCache.Set(cacheKey, finalTempTableName)
	return finalTempTableName, nil
}

func CreateAndFillTempTable(omopDataSource *utils.DbAndSchema, query *gorm.DB, tempTableName string, querySQL string, filterConceptDef utils.CustomConceptVariableDef) string {
	if filterConceptDef.Transformation != "" {
		switch filterConceptDef.Transformation {
		case "log":
			tempTableSQL, finalTempTableName := TempTableSQLAndFinalName(omopDataSource, tempTableName,
				"person_id, observation_concept_id, CASE WHEN value_as_number > 0 THEN LOG(value_as_number) ELSE NULL END AS value_as_number",
				querySQL, "value_as_number > 0")
			log.Printf("Creating new temp table: %s", tempTableName)

			// Execute the SQL to create and fill the temp table
			if err := query.Exec(tempTableSQL).Error; err != nil {
				log.Fatalf("Failed to create temp table: %v", err)
				panic("error")
			}
			return finalTempTableName
		case "inverse_normal_rank":
			// Implement a custom SQL function or transformation if needed
			log.Printf("inverse_normal_rank transformation logic needs to be implemented")
			panic("TODO")
		case "z_score":
			stdDevFunc := "STDDEV"
			if omopDataSource.Vendor == "sqlserver" {
				stdDevFunc = "STDEV"
			}
			tempTableSQL, finalTempTableName := TempTableSQLAndFinalName(omopDataSource, tempTableName,
				"person_id, observation_concept_id, (value_as_number-AVG(value_as_number) OVER ()) / "+stdDevFunc+"(value_as_number) OVER () as value_as_number",
				querySQL, "")
			log.Printf("Creating new temp table: %s", tempTableName)

			// Execute the SQL to create and fill the temp table
			if err := query.Exec(tempTableSQL).Error; err != nil {
				log.Fatalf("Failed to create temp table: %v", err)
				panic("error")
			}
			return finalTempTableName
		case "box_cox":
			// Placeholder: implement Box-Cox transformation logic as per requirements
			log.Printf("box_cox transformation logic needs to be implemented")
			return ""
		default:
			log.Fatalf("Unsupported transformation type: %s", filterConceptDef.Transformation)
			panic("error")
		}
	}
	panic("error")
}

func TempTableSQLAndFinalName(omopDataSource *utils.DbAndSchema, tempTableName string, selectStatement string, fromSQL string, extraWhereSQL string) (string, string) {
	var tempTableSQL string
	if extraWhereSQL == "" {
		extraWhereSQL = "1=1"
	}
	finalTempTableName := tempTableName
	tempTableSQL = fmt.Sprintf(
		"CREATE TEMPORARY TABLE %s AS (SELECT %s FROM (%s) AS T WHERE %s)",
		tempTableName, selectStatement, fromSQL, extraWhereSQL,
	)
	if omopDataSource.Vendor == "sqlserver" {
		finalTempTableName = "#" + tempTableName // Local temp table for MSSQL
		tempTableSQL = fmt.Sprintf(
			"SELECT %s INTO %s FROM (%s) T WHERE %s",
			selectStatement, finalTempTableName, fromSQL, extraWhereSQL,
		)
	}
	return tempTableSQL, finalTempTableName
}

func QueryFilterByConceptDefsHelper(query *gorm.DB, sourceId int, filterConceptDefs []utils.CustomConceptVariableDef,
	omopDataSource *utils.DbAndSchema, resultSchemaName string, personIdFieldForObservationJoin string) *gorm.DB {
	// iterate over the filterConceptDefs, adding a new filters for each:
	for i, filterConceptDef := range filterConceptDefs {
		observationTableAlias := fmt.Sprintf("observation_filter_%d", i)
		query, _, _ = QueryFilterByConceptDefHelper(query, sourceId, filterConceptDef, omopDataSource, personIdFieldForObservationJoin, observationTableAlias)
	}
	return query
}

// Helper function that adds extra filter clauses to the query, for the given filterCohortPairs, intersecting on the
// right set of tables, excluding data where necessary, etc.
// It basically iterates over the list of filterCohortPairs, adding relevant INTERSECT and EXCEPT clauses, so that the resulting set is the
// set of persons that are part of the intersections of cohortDefinitionId and of one of the cohorts in the filterCohortPairs. The EXCEPT
// clauses exclude the persons that are found in both cohorts of a filterCohortPair.
func QueryFilterByCohortPairsHelper(filterCohortPairs []utils.CustomDichotomousVariableDef, resultsDataSource *utils.DbAndSchema, cohortDefinitionId int, unionAndIntersectSQLAlias string) *gorm.DB {
	finalSetAlias := unionAndIntersectSQLAlias
	finalSQL := "(SELECT subject_id FROM " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? )" + " as " + finalSetAlias + " "
	query := resultsDataSource.Db.Table(finalSQL, cohortDefinitionId)
	for i, item := range filterCohortPairs {
		tableAlias := fmt.Sprintf("filter_%d", i)
		query = QueryFilterByCohortPairHelper(query, item, resultsDataSource, cohortDefinitionId, finalSetAlias+".subject_id", tableAlias)
	}
	return query
}

// TODO - remove code duplication above
func QueryFilterByCohortPairHelper(query *gorm.DB, filterCohortPair utils.CustomDichotomousVariableDef, resultsDataSource *utils.DbAndSchema, cohortDefinitionId int, personIdFieldForObservationJoin string, unionAndIntersectSQLAlias string) *gorm.DB {
	unionAndIntersectSQL := "(" +
		"SELECT subject_id FROM " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? "
	var idsList []interface{}
	idsList = append(idsList, cohortDefinitionId)
	unionAndIntersectSQL = unionAndIntersectSQL +
		"INTERSECT ( " +
		"SELECT subject_id FROM " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? " +
		"UNION " +
		"SELECT subject_id FROM " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? " +
		")"
	idsList = append(idsList, filterCohortPair.CohortDefinitionId1, filterCohortPair.CohortDefinitionId2)
	unionAndIntersectSQL = unionAndIntersectSQL +
		"EXCEPT ( " +
		"SELECT subject_id FROM  " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? " +
		"INTERSECT " +
		"SELECT subject_id FROM  " + resultsDataSource.Schema + ".cohort WHERE cohort_definition_id=? " +
		")"
	idsList = append(idsList, filterCohortPair.CohortDefinitionId1, filterCohortPair.CohortDefinitionId2)

	unionAndIntersectSQL = unionAndIntersectSQL + ") "
	query = query.Joins("INNER JOIN ("+unionAndIntersectSQL+") AS "+unionAndIntersectSQLAlias+" ON "+unionAndIntersectSQLAlias+".subject_id = "+personIdFieldForObservationJoin, idsList...)

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
	} else if conceptInfo.ConceptType == "MVP Nominal" {
		return observationTableAlias + ".value_as_concept_id is not null and " + observationTableAlias + ".value_as_concept_id != 0"
	} else {
		panic(fmt.Sprintf("error: concept type not supported [%s]", conceptInfo.ConceptType))
	}
}
