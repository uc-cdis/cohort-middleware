package tests

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"

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

func GetIntAttributeValue[T any](item T, attributeName string) int {
	r := reflect.ValueOf(item)
	f := reflect.Indirect(r).FieldByName(attributeName)
	return int(f.Int())
}

// returns an int array with the attribute values of the given attribute
// for each item in "items" array.
func MapIntAttr[T any](items []T, attributeName string) []int {
	result := make([]int, len(items))
	for i := range items {
		result[i] = GetIntAttributeValue(items[i], attributeName)
	}
	return result
}

// returns an array with the values returned by applying
// given function "f" for each item in "items" array.
func Map[T, U any](items []T, f func(T) U) []U {
	result := make([]U, len(items))
	for i := range items {
		result[i] = f(items[i])
	}
	return result
}
