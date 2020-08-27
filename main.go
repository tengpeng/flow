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
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB

//TODO: Add flag for dev
func main() {
	useWorker := flag.Bool("worker", false, "Start remote worker")
	flag.Parse()

	os.Remove("flow.db")
	killFlow()

	initDB()
	go watchNewFlow()
	r := server()

	if *useWorker {
		log.Info("Bayesnote flow worker started")
		r.Run(":9000")
	}

	go Forward()
	go pollData()
	log.Info("Bayesnote flow core started")

	r.Run(":9000")
}

//TODO: check if remote running
func Forward() {
	//set all forward to false
	var ts []Target
	db.Model(ts).Update("Forwarded", false)

	for {
		var ts []Target
		db.Find(&ts, "Forwarded = ? AND Deployed = ?", false, true)
		for _, t := range ts {
			t.Forward()

			// tunnel := sshtunnel.NewSSHTunnel(
			// 	"ubuntu@"+t.ServerAddr,
			// 	sshtunnel.PrivateKeyFile(t.Pem),
			// 	t.RemoteAddr,
			// 	"8000",
			// )
			// go tunnel.Start()
			// time.Sleep(100 * time.Millisecond)

			db.Model(t).Update("Forwarded", true)

			log.WithFields(logrus.Fields{
				"remote": t.Name,
			}).Info("Start forwarding")
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
