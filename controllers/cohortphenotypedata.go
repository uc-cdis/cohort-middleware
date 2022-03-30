package controllers

// DEPRECATED!
import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/uc-cdis/cohort-middleware/models"
)

type CohortPhenotypeData struct {
	Format string `form:"format"`
}

var cohortPhenotypeDataModel = new(models.CohortPhenotypeData)

func (u CohortPhenotypeData) Retrieve(c *gin.Context) {
	var f CohortPhenotypeData
	c.Bind(&f)

	sourceName := c.Param("sourcename")

	if sourceName != "" {
		cohort, err := cohortPhenotypeDataModel.GetCohortDataPhenotype(sourceName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error to retrieve cohort", "error": err})
			c.Abort()
			return
		}

		format := strings.ToLower(f.Format)

		if format == "tsv" || format == "csv" {
			b := GenerateCsv(format, cohort)
			c.String(http.StatusOK, b.String())
		} else {
			c.JSON(http.StatusOK, gin.H{"Cohort": cohort})
		}
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
	c.Abort()
	return
}

func GenerateCsv(format string, cohort []*models.CohortPhenotypeData) *bytes.Buffer {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)

	if format == "tsv" {
		w.Comma = '\t'
	}

	var rows [][]string
	rows = append(rows, []string{"sample.id", "age", "gender", "Hare", "CDW_race", "Height"})

	for _, cohortItem := range cohort {
		row := []string{cohortItem.SampleId, fmt.Sprint(cohortItem.Age), cohortItem.Gender, cohortItem.Hare, cohortItem.CDWrace, fmt.Sprint(cohortItem.Height)}
		rows = append(rows, row)
	}
	err := w.WriteAll(rows)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
