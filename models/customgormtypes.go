package models

import (
	"database/sql/driver"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// See https://gorm.io/docs/data_types.html for more details on custom data types

type ConceptType string

func (u ConceptType) Value() (driver.Value, error) {
	return string(u), nil
}

func (u *ConceptType) Scan(value interface{}) error {
	valueAsString, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to unmarshal value: %s", value)
	}
	result := "unexpected missing value: concept_type is supposed to be a derived value from another field value that ends with an underscore followed by a suffix (e.g. '_typeX') "
	if strings.Contains(valueAsString, "_") {
		items := strings.Split(valueAsString, "_")
		// the assumption here is that valueAsString is a value with a _suffixX where suffixX is the concept type
		// we want. So we want only the last part of the value, after it is split by "_":
		result = items[len(items)-1]
	}
	*u = ConceptType(result)
	return nil
}

func (u ConceptType) GormDataType() string {
	return "concepttype"
}

// More type "tranformation" related functions below....

// Very simple function, just to add a prefix in front of the conceptId.
// It is a public method here, since it is needed in different places...
// ...so we need to keep it consistent:
func GetPrefixedConceptId(conceptId int64) string {
	return "ID_" + strconv.FormatInt(conceptId, 10)
}

// The reverse of above function:
func GetConceptId(prefixedConceptId string) int64 {
	// validate: it should start with ID_
	if strings.Index(prefixedConceptId, "ID_") != 0 {
		log.Panicf("Prefixed concept id should start with ID_ . However, found this instead: %s", prefixedConceptId)
	}
	var conceptId = strings.Split(prefixedConceptId, "ID_")[1]
	var result, _ = strconv.ParseInt(conceptId, 10, 64)
	return result
}
