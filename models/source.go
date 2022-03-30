package models

import "github.com/uc-cdis/cohort-middleware/db"

type Source struct {
	SourceId         int
	SourceName       string
	SourceConnection string `json:",omitempty"`
	SourceDialect    string `json:",omitempty"`
	Username         string `json:",omitempty"`
	Password         string `json:",omitempty"`
}

func (h Source) GetSourceById(id int) (*Source, error) {
	db2 := db.GetAtlasDB()
	var dataSource *Source
	db2.Model(&Source{}).
		Select("source_id, source_name").
		Where("source_id = ?", id).
		Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetSourceByIdWithConnection(id int) (*Source, error) {
	db2 := db.GetAtlasDB()
	var dataSource *Source
	db2.Model(&Source{}).
		Select("source_id, source_name, source_connection, source_dialect, username, password").
		Where("source_id = ?", id).
		Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetSourceByName(name string) (*Source, error) {
	db2 := db.GetAtlasDB()
	var dataSource *Source
	db2.Model(&Source{}).
		Select("source_id, source_name").
		Where("source_name = ?", name).
		Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetSourceByNameWithConnection(name string) (*Source, error) {
	db2 := db.GetAtlasDB()
	var dataSource *Source
	db2.Model(&Source{}).
		Select("source_id, source_name, source_connection, source_dialect, username, password").
		Where("source_name = ?", name).
		Scan(&dataSource)
	return dataSource, nil
}

func (h Source) GetAllSources() ([]*Source, error) {
	db2 := db.GetAtlasDB()
	var dataSource []*Source
	db2.Model(&Source{}).
		Select("source_id, source_name").
		Scan(&dataSource)
	return dataSource, nil
}
