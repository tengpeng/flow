package main

import (
	"flag"

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

	//os.Remove("flow.db")
	initDB()
	go watchNewFlow()

	if *useWorker {
		log.Info("Bayesnote flow worker started")
	} else {
		var ts []Target
		db.Find(&ts)
		for _, t := range ts {
			newRemote(t)
		}

		log.Info("Bayesnote flow core started")
		go pollData()
	}

	r := server()
	r.Run(":9000")
}
