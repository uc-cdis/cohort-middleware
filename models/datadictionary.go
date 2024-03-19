package models

import (
	"encoding/json"
	"log"

	"github.com/uc-cdis/cohort-middleware/utils"
)

type DataDictionaryI interface {
	GenerateDataDictionary(observationConceptIdsToCheck []int64) ([]*PersonConceptAndValue, error)
}

type DataDictionary struct {
	cohortDataModel CohortDataI
}

type DataDictionaryModel struct {
	Total int64                  `json:"total"`
	Data  []*DataDictionaryEntry `json:"Data"`
}
type DataDictionaryEntry struct {
	VocabularyID                int64  `json:"vocabularyID"`
	ConceptID                   int64  `json:"conceptID"`
	ConceptCode                 string `json:"conceptCode"`
	ConceptClassId              string `json:"conceptClassId"`
	NumberOfPeopleWithVariable  int64  `json:"NumberOfPeopleWithVariable"`
	NumberOfPeopleWithValue     int64  `json:"NumberOfPeopleWithValue"`
	NumberOfPeopleWithNullValue int64  `json:"NumberOfPeopleWithNullValue"`
	ValueStoredAs               string `json:"valueStoredAs"`
	MinValue                    int64  `json:"minValue"`
	MaxValue                    int64  `json:"maxValue"`
	MeanValue                   int64  `json:"meanValue"`
	StandardDeviation           int64  `json:"standardDeviation"`
	ValueSummary                []byte `json:"valueSummary"`
}

// Generate Data Dictionary Json
func (u DataDictionary) GenerateDataDictionary() (DataDictionaryModel, error) {

	//TODO: Get this from some sort of config file later
	var catchAllCohortId = 404 //qa catch all cohort
	var source = new(Source)
	sources, _ := source.GetAllSources()

	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sources[0].SourceId, Omop)
	//resultsDataSource := dataSourceModel.GetDataSource(source.SourceId, Results)

	var dataDictionaryModel DataDictionaryModel
	var dataDictionaryEntries []*DataDictionaryEntry
	//var selects = []string{"vocabulary_id", "concept_id", "concept_code", "concept_class_id"}
	query := omopDataSource.Db.Table(omopDataSource.Schema + ".data_dictionary" + omopDataSource.GetViewDirective())

	query, cancel := utils.AddTimeoutToQuery(query)
	defer cancel()
	meta_result := query.Scan(&dataDictionaryEntries)
	if meta_result.Error != nil {
		return dataDictionaryModel, meta_result.Error
	} else if len(dataDictionaryEntries) == 0 {
		log.Printf("INFO: no data dictionary entry found")
	} else {
		log.Printf("INFO: Data dictionary entries found.")
	}

	for _, data := range dataDictionaryEntries {
		if data.ConceptClassId == "MVP Continuous" {
			// MVP Continuous #similar to bin items below call cohort-middleware
			// Example call, parameter for cohort definition and source id https://qa-mickey.planx-pla.net/cohort-middleware/histogram/by-source-id/2/by-cohort-definition-id/404/by-histogram-concept-id/2000006886
			/*
				[{
				start: number
				end: number
				personCount: number
				},]
			*/
			cohortData, _ := u.cohortDataModel.RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sources[0].SourceId, catchAllCohortId, data.ConceptID, []int64{}, []utils.CustomDichotomousVariableDef{})

			conceptValues := []float64{}
			for _, personData := range cohortData {
				conceptValues = append(conceptValues, float64(*personData.ConceptValueAsNumber))
			}

			histogramData := utils.GenerateHistogramData(conceptValues)
			data.ValueSummary, _ = json.Marshal(histogramData)
		} else {
			//Get Value Summary from bar graph method
			// MVP ordinal use this structure , bin people based on value_as_concept_id and get the count
			/*{
			    name: string
			    personCount: number
			    valueAsString: number
			    valueAsConceptID: number
			}*/
			ordinalValueData, _ := u.cohortDataModel.RetrieveBarGraphDataBySourceIdAndCohortIdAndConceptIds(sources[0].SourceId, catchAllCohortId, data.ConceptID)

			data.ValueSummary, _ = json.Marshal(ordinalValueData)
		}
	}

	query = omopDataSource.Db.Table(omopDataSource.Schema + ".observation_continuous as observation" + omopDataSource.GetViewDirective()).
		Select("count(distinct observation.observation_concept_id) as total, NULL as data")

	query, cancel = utils.AddTimeoutToQuery(query)
	defer cancel()
	_ = query.Scan(&dataDictionaryModel)
	dataDictionaryModel.Data = dataDictionaryEntries
	return dataDictionaryModel, nil
}
