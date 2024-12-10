package utils

import (
	"log"

	"github.com/montanaflynn/stats"
)

type ConceptStats struct {
	CohortId       int     `json:"cohortId"`
	ConceptId      int64   `json:"conceptId"`
	NumberOfPeople int     `json:"personCount"`
	Min            float64 `json:"min"`
	Max            float64 `json:"max"`
	Avg            float64 `json:"avg"`
	Sd             float64 `json:"sd"`
}

func GenerateStatsData(cohortId int, conceptId int64, conceptValues []float64) *ConceptStats {

	if len(conceptValues) == 0 {
		log.Printf("Data size is zero. Returning nil.")
		return nil
	}

	result := new(ConceptStats)
	result.CohortId = cohortId
	result.ConceptId = conceptId
	result.NumberOfPeople = len(conceptValues)

	minValue, _ := stats.Min(conceptValues)
	result.Min = minValue

	maxValue, _ := stats.Max(conceptValues)
	result.Max = maxValue

	meanValue, _ := stats.Mean(conceptValues)
	result.Avg = meanValue

	sdValue, _ := stats.StandardDeviation(conceptValues)
	result.Sd = sdValue

	return result
}
