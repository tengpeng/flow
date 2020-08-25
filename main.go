package main

import (
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

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

func watchNewFlow() {
	for {
		var f Flow
		db.Find(&f, "status = ?", "")

		if f.Schedule != "" {
			db.Model(&f).Update("Status", "STARTED")

			log.WithFields(logrus.Fields{
				"flow": f.FlowName,
			}).Info("Get new flow")

			c := cron.New()
			c.Start()
			c.AddFunc(f.Schedule, func() { f.start() })

			log.WithFields(logrus.Fields{
				"schedule": f.Schedule,
			}).Info("Add cron job")
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
