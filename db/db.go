package db

import (
	"fmt"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var atlasDB *utils.DbAndSchema

func Init() {
	c := config.GetConfig()

	host := c.GetString("atlas_db.host")
	user := c.GetString("atlas_db.username")
	password := c.GetString("atlas_db.password")
	dbname := c.GetString("atlas_db.db")
	port := c.GetString("atlas_db.port")
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host,
		user,
		password,
		dbname,
		port)

	dbSchema := c.GetString("atlas_db.schema")
	db, _ := gorm.Open(postgres.New(
		postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true,
		}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   fmt.Sprintf("%s.", dbSchema),
			SingularTable: true,
		}})
	atlasDB = new(utils.DbAndSchema)
	atlasDB.Db = db
	atlasDB.Schema = dbSchema
}

func GetAtlasDB() *utils.DbAndSchema {
	return atlasDB
}
