package main

import (
	"flag"
	"os"
	"os/exec"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

var db *gorm.DB
var p string

//TODO: Add flag for dev
func main() {
	Worker := flag.Bool("worker", false, "Start remote worker")
	Port := flag.String("p", "22", "Set SSH port")
	DB := flag.Bool("db", false, "Remove db")
	flag.Parse()
	p = *Port

	if *DB {
		os.Remove("flow.db")
	}

	initDB()
	go flowWatcher()
	r := server()

	if *Worker {
		go runJupyter()
		log.Info("Bayesnote flow worker started")
		r.Run(":9000")
		return
	}

	go tunnelWatcher()

	log.Info("Bayesnote flow core started")

	r.Run(":9000")
}

func runJupyter() {
	cmd := exec.Command("sh", "-c", "jupyter notebook --ip='*' --NotebookApp.token='' --NotebookApp.password='' --allow-root")
	err := cmd.Run()
	if err != nil {
		log.Error(err)
	}
}
