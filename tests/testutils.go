package tests

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
)

func ExecAtlasSQLScript(sqlFilePath string) {
	ExecSQLScript(sqlFilePath, -1)
}

func ExecSQLScript(sqlFilePath string, sourceId int) {
	log.Printf("Loading %s...", sqlFilePath)

	path, err := filepath.Abs(sqlFilePath)
	if err != nil {
		panic(err)
	}
	fileContents, err2 := ioutil.ReadFile(path)
	if err2 != nil {
		panic(err)
	}
	// just take a random model:
	var dataSourceModel = new(models.Source)
	if sourceId == -1 {
		// assume Atlas DB:
		var atlasDB = db.GetAtlasDB()
		atlasDB.Db.Model(models.Source{}).Exec(string(fileContents))

	} else {
		// look up the data source in source table:
		omopDataSource := dataSourceModel.GetDataSource(sourceId, models.Omop)
		omopDataSource.Db.Model(models.Source{}).Exec(string(fileContents))
	}
}
