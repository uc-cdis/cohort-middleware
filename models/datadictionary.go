package models

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type DataDictionaryI interface {
	GenerateDataDictionary()
	GetDataDictionary() (*DataDictionaryModel, error)
}

type DataDictionary struct {
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
	ConceptClassId                   string          `json:"conceptClassID"`
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

type DataDictionaryResult struct {
	VocabularyID                     string          `json:"vocabularyID"`
	ConceptID                        int64           `json:"conceptID"`
	ConceptCode                      string          `json:"conceptCode"`
	ConceptName                      string          `json:"conceptName"`
	ConceptClassId                   string          `json:"conceptClassID"`
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

var ResultCache *DataDictionaryModel = nil

func (u DataDictionary) GetDataDictionary() (*DataDictionaryModel, error) {
	//Read from cache
	if ResultCache != nil {
		return ResultCache, nil
	} else {
		//Read from DB
		var source = new(Source)
		sources, _ := source.GetAllSources()
		if len(sources) < 1 {
			panic("Error: No data source found")
		} else if len(sources) > 1 {
			panic("More than one data source! Exiting")
		}
		var dataSourceModel = new(Source)
		omopDataSource := dataSourceModel.GetDataSource(sources[0].SourceId, Omop)
		miscDataSource := dataSourceModel.GetDataSource(sources[0].SourceId, Misc)

		if u.CheckIfDataDictionaryIsFilled(miscDataSource) {
			var newDataDictionary DataDictionaryModel
			var dataDictionaryEntries []*DataDictionaryResult
			//Get total number of person ids
			query := omopDataSource.Db.Table(omopDataSource.Schema + ".observation as observation").
				Select("count(distinct observation.person_id) as total, null as data")

			query, cancel := utils.AddSpecificTimeoutToQuery(query, 600*time.Second)
			defer cancel()
			meta_result := query.Scan(&newDataDictionary)

			if meta_result.Error != nil {
				log.Printf("ERROR: Failed to get number of person_ids")
				return nil, errors.New("data dictionary is not available yet")
			} else {
				log.Printf("INFO: Total number of person_ids from observation view is %v.", newDataDictionary.Total)
			}

			//get data dictionary entires saved in table
			query = miscDataSource.Db.Table(miscDataSource.Schema + ".data_dictionary_result")
			query, cancel = utils.AddSpecificTimeoutToQuery(query, 600*time.Second)
			defer cancel()
			meta_result = query.Scan(&dataDictionaryEntries)

			if meta_result.Error != nil {
				log.Printf("ERROR: Failed to get data entries")
				return nil, errors.New("data dictionary is not available yet")
			} else {
				log.Printf("INFO: Got data entries")
			}

			newDataDictionary.Data, _ = json.Marshal(dataDictionaryEntries)
			//set in cache
			ResultCache = &newDataDictionary
			return &newDataDictionary, nil
		} else {
			return nil, errors.New("data dictionary is not available yet")
		}
	}
}

// Generate Data Dictionary Json
func (u DataDictionary) GenerateDataDictionary() {
	conf := config.GetConfig()
	var maxWorkerSize int = conf.GetInt("worker_pool_size")
	log.Printf("maxWorkerSize is %v", maxWorkerSize)
	var batchSize int = conf.GetInt("batch_size")
	log.Printf("Batch Size is %v", batchSize)

	entryCh := make(chan *DataDictionaryResult, maxWorkerSize)

	var source = new(Source)
	sources, _ := source.GetAllSources()
	if len(sources) < 1 {
		panic("Error: No data source found")
	} else if len(sources) > 1 {
		panic("More than one data source! Exiting")
	}
	var dataSourceModel = new(Source)
	miscDataSource := dataSourceModel.GetDataSource(sources[0].SourceId, Misc)

	if u.CheckIfDataDictionaryIsFilled(miscDataSource) {
		log.Print("Data Dictionary Result already filled. Skipping generation.")
		return
	} else {
		var dataDictionaryEntries []*DataDictionaryEntry
		//see ddl_results_and_cdm.sql Data_Dictionary view
		query := miscDataSource.Db.Table(miscDataSource.Schema + ".data_dictionary")

		query, cancel := utils.AddSpecificTimeoutToQuery(query, 600*time.Second)
		defer cancel()
		meta_result := query.Scan(&dataDictionaryEntries)
		if meta_result.Error != nil {
			log.Printf("Error: db read error: %v", meta_result.Error)
			return
		} else if len(dataDictionaryEntries) == 0 {
			log.Printf("INFO: no data dictionary view entry found")
		} else {
			log.Printf("INFO: Data dictionary view entries found.")
		}

		log.Printf("Get all histogram/bar graph data")
		var partialDataList []*DataDictionaryEntry
		var resultDataList []*DataDictionaryResult = []*DataDictionaryResult{}
		for len(dataDictionaryEntries) > 0 {
			wg := sync.WaitGroup{}
			partialResultList := []*DataDictionaryResult{}
			if len(dataDictionaryEntries) < maxWorkerSize {
				partialDataList = dataDictionaryEntries
				dataDictionaryEntries = []*DataDictionaryEntry{}
			} else {
				partialDataList = dataDictionaryEntries[:maxWorkerSize-1]
				dataDictionaryEntries = dataDictionaryEntries[maxWorkerSize-1:]
			}

			for _, d := range partialDataList {
				wg.Add(1)
				go GenerateData(d, sources[0].SourceId, &wg, entryCh)
				resultEntry := <-entryCh
				partialResultList = append(partialResultList, resultEntry)
			}
			wg.Wait()
			resultDataList = append(resultDataList, partialResultList...)
			if len(resultDataList) >= batchSize {
				log.Printf("%v row of results reached, flush to db.", batchSize)
				u.WriteResultToDB(miscDataSource, resultDataList)
				resultDataList = []*DataDictionaryResult{}
			}
		}

		if len(resultDataList) > 0 {
			u.WriteResultToDB(miscDataSource, resultDataList)
		}

		log.Printf("INFO: Data dictionary generation complete")
		return
	}
}

func GenerateData(data *DataDictionaryEntry, sourceId int, wg *sync.WaitGroup, ch chan *DataDictionaryResult) {
	var c = new(CohortData)

	if data.ValueStoredAs == "Number" {
		//If histogram concept classes
		log.Printf("Generate histogram for Concept id %v.", data.ConceptClassId)
		var cohortData []*PersonConceptAndValue

		cohortData, _ = c.RetrieveHistogramDataBySourceIdAndConceptId(sourceId, data.ConceptID)

		conceptValues := []float64{}
		for _, personData := range cohortData {
			conceptValues = append(conceptValues, float64(*personData.ConceptValueAsNumber))
		}
		log.Printf("INFO: concept id %v data size is %v", data.ConceptID, len(conceptValues))
		histogramData := utils.GenerateHistogramData(conceptValues)
		data.ValueSummary, _ = json.Marshal(histogramData)
	} else if data.ValueStoredAs == "Concept Id" {
		//If bar graph concept classes
		log.Printf("Generate bar graph for Concept id %v.", data.ConceptClassId)
		nominalValueData, _ := c.RetrieveBarGraphDataBySourceIdAndCohortIdAndConceptIds(sourceId, data.ConceptID)
		data.ValueSummary, _ = json.Marshal(nominalValueData)
	}
	result := DataDictionaryResult(*data)

	//send result to channel
	ch <- &result
	wg.Done()
}

func (u DataDictionary) WriteResultToDB(dbSource *utils.DbAndSchema, resultDataList []*DataDictionaryResult) bool {

	result := dbSource.Db.Create(resultDataList)
	if result.Error != nil {
		log.Printf("ERROR: Failed to insert data into table")
		panic("")
	}
	log.Printf("Write to DB succeeded.")
	return true
}

func (u DataDictionary) CheckIfDataDictionaryIsFilled(dbSource *utils.DbAndSchema) bool {
	var dataDictionaryResult []*DataDictionaryResult
	query := dbSource.Db.Table(dbSource.Schema + ".data_dictionary_result")

	query, cancel := utils.AddSpecificTimeoutToQuery(query, 600*time.Second)
	defer cancel()
	meta_result := query.Scan(&dataDictionaryResult)
	if meta_result.Error != nil {
		log.Printf("ERROR: Failed to get data dictionary result")
		panic("")
	} else if len(dataDictionaryResult) > 0 {
		log.Printf("INFO: Data Dictionary Result Table is filled.")
		return true
	} else {
		log.Printf("INFO: Data Dictionary Result Table is empty.")
		return false
	}
}
