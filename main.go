package main

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

func main() {
	log.Info("Bayesnote flow started")

	//os.Remove("flow.db")
	initDB()

	go testForward()
	//go localTestForward()

	go watchNewFlow()

	r := server()
	r.Run(":9000")
}

func watchNewFlow() {
	for {
		var f flow
		db.Find(&f, "status = ?", nil)
		//TODO: update status
		//add cron trigger
		if f.Schedule != "" {
			c := cron.New()
			c.AddFunc(f.Schedule, func() { f.start() })
		}
		time.Sleep(time.Second)
	}
}

/*
	os.Remove("flow.db")
	setUpTestDB()
	testFlow()
	cronTrigger()
	deploy()
	testForward()
*/
