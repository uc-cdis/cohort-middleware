package tests

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"path/filepath"
	"reflect"
	"time"

	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/gorm"
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

func ExecSQLString(sqlString string, sourceId int) (tx *gorm.DB) {
	// just take a random model:
	var dataSourceModel = new(models.Source)
	if sourceId == -1 {
		// assume Atlas DB:
		var atlasDB = db.GetAtlasDB()
		return atlasDB.Db.Model(models.Source{}).Exec(sqlString)

	} else {
		// look up the data source in source table:
		fmt.Println("here before")
		omopDataSource := dataSourceModel.GetDataSource(sourceId, models.Omop)
		fmt.Println("here after", omopDataSource)
		fmt.Println("sql here after", sqlString)
		return omopDataSource.Db.Model(models.Source{}).Exec(sqlString)
	}
}

// Same as ExecSQLString above, but panics if the SQL statement fails
func ExecSQLStringOrFail(sqlString string, sourceId int) {
	result := ExecSQLString(sqlString, sourceId)
	if result.Error != nil {
		panic(fmt.Sprintf("Error while executing DB statement: %v", result.Error))
	}
}

func GetCount(dataSource *utils.DbAndSchema, tableName string) int64 {
	var count int64
	result := dataSource.Db.Table(fmt.Sprintf("%s.%s", dataSource.Schema, tableName))
	result.Count(&count)
	return count
}

func EmptyTable(dataSource *utils.DbAndSchema, tableName string) {
	dataSource.Db.Model(models.Source{}).Exec(
		fmt.Sprintf("Delete from %s.%s", dataSource.Schema, tableName))
}

func GetLastCohortId() int {
	dataSource := db.GetAtlasDB()
	var lastCohortDefinition models.CohortDefinition
	dataSource.Db.Last(&lastCohortDefinition)
	return lastCohortDefinition.Id
}

func GetLastConceptId(sourceId int) int64 {
	dataSource := GetOmopDataSourceForSourceId(sourceId)
	var lastConcept models.Concept
	dataSource.Db.Last(&lastConcept)
	return lastConcept.ConceptId
}

func GetLastObservationId(sourceId int) int64 {
	dataSource := GetOmopDataSourceForSourceId(sourceId)
	var lastObservation models.Observation
	dataSource.Db.Last(&lastObservation)
	return lastObservation.ObservationId
}

func GetLastPersonId(sourceId int) int64 {
	dataSource := GetOmopDataSourceForSourceId(sourceId)
	var lastPerson models.Person
	dataSource.Db.Last(&lastPerson)
	return lastPerson.PersonId
}

func GetOmopDataSource() *utils.DbAndSchema {
	return GetOmopDataSourceForSourceId(GetTestSourceId())
}

func GetOmopDataSourceForSourceId(sourceId int) *utils.DbAndSchema {
	var dataSourceModel = new(models.Source)
	fmt.Println(dataSourceModel)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, models.Omop)
	fmt.Println(omopDataSource)
	return omopDataSource
}

func GetResultsDataSource() *utils.DbAndSchema {
	return GetResultsDataSourceForSourceId(GetTestSourceId())
}

func GetResultsDataSourceForSourceId(sourceId int) *utils.DbAndSchema {
	var dataSourceModel = new(models.Source)
	dataSource := dataSourceModel.GetDataSource(sourceId, models.Results)
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
	var lastConceptId = GetLastConceptId(GetTestSourceId())
	conceptId := lastConceptId + 1
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

func GetRandomSubset(values []int64, subsetSize int) []int64 {
	copyValues := make([]int64, len(values))
	copy(copyValues, values)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(copyValues),
		func(i, j int) {
			copyValues[i], copyValues[j] = copyValues[j], copyValues[i]
		})
	subset := make([]int64, subsetSize)
	for i := 0; i < subsetSize; i++ {
		subset[i] = copyValues[i]
	}
	return subset
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
