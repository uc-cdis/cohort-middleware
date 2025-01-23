package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
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
func ToSQL2(query *gorm.DB) string {
	// Use db.ToSQL to generate the SQL string for the existing query
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Session(&gorm.Session{DryRun: true}).Find([]interface{}{})
	})

	return sql
}

func ToSQL(query *gorm.DB) string {
	// Clone the query object to avoid altering the original
	//tempQuery := query.Session(&gorm.Session{DryRun: true})
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find([]interface{}{})
	})
	return sql
}

func TableExists(tx *gorm.DB, tableName string) bool {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE 1 = 2", tableName)
	err := tx.Exec(query).Error
	if err != nil {
		log.Printf("TableExists check failed for: %s, error: %v", tableName, err)
		return false
	}
	// If the query succeeds, the table exists:
	return true
}
