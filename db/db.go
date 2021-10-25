package db

import (
	"fmt"
	"github.com/uc-cdis/cohort-middleware/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var atlasDB *gorm.DB

func Init() {
	c := config.GetConfig()
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.GetString("atlas_db.host"),
		c.GetString("atlas_db.username"),
		c.GetString("atlas_db.password"),
		c.GetString("atlas_db.db"),
		c.GetString("atlas_db.port"))

	atlasDB, _ = gorm.Open(postgres.New(
		postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true,
		}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   c.GetString("atlas_db.schema") + ".",
			SingularTable: true,
		}})
}

func GetAtlasDB() *gorm.DB {
	return atlasDB
}
