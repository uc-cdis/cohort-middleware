package tests

import (
	"testing"

	"github.com/uc-cdis/cohort-middleware/models"
)

func TestGetConceptId(t *testing.T) {
	conceptModel := new(models.Concept)
	conceptId := conceptModel.GetConceptId("ID_12345")
	if conceptId != 12345 {
		t.Error()
	}
}
