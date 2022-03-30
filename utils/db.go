package utils

import (
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func GetDataSourceDB(sourceConnectionString string, dbSchema string) *gorm.DB {
	dsn := GenerateDsn(sourceConnectionString)

	omopDataSource, _ := gorm.Open(sqlserver.Open(dsn),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   dbSchema,
				SingularTable: true,
			}})
	return omopDataSource
}
