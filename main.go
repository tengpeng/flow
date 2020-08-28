package main

import (
	"flag"
	"os"
	"os/exec"
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
		go watchCmd()
		go runJupyter()
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
	var ts []Tunnel
	db.Model(&ts).Update("Forwarded", false)

	for {
		forward()
		time.Sleep(time.Second)
	}
}

func forward() {
	var t Tunnel
	if db.First(&t, "forwarded = ?", false).RecordNotFound() {
		return
	}

	t.Forward()

	db.Model(&t).Update("forwarded", true)

	log.WithFields(logrus.Fields{
		"RemoteAddr": t.RemoteAddr,
	}).Info("Start forwarding")
}

func runJupyter() {
	cmd := exec.Command("sh", "-c", "jupyter notebook --ip='*' --NotebookApp.token='' --NotebookApp.password='' --allow-root")
	err := cmd.Run()
	if err != nil {
		log.Error(err)
	}
}

//TODO: kill ->
