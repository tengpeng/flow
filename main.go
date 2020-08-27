package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mitchellh/go-ps"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

func main() {
	useWorker := flag.Bool("worker", false, "Start remote worker")
	flag.Parse()

	//os.Remove("flow.db")
	killFlow()

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
		db.Find(&ts, "Forwarded = ? AND Deployed = ?", false, true)
		for _, t := range ts {
			t.Forward()
			db.Model(t).Update("Forwarded", true)
		}
		time.Sleep(time.Second)
	}
}

func killFlow() {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Error(err)
	}

	cmd := filepath.Join(homePath, "flow")

	prs, err := ps.Processes()
	if err != nil {
		log.Info(err)
	}

	for _, v := range prs {
		if strings.Contains(v.Executable(), cmd) {
			p, err := os.FindProcess(v.Pid())
			if err != nil {
				log.Error(err)
			}
			p.Kill()
		}
	}
}
