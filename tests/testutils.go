package tests

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/uc-cdis/cohort-middleware/models"
)

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
	var dataSourceModel = new(models.Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, models.Omop)

	omopDataSource.Db.Model(models.Source{}).Exec(string(fileContents))
}
