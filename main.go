package main

import (
	"flag"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

//TODO: remove fatal
//TODO: when to create new remote
func main() {
	useWorker := flag.Bool("worker", false, "Start remote worker")
	flag.Parse()

	os.Remove("flow.db")
	initDB()
	go watchNewFlow()

	if *useWorker {
		log.Info("Bayesnote flow worker started")
	} else {
		go Forward()
		go pollData()
		log.Info("Bayesnote flow core started")
	}

	r := server()
	r.Run(":9000")
}

func Forward() {
	//set all forward to false
	var ts []Target
	db.Model(ts).Update("Forwarded", false)

	for {
		var ts []Target
		db.Find(&ts, "Forwarded = ?", false)
		for _, t := range ts {
			t.Forward()
			db.Model(t).Update("Forwarded", true)
		}
		time.Sleep(time.Second)
	}
}
