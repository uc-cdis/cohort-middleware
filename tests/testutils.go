package tests

import (
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
)

func GetTestSourceId() int {
	return 1 // TODO - this should also be used when populating "source" tables in test Atlas DB in the first place...see also comment in setupSuite...
}

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
	ExecSQLString(string(fileContents), sourceId)
}

func ExecAtlasSQLString(sqlString string) {
	ExecSQLScript(sqlString, -1)
}

func ExecSQLString(sqlString string, sourceId int) {
	// just take a random model:
	var dataSourceModel = new(models.Source)
	if sourceId == -1 {
		// assume Atlas DB:
		var atlasDB = db.GetAtlasDB()
		atlasDB.Db.Model(models.Source{}).Exec(sqlString)

	} else {
		// look up the data source in source table:
		omopDataSource := dataSourceModel.GetDataSource(sourceId, models.Omop)
		omopDataSource.Db.Model(models.Source{}).Exec(sqlString)
	}
}

func GetSchemaNameForType(sourceType models.SourceType) string {
	sourceModel := new(models.Source)
	dbSchema, _ := sourceModel.GetSourceSchemaNameBySourceIdAndSourceType(GetTestSourceId(), sourceType)
	return dbSchema.SchemaName
}

// utility function to break something in the DB to be able to simulate DB issues
func BreakSomething(sourceType models.SourceType, tableName string, fieldName string) {
	ExecSQLString("ALTER TABLE IF EXISTS "+GetSchemaNameForType(sourceType)+"."+tableName+
		" RENAME "+fieldName+" TO "+fieldName+"_broken", GetTestSourceId())
}

// utility function to fix breaks made with BreakSomething()
func FixSomething(sourceType models.SourceType, tableName string, fieldName string) {
	ExecSQLString("ALTER TABLE IF EXISTS "+GetSchemaNameForType(sourceType)+"."+tableName+
		" RENAME "+fieldName+"_broken TO "+fieldName, GetTestSourceId())
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

type CustomResponseWriter struct {
	CustomResponseWriterOut string
}

func (w *CustomResponseWriter) Header() http.Header {
	result := make(http.Header)
	result.Add("Content-Type", "test")
	return result
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {

	w.CustomResponseWriterOut = string(b)
	return 0, nil
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	// do nothing
}

func (w *CustomResponseWriter) CloseNotify() <-chan bool {
	return nil
}

func (w *CustomResponseWriter) Flush() {
	//do nothing
}

func (w *CustomResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func (w *CustomResponseWriter) Pusher() (pusher http.Pusher) {
	return nil
}

func (w *CustomResponseWriter) Status() int {
	return 0
}

func (w *CustomResponseWriter) Size() int {
	return len(w.CustomResponseWriterOut)
}
func (w *CustomResponseWriter) WriteHeaderNow() {
	//do nothing
}
func (w *CustomResponseWriter) WriteString(s string) (n int, err error) {
	return 0, nil
}
func (w *CustomResponseWriter) Written() bool {
	return true
}
