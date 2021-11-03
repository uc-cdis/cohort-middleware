package main

import (
	"flag"
	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/server"
)

// @title Cohort Middleware
// @version 0.0
// @description Simple Cohort Middleware for accessing OMOP data
// @termsOfService http://swagger.io/terms/

// @contact.name
// @contact.url
// @contact.email

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /
// @schemes http
func main() {
	environment := flag.String("e", "development", "")
	config.Init(*environment)
	db.Init()
	server.Init()
}
