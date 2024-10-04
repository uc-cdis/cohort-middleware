package utils

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

func GenerateDsn(sourceConnectionString string) string {
	sourceConnectionParts := strings.FieldsFunc(sourceConnectionString, func(r rune) bool {
		separators := ":/;=?&"
		return strings.ContainsRune(separators, r)
	})
	dbVendor := sourceConnectionParts[1]
	log.Printf("Found db vendor %s", dbVendor)
	host := sourceConnectionParts[2]
	port := sourceConnectionParts[3]
	dbname := ""
	username := ""
	password := ""
	if len(sourceConnectionParts) == 9 {
		// expecting a string like this: jdbc:postgresql://hostname.com:5432/dbname?user=username&etc
		dbname = sourceConnectionParts[4]
		username = sourceConnectionParts[6]
		password = sourceConnectionParts[8]
	} else if len(sourceConnectionParts) == 10 {
		// expecting a string like this: jdbc:sqlserver://hostname.com:1433;databaseName=dbname;user=username;etc
		dbname = sourceConnectionParts[5]
		username = sourceConnectionParts[7]
		password = sourceConnectionParts[9]
	} else {
		panic("connection string format not supported")
	}

	dsn := fmt.Sprintf(dbVendor+"://%s:%s@%s:%s?database=%s",
		username,
		url.QueryEscape(password),
		host,
		port,
		dbname)
	return dsn
}
