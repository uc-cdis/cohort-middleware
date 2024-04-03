package models

import (
	"encoding/json"
	"log"
	"time"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type DataDictionaryI interface {
	GenerateDataDictionary() (*DataDictionaryModel, error)
}

type DataDictionary struct {
	CohortDataModel CohortDataI
}

type DataDictionaryModel struct {
	Total int64           `json:"total"`
	Data  json.RawMessage `json:"data"`
}

type DataDictionaryEntry struct {
	VocabularyID                     string          `json:"vocabularyID"`
	ConceptID                        int64           `json:"conceptID"`
	ConceptCode                      string          `json:"conceptCode"`
	ConceptName                      string          `json:"conceptName"`
	ConceptClassId                   string          `json:"conceptClassId"`
	NumberOfPeopleWithVariable       int64           `json:"numberOfPeopleWithVariable"`
	NumberOfPeopleWhereValueIsFilled int64           `json:"numberOfPeopleWhereValueIsFilled"`
	NumberOfPeopleWhereValueIsNull   int64           `json:"numberOfPeopleWhereValueIsNull"`
	ValueStoredAs                    string          `json:"valueStoredAs"`
	MinValue                         float64         `json:"minValue"`
	MaxValue                         float64         `json:"maxValue"`
	MeanValue                        float64         `json:"meanValue"`
	StandardDeviation                float64         `json:"standardDeviation"`
	ValueSummary                     json.RawMessage `json:"valueSummary"`
}

// Generate Data Dictionary Json
func (u DataDictionary) GenerateDataDictionary() (*DataDictionaryModel, error) {
	log.Printf("Generating Data Dictionary...")
	conf := config.GetConfig()
	var catchAllCohortId = conf.GetInt("catch_all_cohort_id")
	log.Printf("catch all cohort id is %v", catchAllCohortId)
	var source = new(Source)
	sources, _ := source.GetAllSources()

	var dataSourceModel = new(Source)
	omopDataSource := dataSourceModel.GetDataSource(sources[0].SourceId, Omop)

	var dataDictionaryModel DataDictionaryModel
	var dataDictionaryEntries []*DataDictionaryEntry
	//see ddl_results_and_cdm.sql Data_Dictionary view
	query := omopDataSource.Db.Table(omopDataSource.Schema + ".data_dictionary")

	query, cancel := utils.AddSpecificTimeoutToQuery(query, 600*time.Second)
	defer cancel()
	meta_result := query.Scan(&dataDictionaryEntries)
	if meta_result.Error != nil {
		return &dataDictionaryModel, meta_result.Error
	} else if len(dataDictionaryEntries) == 0 {
		log.Printf("INFO: no data dictionary entry found")
	} else {
		log.Printf("INFO: Data dictionary entries found.")
	}

	//Get total number of concept ids
	query = omopDataSource.Db.Table(omopDataSource.Schema + ".observation_continuous as observation" + omopDataSource.GetViewDirective()).
		Select("count(distinct observation.observation_concept_id) as total, null as data")

	query, cancel = utils.AddSpecificTimeoutToQuery(query, 600*time.Second)
	defer cancel()
	meta_result = query.Scan(&dataDictionaryModel)

	if meta_result.Error != nil {
		log.Printf("ERROR: Failed to get number of concepts")
	} else {
		log.Printf("INFO: Got total number of concepts from observation view.")
	}

	log.Printf("Get all histogram/bar graph data")

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
			var filterConceptIds = []int64{}
			var filterCohortPairs = []utils.CustomDichotomousVariableDef{}
			if u.CohortDataModel == nil {
				u.CohortDataModel = new(CohortData)
			}
			cohortData, _ := u.CohortDataModel.RetrieveHistogramDataBySourceIdAndCohortIdAndConceptIdsAndCohortPairs(sources[0].SourceId, catchAllCohortId, data.ConceptID, filterConceptIds, filterCohortPairs)
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
			ordinalValueData, _ := u.CohortDataModel.RetrieveBarGraphDataBySourceIdAndCohortIdAndConceptIds(sources[0].SourceId, catchAllCohortId, data.ConceptID)
			data.ValueSummary, _ = json.Marshal(ordinalValueData)
		}
	}

	log.Printf("INFO: Data dictionary generation complete")
	dataDictionaryModel.Data, _ = json.Marshal(dataDictionaryEntries)

	return &dataDictionaryModel, nil
}
