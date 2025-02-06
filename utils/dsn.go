package utils

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

func GenerateDsn(sourceConnectionString string) string {
	sourceConnectionParts := strings.FieldsFunc(sourceConnectionString, func(r rune) bool {
		separators := ":/;="
		return strings.ContainsRune(separators, r)
	})
	dbVendor := sourceConnectionParts[1]
	log.Printf("Found db vendor %s", dbVendor)
	if !(dbVendor == "postgresql" || dbVendor == "sqlserver") {
		panic("db vendor not supported")
	}
	host := sourceConnectionParts[2]
	port := sourceConnectionParts[3]
	dbname := sourceConnectionParts[5]
	username := sourceConnectionParts[7]
	password := sourceConnectionParts[9]

	dsn := fmt.Sprintf(dbVendor+"://%s:%s@%s:%s?database=%s",
		username,
		url.QueryEscape(password),
		host,
		port,
		dbname)
	return dsn
}
