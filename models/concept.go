package models

import (
	"github.com/uc-cdis/cohort-middleware/utils"
)

type Concept struct {
	ConceptId   int
	ConceptName string
	DomainId    int
	DomainName  string
}

func (h Concept) RetriveAllBySourceId(sourceId int) ([]*Concept, error) {
	var dataSourceModel = new(Source)
	dataSource, _ := dataSourceModel.GetSourceByIdWithConnection(sourceId)

	sourceConnectionString := dataSource.SourceConnection
	dbSchema := "OMOP." // TODO - verify w/ Andrew that this part of the schema is really called "OMOP"
	omopDataSource := utils.GetDataSourceDB(sourceConnectionString, dbSchema)

	var concept []*Concept
	omopDataSource.Model(&Concept{}).
		Select("concept_id, concept_name, domain_id, '' as domain_name"). // TODO - find how to do join
		Order("concept_name").
		Scan(&concept)
	return concept, nil
}
