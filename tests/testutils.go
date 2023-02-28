package tests

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
)

func GetTestSourceId() int {
	return 1 // TODO - ideally this should also be used when populating "source" tables in test Atlas DB in the first place...
}

func GetTestDummyContinuousConceptId() int64 {
	return 2000000324
}

func GetTestHistogramConceptId() int64 {
	return 2000006885
}

func GetTestHareConceptId() int64 {
	return 2000007027
}

func GetTestAsnHareConceptId() int64 {
	// one of the HARE codes, this one being for "ASN-Asian":
	return 2000007029
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
	ExecSQLString(sqlString, -1)
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

func GetOmopDataSource() *utils.DbAndSchema {
	var dataSourceModel = new(models.Source)
	omopDataSource := dataSourceModel.GetDataSource(GetTestSourceId(), models.Omop)
	return omopDataSource
}

func GetResultsDataSource() *utils.DbAndSchema {
	var dataSourceModel = new(models.Source)
	dataSource := dataSourceModel.GetDataSource(GetTestSourceId(), models.Results)
	return dataSource
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

// utility function to fix things broken by BreakSomething()
func FixSomething(sourceType models.SourceType, tableName string, fieldName string) {
	ExecSQLString("ALTER TABLE IF EXISTS "+GetSchemaNameForType(sourceType)+"."+tableName+
		" RENAME "+fieldName+"_broken TO "+fieldName, GetTestSourceId())
}

// utility function that adds a new concept item with some invalid concept_class_id on the fly
func AddInvalidTypeConcept(sourceType models.SourceType) int64 {
	omopDataSource := GetOmopDataSource()
	var lastConcept models.Concept
	omopDataSource.Db.Last(&lastConcept)
	conceptId := lastConcept.ConceptId + 1
	ExecSQLString(fmt.Sprintf("INSERT into "+GetSchemaNameForType(sourceType)+".concept (concept_id,concept_name,concept_class_id,domain_id,concept_code) "+
		" values (%v, 'dummy concept', 'Invalid type class', 1, 'dummy')", conceptId), GetTestSourceId())
	return conceptId
}

func RemoveConcept(sourceType models.SourceType, conceptId int64) {
	ExecSQLString(fmt.Sprintf("DELETE FROM "+GetSchemaNameForType(sourceType)+".concept  "+
		" where concept_id =%v", conceptId), GetTestSourceId())
}

func GetInt64AttributeValue[T any](item T, attributeName string) int64 {
	r := reflect.ValueOf(item)
	f := reflect.Indirect(r).FieldByName(attributeName)
	return f.Int()
}

// returns an int array with the attribute values of the given attribute
// for each item in "items" array.
// TODO - can also simply be done wit something like: db.Model(&users).Pluck("age", &ages), where var ages []int64
func MapIntAttr[T any](items []T, attributeName string) []int64 {
	result := make([]int64, len(items))
	for i := range items {
		result[i] = GetInt64AttributeValue(items[i], attributeName)
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

// to use when mocking request context (gin.Context) in controller tests:
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
