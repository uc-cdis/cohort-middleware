package utils

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseNumericId(c *gin.Context, paramName string) (int, error) {
	// parse and validate:
	numericIdStr := c.Param(paramName)
	log.Printf("Querying %s: %s", paramName, numericIdStr)
	if numericId, err := strconv.Atoi(numericIdStr); err != nil {
		log.Printf("bad request - %s should be a number", paramName)
		return -1, fmt.Errorf("bad request - %s should be a number", paramName)
	} else {
		return numericId, nil
	}
}
