package utils

import (
	"context"
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

type SourceConnection struct {
	SourceConnection string `json:",omitempty"`
	Username         string `json:",omitempty"`
	Password         string `json:",omitempty"`
}

var dataSourceDbMap = make(map[string]*DbAndSchema)

func GetDataSourceDB(source SourceConnection, dbSchema string) *DbAndSchema {
	sourceAndSchemaKey := "source:" + source.SourceConnection + ",schema:" + dbSchema
	if dataSourceDbMap[sourceAndSchemaKey] != nil {
		// return the already initialized object:
		return dataSourceDbMap[sourceAndSchemaKey]
	}
	// otherwise, open a new connection:
	dsn := GenerateDsn(source)
	dataSourceDb := new(DbAndSchema)
	if strings.Contains(source.SourceConnection, "postgresql") {
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
