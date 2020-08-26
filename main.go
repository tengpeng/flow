package main

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

//TODO: remove fatal
func main() {
	log.Info("Bayesnote flow started")

	os.Remove("flow.db")
	initDB()
	//setUpTestDB()

	//go testForward()
	//go localTestForward()

	go watchNewFlow()

	r := server()
	r.Run(":9000")
}

/*
	os.Remove("flow.db")
	setUpTestDB()
	testFlow()
	cronTrigger()
	deploy()
	testForward()
*/
