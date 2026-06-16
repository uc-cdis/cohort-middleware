package models

import (
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type Source struct {
	SourceId                     int    `json:"source_id"`
	SourceName                   string `json:"source_name"`
	Description                  string `json:"description,omitempty" gorm:"column:description"`
	SourceConnection             string `json:",omitempty"`
	SourceDialect                string `json:",omitempty"`
	Username                     string `json:",omitempty"`
	Password                     string `json:",omitempty"`
	TeamProject                  string `json:",omitempty" gorm:"column:team_project"`
	CurrentTeamProjectAccessible string `json:",omitempty" gorm:"column:current_team_project_accessible"`
}

type SourceI interface {
	GetSourceById(id int) (*Source, error)
	GetSourceByName(name string) (*Source, error)
	GetAllSources() ([]*Source, error)
	GetAllSourcesWithTeamProject(teamName string) ([]*Source, error)
}

func (h Source) GetSourceById(id int) (*Source, error) {
	db2 := db.GetAtlasDB().Db
	var dataSource *Source
	query := db2.Model(&Source{}).
		Select("source_id, source_name").
		Where("source_id = ?", id).
		Where("deleted_date is null")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	query.Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetSourceByIdWithConnection(id int) (*Source, error) {
	db2 := db.GetAtlasDB().Db
	var dataSource *Source
	query := db2.Model(&Source{}).
		Select("source_id, source_name, source_connection, source_dialect, username, password").
		Where("source_id = ?", id).
		Where("deleted_date is null")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	query.Scan(&dataSource)
	return dataSource, nil
}

type SourceSchema struct {
	SchemaName string
}

func (h Source) GetSourceSchemaNameBySourceIdAndSourceType(id int, sourceType SourceType) (*SourceSchema, error) {
	// special handling of sourceType "Misc", as it is not stored in source_daimon table
	if sourceType == Misc {
		return &SourceSchema{SchemaName: "MISC"}, nil
	}

	if sourceType == Dbo {
		return &SourceSchema{SchemaName: "DBO"}, nil
	}

	// otherwise, get the schema name from source_daimon table
	atlasDb := db.GetAtlasDB()
	db2 := atlasDb.Db
	var sourceSchema *SourceSchema
	query := db2.Model(&Source{}).
		Select("source_daimon.table_qualifier as schema_name").
		Joins("INNER JOIN "+atlasDb.Schema+".source_daimon ON source.source_id = source_daimon.source_id").
		Where("source.source_id = ?", id).
		Where("source_daimon.daimon_type = ?", sourceType).
		Where("source.deleted_date is null")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	query.Scan(&sourceSchema)
	return sourceSchema, nil
}

type SourceType int64

const (
	Omop    SourceType = 0 //TODO - we might have to split up into OmopData and OmopVocab in future...
	Results SourceType = 2
	Temp    SourceType = 5
	Misc    SourceType = 6
	Dbo     SourceType = 7
)

// Get the data source details for given source id and source type.
// The source type can be one of the type SourceType.
func (h Source) GetDataSource(sourceId int, sourceType SourceType) *utils.DbAndSchema {
	dataSource, _ := h.GetSourceByIdWithConnection(sourceId)

	dbSchema, _ := h.GetSourceSchemaNameBySourceIdAndSourceType(sourceId, sourceType)
	dbSchemaName := dbSchema.SchemaName
	sourceConnection := utils.SourceConnection{SourceConnection: dataSource.SourceConnection,
		Username: dataSource.Username,
		Password: dataSource.Password, // pragma: allowlist secret
	}
	dbAndSchema := utils.GetDataSourceDB(sourceConnection, dbSchemaName)
	return dbAndSchema
}

func (h Source) GetSourceByName(name string) (*Source, error) {
	db2 := db.GetAtlasDB().Db
	var dataSource *Source
	query := db2.Model(&Source{}).
		Select("source_id, source_name").
		Where("source_name = ?", name).
		Where("deleted_date is null")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	query.Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetAllSources() ([]*Source, error) {
	db2 := db.GetAtlasDB().Db
	var dataSource []*Source
	query := db2.Model(&Source{}).
		Select("source_id, source_name").
		Where("deleted_date is null")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	query.Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetAllSourcesWithTeamProject(teamName string) ([]*Source, error) {
	atlasDb := db.GetAtlasDB()
	db2 := atlasDb.Db
	var dataSource []*Source
	query := db2.Table(atlasDb.Schema+".source AS s").
		Select(`
			s.source_id AS source_id,
			s.source_name AS source_name,
			sr.name AS team_project,
			(sr.name = ?) AS current_team_project_accessible
		`, teamName).
		Joins(`
			JOIN `+atlasDb.Schema+`.sec_permission sp
			  ON s.source_key = SUBSTRING(sp.value FROM 'generate:(.*?):get')
		`).
		Joins(`
			JOIN `+atlasDb.Schema+`.sec_role_permission srp
			  ON sp.id = srp.permission_id
		`).
		Joins(`
			JOIN `+atlasDb.Schema+`.sec_role sr
			  ON srp.role_id = sr.id
		`).
		Where("sr.name LIKE ?", "/gwas_projects/%").
		Where("s.deleted_date is null")
	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	metaResult := query.Scan(&dataSource)
	if metaResult.Error != nil {
		return nil, metaResult.Error
	}

	for _, source := range dataSource {
		var meta struct {
			Description string `gorm:"column:description"`
		}

		omopDataSource := h.GetDataSource(source.SourceId, Omop)

		query := omopDataSource.Db.Table(omopDataSource.Schema + ".cdm_source").
			Select("source_description as description").
			Limit(1)

		query, cancel := utils.AddTimeoutToQuery(query)
		metaResult := query.Scan(&meta)
		cancel()

		if metaResult.Error != nil {
			return nil, metaResult.Error
		}

		source.Description = meta.Description
	}

	return dataSource, nil
}
