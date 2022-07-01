package models

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

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

func GetCohortPairKey(firstCohortDefinitionId int, secondCohortDefinitionId int) string {
	return fmt.Sprintf("ID_%v_%v", firstCohortDefinitionId, secondCohortDefinitionId)
}
