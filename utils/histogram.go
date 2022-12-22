package utils

import (
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

func GenerateHistogramData(conceptValues []float64) []HistogramColumn {

	if len(conceptValues) == 0 {
		return nil
	}

	sort.Float64s(conceptValues)

	width := FreedmanDiaconis(conceptValues)

	startValue := conceptValues[0]
	endValue := conceptValues[len(conceptValues)-1]

	binIndexToHistogramColumn := make(map[int]HistogramColumn)

	numBins := 0
	if width > 0 {
		numBins = int((endValue-startValue)/width) + 1
	} else {
		numBins = 1
		width = endValue + 1 - startValue
	}

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

	return histogram
}

// This function returns the bin width upon the Freedman Diaconis formula: https://en.wikipedia.org/wiki/Freedman%E2%80%93Diaconis_rule
// Can return 0 if IQR(values) is 0.
func FreedmanDiaconis(values []float64) float64 {

	sort.Float64s(values)

	valuesInterQuartileRange := IQR(values)
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
