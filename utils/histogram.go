package utils

import (
	"encoding/json"
	"log"
	"math"
	"sort"

	"github.com/montanaflynn/stats"
)

type HistogramColumn struct {
	Start          float64 `json:"start"`
	End            float64 `json:"end"`
	NumberOfPeople int     `json:"nr_persons"`
}

func GenerateHistogramData(conceptValues []float64) string {

	if len(conceptValues) == 0 {
		return ""
	}

	sort.Float64s(conceptValues)

	width := FreedmanDiaconis(conceptValues)

	startValue := conceptValues[0]
	endValue := conceptValues[len(conceptValues)-1]

	binIndexToHistogramColumn := make(map[int]HistogramColumn)

	numBins := int((endValue-startValue)/width) + 1
	log.Printf("here is num bins %v", numBins)
	for binIndex := 0; binIndex < numBins; binIndex++ {
		binStart := (float64(binIndex) * width) + startValue
		binEnd := binStart + width
		binIndexToHistogramColumn[binIndex] = HistogramColumn{
			Start:          binStart,
			End:            binEnd,
			NumberOfPeople: 0,
		}
	}

	for _, value := range conceptValues {
		valueBinIndex := int((value - startValue) / width)
		if histogramColumn, ok := binIndexToHistogramColumn[valueBinIndex]; ok {
			histogramColumn.NumberOfPeople += 1
			binIndexToHistogramColumn[valueBinIndex] = histogramColumn
		}
	}

	histogram := []HistogramColumn{}
	for binIndex := 0; binIndex < numBins; binIndex++ {
		log.Printf("here is the amount of people %v in bin %v", binIndexToHistogramColumn[binIndex].NumberOfPeople, binIndex)
		histogram = append(histogram, binIndexToHistogramColumn[binIndex])
	}

	histogramJson, _ := json.Marshal(histogram)
	return string(histogramJson)
}

// This function returns the bin width upon the Freedman Diaconis formula: https://en.wikipedia.org/wiki/Freedman%E2%80%93Diaconis_rule
func FreedmanDiaconis(values []float64) float64 {

	sort.Float64s(values)

	valuesInterQuartileRange, _ := stats.InterQuartileRange(values)
	n := len(values)
	width := (2 * valuesInterQuartileRange) / math.Cbrt(float64(n))

	log.Printf("here is the width for freedman diaconis calculation %v", width)

	return width
}

func IQR(values []float64) float64 {
	sort.Float64s(values)
	valuesInterQuartileRange, _ := stats.InterQuartileRange(values)
	return valuesInterQuartileRange
}
