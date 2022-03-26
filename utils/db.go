package utils

import (
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func GetDataSourceDB(sourceConnectionString string, dbSchema string) *gorm.DB {
	dsn := GenerateDsn(sourceConnectionString)

	if strings.Contains(sourceConnectionString, "postgresql") {
		log.Printf("trying to connect to 'postgresql' db...")
		// workaround for schema names in postgres (can't be uppercase):
		dbSchema = strings.ToLower(dbSchema)
		omopDataSource, _ := gorm.Open(postgres.Open(dsn),
			&gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix:   dbSchema,
					SingularTable: true,
				}})
		return omopDataSource
	} else {
		omopDataSource, _ := gorm.Open(sqlserver.Open(dsn),
			&gorm.Config{
				NamingStrategy: schema.NamingStrategy{
					TablePrefix:   dbSchema,
					SingularTable: true,
				}})
		// TODO - should throw error if db connection fails! Currently fails "silently" by printing error to log and then just returning ...
		return omopDataSource
	}
}
