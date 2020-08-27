package main

import (
	"flag"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

//TODO: Add flag for dev
func main() {
	useWorker := flag.Bool("worker", false, "Start remote worker")
	flag.Parse()

	os.Remove("flow.db")

	initDB()
	go watchNewFlow()
	r := server()

	if *useWorker {
		log.Info("Bayesnote flow worker started")
		r.Run(":9000")
	}

	go Forward()

	log.Info("Bayesnote flow core started")

	r.Run(":9000")
}

//TODO: check if remote running
func Forward() {
	//set all forward to false
	var ts []Target
	db.Model(&ts).Update("Forwarded", false).Update("db_forwarded", false)

	for {
		forwardDB()
		forwardJupyter()
		time.Sleep(time.Second)
	}
}

func forwardDB() {
	var t Target
	db.First(&t, "database = ? AND deployed = ? AND db_forwarded = ?", true, true, false)

	if t.IP == "" {
		return
	}

	t.Forward()

	db.Model(&t).Update("DBForwarded", true)

	log.WithFields(logrus.Fields{
		"remote": t.Name,
	}).Info("Start forwarding for db")
}

func forwardJupyter() {
	var ts []Target
	db.Find(&ts, "Deployed = ? AND Forwarded = ?", true, true).Not("JupyterAddr = ?", "")

	if len(ts) == 0 {
		return
	}

	for i := range ts {
		ts[i].Forward()

		db.Model(&ts[i]).Update("Forwarded", true)

		log.WithFields(logrus.Fields{
			"remote": ts[i].Name,
		}).Info("Start forwarding for jupyter")
	}
}
