package utils

import (
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DbAndSchema struct {
	Db     *gorm.DB
	Schema string
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
		log.Printf("trying to connect to 'postgresql' db...")
		// workaround for schema names in postgres (can't be uppercase):
		dbSchema = strings.ToLower(dbSchema)
		omopDataSource, _ := gorm.Open(postgres.Open(dsn),
			&gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix:   dbSchema + ".",
					SingularTable: true,
				}})
		dataSourceDb.Db = omopDataSource
	} else {
		omopDataSource, _ := gorm.Open(sqlserver.Open(dsn),
			&gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix:   dbSchema + ".",
					SingularTable: true,
				}})
		// TODO - should throw error if db connection fails! Currently fails "silently" by printing error to log and then just returning ...
		dataSourceDb.Db = omopDataSource
	}
	dataSourceDb.Schema = dbSchema
	dataSourceDbMap[sourceAndSchemaKey] = dataSourceDb
	return dataSourceDb
}
