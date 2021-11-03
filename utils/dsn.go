package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func GenerateDsn(sourceConnectionString string) string {
	sourceConnectionParts := strings.FieldsFunc(sourceConnectionString, func(r rune) bool {
		separators := ":/;="
		if strings.ContainsRune(separators, r) {
			return true
		}
		return false
	})
	host := sourceConnectionParts[2]
	port := sourceConnectionParts[3]
	dbname := sourceConnectionParts[5]
	username := sourceConnectionParts[7]
	password := sourceConnectionParts[9]

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		username,
		url.QueryEscape(password),
		host,
		port,
		dbname)
	return dsn
}
