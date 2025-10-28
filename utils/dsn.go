package utils

import (
	"fmt"
	"log"
	"net/url"
	"strings"
)

// Translates a sourceConnectionString of the form
//
//	jdbc:postgresql://localhost:5434/mydbname
//
// or
//
//	jdbc:sqlserver://localhost:1433;databaseName=mydbname
//
// into a connection string of the form
//
//	postgresql://username:password@localhost:5434?database=mydbname //pragma: allowlist secret
func GenerateDsn(source SourceConnection) string {
	sourceConnectionParts := strings.FieldsFunc(source.SourceConnection, func(r rune) bool {
		separators := ":/;"
		return strings.ContainsRune(separators, r)
	})
	dbVendor := sourceConnectionParts[1]
	log.Printf("Found db vendor %s", dbVendor)
	// validate:
	if len(sourceConnectionParts) != 5 {
		panic(fmt.Sprintf("Expected a connection string with 5 parts, found %d", len(sourceConnectionParts)))
	}
	host := sourceConnectionParts[2]
	port := sourceConnectionParts[3]
	dbname := sourceConnectionParts[4]

	// Check if string contains '=' (e.g. can be in sqlserver connection string)
	if strings.Contains(dbname, "=") {
		parts := strings.SplitN(dbname, "=", 2)
		// parts[1] will be your dbname ("mydbname")
		dbname = parts[1]
	}

	dsn := fmt.Sprintf(dbVendor+"://%s:%s@%s:%s?database=%s", // note: the "?database=" part is fine for mssql, not really the standard for postgres...but gorm/postgres seem to be flexible about it
		source.Username,
		url.QueryEscape(source.Password),
		host,
		port,
		dbname)
	return dsn
}
