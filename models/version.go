package models

import (
	"time"

	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/utils"
	"github.com/uc-cdis/cohort-middleware/version"
)

type Version struct {
	GitCommit  string
	GitVersion string
}

type DbSchemaVersion struct {
	AtlasSchemaVersion string
	DataSchemaVersion  int
}

type SchemaVersion struct {
	InstalledRank int
	Version       string
	Description   string
	Type          string
	Script        string
	Checksum      int
	InstalledBy   string
	InstalledOn   time.Time
	ExecutionTime int
	Success       bool
}

type VersionInfo struct {
	Version     int
	AppliedOn   time.Time
	Description string
}

func (h Version) GetVersion() *Version {
	return &Version{GitCommit: version.GitCommit, GitVersion: version.GitVersion}
}

func (h Version) GetSchemaVersion() *DbSchemaVersion {
	dbSchemaVersion := &DbSchemaVersion{"error", -1}

	atlasDb := db.GetAtlasDB().Db
	var atlasSchemaVersion *SchemaVersion
	query := atlasDb.Model(&SchemaVersion{}).
		Limit(1).
		Select("Version").
		Order("Version Desc")

	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&atlasSchemaVersion)
	if meta_result.Error == nil {
		dbSchemaVersion.AtlasSchemaVersion = atlasSchemaVersion.Version
	}

	var source = new(Source)
	sources, _ := source.GetAllSources()
	if len(sources) < 1 {
		panic("Error: No data source found")
	} else if len(sources) > 1 {
		panic("More than one data source! Exiting")
	}
	var dataSourceModel = new(Source)
	dboDataSource := dataSourceModel.GetDataSource(sources[0].SourceId, Dbo)

	var versionInfo *VersionInfo
	query = dboDataSource.Db.Table(dboDataSource.Schema + ".versioninfo").
		Limit(1).
		Select("Version").
		Order("Version Desc")

	query, cancel = utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result = query.Scan(&versionInfo)
	if meta_result.Error == nil {
		dbSchemaVersion.DataSchemaVersion = versionInfo.Version
	}

	return dbSchemaVersion
}
