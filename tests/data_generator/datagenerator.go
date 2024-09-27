package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/spf13/viper"
	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/tests"
	"github.com/uc-cdis/cohort-middleware/utils"
)

type Concepts struct {
	Concepts []Concept
}

type Concept struct {
	Concept           string
	Id                string
	ConceptName       string   `mapstructure:"concept_name"`
	ValueType         string   `mapstructure:"value_type"`
	PersonIds         []int64  `mapstructure:"person_ids"`
	PersonIdsOverride []int64  // we use this one for dynamic personIds from cohorts that are created on the fly
	PossibleValues    []string `mapstructure:"possible_values"`
	ConceptValueName  string   `mapstructure:"concept_value_name"`
	RatioOfPersons    float32  `mapstructure:"ratio_of_persons"`
	CloneCount        int      `mapstructure:"clone_count"`
}

type Cohorts struct {
	Cohorts []Cohort
}

type Cohort struct {
	Cohort          string
	PersonIds       []int64 `mapstructure:"person_ids"`
	NumberOfPersons int     `mapstructure:"number_of_persons"`
	CloneCount      int     `mapstructure:"clone_count"`
	Concepts        []Concept
}

func RunDataGeneration(testDataConfigFilePrefix string) {
	conf := GetTestDataConfig(testDataConfigFilePrefix)

	concepts := Concepts{}
	error1 := conf.Unmarshal(&concepts)
	AddConcepts(concepts.Concepts)

	cohorts := Cohorts{}
	error2 := conf.Unmarshal(&cohorts)
	AddCohorts(cohorts.Cohorts)

	if error1 != nil || error2 != nil {
		log.Fatalf("Error while parsing configuration files: %v, %v", error1, error2)
	}
	ResetCohortCache()
}

func ResetCohortCache() {
	tests.ExecSQLStringOrFail(
		fmt.Sprintf(
			"DELETE from %s.cohort_cache ",
			tests.GetResultsDataSourceForSourceId(sourceId).Schema),
		sourceId)
}

func GetTestDataConfig(configFilePrefix string) *viper.Viper {
	var err error
	config := viper.New()
	config.SetConfigType("yaml")
	log.Printf("Reading test data config \"%s\"...", configFilePrefix)
	config.SetConfigName(configFilePrefix)
	config.AddConfigPath(".")
	err = config.ReadInConfig()
	if err != nil {
		log.Fatalf("Error on parsing test data configuration file: %v", err)
	}
	return config
}

func AddCohorts(cohorts []Cohort) {
	log.Printf("Processing %d cohorts...", len(cohorts))
	for _, cohort := range cohorts {
		var nrClones = 1
		if cohort.CloneCount > 0 {
			nrClones = cohort.CloneCount
			log.Printf("** Clone directive found! Cloning this cohort %d times **", cohort.CloneCount)
		}
		for i := 0; i < nrClones; i++ {
			AddCohort(cohort)
		}
	}
}

func AddCohort(cohort Cohort) {
	cohortId := tests.GetNextCohortId()
	cohortName := fmt.Sprintf("%s (%d)", cohort.Cohort, cohortId)
	log.Printf("Adding cohort_definition '%s'...", cohortName)
	// We should probably remove all cohort parts...or add them to sec_ table... TODO
	// tests.ExecSQLStringOrFail(
	// 	fmt.Sprintf(
	// 		"INSERT into %s.cohort_definition "+
	// 			"(id,name,description) "+
	// 			"values "+
	// 			"(%d,'%s','%s')",
	// 		db.GetAtlasDB().Schema,
	// 		cohortId, cohortName, cohortName),
	// 	-1)
	AddCohortPersonsAndObservations(cohortId, cohort)
}

func AddCohortPersonsAndObservations(cohortId int, cohort Cohort) {
	// Add also to results.cohort table:
	personIdsAdded := make([]int64, 0)
	if cohort.PersonIds != nil {
		log.Printf("Adding %d persons to cohort...", len(cohort.PersonIds))
		for _, personId := range cohort.PersonIds {
			AddPersonToCohort(cohortId, personId)
		}
		personIdsAdded = cohort.PersonIds
	} else if cohort.NumberOfPersons > 0 {
		// generate cohort.NumberOfPersons person records:
		log.Printf("Adding %d persons to cohort...", cohort.NumberOfPersons)
		nextPersonId := tests.GetLastPersonId(sourceId) + 1
		for i := 0; i < cohort.NumberOfPersons; i++ {
			personId := nextPersonId + int64(i)
			AddPersonToCohort(cohortId, personId)
			personIdsAdded = append(personIdsAdded, personId)
		}
	} else {
		panic("Invalid config. A cohort must have persons")
	}
	// special case: we can have concepts listed in a cohort, meaning we
	// want persons of this cohort to have observations registered
	// for these concepts:
	if cohort.Concepts != nil {
		for i := range cohort.Concepts {
			concept := &cohort.Concepts[i]
			if concept.RatioOfPersons > 0 {
				log.Printf("Adding concept and observation to only a fraction of the persons in this cohort. Fraction: %f",
					concept.RatioOfPersons)
				concept.PersonIdsOverride = tests.GetRandomSubset(personIdsAdded,
					int(float32(len(personIdsAdded))*concept.RatioOfPersons))
			} else {
				if concept.PersonIds == nil {
					// just copy over all personIds:
					concept.PersonIdsOverride = personIdsAdded
				}
			}
		}
		// now that the personIds are copied over, we can add the concepts (and respective observations)
		// for the persons of this cohort:
		AddConcepts(cohort.Concepts)
	}
}

func AddPersonToCohort(cohortId int, personId int64) {
	AddPerson(personId)
	// We should probably remove all cohort parts...or add them to sec_ table... TODO
	// tests.ExecSQLStringOrFail(
	// 	fmt.Sprintf(
	// 		"INSERT into %s.cohort "+
	// 			"(cohort_definition_id,subject_id,cohort_start_date,cohort_end_date) "+
	// 			"values "+
	// 			"(%d,%d,'1970-01-01','2999-01-01')",
	// 		tests.GetResultsDataSourceForSourceId(sourceId).Schema,
	// 		cohortId, personId),
	// 	sourceId)
}

func AddPerson(personId int64) {
	genderConceptIds := []int{8507, 8532}

	// only add if this personId is new:
	if utils.Pos(personId, personIds) == -1 {
		tests.ExecSQLStringOrFail(
			fmt.Sprintf(
				"INSERT into %s.person "+
					"(person_id,year_of_birth,month_of_birth,day_of_birth,gender_concept_id,race_concept_id,ethnicity_concept_id) "+
					"values "+
					"(%d,%d,%d,%d,%d,0,0)",
				tests.GetOmopDataSourceForSourceId(sourceId).Schema,
				personId, rand.Intn(100)+1900, rand.Intn(12), rand.Intn(28), genderConceptIds[rand.Intn(len(genderConceptIds))]),
			sourceId)
		// keep track of added persons, so we don't add them twice:
		personIds = append(personIds, personId)
	}
}

func AddConcepts(concepts []Concept) {
	log.Printf("Processing %d concepts...", len(concepts))
	nextConceptId := tests.GetLastConceptId(sourceId) + 1
	for _, concept := range concepts {
		var nrClones = 1
		if concept.CloneCount > 0 {
			nrClones = concept.CloneCount
			log.Printf("** Clone directive found! Cloning this concept %d times **", concept.CloneCount)
		}
		for i := 0; i < nrClones; i++ {
			AddConceptAndMaybeAddObservations(nextConceptId, concept)
			nextConceptId++
		}
	}
}

func AddConceptAndMaybeAddObservations(nextConceptId int64, concept Concept) {
	var conceptId = nextConceptId
	var conceptName = concept.ConceptName
	var conceptClassId = "MVP Continuous"
	if concept.Id != "" {
		conceptId = utils.ParseInt64(concept.Id)
	}
	if concept.ConceptValueName != "" {
		conceptName = concept.ConceptValueName
		// TODO - validate there is only one:
		conceptId = utils.ParseInt64(concept.PossibleValues[0])
	}
	if concept.ValueType == "concept" {
		conceptClassId = "MVP Nominal"
	}
	// If still empty, generate a name:
	if conceptName == "" {
		conceptName = fmt.Sprintf("%s-%d", concept.Concept, conceptId)
	}
	vocabularyId := 124 // just use "OMOP Vocabulary" for now...
	// If concept id was already added in this session before, skip inserting it:
	if utils.Pos(conceptId, conceptIds) == -1 {
		// just in case, remove it if it already exists in DB:
		tests.RemoveConcept(models.Omop, conceptId)
		// add:
		tests.ExecSQLStringOrFail(
			fmt.Sprintf(
				"INSERT into %s.concept "+
					"(concept_id,concept_code,concept_name,domain_id,concept_class_id,standard_concept,valid_start_date,valid_end_date,invalid_reason,vocabulary_id) "+
					"values "+
					"(%d,'%s','%s','%s','%s','%s','%s','%s',NULL,%d)",
				tests.GetOmopDataSourceForSourceId(sourceId).Schema,
				conceptId, concept.Concept, conceptName, "Observation", conceptClassId, "S", "1970-01-01", "2097-12-31", vocabularyId),
			sourceId)
		// keep track of added concepts, so we don't add them twice:
		conceptIds = append(conceptIds, conceptId)
	}
	// Check if there are also personIds:
	var personIds = concept.PersonIdsOverride
	if personIds == nil {
		personIds = concept.PersonIds
	}
	// If there are any personIds, add also entries linking the persons and the concept in `observation` table:
	for _, personId := range personIds {
		AddObservationForPerson(conceptId, concept, personId)
	}
}

func AddObservationForPerson(conceptId int64, concept Concept, personId int64) {
	AddPerson(personId)
	var valueAsNumber = "NULL"
	var valueAsConceptId = "NULL"
	if concept.ValueType == "number" {
		if concept.PossibleValues == nil {
			valueAsNumber = strconv.Itoa(rand.Intn(10000))
		}
		if len(concept.PossibleValues) > 0 {
			max := len(concept.PossibleValues)
			randIndex := 0
			if max > 1 {
				randIndex = rand.Intn(max - 1)
			}
			valueAsNumber = concept.PossibleValues[randIndex]
		}
	} else if concept.ValueType == "concept" && len(concept.PossibleValues) > 0 {
		max := len(concept.PossibleValues)
		randIndex := 0
		if max > 1 {
			randIndex = rand.Intn(max - 1)
		}
		valueAsConceptId = concept.PossibleValues[randIndex]
	}
	tests.ExecSQLStringOrFail(
		fmt.Sprintf(
			"INSERT into %s.observation "+
				"(observation_id,person_id,observation_concept_id,value_as_concept_id,value_as_number,observation_date,observation_datetime,observation_type_concept_id) "+
				"values "+
				"(%d,%d,%d,%s,%s,'2000-01-01','2000-01-01 00:00:00',0)",
			tests.GetOmopDataSourceForSourceId(sourceId).Schema,
			lastObservationId+1, personId, conceptId, valueAsConceptId, valueAsNumber),
		sourceId)

	// add observation period as well:
	tests.ExecSQLStringOrFail(
		fmt.Sprintf(
			"INSERT into %s.observation_period "+
				"(observation_period_id,person_id,observation_period_start_date,observation_period_end_date,period_type_concept_id) "+
				"values "+
				"(%d,%d,'1999-01-01','2099-01-01',0)",
			tests.GetOmopDataSourceForSourceId(sourceId).Schema,
			lastObservationId+1, personId),
		sourceId)

	lastObservationId++
	countObservations++
}

var sourceId int
var lastObservationId int64
var countObservations int64
var personIds []int64
var conceptIds []int64

func Init(environment string, omopSourceId int) {
	config.Init(environment)
	db.Init()
	sourceId = omopSourceId
	lastObservationId = tests.GetLastObservationId(sourceId)
	countObservations = 0
	personIds = make([]int64, 0)
	conceptIds = make([]int64, 0)
}

func main() {
	environment := flag.String("e", "development", "Environment/prefix of config .yaml file name")
	testData := flag.String("d", "models_tests_data_config", "Prefix of test data config .yaml file name")
	sourceId := flag.Int("s", 1, "Source id for Omop DB")
	flag.Parse()
	Init(*environment, *sourceId)
	log.Printf("\n\n=============== GENERATING TEST DATA BASED ON CONFIG ============================\n\n")
	RunDataGeneration(*testData)
	log.Printf("\n\n============================= DONE! =============================================\n\n")
	log.Printf("Added this to your DB: \n - Persons: %d \n - Concepts: %d \n - Observations: %d\n\n",
		len(personIds), len(conceptIds), countObservations)
}
