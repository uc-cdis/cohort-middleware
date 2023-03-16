package models

import (
	"fmt"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type Source struct {
	SourceId         int    `json:"source_id"`
	SourceName       string `json:"source_name"`
	SourceConnection string `json:",omitempty"`
	SourceDialect    string `json:",omitempty"`
	Username         string `json:",omitempty"`
	Password         string `json:",omitempty"`
}

func (h Source) GetSourceById(id int) (*Source, error) {
	db2 := db.GetAtlasDB().Db
	var dataSource *Source
	db2.Model(&Source{}).
		Select("source_id, source_name").
		Where("source_id = ?", id).
		Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetSourceByIdWithConnection(id int) (*Source, error) {
	db2 := db.GetAtlasDB().Db
	var dataSource *Source
	db2.Model(&Source{}).
		Select("source_id, source_name, source_connection, source_dialect, username, password").
		Where("source_id = ?", id).
		Scan(&dataSource)
	return dataSource, nil
}

type SourceSchema struct {
	SchemaName string
}

func (h Source) GetSourceSchemaNameBySourceIdAndSourceType(id int, sourceType SourceType) (*SourceSchema, error) {
	atlasDb := db.GetAtlasDB()
	db2 := atlasDb.Db
	var sourceSchema *SourceSchema
	db2.Model(&Source{}).
		Select("source_daimon.table_qualifier as schema_name").
		Joins("INNER JOIN "+atlasDb.Schema+".source_daimon ON source.source_id = source_daimon.source_id").
		Where("source.source_id = ?", id).
		Where("source_daimon.daimon_type = ?", sourceType).
		Scan(&sourceSchema)
	return sourceSchema, nil
}

type SourceType int64

const (
	Omop    SourceType = 0 //TODO - we might have to split up into OmopData and OmopVocab in future...
	Results SourceType = 2
	Temp    SourceType = 5
)

// Get the data source details for given source id and source type.
// The source type can be one of the type SourceType.
func (h Source) GetDataSource(sourceId int, sourceType SourceType) *utils.DbAndSchema {
	fmt.Println(sourceId, sourceType)
	dataSource, _ := h.GetSourceByIdWithConnection(sourceId)

	sourceConnectionString := dataSource.SourceConnection
	dbSchema, _ := h.GetSourceSchemaNameBySourceIdAndSourceType(sourceId, sourceType)
	dbSchemaName := dbSchema.SchemaName
	dbAndSchema := utils.GetDataSourceDB(sourceConnectionString, dbSchemaName)
	return dbAndSchema
}

func (h Source) GetSourceByName(name string) (*Source, error) {
	db2 := db.GetAtlasDB().Db
	var dataSource *Source
	db2.Model(&Source{}).
		Select("source_id, source_name").
		Where("source_name = ?", name).
		Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetAllSources() ([]*Source, error) {
	db2 := db.GetAtlasDB().Db
	var dataSource []*Source
	db2.Model(&Source{}).
		Select("source_id, source_name").
		Scan(&dataSource)
	return dataSource, nil
}
