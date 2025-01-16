package utils

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type DbAndSchema struct {
	Db     *gorm.DB
	Schema string
	Vendor string
}

var dataSourceDbMap = make(map[string]*DbAndSchema)

func GetDataSourceDB(sourceConnectionString string, dbSchema string) *DbAndSchema {
	sourceAndSchemaKey := "source:" + sourceConnectionString + ",schema:" + dbSchema
	if dataSourceDbMap[sourceAndSchemaKey] != nil {
		// return the already initialized object:
		return dataSourceDbMap[sourceAndSchemaKey]
	}
	// otherwise, open a new connection:
	dsn := GenerateDsn(sourceConnectionString)
	dataSourceDb := new(DbAndSchema)
	if strings.Contains(sourceConnectionString, "postgresql") {
		log.Printf("connecting to cohorts 'postgresql' db...")
		// workaround for schema names in postgres (can't be uppercase):
		dbSchema = strings.ToLower(dbSchema)
		dataSource, _ := gorm.Open(postgres.Open(dsn),
			&gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix:   dbSchema + ".",
					SingularTable: true,
				}})
		dataSourceDb.Db = dataSource
		dataSourceDb.Vendor = "postgresql"
	} else {
		log.Printf("connecting to cohorts 'sqlserver' db...")
		dataSource, _ := gorm.Open(sqlserver.Open(dsn),
			&gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix:   dbSchema + ".",
					SingularTable: true,
				}})
		// TODO - should throw error if db connection fails! Currently fails "silently" by printing error to log and then just returning ...
		dataSourceDb.Db = dataSource
		dataSourceDb.Vendor = "sqlserver"
	}
	dataSourceDb.Schema = dbSchema
	dataSourceDbMap[sourceAndSchemaKey] = dataSourceDb
	return dataSourceDb
}

// Adds a default timeout to a query
func AddTimeoutToQuery(query *gorm.DB) (*gorm.DB, context.CancelFunc) {
	// default timeout of 3 minutes:
	query, cancel := AddSpecificTimeoutToQuery(query, 180*time.Second)
	return query, cancel
}

// Adds a specific timeout to a query
func AddSpecificTimeoutToQuery(query *gorm.DB, timeout time.Duration) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	query = query.WithContext(ctx)
	return query, cancel
}

// Returns extra DB dialect specific directives to optimize performance when using views:
func (h DbAndSchema) GetViewDirective() string {
	if h.Vendor == "sqlserver" {
		return " WITH (NOEXPAND) "
	} else {
		return ""
	}
}
func ToSQL2(query *gorm.DB) (string, error) {
	// Use db.ToSQL to generate the SQL string for the existing query
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Session(&gorm.Session{DryRun: true}).Find([]interface{}{})
	})

	return sql, nil
}

func ToSQL(query *gorm.DB) (string, error) {
	//var dummy []interface{}
	//log.Printf("Statement.SQL: %s", query.Statement.SQL.String())
	log.Printf("TableExpr.SQL: %s", query.Statement.TableExpr.SQL)
	log.Printf("Clauses: %+v", query.Statement.Clauses) // TODO - REMOVE?
	//sqlString := query.Session(&gorm.Session{DryRun: true}).Find(&dummy).Statement.SQL.String()
	// if sqlString == "" {
	// 	sqlString, _ = ToSQL2(query)
	// }
	sqlString2, _ := ToSQL2(query)
	log.Printf("Statement.SQL: %s", sqlString2)

	//interpolatedSQL, err := InterpolateSQL(query, sqlString2)

	return sqlString2, nil
}

// Replaces placeholders in SQL with actual values,
// exploring all potential variable sources in the GORM query object.
func InterpolateSQL(query *gorm.DB, sqlString string) (string, error) {

	// Collect variables from query.Statement.TableExpr.Vars
	allVars := query.Statement.TableExpr.Vars

	// Collect variables from query.Statement.Clauses
	for _, c := range query.Statement.Clauses {
		if len(c.Expression.(clause.Where).Exprs) > 0 {
			for _, expr := range c.Expression.(clause.Where).Exprs {
				switch v := expr.(type) {
				case clause.Expr:
					allVars = append(allVars, v.Vars...)
				}
			}
		}
	}

	// Use regular expression to find all occurrences of placeholders like $1, $2, etc.
	placeholderRegex := regexp.MustCompile(`\$(\d+)`)
	placeholders := placeholderRegex.FindAllString(sqlString, -1)

	// Check if the number of placeholders matches the length of allVars
	if len(placeholders) != len(allVars) {
		return "", fmt.Errorf("mismatch between number of placeholders (%d) and number of variables (%d)", len(placeholders), len(allVars))
	}

	// Replace placeholders ($1, $2, etc.) with actual variable values
	resultSQL := sqlString
	for i, v := range allVars {
		placeholder := fmt.Sprintf("$%d", i+1)
		var value string

		// Handle different variable types
		switch val := v.(type) {
		case string:
			value = fmt.Sprintf("'%s'", val) // Wrap strings in quotes
		case int, int64, float64:
			value = fmt.Sprintf("%v", val) // Numeric values
		case nil:
			value = "NULL" // NULL values
		default:
			value = fmt.Sprintf("'%v'", val) // Fallback for other types
		}
		// Replace the placeholder with the actual value
		resultSQL = strings.Replace(resultSQL, placeholder, value, 1)
	}
	return resultSQL, nil
}
