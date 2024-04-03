package main

import (
	"flag"
	"log"

	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/models"
	"github.com/uc-cdis/cohort-middleware/server"
	"github.com/uc-cdis/cohort-middleware/utils"
)

func runDataValidation() {
	conf := config.GetConfig()
	observationConceptIdsToCheck, _ := utils.SliceAtoi(conf.GetStringSlice("validate.single_observation_for_concept_ids"))
	var cohortDataModel = new(models.CohortData)
	nrIssues, _ := cohortDataModel.ValidateObservationData(observationConceptIdsToCheck)
	if nrIssues > 0 {
		log.Printf("WARNING: found %d data issues!", nrIssues)
	}
}

func runDataDictionaryGeneration() {
	var cohortDataModel = new(models.CohortData)
	var dataDictionaryModel = new(models.DataDictionary)
	dataDictionaryModel.CohortDataModel = cohortDataModel
	log.Printf("Generating Data Dictionary...")
	_, error := dataDictionaryModel.GenerateDataDictionary()
	if error != nil {
		log.Printf("WARNING: found %d data issues!", error)
	}
}

func main() {
	environment := flag.String("e", "development", "Environment/prefix of config file name")
	flag.Parse()
	config.Init(*environment)
	db.Init()
	runDataValidation()
	go runDataDictionaryGeneration()
	server.Init()
}
