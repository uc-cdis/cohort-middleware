package tests

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"

	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/utils"
	"gorm.io/gorm"
)

// Global variable
var testSourceId int = 1

func GetTestSourceId() int {
	log.Printf("Using source id %d...", testSourceId)
	return testSourceId
}

func SetTestSourceId(id int) {
	log.Printf("Setting source id %d...", id)
	testSourceId = id
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
	fileContents, err2 := os.ReadFile(path)
	if err2 != nil {
		panic(err2)
	}
	result := ExecSQLString(string(fileContents), sourceId)
	if result.Error != nil && len(result.Error.Error()) > 0 {
		errString := result.Error.Error()
		fmt.Printf("Errors: %s\n", errString)
		panic(errString)
	}
}

func ExecAtlasSQLString(sqlString string) {
	ExecSQLString(sqlString, -1)
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
		omopDataSource := dataSourceModel.GetDataSource(sourceId, models.Omop)
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

func GetCountWhere(dataSource *utils.DbAndSchema, tableName string, whereClause string) int64 {
	var count int64
	result := dataSource.Db.Table(fmt.Sprintf("%s.%s where %s", dataSource.Schema, tableName, whereClause))
	result.Count(&count)
	return count
}

func EmptyTable(dataSource *utils.DbAndSchema, tableName string) {
	dataSource.Db.Model(models.Source{}).Exec(
		fmt.Sprintf("Delete from %s.%s", dataSource.Schema, tableName))
}

func GetLastCohortId() int {
	dataSource := db.GetAtlasDB()
	var lastCohortDefinitionId int
	dataSource.Db.Raw("SELECT MAX(id) FROM " + dataSource.Schema + ".cohort_definition").Scan(&lastCohortDefinitionId)
	return lastCohortDefinitionId
}

func GetNextCohortId() int {
	dataSource := db.GetAtlasDB()
	var nextCohortId int
	// Use raw SQL to get the next value from the sequence
	dataSource.Db.Raw("SELECT NEXTVAL('" + dataSource.Schema + ".cohort_definition_sequence')").Scan(&nextCohortId)
	return nextCohortId
}

func GetLastConceptId(sourceId int) int64 {
	dataSource := GetOmopDataSourceForSourceId(sourceId)
	var lastConceptId int64
	dataSource.Db.Raw("SELECT MAX(concept_id) FROM " + dataSource.Schema + ".concept").Scan(&lastConceptId)
	log.Printf("Last concept id found %d",
		lastConceptId)
	return lastConceptId
}

func GetLastObservationId(sourceId int) int64 {
	dataSource := GetOmopDataSourceForSourceId(sourceId)
	var lastObservationId int64
	dataSource.Db.Raw("SELECT MAX(observation_id) FROM " + dataSource.Schema + ".observation").Scan(&lastObservationId)
	return lastObservationId
}

func GetLastPersonId(sourceId int) int64 {
	dataSource := GetOmopDataSourceForSourceId(sourceId)
	var lastPersonId int64
	dataSource.Db.Raw("SELECT MAX(person_id) FROM " + dataSource.Schema + ".person").Scan(&lastPersonId)
	return lastPersonId
}

func GetOmopDataSource() *utils.DbAndSchema {
	return GetOmopDataSourceForSourceId(GetTestSourceId())
}

func GetOmopDataSourceForSourceId(sourceId int) *utils.DbAndSchema {
	var dataSourceModel = new(models.Source)
	omopDataSource := dataSourceModel.GetDataSource(sourceId, models.Omop)
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

func ConceptExists(sourceType models.SourceType, conceptId int64) bool {
	var dataSourceModel = new(models.Source)
	dataSource := dataSourceModel.GetDataSource(GetTestSourceId(), sourceType)
	count := 0
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s.concept WHERE concept_id = ?", GetSchemaNameForType(sourceType))
	dataSource.Db.Raw(query, conceptId).Scan(&count)
	return count > 0
}

func GetInt64AttributeValue[T any](item T, attributeName string) int64 {
	r := reflect.ValueOf(item)
	f := reflect.Indirect(r).FieldByName(attributeName)
	return f.Int()
}

func GetRandomSubset(values []int64, subsetSize int) []int64 {
	copyValues := make([]int64, len(values))
	copy(copyValues, values)
	log.Printf("Getting a random subset of size %d from a set of values of size %d",
		subsetSize, len(values))
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
	StatusCode              int
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
	// Store the status code
	w.StatusCode = statusCode
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
	return w.StatusCode
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
