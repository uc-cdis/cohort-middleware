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
	NumberOfPeople int     `json:"personCount"`
}

func GenerateHistogramData(conceptValues []float64) []HistogramColumn {

	if len(conceptValues) == 0 {
		return nil
	}
	numBins, width := GetBinsAndWidthUsingFreedmanDiaconisAndSortValues(conceptValues) //conceptValues will get sorted as a side-effect, which is useful in this case
	startValue := conceptValues[0]
	binIndexToHistogramColumn := make(map[int]HistogramColumn)

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
		histogram = append(histogram, binIndexToHistogramColumn[binIndex])
	}

	return histogram
}

// Sorts the given values, and returns the number of bins, the width of the bins using FreedmanDiaconis
func GetBinsAndWidthUsingFreedmanDiaconisAndSortValues(values []float64) (int, float64) {

	width := FreedmanDiaconis(values) // values will get sorted as a side-effect, which is useful in this case
	startValue := values[0]
	endValue := values[len(values)-1]

	numBins := 0
	if width > 0 {
		numBins = int((endValue-startValue)/width) + 1
	} else {
		numBins = 1
		width = endValue + 1 - startValue
	}
	log.Printf("num bins %v", numBins)

	return numBins, width
}

// This function returns the bin width upon the Freedman Diaconis formula: https://en.wikipedia.org/wiki/Freedman%E2%80%93Diaconis_rule
// Can return 0 if IQR(values) is 0.
func FreedmanDiaconis(values []float64) float64 {

	valuesInterQuartileRange := IQR(values) // values will get sorted as a side-effect, which is useful in this case
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
