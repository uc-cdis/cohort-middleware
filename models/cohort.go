package models

import (
	"time"

	"github.com/uc-cdis/cohort-middleware/db"
)

type Cohort struct {
	CohortDefinitionId int `json:",omitempty"`
	SubjectId          int64
	CohortStartDate    time.Time
	CohortEndDate      time.Time
}

func (h Cohort) GetCohortById(id int) ([]*Cohort, error) {
	db2 := db.GetAtlasDB().Db
	var cohort []*Cohort
	db2.Model(&Cohort{}).Select("cohort_definition_id, subject_id, cohort_start_date, cohort_end_date").Where("cohort_definition_id = ?", id).Scan(&cohort)
	return cohort, nil
}
