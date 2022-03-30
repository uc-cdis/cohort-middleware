package main

import (
	"flag"
	"github.com/uc-cdis/cohort-middleware/config"
	"github.com/uc-cdis/cohort-middleware/db"
	"github.com/uc-cdis/cohort-middleware/server"
)

func main() {
	environment := flag.String("e", "development", "")
	config.Init(*environment)
	db.Init()
	server.Init()
}
